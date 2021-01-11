/*
 *  Copyright (C) 2020 Intel Corporation
 *  SPDX-License-Identifier: BSD-3-Clause
 */

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/antchfx/jsonquery"
)

//To map the conditions in the flavor template with old flavor part
var flavorTemplateConditions = map[string]string{"//host_info/tboot_installed//*[text()='true']": "//meta/description/tboot_installed//*[text()='true']",
	"//host_info/hardware_features/SUEFI/enabled//*[text()='true']": "//hardware/feature/SUEFI/enabled//*[text()='true']",
	"//host_info/hardware_features/cbnt/enabled//*[text()='true']":  "//hardware/feature/CBNT/enabled//*[text()='true']",
	"//host_info/vendor//*[text()='Linux']":                         "//meta/vendor//*[text()='INTEL']",
	"//host_info/tpm_version//*[text()='2.0']":                      "//meta/description/tpm_version//*[text()='2.0']"}

var flavorTemplatePath = "/opt/hvs-flavortemplates"

//getOldFlavorPartFilePath method is used to get the data from argement
func getOldFlavorPartFilePath() (string, error) {

	// return error if there are no correct number of arguments
	if len(os.Args) < 2 {
		return "", fmt.Errorf("Old flavor part json file path is required")
	}

	fileLocation := os.Args[1]

	return fileLocation, nil
}

//getFlavorTemplates method is used to get the flavor templates based on old flavor part file
func getFlavorTemplates(body []byte) ([]FlavorTemplate, error) {

	var defaultFlavorTemplates []string

	//read the flavor template file
	templates, err := ioutil.ReadDir(flavorTemplatePath)
	if err != nil {
		return nil, fmt.Errorf("Error in reading flavor template files")
	}
	for _, template := range templates {
		path := flavorTemplatePath + "/" + template.Name()
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("Error in reading the template file - ", template.Name())
		}
		defaultFlavorTemplates = append(defaultFlavorTemplates, string(data))
	}

	// finding the correct template to apply
	filteredTemplate, err := findTemplatesToApply(body, defaultFlavorTemplates)
	if err != nil {
		return nil, fmt.Errorf("Error in getting the template file based on old flavorpart")
	}

	return filteredTemplate, nil

}

//findTemplatesToApply method is used to find the correct templates to apply to convert flavor part
func findTemplatesToApply(oldFlavorPart []byte, defaultFlavorTemplates []string) ([]FlavorTemplate, error) {
	var filteredTemplates []FlavorTemplate

	oldFlavorPartJson, err := jsonquery.Parse(strings.NewReader(string(oldFlavorPart)))
	if err != nil {
		return nil, fmt.Errorf("Error in parsing the old flavor part json")
	}

	var conditionEval bool

	for _, template := range defaultFlavorTemplates {

		flavorTemplate := FlavorTemplate{}

		err := json.Unmarshal([]byte(template), &flavorTemplate)
		if err != nil {
			return nil, fmt.Errorf("Error in unmarshaling the flavor template")
		}

		if flavorTemplate.Label == "" {
			continue
		}
		conditionEval = false

		for _, condition := range flavorTemplate.Condition {
			conditionEval = true

			flavorPartCondition := flavorTemplateConditions[condition]

			expectedData, _ := jsonquery.Query(oldFlavorPartJson, flavorPartCondition)
			if expectedData == nil {
				conditionEval = false
				break
			}
		}
		if conditionEval == true {
			filteredTemplates = append(filteredTemplates, flavorTemplate)
		}
	}

	return filteredTemplates, nil

}

//checkIfValidFile method is used to check if the given input file path is valid or not
func checkIfValidFile(filename string) (bool, error) {
	// Checking if the input file is json
	if fileExtension := filepath.Ext(filename); fileExtension != ".json" {
		return false, fmt.Errorf("File '%s' is not json", filename)
	}

	// Checking if filepath entered belongs to an existing file
	if _, err := os.Stat(filename); err != nil && os.IsNotExist(err) {
		return false, fmt.Errorf("File %s does not exist", filename)
	}

	// returns true if this is a valid file
	return true, nil
}

//main method implements migration of old format of flavor part to new format
func main() {

	// Getting the file data that was entered by the user
	oldFlavorPartFilePath, err := getOldFlavorPartFilePath()
	if err != nil {
		fmt.Println("Error in getting the old flavor part file path - ", err)
		os.Exit(1)
	}

	// Validating the old flavor part file path entered
	if valid, err := checkIfValidFile(oldFlavorPartFilePath); err != nil && !valid {
		fmt.Println("Error in validating the input file path - ", err)
		os.Exit(1)
	}

	//reading the data from oldFlavorPartFilePath
	body, err := ioutil.ReadFile(oldFlavorPartFilePath)
	if err != nil {
		fmt.Println("Error in reading the old flavor part file data")
		os.Exit(1)
	}

	//get the flavor template based on old flavor part file
	templates, err := getFlavorTemplates(body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var oldFlavorPart OldFlavorPart

	//unmarshaling the old flavor part file into OldFlavorPart struct
	err = json.Unmarshal(body, &oldFlavorPart)
	if err != nil {
		fmt.Println("Error in unmarshaling the old flavor part file")
		os.Exit(1)
	}

	for flavorIndex, flavor := range oldFlavorPart.SignedFlavor {

		//Updating meta section
		if flavor.Flavor.Hardware != nil && flavor.Flavor.Hardware.Feature.CBNT != nil && flavor.Flavor.Hardware.Feature.CBNT.Enabled {
			oldFlavorPart.SignedFlavor[flavorIndex].Flavor.Meta.Description.CbntEnabled = true
		} else if flavor.Flavor.Hardware != nil && flavor.Flavor.Hardware.Feature.SUEFI != nil && flavor.Flavor.Hardware.Feature.SUEFI.Enabled {
			oldFlavorPart.SignedFlavor[flavorIndex].Flavor.Meta.Description.SuefiEnabled = true
		}

		//removing the signature from the flavors
		//since the final flavor part file is not a signed flavor(only the flavor collection)
		oldFlavorPart.SignedFlavor[flavorIndex].Signature = ""

		// Copying the pcrs sections from old flavor part to new flavor part
		if flavor.Flavor.Pcrs == nil {
			continue
		}

		for _, template := range templates {

			oldFlavorPart.SignedFlavor[flavorIndex].Flavor.Meta.Description.FlavorTemplateIds = append(oldFlavorPart.SignedFlavor[flavorIndex].Flavor.Meta.Description.FlavorTemplateIds, template.ID)

			flavorname := flavor.Flavor.Meta.Description.FlavorPart

			pcrsmap := make(map[int]string)

			var rules PcrRules

			if flavorname == flavor.Flavor.Meta.Description.FlavorPart {

				if flavorname == "PLATFORM" && template.FlavorParts.Platform != nil {
					for _, rules := range template.FlavorParts.Platform.PcrRules {
						pcrsmap[rules.Pcr.Index] = rules.Pcr.Bank
					}
					rules = template.FlavorParts.Platform.PcrRules

				} else if flavorname == "OS" && template.FlavorParts.OS != nil {
					for _, rules := range template.FlavorParts.OS.PcrRules {
						pcrsmap[rules.Pcr.Index] = rules.Pcr.Bank
					}
					rules = template.FlavorParts.OS.PcrRules

				} else if flavorname == "HOST_UNIQUE" && template.FlavorParts.HostUnique != nil {
					for _, rules := range template.FlavorParts.HostUnique.PcrRules {
						pcrsmap[rules.Pcr.Index] = rules.Pcr.Bank
					}
					rules = template.FlavorParts.HostUnique.PcrRules
				}
			} else if flavorname != flavor.Flavor.Meta.Description.FlavorPart {
				continue
			}

			var newFlavorPcrs []PcrLogs
			newFlavorPcrs = make([]PcrLogs, len(pcrsmap))

			for bank, pcrMap := range flavor.Flavor.Pcrs {
				index := 0
				for mapIndex, templateBank := range pcrsmap {
					pcrIndex := PcrIndex(mapIndex)

					if SHAAlgorithm(bank) != SHAAlgorithm(templateBank) {
						break
					}
					if expectedPcrEx, ok := pcrMap[pcrIndex.String()]; ok {

						newFlavorPcrs[index].PCR.Index = mapIndex
						newFlavorPcrs[index].PCR.Bank = bank
						newFlavorPcrs[index].Measurement = expectedPcrEx.Value
						newFlavorPcrs[index].PCRMatches = rules[index].PcrMatches

						var newTpmEvents []NewEventLog
						if expectedPcrEx.Event != nil && !reflect.ValueOf(rules[index].EventlogEquals).IsZero() {

							newFlavorPcrs[index].EventlogEqual = new(EventlogEquals)

							if rules[index].EventlogEquals.ExcludingTags != nil {
								newFlavorPcrs[index].EventlogEqual.ExcludeTags = rules[index].EventlogEquals.ExcludingTags
							}

							newTpmEvents = make([]NewEventLog, len(expectedPcrEx.Event))
							for eventIndex, oldEvents := range expectedPcrEx.Event {

								newTpmEvents[eventIndex].Tags = append(newTpmEvents[eventIndex].Tags, oldEvents.Label)
								newTpmEvents[eventIndex].Measurement = oldEvents.Value
								newTpmEvents[eventIndex].TypeID = oldEvents.DigestType
							}
							newFlavorPcrs[index].EventlogEqual.Events = newTpmEvents
							newTpmEvents = nil
						}

						if expectedPcrEx.Event != nil && !reflect.ValueOf(rules[index].EventlogIncludes).IsZero() {

							newTpmEvents = make([]NewEventLog, len(expectedPcrEx.Event))
							for eventIndex, oldEvents := range expectedPcrEx.Event {

								newTpmEvents[eventIndex].Tags = append(newTpmEvents[eventIndex].Tags, oldEvents.Label)
								newTpmEvents[eventIndex].Measurement = oldEvents.Value
								newTpmEvents[eventIndex].TypeID = oldEvents.DigestType
							}
							newFlavorPcrs[index].EventlogIncludes = newTpmEvents

							newTpmEvents = nil
						}
					}

					index++
				}

			}
			flavor.Flavor.PcrLogs = newFlavorPcrs
		}
		oldFlavorPart.SignedFlavor[flavorIndex].Flavor.Pcrs = nil
		oldFlavorPart.SignedFlavor[flavorIndex].Flavor.PcrLogs = flavor.Flavor.PcrLogs
	}

	//getting the final data
	finalFlavorPart, err := json.Marshal(oldFlavorPart.SignedFlavor)
	if err != nil {
		fmt.Println("Error in marshaling the final flavor part file")
		os.Exit(1)
	}

	//Printing the final flavor part file in console
	fmt.Println("New flavor part json:\n", string(finalFlavorPart))

	//writing the new flavor part into the local file
	data := []byte(finalFlavorPart)
	err = ioutil.WriteFile("/opt/newflavorpart.json", data, 0644)
	if err != nil {
		fmt.Println("Error in writing the new flavor part file")
		os.Exit(1)
	}
}

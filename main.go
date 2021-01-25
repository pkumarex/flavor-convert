/*
 *  Copyright (C) 2020 Intel Corporation
 *  SPDX-License-Identifier: BSD-3-Clause
 */

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/antchfx/jsonquery"
)

// EventIDList - define map for event id
var eventIDList = map[string]string{
	"EV_PREBOOT_CERT":                  "0x00000000",
	"EV_POST_CODE":                     "0x00000001",
	"EV_UNUSED":                        "0x00000002",
	"EV_NO_ACTION":                     "0x00000003",
	"EV_SEPARATOR":                     "0x00000004",
	"EV_ACTION":                        "0x00000005",
	"EV_EVENT_TAG":                     "0x00000006",
	"EV_S_CRTM_CONTENTS":               "0x00000007",
	"EV_S_CRTM_VERSION":                "0x00000008",
	"EV_CPU_MICROCODE":                 "0x00000009",
	"EV_PLATFORM_CONFIG_FLAGS":         "0x0000000A",
	"EV_TABLE_OF_DEVICES":              "0x0000000B",
	"EV_COMPACT_HASH":                  "0x0000000C",
	"EV_IPL":                           "0x0000000D",
	"EV_IPL_PARTITION_DATA":            "0x0000000E",
	"EV_NONHOST_CODE":                  "0x0000000F",
	"EV_NONHOST_CONFIG":                "0x00000010",
	"EV_NONHOST_INFO":                  "0x00000011",
	"EV_OMIT_BOOT_DEVICE_EVENTS":       "0x00000012",
	"EV_EFI_EVENT_BASE":                "0x80000000",
	"EV_EFI_VARIABLE_DRIVER_CONFIG":    "0x80000001",
	"EV_EFI_VARIABLE_BOOT":             "0x80000002",
	"EV_EFI_BOOT_SERVICES_APPLICATION": "0x80000003",
	"EV_EFI_BOOT_SERVICES_DRIVER":      "0x80000004",
	"EV_EFI_RUNTIME_SERVICES_DRIVER":   "0x80000005",
	"EV_EFI_GPT_EVENT":                 "0x80000006",
	"EV_EFI_ACTION":                    "0x80000007",
	"EV_EFI_PLATFORM_FIRMWARE_BLOB":    "0x80000008",
	"EV_EFI_HANDOFF_TABLES":            "0x80000009",
	"EV_EFI_PLATFORM_FIRMWARE_BLOB2":   "0x8000000A",
	"EV_EFI_HANDOFF_TABLES2":           "0x8000000B",
	"EV_EFI_VARIABLE_BOOT2":            "0x8000000C",
	"EV_EFI_HCRTM_EVENT":               "0x80000010",
	"EV_EFI_VARIABLE_AUTHORITY":        "0x800000E0",
	"EV_EFI_SPDM_FIRMWARE_BLOB":        "0x800000E1",
	"EV_EFI_SPDM_FIRMWARE_CONFIG":      "0x800000E2",
	"PCR_MAPPING":                      "0x401",
	"HASH_START":                       "0x402",
	"COMBINED_HASH":                    "0x403",
	"MLE_HASH":                         "0x404",
	"BIOSAC_REG_DATA":                  "0x40a",
	"CPU_SCRTM_STAT":                   "0x40b",
	"LCP_CONTROL_HASH":                 "0x40c",
	"ELEMENTS_HASH":                    "0x40d",
	"STM_HASH":                         "0x40e",
	"OSSINITDATA_CAP_HASH":             "0x40f",
	"SINIT_PUBKEY_HASH":                "0x410",
	"LCP_HASH":                         "0x411",
	"LCP_DETAILS_HASH":                 "0x412",
	"LCP_AUTHORITIES_HASH":             "0x413",
	"NV_INFO_HASH":                     "0x414",
	"EVTYPE_KM_HASH":                   "0x416",
	"EVTYPE_BPM_HASH":                  "0x417",
	"EVTYPE_KM_INFO_HASH":              "0x418",
	"EVTYPE_BPM_INFO_HASH":             "0x419",
	"EVTYPE_BOOT_POL_HASH":             "0x41a",
	"CAP_VALUE":                        "0x4ff",
	"tb_policy":                        "0x501",
	"vmlinuz":                          "0x501",
	"initrd":                           "0x501",
	"asset-tag":                        "0x501",
}

const (
	intelVendor  = "INTEL"
	vmwareVendor = "VMWARE"

	platformFlavor   = "PLATFORM"
	osFlavor         = "OS"
	hostUniqueFlavor = "HOST_UNIQUE"
)

var BuildVersion string

const helpStr = `Usage:
flavor-convert <command> [argument]
	
Available Command:
	-o                To provide old flavor part json file
	-h|--help         Show this help message
	-version          Print the current version
`

//To map the conditions in the flavor template with old flavor part
var flavorTemplateConditions = map[string]string{"//host_info/tboot_installed//*[text()='true']": "//meta/description/tboot_installed//*[text()='true']",
	"//host_info/hardware_features/SUEFI/enabled//*[text()='true']": "//hardware/feature/SUEFI/enabled//*[text()='true']",
	"//host_info/hardware_features/cbnt/enabled//*[text()='true']":  "//hardware/feature/CBNT/enabled//*[text()='true']",
	"//host_info/vendor//*[text()='Linux']":                         "//meta/vendor//*[text()='INTEL']",
	"//host_info/tpm_version//*[text()='2.0']":                      "//meta/description/tpm_version//*[text()='2.0']"}

var flavorTemplatePath = "/opt/hvs-flavortemplates"

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

	oldFlavorPartFilePath := flag.String("o", "", "old flavor part json file")
	versionFlag := flag.Bool("version", false, "Print the current version and exit")

	// Showing useful information when the user enters the -h|--help option
	flag.Usage = func() {
		if len(os.Args) <= 2 && !*versionFlag && *oldFlavorPartFilePath == "" {
			fmt.Println(helpStr)
		} else {
			fmt.Println("Invalid Command Usage")
			fmt.Printf(helpStr)
		}
	}

	flag.Parse()

	// Show the current version when the user enters the -version option
	if *versionFlag && *oldFlavorPartFilePath != "" {
		fmt.Println("Invalid Command Usage")
		fmt.Printf(helpStr)
		os.Exit(1)
	} else if *versionFlag && *oldFlavorPartFilePath == "" {
		fmt.Println("Current build version: ", BuildVersion)
		os.Exit(1)
	} else if *oldFlavorPartFilePath == "" {
		// Checks for the file data that was entered by the user
		fmt.Println("Error: Old flavor part json file path is missing")
		fmt.Printf(helpStr)
		os.Exit(1)
	}

	// Validating the old flavor part file path entered
	if valid, err := checkIfValidFile(*oldFlavorPartFilePath); err != nil && !valid {
		fmt.Println("Error in validating the input file path - ", err)
		os.Exit(1)
	}

	//reading the data from oldFlavorPartFilePath
	body, err := ioutil.ReadFile(*oldFlavorPartFilePath)
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
		fmt.Println("Error in unmarshaling the old flavor part file", err)
		os.Exit(1)
	}

	for flavorIndex, flavor := range oldFlavorPart.SignedFlavor {

		//Updating meta section
		if flavor.Flavor.Hardware != nil && flavor.Flavor.Hardware.Feature.CBNT.Enabled != nil && flavor.Flavor.Hardware.Feature.CBNT.Enabled.(bool) {
			oldFlavorPart.SignedFlavor[flavorIndex].Flavor.Meta.Description.CbntEnabled = true
		} else if flavor.Flavor.Hardware != nil && flavor.Flavor.Hardware.Feature.SUEFI != nil && flavor.Flavor.Hardware.Feature.SUEFI.Enabled {
			oldFlavorPart.SignedFlavor[flavorIndex].Flavor.Meta.Description.SuefiEnabled = true
		}

		if flavor.Flavor.Meta.Vendor == intelVendor {
			oldFlavorPart.SignedFlavor[flavorIndex].Flavor.Meta.Description.Vendor = "Linux"
		} else if flavor.Flavor.Meta.Vendor == vmwareVendor {
			oldFlavorPart.SignedFlavor[flavorIndex].Flavor.Meta.Description.Vendor = "VMware"
		}

		//Updating hardware section
		if flavor.Flavor.Hardware != nil {
			//TXT
			if flavor.Flavor.Hardware.Feature.TXT.Enabled != nil {
				oldFlavorPart.SignedFlavor[flavorIndex].Flavor.Hardware.Feature.TXT.Enabled = strconv.FormatBool(flavor.Flavor.Hardware.Feature.TXT.Enabled.(bool))
			} else {
				//if the TXT section not present in oldflavorpart json,assign false to it
				oldFlavorPart.SignedFlavor[flavorIndex].Flavor.Hardware.Feature.TXT.Enabled = "false"
			}

			//TPM
			if flavor.Flavor.Hardware.Feature.TPM.Enabled != nil {
				oldFlavorPart.SignedFlavor[flavorIndex].Flavor.Hardware.Feature.TPM.Enabled = strconv.FormatBool(flavor.Flavor.Hardware.Feature.TPM.Enabled.(bool))
				oldFlavorPart.SignedFlavor[flavorIndex].Flavor.Hardware.Feature.TPM.Meta.TPMVersion = flavor.Flavor.Hardware.Feature.TPM.Version
				flavor.Flavor.Hardware.Feature.TPM.Version = ""
				oldFlavorPart.SignedFlavor[flavorIndex].Flavor.Hardware.Feature.TPM.Meta.PCRBanks = flavor.Flavor.Hardware.Feature.TPM.PcrBanks
				flavor.Flavor.Hardware.Feature.TPM.PcrBanks = nil
			} else {
				//if the TPM section not present in oldflavorpart json,assign false to it
				oldFlavorPart.SignedFlavor[flavorIndex].Flavor.Hardware.Feature.TPM.Enabled = "false"
			}

			//CBNT
			if flavor.Flavor.Hardware.Feature.CBNT.Enabled != nil {
				oldFlavorPart.SignedFlavor[flavorIndex].Flavor.Hardware.Feature.CBNT.Enabled = strconv.FormatBool(flavor.Flavor.Hardware.Feature.CBNT.Enabled.(bool))
				oldFlavorPart.SignedFlavor[flavorIndex].Flavor.Hardware.Feature.CBNT.Meta.Profile = flavor.Flavor.Hardware.Feature.CBNT.Profile
				flavor.Flavor.Hardware.Feature.CBNT.Profile = ""
			} else {
				//if the CBNT section not present in oldflavorpart json,assign false to it
				oldFlavorPart.SignedFlavor[flavorIndex].Flavor.Hardware.Feature.CBNT.Enabled = "false"
			}

			//UEFI
			if flavor.Flavor.Hardware.Feature.SUEFI != nil {
				oldFlavorPart.SignedFlavor[flavorIndex].Flavor.Hardware.Feature.UEFI.Enabled = flavor.Flavor.Hardware.Feature.SUEFI.Enabled
				oldFlavorPart.SignedFlavor[flavorIndex].Flavor.Hardware.Feature.UEFI.Meta.SecureBootEnabled = flavor.Flavor.Hardware.Feature.SUEFI.Enabled
				flavor.Flavor.Hardware.Feature.SUEFI = nil
			}
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

				if flavorname == platformFlavor && template.FlavorParts.Platform != nil {
					for _, rules := range template.FlavorParts.Platform.PcrRules {
						pcrsmap[rules.Pcr.Index] = rules.Pcr.Bank
					}
					rules = template.FlavorParts.Platform.PcrRules

				} else if flavorname == osFlavor && template.FlavorParts.OS != nil {
					for _, rules := range template.FlavorParts.OS.PcrRules {
						pcrsmap[rules.Pcr.Index] = rules.Pcr.Bank
					}
					rules = template.FlavorParts.OS.PcrRules

				} else if flavorname == hostUniqueFlavor && template.FlavorParts.HostUnique != nil {
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
								newTpmEvents[eventIndex].TypeName = oldEvents.Label
								newTpmEvents[eventIndex].Tags = append(newTpmEvents[eventIndex].Tags, oldEvents.Label)
								newTpmEvents[eventIndex].Measurement = oldEvents.Value
								newTpmEvents[eventIndex].TypeID = eventIDList[oldEvents.Label]
							}
							newFlavorPcrs[index].EventlogEqual.Events = newTpmEvents
							newTpmEvents = nil
						}

						if expectedPcrEx.Event != nil && !reflect.ValueOf(rules[index].EventlogIncludes).IsZero() {
							newTpmEvents = make([]NewEventLog, len(expectedPcrEx.Event))
							for eventIndex, oldEvents := range expectedPcrEx.Event {
								newTpmEvents[eventIndex].TypeName = oldEvents.Label
								newTpmEvents[eventIndex].Tags = append(newTpmEvents[eventIndex].Tags, oldEvents.Label)
								newTpmEvents[eventIndex].Measurement = oldEvents.Value
								newTpmEvents[eventIndex].TypeID = eventIDList[oldEvents.Label]
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

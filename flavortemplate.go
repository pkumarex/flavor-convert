/*
 * Copyright (C) 2020 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package main

import (
	"github.com/google/uuid"
)

type FVMeta struct {
	FlavorPart     string     `json:"flavor_part,omitempty"`
	Source         string     `json:"source,omitempty"`
	Label          string     `json:"label,omitempty"`
	Vendot         string     `json:"vendor,omitempty"`
	IPAddress      string     `json:"ip_address,omitempty"`
	BiosName       string     `json:"bios_name,omitempty"`
	BiosVersion    string     `json:"bios_version,omitempty"`
	OsName         string     `json:"os_name,omitempty"`
	OsVersion      string     `json:"os_version,omitempty"`
	VmmName        string     `json:"vmm_name,omitempty"`
	VmmVersion     string     `json:"vmm_version,omitempty"`
	TpmVersion     string     `json:"tpm_version,omitempty"`
	HardwareUUID   *uuid.UUID `json:"hardware_uuid,omitempty"`
	Comment        string     `json:"comment,omitempty"`
	TbootInstalled bool       `json:"tboot_installed,omitempty"`
	CBNTEnabled    bool       `json:"cbnt_enabled,omitempty"`
	UEFIEnabled    bool       `json:"uefi_enabled,omitempty"`

	DigestAlgorithm string `json:"digest_algorithm,omitempty"`
}

type PCR struct {
	Index int    `json:"index"`
	Bank  string `json:"bank"`
}

type EventLogEquals struct {
	ExcludingTags []string `json:"excluding_tags"`
}

type PcrRules []struct {
	Pcr              PCR            `json:"pcr"`
	PcrMatches       bool           `json:"pcr_matches"`
	EventlogEquals   EventLogEquals `json:"eventlog_equals,omitempty"`
	EventlogIncludes []string       `json:"eventlog_includes,omitempty"`
}

type FlavorPart struct {
	FVMeta   *FVMeta  `json:"meta,omitempty"`
	PcrRules PcrRules `json:"pcr_rules"`
}

type FlavorParts struct {
	Platform   *FlavorPart `json:"PLATFORM,omitempty"`
	OS         *FlavorPart `json:"OS,omitempty"`
	Software   *FlavorPart `json:"SOFTWARE,omitempty"`
	HostUnique *FlavorPart `json:"HOST_UNIQUE,omitempty"`
	AssetTag   *FlavorPart `json:"ASSET_TAG,omitempty"`
}

type FlavorTemplate struct {
	ID          uuid.UUID   `json:"id,omitempty"`
	Label       string      `json:"label"`
	Condition   []string    `json:"condition"`
	FlavorParts FlavorParts `json:"flavor_parts,omitempty"`
}

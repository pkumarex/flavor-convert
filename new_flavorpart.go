/*
 * Copyright (C) 2020 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package main

import "fmt"

type PcrLogs struct {
	PCR              PCR             `json:"pcr"`         //required
	Measurement      string          `json:"measurement"` //required
	PCRMatches       bool            `json:"pcr_matches,omitempty"`
	EventlogEqual    *EventlogEquals `json:"eventlog_equals,omitempty"`
	EventlogIncludes []NewEventLog   `json:"eventlog_includes,omitempty"`
}

type EventlogEquals struct {
	Events      []NewEventLog `json:"events,omitempty"`
	ExcludeTags []string      `json:"exclude_tags,omitempty"`
}

type NewEventLog struct {
	TypeID      string   `json:"type_id,omiempty"`
	TypeName    string   `json:"type_name,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Measurement string   `json:"measurement"`
}

type PcrIndex int

// String returns the string representation of the PcrIndex
func (p PcrIndex) String() string {
	return fmt.Sprintf("pcr_%d", p)
}

const (
	PCR0 PcrIndex = iota
	PCR1
	PCR2
	PCR3
	PCR4
	PCR5
	PCR6
	PCR7
	PCR8
	PCR9
	PCR10
	PCR11
	PCR12
	PCR13
	PCR14
	PCR15
	PCR16
	PCR17
	PCR18
	PCR19
	PCR20
	PCR21
	PCR22
	PCR23
	INVALID_INDEX = -1
)

type SHAAlgorithm string

const (
	SHA1    SHAAlgorithm = "SHA1"
	SHA256  SHAAlgorithm = "SHA256"
	SHA384  SHAAlgorithm = "SHA384"
	SHA512  SHAAlgorithm = "SHA512"
	UNKNOWN SHAAlgorithm = "unknown"
)

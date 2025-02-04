/*
 *     Copyright 2022 The Dragonfly Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package types

import (
	"encoding/json"
	"errors"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// PEMContent supports load PEM format from file or just inline PEM format content
type PEMContent string

func (p *PEMContent) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	return p.loadPEM(s)
}

func (p *PEMContent) UnmarshalYAML(node *yaml.Node) error {
	var s string
	switch node.Kind {
	case yaml.ScalarNode:
		if err := node.Decode(&s); err != nil {
			return err
		}
	default:
		return errors.New("invalid pem content")
	}

	return p.loadPEM(s)
}

func (p *PEMContent) loadPEM(content string) error {
	if content == "" {
		*p = PEMContent("")
		return nil
	}

	// inline PEM, just return
	if strings.HasPrefix(strings.TrimSpace(content), "-----BEGIN ") {
		val := strings.TrimSpace(content)
		*p = PEMContent(val)
		return nil
	}

	file, err := os.ReadFile(content)
	if err != nil {
		return err
	}
	val := strings.TrimSpace(string(file))
	*p = PEMContent(val)
	return nil
}

// HostType is the type of host.
type HostType int

const (
	// HostTypeNormal is the normal type of host.
	HostTypeNormal HostType = iota

	// HostTypeSuperSeed is the super seed type of host.
	HostTypeSuperSeed

	// HostTypeStrongSeed is the strong seed type of host.
	HostTypeStrongSeed

	// HostTypeWeakSeed is the weak seed type of host.
	HostTypeWeakSeed
)

const (
	// HostTypeNormalName is the name of normal host type.
	HostTypeNormalName = "normal"

	// HostTypeSuperSeedName is the name of super host type.
	HostTypeSuperSeedName = "super"

	// HostTypeStrongSeedName is the name of strong host type.
	HostTypeStrongSeedName = "strong"

	// HostTypeWeakSeedName is the name of weak host type.
	HostTypeWeakSeedName = "weak"
)

// Name returns the name of host type.
func (h *HostType) Name() string {
	switch *h {
	case HostTypeSuperSeed:
		return HostTypeSuperSeedName
	case HostTypeStrongSeed:
		return HostTypeStrongSeedName
	case HostTypeWeakSeed:
		return HostTypeWeakSeedName
	}

	return HostTypeNormalName
}

// ParseHostType parses host type by name.
func ParseHostType(name string) HostType {
	switch name {
	case HostTypeSuperSeedName:
		return HostTypeSuperSeed
	case HostTypeStrongSeedName:
		return HostTypeStrongSeed
	case HostTypeWeakSeedName:
		return HostTypeWeakSeed
	}

	return HostTypeNormal
}

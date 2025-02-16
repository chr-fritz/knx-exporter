// Copyright © 2020-2025 Christian Fritz <mail@chr-fritz.de>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package knx

import (
	"encoding/json"
	"strings"

	"github.com/vapourismo/knx-go/knx/cemi"
)

// PhysicalAddress defines an individual address of a knx device.
type PhysicalAddress cemi.IndividualAddr

const InvalidPhysicalAddress = PhysicalAddress(0)

// NewPhysicalAddress creates a new knx device PhysicalAddress by parsing the given string. It either returns the parsed
// PhysicalAddress or an error if it is not possible to parse the string.
func NewPhysicalAddress(str string) (PhysicalAddress, error) {
	pa, e := cemi.NewIndividualAddrString(str)
	if e != nil {
		return InvalidPhysicalAddress, e
	}
	return PhysicalAddress(pa), nil
}

func (g PhysicalAddress) String() string {
	return cemi.IndividualAddr(g).String()
}

func (g PhysicalAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(g.String())
}

func (g PhysicalAddress) MarshalText() ([]byte, error) {
	return []byte(g.String()), nil
}

func (g *PhysicalAddress) UnmarshalJSON(data []byte) error {
	str := string(data)
	str = strings.Trim(str, "\"'")
	ga, e := cemi.NewIndividualAddrString(str)
	if e != nil {
		return e
	}
	*g = PhysicalAddress(ga)
	return nil
}
func (g *PhysicalAddress) UnmarshalText(data []byte) error {
	return g.UnmarshalJSON(data)
}

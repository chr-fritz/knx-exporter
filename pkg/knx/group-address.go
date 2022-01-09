// Copyright © 2020-2022 Christian Fritz <mail@chr-fritz.de>
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

// GroupAddress defines a single group address. It do not contain any additional information about purpose, data types
// or allowed telegram types.
type GroupAddress cemi.GroupAddr

// InvalidGroupAddress defines the nil group address.
const InvalidGroupAddress = GroupAddress(0)

// NewGroupAddress creates a new GroupAddress by parsing the given string. It either returns the parsed GroupAddress or
// an error if it is not possible to parse the string.
func NewGroupAddress(str string) (GroupAddress, error) {
	ga, e := cemi.NewGroupAddrString(str)
	if e != nil {
		return InvalidGroupAddress, e
	}
	return GroupAddress(ga), nil
}

func (g GroupAddress) String() string {
	return cemi.GroupAddr(g).String()
}

func (g GroupAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(g.String())
}

func (g GroupAddress) MarshalText() ([]byte, error) {
	return []byte(g.String()), nil
}

func (g *GroupAddress) UnmarshalJSON(data []byte) error {
	str := string(data)
	str = strings.Trim(str, "\"'")
	ga, e := cemi.NewGroupAddrString(str)
	if e != nil {
		return e
	}
	*g = GroupAddress(ga)
	return nil
}
func (g *GroupAddress) UnmarshalText(data []byte) error {
	return g.UnmarshalJSON(data)
}

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

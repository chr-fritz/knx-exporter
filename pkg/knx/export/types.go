package export

import (
	"encoding/xml"
)

// GroupAddressExport is the root element of the xml file which generates the ETS 5 application
// while the group address export.
type GroupAddressExport struct {
	XMLName    xml.Name     `xml:"GroupAddress-Export"`
	Xmlns      string       `xml:"xmlns,attr"`
	GroupRange []GroupRange `xml:"GroupRange"`
}

// GroupRange defines a rage of group addresses in the ETS 5 group address export.
type GroupRange struct {
	Name         string         `xml:"Name,attr"`
	RangeStart   uint16         `xml:"RangeStart,attr"`
	RangeEnd     uint16         `xml:"RangeEnd,attr"`
	GroupRange   []GroupRange   `xml:"GroupRange"`
	GroupAddress []GroupAddress `xml:"GroupAddress"`
}

// GroupAddress defines a a single group address in the ETS 5 group address export.
type GroupAddress struct {
	Name        string `xml:"Name,attr"`
	Address     string `xml:"Address,attr"`
	Central     bool   `xml:"Central,attr"`
	Unfiltered  bool   `xml:"Unfiltered,attr"`
	DPTs        string `xml:"DPTs,attr"`
	Description string `xml:"Description,attr"`
}

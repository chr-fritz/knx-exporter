package knx

import "github.com/vapourismo/knx-go/knx"

// GroupClient is a super interface for the knx.GroupClient interface to also export the Close() function.
type GroupClient interface {
	Send(event knx.GroupEvent) error
	Inbound() <-chan knx.GroupEvent
	Close()
}

// DPT is wrapper interface for all types under github.com/vapourismo/knx-go/knx/dpt to simplifies working with them.
type DPT interface {
	Pack() []byte
	Unpack(data []byte) error
	Unit() string
	String() string
}

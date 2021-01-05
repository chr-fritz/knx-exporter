package knx

import "github.com/vapourismo/knx-go/knx"

type GroupClient interface {
	Send(event knx.GroupEvent) error
	Inbound() <-chan knx.GroupEvent
	Close()
}

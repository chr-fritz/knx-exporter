package knx

//go:generate mockgen -destination=adapterMocks_test.go -package=knx -source=adapter.go

import (
	"github.com/sirupsen/logrus"
	"github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/util"
)

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

func init() {
	util.Logger = logger{}
}

type logger struct{}

func (l logger) Printf(format string, args ...interface{}) {
	logrus.Debugf(format, args...)
}

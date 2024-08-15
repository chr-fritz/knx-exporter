// Copyright © 2020-2024 Christian Fritz <mail@chr-fritz.de>
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

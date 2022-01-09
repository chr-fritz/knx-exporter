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

package cmd

import (
	"fmt"
	"testing"
	"time"

	knxFake "github.com/chr-fritz/knx-exporter/pkg/knx/fake"
	metricsFake "github.com/chr-fritz/knx-exporter/pkg/metrics/fake"
	"github.com/golang/mock/gomock"
)

func TestRunOptions_aliveCheck(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	i := &RunOptions{
		restart:            "exit",
		aliveCheckInterval: 10 * time.Millisecond,
	}

	exporter := metricsFake.NewMockExporter(ctrl)
	knxExporter := knxFake.NewMockMetricsExporter(ctrl)

	knxExporter.EXPECT().
		IsAlive().
		MinTimes(1).
		Return(nil).
		Return(fmt.Errorf("no connection"))

	exporter.EXPECT().
		Shutdown().
		Times(1)
	i.aliveCheck(exporter, knxExporter)
}

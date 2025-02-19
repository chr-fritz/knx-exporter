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

package cmd

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/spf13/viper"

	knxFake "github.com/chr-fritz/knx-exporter/pkg/knx/fake"
	"github.com/golang/mock/gomock"
)

func TestRunOptions_aliveCheck(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx, cancelFunc := context.WithTimeout(context.Background(), 15*time.Millisecond)
	defer cancelFunc()

	viper.Set(RunRestartParm, "exit")
	i := &RunOptions{
		aliveCheckInterval: 10 * time.Millisecond,
	}

	knxExporter := knxFake.NewMockMetricsExporter(ctrl)

	knxExporter.EXPECT().
		IsAlive().
		MinTimes(1).
		Return(nil).
		Return(fmt.Errorf("no connection"))

	go i.aliveCheck(ctx, cancelFunc, knxExporter)
	time.Sleep(20 * time.Millisecond)
}

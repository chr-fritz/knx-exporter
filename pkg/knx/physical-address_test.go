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

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhysicalAddress_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		g       PhysicalAddress
		want    []byte
		wantErr bool
	}{
		{"0.0.0", PhysicalAddress(0), []byte("\"0.0.0\""), false},
		{"0.0.1", PhysicalAddress(1), []byte("\"0.0.1\""), false},
		{"0.1.0", PhysicalAddress(0x100), []byte("\"0.1.0\""), false},
		{"15.15.0", PhysicalAddress(0xFF00), []byte("\"15.15.0\""), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.g.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPhysicalAddress_MarshalText(t *testing.T) {
	tests := []struct {
		name    string
		g       PhysicalAddress
		want    []byte
		wantErr bool
	}{
		{"0.0.0", PhysicalAddress(0), []byte("0.0.0"), false},
		{"0.0.1", PhysicalAddress(1), []byte("0.0.1"), false},
		{"0.1.0", PhysicalAddress(0x100), []byte("0.1.0"), false},
		{"15.15.0", PhysicalAddress(0xFF00), []byte("15.15.0"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.g.MarshalText()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPhysicalAddress_String(t *testing.T) {
	tests := []struct {
		name string
		g    PhysicalAddress
		want string
	}{
		{"0.0.0", PhysicalAddress(0), "0.0.0"},
		{"0.0.1", PhysicalAddress(1), "0.0.1"},
		{"0.1.0", PhysicalAddress(0x100), "0.1.0"},
		{"15.15.0", PhysicalAddress(0xFF00), "15.15.0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.g.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPhysicalAddress_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    PhysicalAddress
		wantErr bool
	}{
		{"0.0.0", []byte("0.0.0"), PhysicalAddress(0), true},
		{"0.0.1", []byte("0.0.1"), PhysicalAddress(1), false},
		{"0.1.0", []byte("0.1.0"), PhysicalAddress(0x100), false},
		{"15.15.0", []byte("15.15.0"), PhysicalAddress(0xFF00), false},
		{"a.b.c", []byte("a.b.c"), PhysicalAddress(0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := PhysicalAddress(0)
			if err := g.UnmarshalJSON(tt.data); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, g)
		})
	}
}

func TestPhysicalAddress_UnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    PhysicalAddress
		wantErr bool
	}{
		{"0.0.0", []byte("0.0.0"), PhysicalAddress(0), true},
		{"0.0.1", []byte("0.0.1"), PhysicalAddress(1), false},
		{"0.1.0", []byte("0.1.0"), PhysicalAddress(0x100), false},
		{"15.15.0", []byte("15.15.0"), PhysicalAddress(0xFF00), false},
		{"a.b.c", []byte("a.b.c"), PhysicalAddress(0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := PhysicalAddress(0)
			if err := g.UnmarshalText(tt.data); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, g)
		})
	}
}

func TestNewPhysicalAddress(t *testing.T) {
	tests := []struct {
		name    string
		want    PhysicalAddress
		wantErr bool
	}{
		{"0.0.0", PhysicalAddress(0), true},
		{"0.0.1", PhysicalAddress(1), false},
		{"0.0.1", PhysicalAddress(1), false},
		{"0.1.0", PhysicalAddress(0x100), false},
		{"15.15.0", PhysicalAddress(0xFF00), false},
		{"31.7", PhysicalAddress(0x1f07), false},
		{"31", PhysicalAddress(0x1f), false},
		{"a.b.c", PhysicalAddress(0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPhysicalAddress(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPhysicalAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

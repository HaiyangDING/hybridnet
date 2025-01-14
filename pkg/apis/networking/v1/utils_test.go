/*
 Copyright 2021 The Hybridnet Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package v1

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateAddressRange(t *testing.T) {
	tests := []struct {
		name         string
		addressRange *AddressRange
		expectError  error
	}{
		{
			"unknown version",
			&AddressRange{
				Version: IPVersion("123"),
			},
			fmt.Errorf("unsupported IP Version 123"),
		},
		{
			"wrong start",
			&AddressRange{
				Version: IPv4,
				Start:   "192.168.8",
			},
			fmt.Errorf("invalid range start 192.168.8"),
		},
		{
			"wrong end",
			&AddressRange{
				Version: IPv4,
				End:     "192.168.8",
			},
			fmt.Errorf("invalid range end 192.168.8"),
		},
		{
			"wrong gateway",
			&AddressRange{
				Version: IPv4,
				CIDR:    "192.168.8.0/24",
				Gateway: "192.168.8",
			},
			fmt.Errorf("invalid range gateway 192.168.8"),
		},
		{
			"IP family mismatch",
			&AddressRange{
				Version: IPv6,
				CIDR:    "192.168.8.0/24",
				Gateway: "192.168.8.254",
			},
			fmt.Errorf("address families of ip version and gateway mismatch"),
		},
		{
			"wrong CIDR",
			&AddressRange{
				Version: IPv4,
				CIDR:    "192.168.8/24",
				Gateway: "192.168.8.254",
			},
			fmt.Errorf("invalid range CIDR 192.168.8/24"),
		},
		{
			"non-standard CIDR",
			&AddressRange{
				Version: IPv4,
				CIDR:    "192.168.8.10/24",
				Gateway: "192.168.8.254",
			},
			fmt.Errorf("CIDR notation is not standard, should start from 192.168.8.0 but from 192.168.8.10"),
		},
		{
			"start is not in range",
			&AddressRange{
				Version: IPv4,
				Start:   "192.168.9.1",
				CIDR:    "192.168.8.0/24",
				Gateway: "192.168.8.254",
			},
			fmt.Errorf("start 192.168.9.1 is not in CIDR 192.168.8.0/24"),
		},
		{
			"end is not in range",
			&AddressRange{
				Version: IPv4,
				End:     "192.168.9.1",
				CIDR:    "192.168.8.0/24",
				Gateway: "192.168.8.254",
			},
			fmt.Errorf("end 192.168.9.1 is not in CIDR 192.168.8.0/24"),
		},
		{
			"start is big than end",
			&AddressRange{
				Version: IPv4,
				Start:   "192.168.8.11",
				End:     "192.168.8.10",
				CIDR:    "192.168.8.0/24",
				Gateway: "192.168.8.254",
			},
			fmt.Errorf("subnet should have at least one available IP. start=192.168.8.11, end=192.168.8.10"),
		},
		{
			"gateway is not in range",
			&AddressRange{
				Version: IPv4,
				CIDR:    "192.168.8.0/24",
				Gateway: "192.168.9.254",
			},
			fmt.Errorf("gateway 192.168.9.254 is not in CIDR 192.168.8.0/24"),
		},
		{
			"wrong reserved ip",
			&AddressRange{
				Version: IPv4,
				CIDR:    "192.168.8.0/24",
				Gateway: "192.168.8.254",
				ReservedIPs: []string{
					"192.168.8",
				},
			},
			fmt.Errorf("invalid reserved ip 192.168.8"),
		},
		{
			"reserved ip is not in range",
			&AddressRange{
				Version: IPv4,
				CIDR:    "192.168.8.0/24",
				Gateway: "192.168.8.254",
				ReservedIPs: []string{
					"192.168.9.100",
				},
			},
			fmt.Errorf("reserved ip 192.168.9.100 is not in CIDR 192.168.8.0/24"),
		},
		{
			"wrong excluded ip",
			&AddressRange{
				Version: IPv4,
				CIDR:    "192.168.8.0/24",
				Gateway: "192.168.8.254",
				ExcludeIPs: []string{
					"192.168.8",
				},
			},
			fmt.Errorf("invalid excluded ip 192.168.8"),
		},
		{
			"excluded ip is not in range",
			&AddressRange{
				Version: IPv4,
				CIDR:    "192.168.8.0/24",
				Gateway: "192.168.8.254",
				ExcludeIPs: []string{
					"192.168.9.100",
				},
			},
			fmt.Errorf("excluded ip 192.168.9.100 is not in CIDR 192.168.8.0/24"),
		},
		{
			"normal",
			&AddressRange{
				Version: IPv4,
				Start:   "192.168.8.10",
				End:     "192.168.8.100",
				CIDR:    "192.168.8.0/24",
				Gateway: "192.168.8.254",
				ReservedIPs: []string{
					"192.168.8.50",
				},
				ExcludeIPs: []string{
					"192.168.8.90",
				},
			},
			nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateAddressRange(test.addressRange)
			if test.expectError == nil {
				assert.Nil(t, err)
			} else {
				assert.EqualError(t, err, test.expectError.Error())
			}
		})
	}
}

func TestCalculateCapacity(t *testing.T) {
	tests := []struct {
		name         string
		addressRange *AddressRange
		expected     int64
	}{
		{
			"invalid cidr",
			&AddressRange{
				CIDR: "fake",
			},
			math.MaxInt64,
		},
		{
			"only cidr",
			&AddressRange{
				CIDR: "192.168.0.0/24",
			},
			254,
		},
		{
			"cidr and start",
			&AddressRange{
				Start: "192.168.0.100",
				CIDR:  "192.168.0.0/24",
			},
			155,
		},
		{
			"cidr, start and end",
			&AddressRange{
				Start: "192.168.0.100",
				End:   "192.168.0.200",
				CIDR:  "192.168.0.0/24",
			},
			101,
		},
		{
			"all set",
			&AddressRange{
				Start: "192.168.0.100",
				End:   "192.168.0.200",
				CIDR:  "192.168.0.0/24",
				ExcludeIPs: []string{
					"192.168.0.105",
					"192.168.0.107",
				},
			},
			99,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if calculated := CalculateCapacity(test.addressRange); calculated != test.expected {
				t.Fatalf("test %s fails, expected %d but calculated %d", test.name, test.expected, calculated)
			}
		})
	}
}

func TestGetNetworkType(t *testing.T) {
	tests := []struct {
		name        string
		network     *Network
		networkType NetworkType
	}{
		{
			"empty",
			&Network{
				Spec: NetworkSpec{
					Type: "",
				},
			},
			NetworkTypeUnderlay,
		},
		{
			"nil",
			nil,
			NetworkTypeUnderlay,
		},
		{
			"overlay",
			&Network{
				Spec: NetworkSpec{
					Type: "Overlay",
				},
			},
			NetworkTypeOverlay,
		},
		{
			"others",
			&Network{
				Spec: NetworkSpec{
					Type: "others",
				},
			},
			NetworkType("others"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if GetNetworkType(test.network) != test.networkType {
				t.Errorf("test %s fails, expect %s but got %s", test.name, test.networkType, GetNetworkType(test.network))
			}
		})
	}
}

// Copyright (c) 2020 Cisco Systems, Inc.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package kerneltap provides networkservice chain elements that support the kernel Mechanism via tapv2
package kerneltap

import (
	"context"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/trace"

	"github.com/ligato/vpp-agent/api/models/linux"
	linuxinterfaces "github.com/ligato/vpp-agent/api/models/linux/interfaces"
	linuxnamespace "github.com/ligato/vpp-agent/api/models/linux/namespace"
	vppinterfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"

	"github.com/networkservicemesh/api/pkg/api/connection"
	"github.com/networkservicemesh/api/pkg/api/connection/mechanisms/common"
	"github.com/networkservicemesh/api/pkg/api/connection/mechanisms/kernel"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
	"github.com/networkservicemesh/sdk-vppagent/pkg/tools/netnsinode"
)

func appendInterfaceConfig(ctx context.Context, conn *connection.Connection, name string) error {
	if mechanism := kernel.ToMechanism(conn.GetMechanism()); mechanism != nil {
		conf := vppagent.Config(ctx)
		// We append an Interfaces.  Interfaces creates the vpp side of an interface.
		//   In this case, a Tapv2 interface that has one side in vpp, and the other
		//   as a Linux kernel interface
		conf.GetVppConfig().Interfaces = append(conf.GetVppConfig().Interfaces, &vppinterfaces.Interface{
			Name:    name,
			Type:    vppinterfaces.Interface_TAP,
			Enabled: true,
			Link: &vppinterfaces.Interface_Tap{
				Tap: &vppinterfaces.TapLink{
					Version: 2,
				},
			},
		})
		filepath, err := netnsinode.LinuxNetNsFileName(mechanism.GetParameters()[common.NetNSInodeKey])
		if err != nil {
			return err
		}
		trace.Log(ctx).Info("Found /dev/vhost-net - using tapv2")
		// We apply configuration to LinuxInterfaces
		// Important details:
		//    - LinuxInterfaces.HostIfName - must be no longer than 15 chars (linux limitation)
		conf.GetLinuxConfig().Interfaces = append(conf.GetLinuxConfig().Interfaces, &linux.Interface{
			Name:    name,
			Type:    linuxinterfaces.Interface_TAP_TO_VPP,
			Enabled: true,
			// TODO - fix this to have a proper getter in the mechanisms/kernel package and use it here
			HostIfName: mechanism.GetParameters()[common.InterfaceNameKey],
			Namespace: &linuxnamespace.NetNamespace{
				Type:      linuxnamespace.NetNamespace_FD,
				Reference: filepath,
			},
			Link: &linuxinterfaces.Interface_Tap{
				Tap: &linuxinterfaces.TapLink{
					VppTapIfName: name,
				},
			},
		})
	}
	return nil
}

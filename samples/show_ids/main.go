/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */
package main

import (
	"fmt"
	"os"

	"github.com/HYGON-AI/hcu-dcgm/v3/pkg/dcgm"
)

func main() {
	if err := dcgm.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "dcgm.Init 失败: %v\n", err)
		os.Exit(1)
	}
	defer dcgm.ShutDown()

	numHcus, err := dcgm.NumMonitorDevices()
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取设备数量失败: %v\n", err)
		os.Exit(1)
	}

	line := "+--------+----------------------+--------------+----------------------+"
	fmt.Println(line)
	fmt.Printf("| %-6s | %-20s | %-12s | %-20s |\n", "HCU ID", "Unique ID", "PCI Bus ID", "Serial Number")
	fmt.Println(line)

	for i := 0; i < numHcus; i++ {
		uniqueId, err1 := dcgm.GetDeviceUniqueId(i)
		if err1 != nil {
			uniqueId = "N/A"
		}

		pciId, err2 := dcgm.PicBusInfo(i)
		if err2 != nil {
			pciId = "N/A"
		}

		serial, err3 := dcgm.GetDeviceId(i)
		if err3 != nil {
			serial = "N/A"
		}

		fmt.Printf("| %-6d | %-20s | %-12s | %-20s |\n", i, uniqueId, pciId, serial)
	}
	fmt.Println(line)
}

// Copyright (c) 2022 Vasiliy Vasilyuk. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package freeport

import (
	"errors"
	"fmt"
	"net"
)

// Much returns a slice of free ports of the specified size.
func Much(count int) ([]int, error) {
	if count < 1 {
		return nil, errors.New("freeport: not possible to create less than one port")
	}

	ports := make([]int, 0, count)

	for i := 0; i < count; i++ {
		addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
		if err != nil {
			const format = "freeport: cannot resolve tcp addr: %v"
			return nil, fmt.Errorf(format, err)
		}

		listener, err := net.ListenTCP("tcp", addr)
		if err != nil {
			const format = "freeport: cannot start listen: %v"
			return nil, fmt.Errorf(format, err)
		}

		if err := listener.Close(); err != nil {
			const format = "freeport: cannot close listener: %v"
			return nil, fmt.Errorf(format, err)
		}

		tcpAddr := listener.Addr().(*net.TCPAddr)
		ports = append(ports, tcpAddr.Port)
	}

	return ports, nil
}

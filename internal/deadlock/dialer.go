// Copyright (c) 2022 Vasiliy Vasilyuk. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package deadlock

import (
	"context"
	"errors"
	"fmt"
	"net"
)

var ErrDeadlockDial = errors.New("deadlock: impossible dial")

type Dealer struct{}

func (e Dealer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	format := "context %v, network %q, address %q: %w"
	return nil, fmt.Errorf(format, ctx, network, address, ErrDeadlockDial)
}

func (e Dealer) Dial(network, address string) (net.Conn, error) {
	format := "network %q, address %q: %w"
	return nil, fmt.Errorf(format, network, address, ErrDeadlockDial)
}

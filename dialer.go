// Copyright (c) 2022 Vasiliy Vasilyuk. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tornado

import (
	"context"
	"net"
)

// A ContextDialer dials using a context.
type ContextDialer interface {
	// DialContext connects to the address on the named network using
	// the provided context.
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// A dialer is a means to establish a connection.
type dialer interface {
	// Dial connects to the given address via the proxy.
	Dial(network, addr string) (c net.Conn, err error)
}

type comboDialer interface {
	dialer
	ContextDialer
}

type comboDialAdapter func(ctx context.Context, network, address string) (net.Conn, error)

func (f comboDialAdapter) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return f(ctx, network, address)
}

func (f comboDialAdapter) Dial(network, address string) (net.Conn, error) {
	return f(context.Background(), network, address)
}

// Copyright (c) 2022 Vasiliy Vasilyuk. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tornado

import (
	"context"
	"net"

	"golang.org/x/net/proxy"
)

// A ContextDialer dials using a context.
type ContextDialer interface {
	// DialContext connects to the address on the named network using
	// the provided context.
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type comboDialer interface {
	ContextDialer
	proxy.Dialer
}

type comboDialAdapter func(ctx context.Context, network, address string) (net.Conn, error)

func (f comboDialAdapter) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return f(ctx, network, address)
}

func (f comboDialAdapter) Dial(network, address string) (net.Conn, error) {
	return f(context.Background(), network, address)
}

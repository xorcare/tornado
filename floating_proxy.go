// Copyright (c) 2022 Vasiliy Vasilyuk. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tornado

import (
	"context"
	"net"

	"golang.org/x/net/proxy"
)

var _ proxy.Dialer = (*FloatingProxy)(nil)
var _ proxy.ContextDialer = (*FloatingProxy)(nil)

// NewFloatingProxy creates new instance of FloatingProxy.
func NewFloatingProxy(pool *Pool) *FloatingProxy {
	return &FloatingProxy{pool: pool}
}

// A FloatingProxy is an abstraction for making it easy to create connections
// through different proxies but within the same instance of the dial function.
//
// But do not forget that different implementations of clients for network
// interaction can keep the connection open for different requests while using
// one common Proxy chain, for example http.Client with http.DefaultTransport.
type FloatingProxy struct {
	pool *Pool
}

// Dial connects to the address on the named network.
//
// Dial uses context.Background internally; to specify the context, use
// DialContext.
//
// See func Dial of the net package of standard library for a
// description of the network and address parameters.
func (p *FloatingProxy) Dial(network, address string) (c net.Conn, err error) {
	return p.DialContext(context.Background(), network, address)
}

// DialContext connects to the address on the named network using
// the provided context.
//
// See func Dial of the net package of standard library for a
// description of the network and address parameters.
func (p *FloatingProxy) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if ctx == nil {
		panic("tornado: nil context")
	}

	prx := p.pool.Get()
	conn, err := prx.DialContext(ctx, network, address)
	p.pool.Put(prx)

	return conn, err
}

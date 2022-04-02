// Copyright (c) 2022 Vasiliy Vasilyuk. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tornado

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"sync"

	"golang.org/x/net/proxy"
)

var _ proxy.Dialer = (*Proxy)(nil)
var _ proxy.ContextDialer = (*Proxy)(nil)
var _ io.Closer = (*Proxy)(nil)

// NewProxy creates new instance of Proxy.
//
// If the Proxy was created using NewProxy, it must be closed wia using Close
// method after end usage to prevent memory leak and tor demon process leak.
//
// Pool must be closed wia using Close method after end usage to prevent memory
// leak and tor demon process leak, but keep in mind that all proxies will stop
// working immediately after the Pool is closed.
func NewProxy(ctx context.Context, ops ...Option) (*Proxy, error) {
	if ctx == nil {
		panic("tornado: nil context")
	}

	state := options{
		numberOfProxy: 1,
	}

	for _, option := range ops {
		option.apply(&state)
	}

	trc, err := newTorrcFromState(state)
	if err != nil {
		return nil, err
	}

	cmd, err := launchBackgroundTorDemon(ctx, trc)
	if err != nil {
		const format = "cannot run tor demon for a single proxy: %v"
		return nil, fmt.Errorf(format, err)
	}

	closeFunc := makeCloseFunc(cmd)

	prx, err := openSOCKS5Proxy(trc.socksPort[0], state.forwardDialer, closeFunc)
	if err != nil {
		const format = "cannot create proxy instance: %v"
		return nil, fmt.Errorf(format, err)
	}

	return prx, nil
}

// Proxy returns a ContextDialer that makes connections to the given
// address over tor network.
type Proxy struct {
	proxy proxy.ContextDialer

	valid     bool
	closeFunc func() error
	closeOnce sync.Once
}

// DialContext connects to the address on the named network over
// tor network using the provided context.
//
// The provided Context must be non-nil. If the context expires before
// the connection is complete, an error is returned. Once successfully
// connected, any expiration of the context will not affect the
// connection.
//
// See func Dial of the net package of standard library for a
// description of the network and address parameters.
func (p *Proxy) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if ctx == nil {
		panic("tornado: nil context")
	}

	return p.proxy.DialContext(ctx, network, address)
}

// Dial connects to the address on the named network over tor network.
//
// See func Dial of the net package of standard library for a
// description of the network and address parameters.
//
// Dial uses context.Background internally; to specify the context, use
// DialContext.
func (p *Proxy) Dial(network, address string) (c net.Conn, err error) {
	return p.proxy.DialContext(context.Background(), network, address)
}

// Close stops the tor demon running in the background.
// If the Proxy was created using NewPool directly instead of NewProxy,
// Close has no effect.
//
// This operation will not wait for active connections to close,
// they will be aborted.
func (p *Proxy) Close() (err error) {
	p.closeOnce.Do(func() {
		if p.closeFunc != nil {
			err = p.closeFunc()
		}

		// no need for a finalizer anymore.
		runtime.SetFinalizer(p, nil)
	})

	return err
}

func (p *Proxy) isValid() bool {
	return p != nil && p.valid
}

func openSOCKS5Proxy(port int, forward proxy.Dialer, closeFunc func() error) (*Proxy, error) {
	const socksFormat = "socks5://localhost:%d"
	socks5URL, err := url.Parse(fmt.Sprintf(socksFormat, port))
	if err != nil {
		const format = "cannot create socks5 url: %v"
		return nil, fmt.Errorf(format, err)
	}

	dialer, err := proxy.FromURL(socks5URL, forward)
	if err != nil {
		const format = "cannot create proxy dialer: %v"
		return nil, fmt.Errorf(format, err)
	}

	prx := &Proxy{
		proxy:     dialer.(proxy.ContextDialer),
		valid:     true,
		closeFunc: closeFunc,
	}
	runtime.SetFinalizer(prx, (*Proxy).Close)

	return prx, nil
}

func makeCloseFunc(cmd *exec.Cmd) func() error {
	return func() error {
		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			format := "an error occurred while sending, a signal to interrupt" +
				" the operation of the tor demon: %v"
			return fmt.Errorf(format, err)
		}

		if err := cmd.Wait(); err != nil {
			const format = "error while waiting is the result of sending," +
				" a signal to interrupt the command: %v"
			return fmt.Errorf(format, err)
		}

		return nil
	}
}

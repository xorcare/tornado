// Copyright (c) 2022 Vasiliy Vasilyuk. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package tornado provides support for an easy launch of tor proxy on golang.
package tornado // import "github.com/xorcare/tornado"

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"sync"
)

var _ io.Closer = (*Pool)(nil)

// NewPool creates new instance of proxy Pool.
//
// Pool must be closed wia using Close method after end usage to prevent memory
// leak and tor demon process leak, but keep in mind that all proxies will stop
// working immediately after the Pool is closed.
func NewPool(ctx context.Context, size int, ops ...Option) (*Pool, error) {
	if ctx == nil {
		panic("tornado: nil context")
	}

	state := options{
		numberOfProxy: size,
	}

	for _, option := range ops {
		option.apply(&state)
	}

	trc, err := newTorrcFromState(state)
	if err != nil {
		const format = "failed to create torrc: %v"
		return nil, fmt.Errorf(format, err)
	}

	cmd, err := launchBackgroundTorDemon(ctx, trc)
	if err != nil {
		const format = "failed to launch tor demon for the pool: %v"
		return nil, fmt.Errorf(format, err)
	}

	closeFunc := makeCloseFunc(cmd)
	pool := newFreePool(len(trc.socksPort), closeFunc)

	for _, port := range trc.socksPort {
		prx, err := openSOCKS5Proxy(port, state.forwardDialer, nil)
		if err != nil {
			const format = "cannot create proxy instance for pool: %v"
			return nil, fmt.Errorf(format, err)
		}

		pool.Put(prx)
	}

	return pool, nil
}

// A Pool is an abstraction for creating multiple proxy instances using
// a single tor process to reduce resource usage.
//
// Also, Pool provides a minimal interface for managing a set of proxies
// and their reuse.
type Pool struct {
	ch chan *Proxy

	closeFunc func() error
	closeOnce sync.Once
}

// Get gets a proxy instance from the pool.
//
// This operation can block the goroutine until a new proxy instance appears
// in the pool.
func (p *Pool) Get() *Proxy {
	return <-p.ch
}

// Put puts the proxy instance back in the pool.
//
// Do not try to manually fill the Pool in excess of the size specified when
// creating the pool, this can lead to blocking of the goroutine when pool
// overflows.
func (p *Pool) Put(prx *Proxy) {
	if !prx.isValid() {
		panic("tornado: not possible to put an invalid proxy instance in the proxy pool")
	}

	p.ch <- prx
}

// Close stops the tor demon running in the background.
//
// This operation will not wait for active connections to close,
// they will be aborted.
func (p *Pool) Close() (err error) {
	p.closeOnce.Do(func() {
		if p.closeFunc != nil {
			err = p.closeFunc()
		}

		// no need for a finalizer anymore.
		runtime.SetFinalizer(p, nil)
	})

	return err
}

func newFreePool(number int, closeFunc func() error) *Pool {
	pool := &Pool{
		ch:        make(chan *Proxy, number),
		closeFunc: closeFunc,
	}
	runtime.SetFinalizer(pool, (*Pool).Close)

	return pool
}

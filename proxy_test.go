// Copyright (c) 2022 Vasiliy Vasilyuk. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tornado

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/xorcare/tornado/internal/deadlock"
	"github.com/xorcare/tornado/internal/torproject"
)

const TestProxyServerStartupTimeout = 90 * time.Second

var WithTestTorrOptions = WithTorrcOption(os.Getenv("TORNADO_TEST_TORRC_OPTIONS"))

func TestNewProxy(t *testing.T) {
	t.Run("Should be get error impossible dial because set deadlock context dialer", func(t *testing.T) {
		// arrange
		ctx, done := context.WithTimeout(context.Background(), TestProxyServerStartupTimeout)
		t.Cleanup(done)

		prx, err := NewProxy(
			ctx,
			WithTestTorrOptions,
			WithForwardContextDialer(deadlock.Dealer{}),
		)
		if err != nil {
			t.Fatalf("cannot make proxy: %v", err)
		}

		// act
		_, err = prx.DialContext(ctx, "tcp", "localhost:80")

		// assert
		if !errors.Is(err, deadlock.ErrDeadlockDial) {
			const format = "deadlock.ErrDeadlockDial error was expected, but got another one: %v"
			log.Fatalf(format, err)
		}
	})

	t.Run("Should get a successful tor ip checking result", func(t *testing.T) {
		// arrange
		ctx, done := context.WithTimeout(context.Background(), TestProxyServerStartupTimeout)
		t.Cleanup(done)

		prx, err := NewProxy(
			ctx,
			WithTestTorrOptions,
		)
		if err != nil {
			t.Fatalf("cannot make proxy: %v", err)
		}
		defer prx.Close()

		t.Run("Check that the dial is working", func(t *testing.T) {
			// act
			cr, err := torproject.CheckDialer(prx)
			if err != nil {
				t.Fatalf("failed make torproject check: %v", err)
			}

			// assert
			t.Log("ip address received as a result of checking", cr.IP)
			if !cr.IsTor {
				t.Fatal("tor proxy server was not used")
			}
		})

		t.Run("Check that the dial context is working", func(t *testing.T) {
			// act
			cr, err := torproject.CheckContextDialer(prx)
			if err != nil {
				t.Fatalf("failed make torproject check: %v", err)
			}

			// assert
			t.Log("ip address received as a result of checking", cr.IP)
			if !cr.IsTor {
				t.Fatal("tor proxy server was not used")
			}
		})
	})
}

func TestProxy_Close(t *testing.T) {
	newProxy := func(t *testing.T) *Proxy {
		ctx, done := context.WithTimeout(context.Background(), TestProxyServerStartupTimeout)
		t.Cleanup(done)
		pool, err := NewProxy(ctx, WithTestTorrOptions)
		if err != nil {
			t.Fatal(err)
		}

		return pool
	}

	t.Run("Single calls to the close method do not return an error", func(t *testing.T) {
		// arrange
		pool := newProxy(t)

		// act
		err := pool.Close()

		// assert
		if err != nil {
			t.Fatal("should not get an error:", err)
		}
	})

	t.Run("Multiple calls to the close method do not return an error", func(t *testing.T) {
		// arrange
		pool := newProxy(t)
		var err error

		// act
		for i := 0; i < 10; i++ {
			err = pool.Close()
			if err != nil {
				break
			}
		}

		// assert
		if err != nil {
			t.Fatal("should not get an error:", err)
		}
	})
}

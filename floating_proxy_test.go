// Copyright (c) 2022 Vasiliy Vasilyuk. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tornado

import (
	"context"
	"fmt"
	"testing"

	"golang.org/x/net/proxy"

	"github.com/xorcare/tornado/internal/torproject"
)

var (
	_ proxy.Dialer        = (*FloatingProxy)(nil)
	_ proxy.ContextDialer = (*FloatingProxy)(nil)
)

func TestNewFloatingProxy(t *testing.T) {
	t.Parallel()
	t.Run("Should be successful check ip owner floating tor proxy", func(t *testing.T) {
		t.Parallel()
		// arrange
		ctx, done := context.WithTimeout(context.Background(), TestProxyServerStartupTimeout)
		t.Cleanup(done)

		pool, err := NewPool(
			ctx,
			10, // pool size
			WithTestTorrOptions,
		)
		if err != nil {
			t.Fatalf("cannot make proxy: %v", err)
		}

		prx := NewFloatingProxy(pool)

		for i := 0; i < 5; i++ {
			t.Run(fmt.Sprintf("Check %d", i), func(t *testing.T) {
				t.Parallel()
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
		}
	})
	t.Run("Proxy should work even if the pool is smaller than the number of requests", func(t *testing.T) {
		t.Parallel()
		// arrange
		ctx, done := context.WithTimeout(context.Background(), TestProxyServerStartupTimeout)
		t.Cleanup(done)

		pool, err := NewPool(
			ctx,
			1, // pool size
			WithTestTorrOptions,
		)
		if err != nil {
			t.Fatalf("cannot make proxy: %v", err)
		}

		prx := NewFloatingProxy(pool)

		for i := 0; i < 5; i++ {
			t.Run(fmt.Sprintf("Check %d", i), func(t *testing.T) {
				t.Parallel()
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
		}
	})
}

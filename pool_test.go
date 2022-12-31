// Copyright (c) 2022 Vasiliy Vasilyuk. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tornado

import (
	"context"
	"io"
	"testing"

	"github.com/xorcare/tornado/internal/torproject"
)

var _ io.Closer = (*Pool)(nil)

func TestNewPool(t *testing.T) {
	t.Parallel()
	t.Run("Should be successful check ip owner tor proxy from pool", func(t *testing.T) {
		// arrange
		ctx, done := context.WithTimeout(context.Background(), TestProxyServerStartupTimeout)
		t.Cleanup(done)

		pool, err := NewPool(ctx, 3 /* pool size */, WithTestTorrOptions)
		if err != nil {
			t.Fatalf("cannot make proxy: %v", err)
		}

		defer pool.Close()

		// act
		cr, err := torproject.CheckContextDialer(pool.Get())
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

func TestPool_Close(t *testing.T) {
	t.Parallel()
	newPool := func(t *testing.T) *Pool {
		t.Helper()

		ctx, done := context.WithTimeout(context.Background(), TestProxyServerStartupTimeout)
		t.Cleanup(done)

		pool, err := NewPool(ctx, 3 /* pool size */, WithTestTorrOptions)
		if err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() {
			t.Log(pool.Close())
		})

		return pool
	}

	t.Run("Single calls to the close method do not return an error", func(t *testing.T) {
		t.Parallel()
		// arrange
		pool := newPool(t)

		// act
		err := pool.Close()

		// assert
		if err != nil {
			t.Fatal("should not get an error:", err)
		}
	})

	t.Run("Multiple calls to the close method do not return an error", func(t *testing.T) {
		t.Parallel()
		// arrange
		pool := newPool(t)
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

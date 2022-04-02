// Copyright (c) 2022 Vasiliy Vasilyuk. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tornado_test

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/xorcare/tornado"
)

func ExampleNewProxy() {
	const proxyServerStartupTimeout = 15 * time.Second
	ctx, done := context.WithTimeout(context.Background(), proxyServerStartupTimeout)
	defer done()

	prx, err := tornado.NewProxy(ctx)
	if err != nil {
		log.Fatalln("failed to create new instance of proxy:", err)
	}

	httpcli := &http.Client{
		Transport: &http.Transport{
			DialContext: prx.DialContext,
		},
		Timeout: time.Second * 15,
	}

	resp, err := httpcli.Get("https://check.torproject.org/api/ip")
	if err != nil {
		log.Fatalln("failed to execute http request to tor project api:", err)
	}

	text, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Fatalln("failed to dump full response info:", err)
	}

	log.Println(string(text))
}

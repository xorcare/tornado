# tornado

[![Go](https://github.com/xorcare/tornado/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/xorcare/tornado/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/xorcare/tornado/branch/main/graph/badge.svg?branch=main)](https://codecov.io/gh/xorcare/tornado)
[![Go Report Card](https://goreportcard.com/badge/github.com/xorcare/tornado)](https://goreportcard.com/report/github.com/xorcare/tornado)
[![Go Reference](https://pkg.go.dev/badge/github.com/xorcare/tornado.svg)](https://pkg.go.dev/github.com/xorcare/tornado)

Library for easy launch of tor proxy on golang.

# Examples

## Small [example][eprx] of using [single proxy][nprx] server

```go
package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/xorcare/tornado"
)

func main() {
	const proxyServerStartupTimeout = 15 * time.Second
	ctx, done := context.WithTimeout(context.Background(), proxyServerStartupTimeout)
	defer done()

	prx, err := tornado.NewProxy(ctx)
	if err != nil {
		log.Panicln("failed to create new instance of proxy:", err)
	}
	// After usage Proxy must be closed to prevent memory leak and tor
	// demon process leak.
	defer prx.Close()
	
	httpcli := &http.Client{
		Transport: &http.Transport{
			DialContext: prx.DialContext,
		},
		Timeout: time.Second * 15,
	}

	resp, err := httpcli.Get("https://check.torproject.org/api/ip")
	if err != nil {
		log.Panicln("failed to execute http request to tor project api:", err)
	}

	text, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Panicln("failed to dump full response info:", err)
	}

	log.Println(string(text))
}
```

**Output:**

```text
2022/04/05 23:50:18 HTTP/1.1 200 OK
Content-Length: 37
Content-Type: application/json
Date: Tue, 05 Apr 2022 20:50:18 GMT
Referrer-Policy: no-referrer
Server: Apache
Strict-Transport-Security: max-age=15768000; preload
X-Content-Type-Options: nosniff
X-Frame-Options: sameorigin
X-Xss-Protection: 1

{"IsTor":true,"IP":"185.220.100.250"}
```

[eprx]: https://pkg.go.dev/github.com/xorcare/tornado#example-NewProxy

[nprx]: https://pkg.go.dev/github.com/xorcare/tornado#NewProxy
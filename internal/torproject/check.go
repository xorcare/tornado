// Copyright (c) 2022 Vasiliy Vasilyuk. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package torproject

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

// A Dialer is a means to establish a connection.
type Dialer interface {
	// Dial connects to the address on the named network.
	Dial(network, address string) (net.Conn, error)
}

// A ContextDialer dials using a context.
type ContextDialer interface {
	// DialContext connects to the address on the named network using
	// the provided context.
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type CheckResponse struct {
	IsTor bool   `json:"IsTor"`
	IP    string `json:"IP"`
}

func CheckDialer(dialer Dialer) (CheckResponse, error) {
	transport := &http.Transport{
		Dial: dialer.Dial,
	}

	return doCheckRequest(transport)
}

func CheckContextDialer(dialer ContextDialer) (CheckResponse, error) {
	transport := &http.Transport{
		DialContext: dialer.DialContext,
	}

	return doCheckRequest(transport)
}

func doCheckRequest(transport http.RoundTripper) (CheckResponse, error) {
	httpcli := &http.Client{
		Transport: transport,
		Timeout:   time.Second * 15,
	}

	req, err := http.NewRequest("GET", "https://check.torproject.org/api/ip", nil)
	if err != nil {
		const format = "cannot create http request: %v"
		return CheckResponse{}, fmt.Errorf(format, err)
	}

	req.Header.Set("user-agent", "tornado")

	resp, err := httpcli.Do(req)
	if err != nil {
		const format = "request sending error: %v"
		return CheckResponse{}, fmt.Errorf(format, err)
	}

	var cr CheckResponse
	if err = json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		const format = "failed decode response for %q: %v"
		return CheckResponse{}, fmt.Errorf(format, resp.Request.URL, err)
	}

	return cr, nil
}

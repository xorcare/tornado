// Copyright (c) 2022 Vasiliy Vasilyuk. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tornado

import (
	"bytes"
	"fmt"
	"os"

	"github.com/xorcare/tornado/internal/freeport"
)

type torrc struct {
	dataDirectory string

	socksPort    []int
	customOption []string
	afterOption  []string

	torrc    string
	filename string
}

func newTorrcFromState(state options) (trc torrc, err error) {
	trc = torrc{
		afterOption: []string{
			// Recognized severity levels are debug, info, notice, warn, and err.
			// Log level must be "notice" for working startup trap.
			"Log notice stderr",
			"RunAsDaemon 0",
		},
	}

	ports, err := freeport.Much(state.numberOfProxy)
	if err != nil {
		const format = "cannot get free ports for tor proxy: %v"
		return torrc{}, fmt.Errorf(format, err)
	}
	trc.socksPort = append(trc.socksPort, ports...)
	trc.customOption = append(trc.customOption, state.torrcOptions...)

	dir := fmt.Sprintf("tornado.%d.*", os.Getpid())
	trc.dataDirectory, err = os.MkdirTemp("", dir)
	if err != nil {
		const format = "cannot create temp dir for tor proxy: %v"
		return torrc{}, fmt.Errorf(format, err)
	}

	buf := bytes.NewBuffer(make([]byte, 0, 4096))

	fmt.Fprintf(buf, "DataDirectory %s\n\n", trc.dataDirectory)
	for _, port := range trc.socksPort {
		fmt.Fprintf(buf, "SocksPort %d\n\n", port)
	}

	for _, option := range trc.customOption {
		buf.WriteString(option)
		buf.WriteString("\n")
	}

	for _, option := range trc.afterOption {
		buf.WriteString(option)
		buf.WriteString("\n")
	}

	trc.torrc = buf.String()

	tempFile, err := os.CreateTemp(trc.dataDirectory, "torrc.*")
	if err != nil {
		const format = "cannot open temp torrc file: %v"
		return torrc{}, fmt.Errorf(format, err)
	}

	if _, err := tempFile.WriteString(trc.torrc); err != nil {
		const format = "cannot write temp torrc file: %v"
		return torrc{}, fmt.Errorf(format, err)
	}

	if err := tempFile.Close(); err != nil {
		const format = "cannot close temp torrc file: %v"
		return torrc{}, fmt.Errorf(format, err)
	}

	trc.filename = tempFile.Name()

	return trc, nil
}

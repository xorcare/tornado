// Copyright (c) 2022 Vasiliy Vasilyuk. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tornado

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func launchBackgroundTorDemon(ctx context.Context, trc torrc) (cmd *exec.Cmd, err error) {
	if ctx == nil {
		panic("tornado: cannot create tor demon by nil context")
	}

	cmd = exec.CommandContext(ctx, "tor", "-f", trc.filename)
	cmd.Dir = trc.dataDirectory

	stdoutPipe, err := cmd.StderrPipe()
	if err != nil {
		const format = "failed to create stdout pipe for exec command %q: %v"
		return nil, fmt.Errorf(format, cmd.String(), err)
	}

	defer stdoutPipe.Close()

	if err := cmd.Start(); err != nil {
		const format = "failed starting the command %q: %v"
		return nil, fmt.Errorf(format, cmd.String(), err)
	}

	launchLog := bytes.NewBuffer(make([]byte, 0, 4096))
	launched := make(chan error, 1)

	go func() {
		defer close(launched)

		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			text := scanner.Text()
			launchLog.WriteString(text)

			if strings.Contains(text, "Bootstrapped 100%") {
				return
			}
		}

		if err := scanner.Err(); err != nil {
			const format = "failed to scan text: %v"
			launched <- fmt.Errorf(format, err)

			return
		}

		launched <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		err = ctx.Err()
	case err = <-launched:
	}

	if err != nil {
		const format = "failed running the command %q: %v" +
			"\n\n# Torrc file:\n%s" +
			"\n\n# Launch log:\n%s"

		return nil, fmt.Errorf(format, cmd.String(), err, trc.torrc, launchLog.String())
	}

	return cmd, nil
}

// Copyright (c) 2022 Vasiliy Vasilyuk. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tornado

type options struct {
	numberOfProxy int
	torrcOptions  []string
	forwardDialer comboDialer
}

// Option is an abstraction on the options.
type Option interface {
	apply(*options)
}

// optionFunc this type simplifies the creation of simple options that
// do not require a separate type.
type optionFunc func(*options)

func (f optionFunc) apply(optionState *options) {
	f(optionState)
}

// WithTorrcOption allows adding arbitrary parameters to torrc.
func WithTorrcOption(ops ...string) Option {
	fun := func(s *options) {
		s.torrcOptions = append(
			s.torrcOptions,
			ops...,
		)
	}

	return optionFunc(fun)
}

// WithForwardContextDialer allows to specify the optional dial function for
// establishing the transport connection.
func WithForwardContextDialer(dialer ContextDialer) Option {
	fun := func(s *options) {
		if dr, ok := dialer.(comboDialer); ok {
			s.forwardDialer = dr
			return
		}

		s.forwardDialer = comboDialAdapter(dialer.DialContext)
	}

	return optionFunc(fun)
}

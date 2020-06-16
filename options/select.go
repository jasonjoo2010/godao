// Copyright 2020 The GoDao Authors. All rights reserved.
// Use of this source code is governed by BSD
// license that can be found in the LICENSE file.

package options

type SelectOptions struct {
	Fields []string
}

type SelectOption func(opts *SelectOptions)

// WithFields indicates fields when selecting
func WithFields(names ...string) SelectOption {
	return func(opts *SelectOptions) {
		opts.Fields = names
	}
}

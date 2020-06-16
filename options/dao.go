// Copyright 2020 The GoDao Authors. All rights reserved.
// Use of this source code is governed by BSD
// license that can be found in the LICENSE file.

package options

type DaoOptions struct {
	Table string
}

type DaoOption func(opts *DaoOptions)

// WithTable specify table name manually
func WithTable(name string) DaoOption {
	return func(opts *DaoOptions) {
		opts.Table = name
	}
}

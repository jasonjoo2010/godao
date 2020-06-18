// Copyright 2020 The GoDao Authors. All rights reserved.
// Use of this source code is governed by BSD
// license that can be found in the LICENSE file.

package options

import "strings"

type InsertOptions struct {
	Ignore, Replace bool
}

type InsertOption func(opts *InsertOptions)

// WithInsertIgnore indicates performing an INSERT IGNORE
func WithInsertIgnore() InsertOption {
	return func(opts *InsertOptions) {
		opts.Ignore = true
	}
}

// WithReplace indicates a REPLACE which cannot be used with `WithInsertIgnore()`
func WithReplace() InsertOption {
	return func(opts *InsertOptions) {
		opts.Replace = true
	}
}

func InsertBaseSQL(table, columns string, cfg *InsertOptions) string {
	b := strings.Builder{}
	if cfg.Replace {
		b.WriteString("replace into ")
	} else {
		b.WriteString("insert ")
		if cfg.Ignore {
			b.WriteString("ignore ")
		}
		b.WriteString("into ")
	}
	b.WriteString("`")
	b.WriteString(table)
	b.WriteString("` (")
	b.WriteString(columns)
	b.WriteString(") values ")
	return b.String()
}

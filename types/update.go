// Copyright 2020 The GoDao Authors. All rights reserved.
// Use of this source code is governed by BSD
// license that can be found in the LICENSE file.

package types

import "fmt"

// UpdateEntry represents two kinds of updating scenarios (only pick one):
//	Simple value updating. Value should be filled and it will be taken as a raw value.
//	Expression updating. Value should be omitted and Expr should be set. It will not be parsed as a whole value and you can use functions, reference other fields.
//	It's DANGEROUS when using Expr method. Possible injections and data damagement could occur.
type UpdateEntry struct {
	Field string
	Value string
	Expr  string
	Args  []interface{} // bind to expr
}

func NewIncrease(field string, step int64) *UpdateEntry {
	return &UpdateEntry{
		Field: field,
		Expr:  fmt.Sprintf("@%s@ + %d", field, step),
	}
}

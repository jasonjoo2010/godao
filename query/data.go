// Copyright 2020 The GoDao Authors. All rights reserved.
// Use of this source code is governed by BSD
// license that can be found in the LICENSE file.

package query

type Op int

const (
	_ Op = iota
	OpEqual
	OpNotEqual
	OpLess
	OpLessOrEqual
	OpGreater
	OpGreaterOrEqual
	OpLike
	OpStartsWith
	OpEndsWith
	OpNil
	OpNotNil
	OpIn
	OpNotIn
	OpExpr
)

type Data struct {
	Conditions    []Condition
	Children      []Data
	Offset, Limit int
	Order         []Order
	Or            bool
}

type Condition struct {
	Op    Op
	Field string
	Value interface{}
}

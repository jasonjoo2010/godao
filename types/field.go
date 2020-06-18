// Copyright 2020 The GoDao Authors. All rights reserved.
// Use of this source code is governed by BSD
// license that can be found in the LICENSE file.

package types

import "reflect"

type ModelField struct {
	// Field index in struct
	Index int
	// Field name in struct
	Name string
	// Column name in table
	Column string
	// Whether is a primary key
	Primary bool
	// Whether is auto increment
	AutoIncrement bool
	Type          reflect.Type
}

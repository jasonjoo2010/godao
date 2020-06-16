// Copyright 2020 The GoDao Authors. All rights reserved.
// Use of this source code is governed by BSD
// license that can be found in the LICENSE file.

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type DemoTable struct {
	Id          int64 `dao:"primary;auto_increment"`
	Name, Value string
	Extra       string `dao:"omit"`
	Cnt         int    `dao:"column=count"`
	CreateTime  int64
}

// None auto_increament primary key
type UserInfo struct {
	Id          int64 `dao:"primary"`
	Name, Value string
	Cnt         int
	CreateTime  int64
}

// Union primary keys
type Relation struct {
	UserId     int64 `dao:"column=uid;primary"`
	Follow     int64 `dao:"primary"`
	CreateTime int64
}

func TestParse(t *testing.T) {
	assert.Panics(t, func() { Parse(nil) })

	fields1 := Parse(DemoTable{})
	fields2 := Parse(&DemoTable{})
	assert.NotNil(t, fields1)
	assert.NotNil(t, fields2)
	assert.Equal(t, fields1, fields2)
	assert.Equal(t, 5, len(fields1))
	assert.True(t, fields1[0].Primary)
	assert.True(t, fields1[0].AutoIncrement)

	fields3 := Parse(UserInfo{})
	assert.NotNil(t, fields3)
	assert.Equal(t, 5, len(fields3))

	fields4 := Parse(Relation{})
	assert.NotNil(t, fields4)
	assert.Equal(t, 3, len(fields4))
	assert.Equal(t, "uid", fields4[0].Column)
	assert.True(t, fields4[0].Primary)
	assert.False(t, fields4[0].AutoIncrement)
	assert.True(t, fields4[1].Primary)
	assert.False(t, fields4[2].Primary)
}

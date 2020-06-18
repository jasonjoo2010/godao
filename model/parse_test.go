// Copyright 2020 The GoDao Authors. All rights reserved.
// Use of this source code is governed by BSD
// license that can be found in the LICENSE file.

package model

import (
	"reflect"
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

func TestRealType(t *testing.T) {
	num := 1
	num1 := &num
	num2 := &num1
	assert.Equal(t, reflect.TypeOf(num), RealType(num1))
	assert.Equal(t, reflect.TypeOf(num), RealType(num2))

	obj := DemoTable{}
	obj1 := &obj
	obj2 := &obj1
	assert.Equal(t, reflect.TypeOf(obj), RealType(obj1))
	assert.Equal(t, reflect.TypeOf(obj), RealType(obj2))

	p := &DemoTable{}
	p1 := &p
	p2 := &p1
	assert.NotEqual(t, reflect.TypeOf(p), RealType(p1))
	assert.NotEqual(t, reflect.TypeOf(p), RealType(p2))
	assert.Equal(t, reflect.TypeOf(*p), RealType(p1))
	assert.Equal(t, reflect.TypeOf(*p), RealType(p2))
}

func TestRealValue(t *testing.T) {
	num := 1
	num1 := &num
	num2 := &num1
	assert.Equal(t, reflect.TypeOf(num), reflect.TypeOf(RealValue(num1)))
	assert.Equal(t, reflect.TypeOf(num), reflect.TypeOf(RealValue(num2)))
	assert.NotEqual(t, reflect.TypeOf(num1), reflect.TypeOf(RealValue(num2)))
	assert.NotEqual(t, reflect.TypeOf(num2), reflect.TypeOf(RealValue(num2)))

	obj := DemoTable{}
	obj1 := &obj
	obj2 := &obj1
	assert.Equal(t, reflect.TypeOf(obj), reflect.TypeOf(RealValue(obj1)))
	assert.NotEqual(t, reflect.TypeOf(obj1), reflect.TypeOf(RealValue(obj1)))
	assert.Equal(t, reflect.TypeOf(obj), reflect.TypeOf(RealValue(obj2)))

	p := &DemoTable{}
	p1 := &p
	p2 := &p1
	assert.Equal(t, reflect.TypeOf(*p), reflect.TypeOf(RealValue(p1)))
	assert.Equal(t, reflect.TypeOf(*p), reflect.TypeOf(RealValue(p2)))
}

func TestRealPointer(t *testing.T) {
	num := 1
	num1 := &num
	num2 := &num1
	assert.Equal(t, reflect.TypeOf(num1), reflect.TypeOf(RealPointer(num1)))
	assert.Equal(t, reflect.TypeOf(num1), reflect.TypeOf(RealPointer(num2)))
	assert.NotEqual(t, reflect.TypeOf(num), reflect.TypeOf(RealPointer(num1)))
	assert.NotEqual(t, reflect.TypeOf(num), reflect.TypeOf(RealPointer(num1)))
}

func TestFlatten(t *testing.T) {
	obj := &DemoTable{
		Id:         321,
		Name:       "n0",
		Value:      "v1",
		Extra:      "333aaddd",
		Cnt:        -1,
		CreateTime: -22222,
	}
	fields := Parse(obj)
	dst := make([]interface{}, len(fields))
	err := Flatten(dst, RealType(obj), fields, &obj)
	assert.Nil(t, err)

	assert.Equal(t, int64(321), dst[0])
	assert.Equal(t, "n0", dst[1])
	assert.Equal(t, "v1", dst[2])
	assert.Equal(t, -1, dst[3])
	assert.Equal(t, int64(-22222), dst[4])
}

func TestPack(t *testing.T) {
	obj := &DemoTable{
		Id:         321,
		Name:       "n0",
		Value:      "v1",
		Extra:      "333aaddd",
		Cnt:        -1,
		CreateTime: -22222,
	}

	obj0 := *obj
	obj0.Extra = ""

	obj1 := DemoTable{}
	obj2 := &DemoTable{}

	fields := Parse(obj)
	values := make([]interface{}, len(fields))
	err := Flatten(values, RealType(obj), fields, &obj)
	assert.Nil(t, err)

	assert.NotEqual(t, obj, obj1)

	err = Pack(obj1, RealType(obj), fields, values)
	assert.NotNil(t, err)

	err = Pack(&obj1, RealType(obj), fields, values)
	assert.Nil(t, err)
	assert.NotEqual(t, *obj, obj1)
	assert.Equal(t, obj0, obj1)

	err = Pack(obj2, RealType(obj), fields, values)
	assert.Nil(t, err)
	assert.NotEqual(t, obj, obj2)
	assert.Equal(t, &obj0, obj2)

	err = Pack(&obj2, RealType(obj), fields, values)
	assert.Nil(t, err)
	assert.NotEqual(t, obj, obj2)
	assert.Equal(t, &obj0, obj2)
}

func TestParseTable(t *testing.T) {
	assert.Panics(t, func() { Parse(nil) })

	var v0 *DemoTable = nil
	v1 := &v0
	assert.Equal(t, "demo_table", ParseTableName(v1))

	assert.Equal(t, "demo_table", ParseTableName(&DemoTable{}))
	assert.Equal(t, "relation", ParseTableName(&Relation{}))
}

func TestParse(t *testing.T) {
	assert.Panics(t, func() { Parse(nil) })

	var v0 *DemoTable = nil
	v1 := &v0
	fields0 := Parse(v1)
	assert.NotNil(t, fields0)

	fields1 := Parse(DemoTable{})
	fields2 := Parse(&DemoTable{})
	assert.NotNil(t, fields1)
	assert.NotNil(t, fields2)
	assert.Equal(t, fields0, fields1)
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

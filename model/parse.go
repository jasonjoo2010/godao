// Copyright 2020 The GoDao Authors. All rights reserved.
// Use of this source code is governed by BSD
// license that can be found in the LICENSE file.

package model

import (
	"errors"
	"reflect"
	"strings"

	"github.com/jasonjoo2010/enhanced-utils/strutils"
	"github.com/jasonjoo2010/godao/types"
)

const (
	internal_TAG_KEY   = "dao"
	internal_TAG_OMIT  = "omit"
	internal_TAG_PRI   = "primary"
	internal_TAG_AUTO  = "auto_increment"
	internal_TAG_FIELD = "column="
)

// parseField parses single field's tag.
func parseField(f reflect.StructField) *types.ModelField {
	arr := strings.Split(f.Tag.Get(internal_TAG_KEY), ";")
	field := &types.ModelField{}
	field.Index = f.Index[0]
	field.Name = f.Name
	field.Column = strutils.ToUnderscore(field.Name)
	field.Type = f.Type
	for _, tag := range arr {
		switch {
		case tag == internal_TAG_AUTO:
			field.AutoIncrement = true
		case tag == internal_TAG_OMIT:
			return nil
		case tag == internal_TAG_PRI:
			field.Primary = true
		case strings.HasPrefix(tag, internal_TAG_FIELD):
			field.Column = tag[len(internal_TAG_FIELD):]
		}
	}
	return field
}

// ParseTableName returns the automatic table name.
func ParseTableName(obj interface{}) string {
	if obj == nil {
		panic("Model for parsing can not be nil")
	}
	t := RealType(obj)
	return strutils.ToUnderscore(t.Name())
}

// Parse parses the struct fileds into a slice(according to the tags)
func Parse(obj interface{}) []*types.ModelField {
	if obj == nil {
		panic("Model for parsing can not be nil")
	}
	t := RealType(obj)
	var fields []*types.ModelField
	for i := 0; i < t.NumField(); i++ {
		field := parseField(t.Field(i))
		if field != nil {
			fields = append(fields, field)
		}
	}
	return fields
}

// RealType returns the root non-pointer type.
func RealType(obj interface{}) reflect.Type {
	t := reflect.TypeOf(obj)
	for t.Kind() == reflect.Ptr {
		// pointer
		t = t.Elem()
	}
	return t
}

// RealValue returns the root value it pointed to.
func RealValue(obj interface{}) interface{} {
	if obj == nil {
		return obj
	}
	val := reflect.ValueOf(obj)
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	return val.Interface()
}

// RealPointer returns single layer pointer of root type.
//	Return nil if obj is not a pointer.
func RealPointer(obj interface{}) interface{} {
	if obj == nil {
		return obj
	}
	val := reflect.ValueOf(obj)
	var last reflect.Value = reflect.Value{}
	for val.Kind() == reflect.Ptr {
		last = val
		val = val.Elem()
	}
	if last.Kind() == reflect.Invalid {
		return nil
	}
	return last.Interface()
}

// Flatten flattens model object into given array.
//	Array should have the length of fields.
func Flatten(dst []interface{}, typ reflect.Type, fields []*types.ModelField, obj interface{}) error {
	if len(dst) != len(fields) {
		return errors.New("dst doesn't have the same length as fields has")
	}
	obj = RealValue(obj)
	if reflect.TypeOf(obj) != typ {
		return errors.New("The type of given object is unexpected")
	}
	val := reflect.ValueOf(obj)
	for i, f := range fields {
		dst[i] = val.Field(f.Index).Interface()
	}
	return nil
}

// Pack packs the flatten values to struct object
//	Pay attention that `dst` should be passed by reference to get the correct state outside.
//	Passing reference to reduce memory footprints in some scenarios.
func Pack(dst interface{}, typ reflect.Type, fields []*types.ModelField, values []interface{}) error {
	if len(values) != len(fields) {
		return errors.New("Array of values doesn't have the same length as fields has")
	}
	dst = RealPointer(dst)
	if dst == nil {
		return errors.New("Please pass a reference instead of struct itself")
	}
	dstTyp := reflect.TypeOf(dst)
	dstVal := reflect.ValueOf(dst)
	if dstTyp.Kind() != reflect.Ptr || dstVal.Elem().Type() != typ {
		return errors.New("The type of dst object is unexpected")
	}
	for i, f := range fields {
		dstVal.Elem().Field(f.Index).Set(reflect.ValueOf(values[i]))
	}
	return nil
}

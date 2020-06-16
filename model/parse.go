// Copyright 2020 The GoDao Authors. All rights reserved.
// Use of this source code is governed by BSD
// license that can be found in the LICENSE file.

package model

import (
	"reflect"
	"strings"

	"github.com/jasonjoo2010/enhanced-utils/strutils"
)

const (
	internal_TAG_KEY   = "dao"
	internal_TAG_OMIT  = "omit"
	internal_TAG_PRI   = "primary"
	internal_TAG_AUTO  = "auto_increment"
	internal_TAG_FIELD = "column="
)

type Field struct {
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
}

func parseField(f reflect.StructField) *Field {
	arr := strings.Split(f.Tag.Get(internal_TAG_KEY), ";")
	field := &Field{}
	field.Index = f.Index[0]
	field.Name = f.Name
	field.Column = strutils.ToUnderscore(field.Name)
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

func Parse(obj interface{}) []*Field {
	if obj == nil {
		panic("Model for parsing can not be nil")
	}
	t := reflect.TypeOf(obj)
	if t.String()[0] == '*' {
		// pointer
		t = t.Elem()
	}
	var fields []*Field
	for i := 0; i < t.NumField(); i++ {
		field := parseField(t.Field(i))
		if field != nil {
			fields = append(fields, field)
		}
	}
	return fields
}

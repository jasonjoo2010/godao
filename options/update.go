// Copyright 2020 The GoDao Authors. All rights reserved.
// Use of this source code is governed by BSD
// license that can be found in the LICENSE file.

package options

import (
	"strings"

	"github.com/jasonjoo2010/godao/query"
	"github.com/jasonjoo2010/godao/types"
	"github.com/sirupsen/logrus"
)

func UpdateSQL(table string, fields []*types.ModelField) string {
	b := strings.Builder{}
	b1 := strings.Builder{} // primary condition
	b2 := strings.Builder{} // fields
	for _, f := range fields {
		if f.Primary {
			if b1.Len() > 0 {
				b1.WriteString(" and ")
			}
			b1.WriteString("`")
			b1.WriteString(f.Column)
			b1.WriteString("` = ?")
		} else {
			if b2.Len() > 0 {
				b2.WriteString(", ")
			}
			b2.WriteString(f.Column)
			b2.WriteString(" = ?")
		}
	}
	b.WriteString("update ")
	b.WriteString("`")
	b.WriteString(table)
	b.WriteString("` set ")
	b.WriteString(b2.String())
	b.WriteString(" where ")
	b.WriteString(b1.String())
	return b.String()
}

func UpdateEntrySQL(
	entries []*types.UpdateEntry,
	byName map[string]*types.ModelField,
	byColumn map[string]*types.ModelField,
) (sqlString string, args []interface{}) {
	b := strings.Builder{}
	for _, entry := range entries {
		f := getField(entry.Field, byName, byColumn)
		if f == nil {
			logrus.Panic("Field not found: ", entry.Field)
		}
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		b.WriteString("`")
		b.WriteString(f.Column)
		b.WriteString("` = ")
		if entry.Value != nil {
			b.WriteString("?")
			args = append(args, entry.Value)
		} else if entry.Expr != "" {
			b.WriteString(query.ParseColumnPlaceholder(entry.Expr, byName, byColumn))
			if len(entry.Args) > 0 {
				args = append(args, entry.Args...)
			}
		} else {
			logrus.Panic("No value or expr specified: ", entry.Field)
		}
	}
	sqlString = b.String()
	return
}

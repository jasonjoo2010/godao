// Copyright 2020 The GoDao Authors. All rights reserved.
// Use of this source code is governed by BSD
// license that can be found in the LICENSE file.

package options

import (
	"regexp"
	"strings"

	"github.com/jasonjoo2010/godao/query"
	"github.com/jasonjoo2010/godao/types"
	"github.com/sirupsen/logrus"
)

var (
	expressionPattern = regexp.MustCompile("(?i)^(.+)\\s+as\\s+[`]?([a-zA-Z0-9_]+)[`]?\\s*$")
)

type SelectField struct {
	Field *types.ModelField
	Expr  string
}

type SelectOptions struct {
	Fields []string
}

type SelectOption func(opts *SelectOptions)

// WithFieldString spedify fields to return
//	"id"
//	"Id"
//	"Id, Name"
//
//	You can use expression:
//	"avg(@Num@) as Num"
//	Which Num is existed field name and @Num@ will be replace with `num`
//	But you can't use extra dot in expression like
//	"concat(@Info@, @Name@) as Num"
//	If so you should use WithFields() instead
func WithFieldString(names string) SelectOption {
	if strings.IndexRune(names, ',') > 0 {
		arr := strings.Split(names, ",")
		for i := 0; i < len(arr); i++ {
			arr[i] = strings.Trim(arr[i], " ")
		}
		return WithFields(arr...)
	} else {
		return WithFields(names)
	}
}

// WithFields indicates fields when selecting
//	"id"
//	"Id"
//	"Id, Name"
//
//	You can use expression:
//	"avg(@Num@) as Num"
//	Which Num is existed field name and @Num@ will be replace with `num`
func WithFields(names ...string) SelectOption {
	return func(opts *SelectOptions) {
		opts.Fields = names
	}
}

func getField(
	str string,
	byName map[string]*types.ModelField,
	byColumn map[string]*types.ModelField,
) *types.ModelField {
	if f, ok := byName[str]; ok {
		return f
	}
	if f, ok := byColumn[str]; ok {
		return f
	}
	return nil
}

func ParseSelectField(
	str string,
	byName map[string]*types.ModelField,
	byColumn map[string]*types.ModelField,
) *SelectField {
	f := getField(str, byName, byColumn)
	if f != nil {
		return &SelectField{
			Field: f,
		}
	}
	// expression support
	arr := expressionPattern.FindStringSubmatch(str)
	if len(arr) != 3 {
		return nil
	}
	f = getField(arr[len(arr)-1], byName, byColumn)
	if f == nil {
		return nil
	}
	return &SelectField{
		Field: f,
		Expr:  query.ParseColumnPlaceholder(arr[1], byName, byColumn),
	}
}

func GenerateSelectFields(
	names []string,
	byName map[string]*types.ModelField,
	byColumn map[string]*types.ModelField,
) (sql string, fields []*types.ModelField) {
	sqlBuilder := strings.Builder{}
	for i, str := range names {
		f := ParseSelectField(str, byName, byColumn)
		if f == nil {
			logrus.Panic("Incorrect field: ", str)
		}
		fields = append(fields, f.Field)
		if i > 0 {
			sqlBuilder.WriteString(", ")
		}
		if f.Expr == "" {
			sqlBuilder.WriteString("`")
			sqlBuilder.WriteString(f.Field.Column)
			sqlBuilder.WriteString("` as `")
			sqlBuilder.WriteString(f.Field.Name)
			sqlBuilder.WriteString("`")
		} else {
			sqlBuilder.WriteString(f.Expr)
			sqlBuilder.WriteString("` as `")
			sqlBuilder.WriteString(f.Field.Name)
			sqlBuilder.WriteString("`")
		}
	}
	sql = sqlBuilder.String()
	return
}

// Copyright 2020 The GoDao Authors. All rights reserved.
// Use of this source code is governed by BSD
// license that can be found in the LICENSE file.

package query

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jasonjoo2010/godao/types"
)

var (
	fieldHolderReg = regexp.MustCompile("@[a-zA-Z_0-9]+@")
)

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

func (o Op) Op() string {
	switch o {
	case OpEqual:
		return "="
	case OpNotEqual:
		return "<>"
	case OpLess:
		return "<"
	case OpLessOrEqual:
		return "<="
	case OpGreater:
		return ">"
	case OpGreaterOrEqual:
		return ">="
	case OpLike:
		return "like"
	case OpStartsWith:
		return "like"
	case OpEndsWith:
		return "like"
	case OpNil:
		return "is null"
	case OpNotNil:
		return "not null"
	case OpIn:
		return "in"
	case OpNotIn:
		return "not in"
	case OpExpr:
	}
	return ""
}

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
	Args  []interface{} // Bind to OpExpr, support deeper placeholders
}

// GetColumn returns the *column name* if there was specific Field or Column
func GetColumn(
	name string,
	fieldsByName map[string]*types.ModelField,
	fieldsByColumn map[string]*types.ModelField,
	should_panic bool,
) string {
	if f, ok := fieldsByName[name]; ok {
		return f.Column
	}
	if f, ok := fieldsByColumn[name]; ok {
		return f.Column
	}
	if should_panic {
		panic("Unknow column: " + name)
	}
	return ""
}

// ParseColumnPlaceholder parses @field@ into `field`
func ParseColumnPlaceholder(str string,
	byName map[string]*types.ModelField,
	byColumn map[string]*types.ModelField,
) string {
	arr := fieldHolderReg.FindAllString(str, 100)
	for _, m := range arr {
		c := GetColumn(strings.Trim(m, "@"), byName, byColumn, false)
		if c == "" {
			continue
		}
		str = strings.ReplaceAll(str, m, "`"+c+"`")
	}
	return str
}

func generateCondition(c *Condition,
	byName map[string]*types.ModelField,
	byColumn map[string]*types.ModelField,
) (string, []interface{}) {
	switch c.Op {
	case OpExpr:
		expr, ok := c.Value.(string)
		if !ok || len(expr) < 1 {
			panic("expr should be a non-empty string")
		}
		return ParseColumnPlaceholder(c.Field, byName, byColumn) + " " + ParseColumnPlaceholder(expr, byName, byColumn), c.Args
	default:
		prefix := "`" + GetColumn(c.Field, byName, byColumn, true) + "` " + c.Op.Op()
		switch c.Op {
		case OpNil, OpNotNil:
			return prefix, nil
		case OpIn, OpNotIn:
			arr, ok := c.Value.([]interface{})
			if !ok {
				panic("`in` / `not int` should take a `[]interface{}` as argument")
			}
			if len(arr) < 1 {
				panic("`in` / `not in` should take non-empty slice")
			}
			return prefix + " (" + strings.TrimLeft(strings.Repeat(", ?", len(arr)), ", ") + ")", arr
		case OpLike, OpStartsWith, OpEndsWith:
			val, ok := c.Value.(string)
			if !ok {
				panic("`like` / `startsWith` / `endsWith` should take a string as argument")
			}
			switch c.Op {
			case OpStartsWith:
				return prefix + " ?", []interface{}{val + "%"}
			case OpEndsWith:
				return prefix + " ?", []interface{}{"%" + val}
			default:
				return prefix + " ?", []interface{}{"%" + val + "%"}
			}
		default:
			return prefix + " ?", []interface{}{c.Value}
		}
	}
}

func whereSQL(
	fieldsByName map[string]*types.ModelField,
	fieldsByColumn map[string]*types.ModelField,
	data *Data,
) (where string, args []interface{}) {
	b := strings.Builder{}
	for _, w := range data.Conditions {
		if b.Len() > 0 {
			if data.Or {
				// or
				b.WriteString(" or ")
			} else {
				// and
				b.WriteString(" and ")
			}
		}
		str, arr := generateCondition(&w, fieldsByName, fieldsByColumn)
		b.WriteString(str)
		if len(arr) > 0 {
			args = append(args, arr...)
		}
	}

	// children
	for _, child := range data.Children {
		str, params := whereSQL(fieldsByName, fieldsByColumn, &child)
		if str != "" {
			if b.Len() > 0 {
				if data.Or {
					// or
					b.WriteString(" or ")
				} else {
					// and
					b.WriteString(" and ")
				}
			}
			b.WriteString("(")
			b.WriteString(str)
			b.WriteString(")")
		}
		if len(params) > 0 {
			args = append(args, params...)
		}
	}

	where = b.String()
	return
}

func ConditionSQL(
	fieldsByName map[string]*types.ModelField,
	fieldsByColumn map[string]*types.ModelField,
	data *Data,
) (string, []interface{}) {
	var args []interface{}
	sql := strings.Builder{}

	// where
	{
		str, params := whereSQL(fieldsByName, fieldsByColumn, data)
		if str != "" {
			sql.WriteString("where ")
			sql.WriteString(str)
		}
		if len(params) > 0 {
			args = append(args, params...)
		}
	}

	// order by
	if len(data.Order) > 0 {
		if sql.Len() > 0 {
			sql.WriteString(" ")
		}
		sql.WriteString("order by ")
		for i, o := range data.Order {
			if i > 0 {
				sql.WriteString(", ")
			}
			sql.WriteString("`")
			sql.WriteString(GetColumn(o.Field, fieldsByName, fieldsByColumn, true))
			if o.Desc {
				sql.WriteString("` desc")
			} else {
				sql.WriteString("` asc")
			}
		}
	}

	// limit
	if data.Limit > 0 {
		if sql.Len() > 0 {
			sql.WriteString(" ")
		}
		sql.WriteString(fmt.Sprint("limit ", data.Offset, ", ", data.Limit))
	}

	return sql.String(), args
}

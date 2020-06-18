// Copyright 2020 The GoDao Authors. All rights reserved.
// Use of this source code is governed by BSD
// license that can be found in the LICENSE file.

package godao

import (
	"github.com/jasonjoo2010/godao/query"
	"github.com/sirupsen/logrus"
)

type Query struct {
	sub_queries   []*Query
	conditions    []query.Condition
	logicalOr     bool
	offset, limit int
	order_by      []query.Order
}

func (q *Query) addCondition(field_name string, op query.Op, val interface{}) *Query {
	q.conditions = append(q.conditions, query.Condition{
		Field: field_name,
		Value: val,
		Op:    op,
	})
	return q
}

// NotIn represents a NOT IN gramma
func (q *Query) NotIn(field_name string, val []interface{}) *Query {
	return q.addCondition(field_name, query.OpNotIn, val)
}

// In represents a IN gramma
func (q *Query) In(field_name string, val []interface{}) *Query {
	return q.addCondition(field_name, query.OpIn, val)
}

// Greater represents a ">" gramma
func (q *Query) Greater(field_name string, val interface{}) *Query {
	return q.addCondition(field_name, query.OpGreater, val)
}

// GreaterOrEqual represents a ">=" gramma
func (q *Query) GreaterOrEqual(field_name string, val interface{}) *Query {
	return q.addCondition(field_name, query.OpGreaterOrEqual, val)
}

// Less represents a "<" gramma
func (q *Query) Less(field_name string, val interface{}) *Query {
	return q.addCondition(field_name, query.OpLess, val)
}

// LessOrEqual represents a "<=" gramma
func (q *Query) LessOrEqual(field_name string, val interface{}) *Query {
	return q.addCondition(field_name, query.OpLessOrEqual, val)
}

// Equal represents a "=" gramma
func (q *Query) Equal(field_name string, val interface{}) *Query {
	return q.addCondition(field_name, query.OpEqual, val)
}

// NotEqual represents a "<>" gramma
func (q *Query) NotEqual(field_name string, val interface{}) *Query {
	return q.addCondition(field_name, query.OpNotEqual, val)
}

// Nil represents a IS NULL gramma
func (q *Query) Nil(field_name string) *Query {
	return q.addCondition(field_name, query.OpNil, nil)
}

// NotNil represents a NOT NULL gramma
func (q *Query) NotNil(field_name string) *Query {
	return q.addCondition(field_name, query.OpNotNil, nil)
}

// Like represents a LIKE '%str%' gramma
//	Pay attention that all indexes will be disabled in general LIKE
func (q *Query) Like(field_name string, val interface{}) *Query {
	return q.addCondition(field_name, query.OpLike, val)
}

// StartsWith represents a RIGHT LIKE 'str%' gramma
func (q *Query) StartsWith(field_name string, val interface{}) *Query {
	return q.addCondition(field_name, query.OpStartsWith, val)
}

// EndsWith represents a LEFT LIKE '%str' gramma
//	Pay attention that all indexes will be disabled in left LIKE
func (q *Query) EndsWith(field_name string, val interface{}) *Query {
	return q.addCondition(field_name, query.OpEndsWith, val)
}

// Expr represents a more like RAW condition.
//	Pay attention possible violations and injections.
//	Fields in same model can be referenced in expr by `@fieldName@`
//	Examples:
//		@Id@ + 1
//		@Id@ + ceil(@Count@)
//	Surely you can use RAW name of fields but it's not recommended
//		Count + 1
//		`Count` + 1
func (q *Query) Expr(field_name, expr string) *Query {
	return q.addCondition(field_name, query.OpExpr, expr)
}

// NoLimit infers no limit at all
//	Please use it carefully
func (q *Query) NoLimit() *Query {
	q.offset = 0
	q.limit = 0
	return q
}

// And joins all conditions in this query by AND
func (q *Query) And() *Query {
	q.logicalOr = false
	return q
}

// Or joins all conditions in this query by OR
func (q *Query) Or() *Query {
	q.logicalOr = true
	return q
}

// Wrap adds specific query as its child query
func (q *Query) Wrap(sub_query *Query) *Query {
	q.sub_queries = append(q.sub_queries, sub_query)
	return q
}

// WrappedTo adds itself to specific query as a child query
func (q *Query) WrappedTo(parent_query *Query) *Query {
	parent_query.sub_queries = append(parent_query.sub_queries, q)
	return q
}

// OrderBy takes the column into ordering.
//	Pay attention the the orders when adding.
func (q *Query) OrderBy(field_name string, desc bool) *Query {
	q.order_by = append(q.order_by, query.Order{
		Field: field_name,
		Desc:  desc,
	})
	return q
}

// Page uses the pagination style to locate offset and limit
func (q *Query) Page(page, page_size int) *Query {
	if page < 1 || page_size < 1 {
		logrus.Warn("illegal page or page_size: ", page, ", ", page_size)
		return q
	}
	q.offset = (page - 1) * page_size
	q.limit = page_size
	return q
}

// Limit limits the maximum records returned
func (q *Query) Limit(limit int) *Query {
	if limit < 1 {
		logrus.Warn("illegal limit: ", limit)
		return q
	}
	q.limit = limit
	return q
}

// Offset sets both the offset and limit
func (q *Query) Offset(offset, limit int) *Query {
	if limit < 1 || offset < 0 {
		logrus.Warn("illegal offset or limit: ", offset, ", ", limit)
		return q
	}
	q.offset = offset
	q.limit = limit
	return q
}

// Data generates the final data object representing the conditions
func (q *Query) Data() query.Data {
	data := query.Data{}

	// conditions
	data.Conditions = q.conditions

	// logical
	data.Or = q.logicalOr

	// order
	data.Order = q.order_by

	// offset, limit
	data.Offset = q.offset
	data.Limit = q.limit

	// sub queries
	data.Children = []query.Data{}
	for _, query := range q.sub_queries {
		data.Children = append(data.Children, query.Data())
	}

	return data
}

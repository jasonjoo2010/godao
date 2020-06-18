// Copyright 2020 The GoDao Authors. All rights reserved.
// Use of this source code is governed by BSD
// license that can be found in the LICENSE file.

package godao

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryLimit(t *testing.T) {
	q := Query{}

	assert.Equal(t, 0, q.offset)
	assert.Equal(t, 0, q.limit)

	data := q.Data()
	assert.Equal(t, 0, data.Offset)
	assert.Equal(t, 20, data.Limit)

	q.Limit(20)
	assert.Equal(t, 0, q.offset)
	assert.Equal(t, 20, q.limit)

	q.Limit(-20)
	assert.Equal(t, 0, q.offset)
	assert.Equal(t, 20, q.limit)

	q.Page(0, 20)
	assert.Equal(t, 0, q.offset)
	assert.Equal(t, 20, q.limit)

	q.Page(2, 40)
	assert.Equal(t, 40, q.offset)
	assert.Equal(t, 40, q.limit)

	data = q.Data()
	assert.Equal(t, 40, data.Limit)
	assert.Equal(t, 40, data.Offset)
}

func TestQueryOrder(t *testing.T) {
	q := Query{}

	assert.Equal(t, 0, len(q.order_by))

	q.OrderBy("a", false)
	assert.Equal(t, 1, len(q.order_by))

	q.OrderBy("a", false)
	assert.Equal(t, 2, len(q.order_by))

	q.OrderBy("b", true)
	assert.Equal(t, 3, len(q.order_by))

	data := q.Data()
	assert.Equal(t, 3, len(data.Order))
}

func TestQueryAnd(t *testing.T) {
	q := Query{}

	assert.False(t, q.logicalOr)

	q.And()
	assert.False(t, q.logicalOr)

	q.Or()
	assert.True(t, q.logicalOr)

	q.And()
	assert.False(t, q.logicalOr)
}

func TestQuerySubQuery(t *testing.T) {
	q := Query{}
	q1 := Query{}
	q2 := Query{}

	q1.Equal("a", 1).
		Equal("b", 1)

	q2.Equal("a", 1).
		Equal("b", 1)

	data := q.Or().
		Wrap(&q1).
		Wrap(&q2).
		Data()

	assert.Equal(t, 2, len(data.Children))
	assert.True(t, data.Or)
	assert.False(t, data.Children[0].Or)
	assert.False(t, data.Children[1].Or)
}

func TestQueryCondition(t *testing.T) {
	q := Query{}

	q.Equal("a", 1)
	q.Less("b", 1)
	q.Expr("length(@Name@)", "< length(@Value@) + ?", 10)
	assert.Equal(t, 3, len(q.conditions))
	fmt.Println(q.conditions)

	data := q.Data()
	assert.Equal(t, 2, len(data.Conditions))
}

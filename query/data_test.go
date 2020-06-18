// Copyright 2020 The GoDao Authors. All rights reserved.
// Use of this source code is governed by BSD
// license that can be found in the LICENSE file.

package query

import (
	"fmt"
	"testing"

	"github.com/jasonjoo2010/godao/model"
	"github.com/jasonjoo2010/godao/types"
	"github.com/stretchr/testify/assert"
)

type userInfo struct {
	Id            int64 `dao:"auto_increment;primary"`
	Name          string
	Password      string
	LastLogin     int64
	AvgOnlineTime float32
	Birth         int64 `dao:"column=b"`
}

func TestConditionSQL(t *testing.T) {
	fields := model.Parse(userInfo{})
	fieldsByName := make(map[string]*types.ModelField, len(fields))
	fieldsByColumn := make(map[string]*types.ModelField, len(fields))
	for _, f := range fields {
		fieldsByName[f.Name] = f
		fieldsByColumn[f.Column] = f
	}
	data := &Data{
		Conditions: []Condition{
			Condition{
				Field: "Id",
				Op:    OpGreater,
				Value: 3,
			},
			Condition{
				Field: "Password",
				Op:    OpNotNil,
				Value: nil,
			},
			Condition{
				Field: "Name",
				Op:    OpStartsWith,
				Value: "admin",
			},
			Condition{
				Field: "Id",
				Op:    OpNotIn,
				Value: []interface{}{1, 2, 34},
			},
			Condition{
				Field: "md5(@Name@)",
				Op:    OpExpr,
				Value: "like concat(@Id@, '%')",
			},
		},
		Order: []Order{
			Order{
				Field: "Name",
				Desc:  true,
			},
			Order{
				Field: "AvgOnlineTime",
			},
			Order{
				Field: "id",
			},
		},
		Offset: 1,
		Limit:  10,
	}
	sql, args := ConditionSQL(fieldsByName, fieldsByColumn, data)
	assert.Contains(t, sql, "where `id` > ?")
	assert.Contains(t, sql, "`password` not null")
	assert.Contains(t, sql, "`name` like ?")
	assert.Contains(t, sql, "`id` not in (")
	assert.Contains(t, sql, "md5(`name`) like concat(`id`, '%')")
	assert.Contains(t, sql, "order by")
	assert.Contains(t, sql, "name desc")
	assert.Contains(t, sql, "id asc")
	assert.Contains(t, sql, "avg_online_time asc")
	assert.Contains(t, sql, "limit 1, 10")
	fmt.Println(sql)
	fmt.Println(args)

	// logical

	// children

}

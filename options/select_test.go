// Copyright 2020 The GoDao Authors. All rights reserved.
// Use of this source code is governed by BSD
// license that can be found in the LICENSE file.

package options

import (
	"testing"

	"github.com/jasonjoo2010/godao/model"
	"github.com/jasonjoo2010/godao/types"
	"github.com/stretchr/testify/assert"
)

type TestSelectTable struct {
	Id      int64
	Name    string
	Content string
	Created int64
}

func TestParseSelectField(t *testing.T) {
	fields := model.Parse(TestSelectTable{})
	byName := make(map[string]*types.ModelField, len(fields))
	byColumn := make(map[string]*types.ModelField, len(fields))
	for _, f := range fields {
		byName[f.Name] = f
		byColumn[f.Column] = f
	}

	field := ParseSelectField("a", byName, byColumn)
	assert.Nil(t, field)

	field = ParseSelectField("Id", byName, byColumn)
	assert.NotNil(t, field)
	assert.Equal(t, "id", field.Field.Column)
	assert.Equal(t, "Id", field.Field.Name)
	assert.Empty(t, field.Expr)

	field = ParseSelectField("id", byName, byColumn)
	assert.NotNil(t, field)
	assert.Equal(t, "id", field.Field.Column)
	assert.Equal(t, "Id", field.Field.Name)
	assert.Empty(t, field.Expr)

	// expression

	field = ParseSelectField("min(id) as Id", byName, byColumn)
	assert.NotNil(t, field)
	assert.Equal(t, "id", field.Field.Column)
	assert.Equal(t, "Id", field.Field.Name)
	assert.Equal(t, "min(id)", field.Expr)

	field = ParseSelectField("min(@id@) as id", byName, byColumn)
	assert.NotNil(t, field)
	assert.Equal(t, "id", field.Field.Column)
	assert.Equal(t, "Id", field.Field.Name)
	assert.Equal(t, "min(`id`)", field.Expr)

	field = ParseSelectField("min(@Id@) as id", byName, byColumn)
	assert.NotNil(t, field)
	assert.Equal(t, "id", field.Field.Column)
	assert.Equal(t, "Id", field.Field.Name)
	assert.Equal(t, "min(`id`)", field.Expr)

	field = ParseSelectField("concat('id-', @Id@) as id", byName, byColumn)
	assert.NotNil(t, field)
	assert.Equal(t, "id", field.Field.Column)
	assert.Equal(t, "Id", field.Field.Name)
	assert.Equal(t, "concat('id-', `id`)", field.Expr)

	field = ParseSelectField("concat('id-', @Id@)", byName, byColumn)
	assert.Nil(t, field)
}

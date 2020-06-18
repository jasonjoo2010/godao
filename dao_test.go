// Copyright 2020 The GoDao Authors. All rights reserved.
// Use of this source code is governed by BSD
// license that can be found in the LICENSE file.

package godao

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jasonjoo2010/godao/types"
	"github.com/stretchr/testify/assert"
)

// Demo table structure:
// CREATE TABLE `demo` (
//   `id` bigint(20) NOT NULL AUTO_INCREMENT,
//   `name` varchar(100) NOT NULL DEFAULT '',
//   `value` varchar(255) NOT NULL DEFAULT '',
//   `cnt` int(11) NOT NULL DEFAULT '0',
//   `created` bigint(20) NOT NULL DEFAULT '0',
//   PRIMARY KEY (`id`),
//   KEY `name` (`name`)
// ) ENGINE=InnoDB

type Demo struct {
	Id          int64 `dao:"primary;auto_increment"`
	Name, Value string
	Cnt         int
	Created     int64
}

func testDB() *sql.DB {
	db, _ := sql.Open("mysql", "root@tcp(127.0.0.1:3306)/test?charset=utf8mb4,utf8")
	return db
}

func TestDaoBasic(t *testing.T) {
	db := testDB()
	defer db.Close()
	dao := NewDao(Demo{}, db)
	// insert, update, get, delete

	demo := Demo{
		Name:    "n1",
		Value:   "v1",
		Created: time.Now().Unix(),
	}

	// create a record
	affected, id, err := dao.Insert(context.Background(), demo)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), affected)
	assert.True(t, id > 0)

	// create another
	var id1 int64
	affected, id1, err = dao.Insert(context.Background(), &demo)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), affected)
	assert.True(t, id1 > 0)
	dao.Delete(context.Background(), id1)

	// update it
	demo.Id = id
	demo.Value = "v3"
	demo.Cnt = 111
	affected, err = dao.Update(context.Background(), demo)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), affected)

	// get back
	obj, err := dao.SelectOne(context.Background(), demo.Id)
	assert.Nil(t, err)
	assert.NotNil(t, obj)
	newDemo, ok := obj.(*Demo)
	assert.True(t, ok)
	assert.Equal(t, newDemo.Id, demo.Id)

	// update none
	demo.Id = -1
	affected, err = dao.Update(context.Background(), demo)
	assert.Nil(t, err)
	assert.Equal(t, int64(0), affected)

	// count
	cnt, err := dao.CountBy(context.Background(), "Id", newDemo.Id)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), cnt)

	// sum
	obj, err = dao.Sum(context.Background(), "Id", (&Query{}).Equal("Id", newDemo.Id).Data())
	assert.Nil(t, err)
	assert.NotNil(t, obj)
	summed := obj.(int64)
	assert.Equal(t, newDemo.Id, summed)

	// avg
	avgVal, err := dao.Avg(context.Background(), "Id", (&Query{}).Equal("Id", newDemo.Id).Data())
	assert.Nil(t, err)
	assert.Equal(t, float64(newDemo.Id), avgVal)

	// update by
	affected, err = dao.UpdateBy(context.Background(), (&Query{}).
		Equal("Id", newDemo.Id).
		Data(),
		&types.UpdateEntry{
			Field: "value",
			Value: "test update",
		},
		types.NewIncrease("cnt", 1),
	)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), affected)

	// refetch to verify
	obj, err = dao.SelectOne(context.Background(), newDemo.Id)
	assert.Nil(t, err)
	assert.NotNil(t, obj)
	obj1 := obj.(*Demo)
	assert.Equal(t, newDemo.Id, obj1.Id)
	assert.Equal(t, "test update", obj1.Value)
	assert.Equal(t, 112, obj1.Cnt)

	// delete it
	affected, err = dao.Delete(context.Background(), newDemo.Id)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), affected)

	// refetch to verify
	obj, err = dao.SelectOne(context.Background(), newDemo.Id)
	assert.Nil(t, err)
	assert.Nil(t, obj)
}

func TestTxnRollback(t *testing.T) {
	db := testDB()
	defer db.Close()
	dao := NewDao(Demo{}, db)

	ctx, err := dao.Txn(nil)
	assert.Nil(t, err)

	demo := Demo{
		Name:    "n1",
		Value:   "v1",
		Created: time.Now().Unix(),
	}

	// create 2 records
	var ids [2]interface{}
	affected, id, err := dao.Insert(ctx, demo)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), affected)
	assert.True(t, id > 0)
	ids[0] = id

	affected, id, err = dao.Insert(ctx, demo)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), affected)
	assert.True(t, id > 0)
	ids[1] = id

	// test records in transaction
	cnt, _ := dao.Count(ctx, (&Query{}).
		In("Id", ids[:]).
		Data(),
	)
	assert.Equal(t, int64(2), cnt)

	// test outside transaction
	cnt, _ = dao.Count(context.Background(), (&Query{}).
		In("Id", ids[:]).
		Data(),
	)
	assert.Equal(t, int64(0), cnt)

	ctx.Txn().Rollback()

	// recheck
	cnt, _ = dao.Count(context.Background(), (&Query{}).
		In("Id", ids[:]).
		Data(),
	)
	assert.Equal(t, int64(0), cnt)
}

func TestTxnCommit(t *testing.T) {
	db := testDB()
	defer db.Close()
	dao := NewDao(Demo{}, db)

	ctx, err := dao.Txn(nil)
	assert.Nil(t, err)

	demo := Demo{
		Name:    "n1",
		Value:   "v1",
		Created: time.Now().Unix(),
	}

	// create 2 records
	var ids [2]interface{}
	affected, id, err := dao.Insert(ctx, demo)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), affected)
	assert.True(t, id > 0)
	ids[0] = id

	affected, id, err = dao.Insert(ctx, demo)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), affected)
	assert.True(t, id > 0)
	ids[1] = id

	// test records in transaction
	cnt, _ := dao.Count(ctx, (&Query{}).
		In("Id", ids[:]).
		Data(),
	)
	assert.Equal(t, int64(2), cnt)

	// test outside transaction
	cnt, _ = dao.Count(context.Background(), (&Query{}).
		In("Id", ids[:]).
		Data(),
	)
	assert.Equal(t, int64(0), cnt)

	ctx.Txn().Commit()

	// recheck
	cnt, _ = dao.Count(context.Background(), (&Query{}).
		In("Id", ids[:]).
		Data(),
	)
	assert.Equal(t, int64(2), cnt)

	dao.Delete(context.Background(), ids[:]...)
}

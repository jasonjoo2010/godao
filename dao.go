// Copyright 2020 The GoDao Authors. All rights reserved.
// Use of this source code is governed by BSD
// license that can be found in the LICENSE file.

package godao

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strings"

	"github.com/jasonjoo2010/enhanced-utils/strutils"
	"github.com/jasonjoo2010/godao/model"
	"github.com/jasonjoo2010/godao/options"
	"github.com/jasonjoo2010/godao/query"
	"github.com/jasonjoo2010/godao/types"
	"github.com/sirupsen/logrus"
)

const (
	internal_TXN = "__TXN__"
)

type DaoTxnContext struct {
	context.Context
}

func (ctx *DaoTxnContext) Txn() *sql.Tx {
	return ctx.Value(internal_TXN).(*sql.Tx)
}

type Dao struct {
	db        *sql.DB
	table     string
	modelType reflect.Type

	// fields
	primaries []*types.ModelField
	fieldMap  map[string]*types.ModelField
	columnMap map[string]*types.ModelField
	fields    []*types.ModelField

	// cache
	selectColumns            []string
	columnsAll, valuesHolder string
}

// NewDao creates a dao object based on given model type.
//	It should be reused as possible to keep efficient.
func NewDao(m interface{}, db *sql.DB, opts ...options.DaoOption) *Dao {
	dao := &Dao{
		db: db,
	}
	// options
	cfg := options.DaoOptions{}
	for _, fn := range opts {
		fn(&cfg)
	}
	if cfg.Table != "" {
		dao.table = cfg.Table
	} else {
		dao.table = strutils.ToUnderscore(model.ParseTableName(m))
	}
	dao.modelType = model.RealType(m)
	// fields
	fields := model.Parse(m)
	if len(fields) < 1 {
		panic("No fields found in model given")
	}
	dao.fields = fields
	dao.primaries = []*types.ModelField{}
	dao.columnMap = make(map[string]*types.ModelField, len(fields))
	dao.fieldMap = make(map[string]*types.ModelField, len(fields))
	columnsBuilder := strings.Builder{}
	holderBuilder := strings.Builder{}
	selectFields := make([]string, 0, len(fields))
	for _, field := range fields {
		dao.columnMap[field.Column] = field
		dao.fieldMap[field.Name] = field
		if field.Primary {
			dao.primaries = append(dao.primaries, field)
		}
		{
			if columnsBuilder.Len() > 0 {
				columnsBuilder.WriteString(", ")
				holderBuilder.WriteString(", ")
			}
			columnsBuilder.WriteByte('`')
			columnsBuilder.WriteString(field.Column)
			columnsBuilder.WriteByte('`')
			holderBuilder.WriteString("?")
			selectFields = append(selectFields, field.Name)
		}
	}
	if len(dao.primaries) < 1 {
		panic("No primary key found")
	}
	dao.columnsAll = columnsBuilder.String()
	dao.valuesHolder = holderBuilder.String()
	dao.selectColumns = selectFields
	return dao
}

// Txn creates a new transaction and wraps it in a context
// which can be used in following invocations.
func (dao *Dao) Txn(opts *sql.TxOptions) (*DaoTxnContext, error) {
	return dao.TxnWithContext(context.Background(), opts)
}

// TxnWithContext creates a new transaction and wrap it in a context
// based on specific context.
func (dao *Dao) TxnWithContext(ctx context.Context, opts *sql.TxOptions) (*DaoTxnContext, error) {
	tx, err := dao.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &DaoTxnContext{context.WithValue(ctx, internal_TXN, tx)}, nil
}

// SelectOne returns the row or nil specified by primary.
// Union primaries are not supported. Please use SelectOneByCondition
func (dao *Dao) SelectOne(ctx context.Context, id interface{}, opts ...options.SelectOption) (interface{}, error) {
	if len(dao.primaries) != 1 {
		panic("SelectOne only support single primary key model: " + dao.table)
	}
	return dao.SelectOneByCondition(ctx,
		(&Query{}).
			Equal(dao.primaries[0].Name, id).
			Limit(1).
			Data(),
		opts...)
}

func (dao *Dao) SelectOneByCondition(ctx context.Context, data query.Data, opts ...options.SelectOption) (interface{}, error) {
	if data.Limit != 1 {
		data.Limit = 1
	}
	rows, err := dao.Select(ctx, data, opts...)
	if err != nil {
		return nil, err
	}
	if len(rows) < 1 {
		return nil, nil
	}
	return rows[0], nil
}

func (dao *Dao) SelectOneBy(ctx context.Context, name string, val interface{}, opts ...options.SelectOption) (interface{}, error) {
	return dao.SelectOneByCondition(ctx,
		(&Query{}).
			Equal(name, val).
			Limit(1).
			Data(),
		opts...)
}

func (dao *Dao) fetchObj(rows *sql.Rows, fields []*types.ModelField) (obj interface{}, err error) {
	args := make([]interface{}, len(fields))
	val := reflect.New(dao.modelType)
	for i, f := range fields {
		args[i] = val.Elem().Field(f.Index).Addr().Interface()
	}
	err = rows.Scan(args...)
	if err == nil {
		obj = val.Interface()
	}
	return
}

func (dao *Dao) Select(ctx context.Context, data query.Data, opts ...options.SelectOption) (result []interface{}, err error) {
	cfg := options.SelectOptions{}
	for _, fn := range opts {
		fn(&cfg)
	}
	condition, args := query.ConditionSQL(dao.fieldMap, dao.columnMap, &data)
	sqlBuilder := strings.Builder{}
	sqlBuilder.WriteString("select ")
	if len(cfg.Fields) == 0 {
		cfg.Fields = dao.selectColumns
	}
	sqlSelect, fieldsSelect := options.GenerateSelectFields(cfg.Fields, dao.fieldMap, dao.columnMap)
	sqlBuilder.WriteString(sqlSelect)
	sqlBuilder.WriteString(" from `")
	sqlBuilder.WriteString(dao.table)
	sqlBuilder.WriteString("`")
	if condition != "" {
		sqlBuilder.WriteString(" ")
		sqlBuilder.WriteString(condition)
	}
	sqlBuilder.WriteString(";")

	var stmt *sql.Stmt
	var txn *sql.Tx
	if txnCtx, ok := ctx.(*DaoTxnContext); ok {
		txn = txnCtx.Txn()
		stmt, err = txn.Prepare(sqlBuilder.String())
	} else {
		stmt, err = dao.db.Prepare(sqlBuilder.String())
	}
	if err != nil {
		return
	}
	defer stmt.Close()

	var rows *sql.Rows
	rows, err = stmt.Query(args...)
	if err != nil {
		return
	}
	for rows.Next() {
		obj, err := dao.fetchObj(rows, fieldsSelect)
		if err != nil {
			logrus.Warn("Convert object failed: ", err.Error())
			continue
		}
		result = append(result, obj)
	}

	return
}

func (dao *Dao) SelectBy(ctx context.Context, name string, val interface{}, limit int, opts ...options.SelectOption) ([]interface{}, error) {
	return dao.Select(ctx,
		(&Query{}).
			Equal(name, val).
			Limit(limit).
			Data(),
		opts...)
}

func (dao *Dao) aggregate(ctx context.Context, data query.Data, aggregation string, values ...interface{}) (err error) {
	conditionSQL, args := query.ConditionSQL(dao.fieldMap, dao.columnMap, &data)

	sqlBuilder := strings.Builder{}
	sqlBuilder.WriteString("select ")
	sqlBuilder.WriteString(aggregation)
	sqlBuilder.WriteString(" from `")
	sqlBuilder.WriteString(dao.table)
	sqlBuilder.WriteString("`")
	if conditionSQL != "" {
		sqlBuilder.WriteString(" ")
		sqlBuilder.WriteString(conditionSQL)
	}

	var stmt *sql.Stmt
	if txnCtx, ok := ctx.(*DaoTxnContext); ok {
		stmt, err = txnCtx.Txn().Prepare(sqlBuilder.String())
	} else {
		stmt, err = dao.db.Prepare(sqlBuilder.String())
	}
	if err != nil {
		return
	}
	defer stmt.Close()

	var row *sql.Row
	row = stmt.QueryRow(args...)
	err = row.Scan(values...)
	return
}

func (dao *Dao) Count(ctx context.Context, data query.Data) (cnt int64, err error) {
	err = dao.aggregate(ctx, data, "count(*)", &cnt)
	return
}

func (dao *Dao) CountBy(ctx context.Context, name string, val interface{}) (int64, error) {
	return dao.Count(ctx,
		(&Query{}).
			Equal(name, val).
			NoLimit().
			Data(),
	)
}

func (dao *Dao) Sum(ctx context.Context, name string, data query.Data) (interface{}, error) {
	columnName := query.GetColumn(name, dao.fieldMap, dao.columnMap, true)
	field := dao.columnMap[columnName]
	fieldSelect := "sum(`" + field.Column + "`)"
	switch field.Type.Kind() {
	case
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		val := int64(0)
		err := dao.aggregate(ctx, data, fieldSelect, &val)
		return val, err
	case
		reflect.Float32,
		reflect.Float64:
		val := float64(0)
		err := dao.aggregate(ctx, data, fieldSelect, &val)
		return val, err
	default:
		panic("Unsupport type for summing")
	}
}

func (dao *Dao) Avg(ctx context.Context, name string, data query.Data) (val float64, err error) {
	columnName := query.GetColumn(name, dao.fieldMap, dao.columnMap, true)
	field := dao.columnMap[columnName]
	err = dao.aggregate(ctx, data, "avg(`"+field.Column+"`)", &val)
	return
}

func (dao *Dao) Insert(ctx context.Context, obj interface{}, opts ...options.InsertOption) (int64, int64, error) {
	affected, ids, err := dao.BatchInsert(ctx, []interface{}{obj}, opts...)
	if len(ids) > 0 {
		return affected, ids[0], err
	}
	return affected, 0, err
}

func (dao *Dao) BatchInsert(ctx context.Context, arr []interface{}, opts ...options.InsertOption) (affected int64, inserted []int64, err error) {
	if len(arr) == 0 {
		return
	}
	inserted = make([]int64, len(arr))
	cfg := &options.InsertOptions{}
	for _, fn := range opts {
		fn(cfg)
	}
	hodler := "(" + dao.valuesHolder + ")"
	sqlBase := options.InsertBaseSQL(dao.table, dao.columnsAll, cfg)
	var stmt *sql.Stmt
	var txn *sql.Tx
	if txnCtx, ok := ctx.(*DaoTxnContext); ok {
		txn = txnCtx.Txn()
	} else {
		txn, err = dao.db.BeginTx(ctx, nil)
		if err != nil {
			return
		}
		defer txn.Commit()
	}
	stmt, err = txn.Prepare(sqlBase + hodler + ";")
	if err != nil {
		return
	}
	defer stmt.Close()

	values := make([]interface{}, len(dao.fields))
	for i, obj := range arr {
		err := model.Flatten(values, dao.modelType, dao.fields, obj)
		if err != nil {
			logrus.Warn("Flatten object failed, ignore: ", err.Error())
			continue
		}
		result, err := stmt.ExecContext(ctx, values...)
		if err != nil {
			logrus.Warn("Insert into table failed: ", err.Error())
			continue
		}
		if result != nil {
			insert_id, err := result.LastInsertId()
			if err != nil {
				logrus.Warn("Fetch insert_id failed: ", err.Error())
			} else {
				inserted[i] = insert_id
			}
			num, err := result.RowsAffected()
			if err != nil {
				logrus.Warn("Fetch affected failed: ", err.Error())
			} else {
				affected += num
			}
		}
	}
	return
}

func (dao *Dao) Update(ctx context.Context, item interface{}) (int64, error) {
	return dao.BatchUpdate(ctx, []interface{}{item})
}

func (dao *Dao) BatchUpdate(ctx context.Context, items []interface{}) (affected int64, err error) {
	var stmt *sql.Stmt
	var txn *sql.Tx
	if txnCtx, ok := ctx.(*DaoTxnContext); ok {
		txn = txnCtx.Txn()
	} else {
		txn, err = dao.db.BeginTx(ctx, nil)
		if err != nil {
			return
		}
		defer txn.Commit()
	}
	stmt, err = txn.Prepare(options.UpdateSQL(dao.table, dao.fields))
	if err != nil {
		return
	}
	defer stmt.Close()
	values := make([]interface{}, len(dao.fields))
	valuesPrimary := make([]interface{}, len(dao.primaries))
	args := make([]interface{}, len(dao.fields))
	for _, item := range items {
		err = model.Flatten(values, dao.modelType, dao.fields, item)
		if err != nil {
			logrus.Warn("Flatten object failed, ignore: ", err.Error())
			continue
		}
		valuesPrimary = valuesPrimary[:0]
		pos := 0
		for i, v := range values {
			if dao.fields[i].Primary {
				valuesPrimary = append(valuesPrimary, v)
			} else {
				args[pos] = v
				pos++
			}
		}
		for _, v := range valuesPrimary {
			args[pos] = v
			pos++
		}
		result, err := stmt.Exec(args...)
		if err != nil {
			logrus.Warn("Update table failed: ", err.Error())
			continue
		}
		num, err := result.RowsAffected()
		if err != nil {
			logrus.Warn("Fetch affected failed: ", err.Error())
		} else {
			affected += num
		}
	}
	return
}

func (dao *Dao) UpdateBy(ctx context.Context, data query.Data, entries ...*types.UpdateEntry) (affected int64, err error) {
	conditionSQL, args := query.ConditionSQL(dao.fieldMap, dao.columnMap, &data)
	if conditionSQL == "" {
		return 0, errors.New("Whole table updating is not allowed")
	}

	sqlBuilder := strings.Builder{}
	sqlBuilder.WriteString("update `")
	sqlBuilder.WriteString(dao.table)
	sqlBuilder.WriteString("` set ")
	updateSQL, values := options.UpdateEntrySQL(entries, dao.fieldMap, dao.columnMap)
	if updateSQL == "" {
		return 0, errors.New("Invalid updating")
	}
	values = append(values, args...)
	sqlBuilder.WriteString(updateSQL)
	sqlBuilder.WriteString(" ")
	sqlBuilder.WriteString(conditionSQL)

	var stmt *sql.Stmt
	if txnCtx, ok := ctx.(*DaoTxnContext); ok {
		stmt, err = txnCtx.Txn().Prepare(sqlBuilder.String())
	} else {
		stmt, err = dao.db.Prepare(sqlBuilder.String())
	}
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(values...)
	if err != nil {
		return
	}
	affected, err = result.RowsAffected()
	return
}

func (dao *Dao) Delete(ctx context.Context, ids ...interface{}) (int64, error) {
	if len(ids) < 1 {
		return 0, nil
	}
	if len(ids) == 1 {
		return dao.DeleteRange(ctx, (&Query{}).
			Equal(dao.primaries[0].Name, ids[0]).
			NoLimit().
			Data())
	}
	return dao.DeleteRange(ctx, (&Query{}).
		In(dao.primaries[0].Name, ids).
		Data())
}

func (dao *Dao) DeleteRange(ctx context.Context, data query.Data) (affected int64, err error) {
	conditionSQL, args := query.ConditionSQL(dao.fieldMap, dao.columnMap, &data)
	if conditionSQL == "" {
		logrus.Panic("Deletion without condition is not allowed")
	}

	sqlBuilder := strings.Builder{}
	sqlBuilder.WriteString("delete from `")
	sqlBuilder.WriteString(dao.table)
	sqlBuilder.WriteString("` ")
	sqlBuilder.WriteString(conditionSQL)

	var stmt *sql.Stmt
	if txnCtx, ok := ctx.(*DaoTxnContext); ok {
		stmt, err = txnCtx.Txn().Prepare(sqlBuilder.String())
	} else {
		stmt, err = dao.db.Prepare(sqlBuilder.String())
	}
	if err != nil {
		return
	}
	defer stmt.Close()

	var result sql.Result
	result, err = stmt.Exec(args...)
	affected, err = result.RowsAffected()
	return
}

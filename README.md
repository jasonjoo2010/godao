# GoDao

GoDao is an implementation of data object access based on `database/sql`.
It provides security and easy for use when accessing data in db.

## Define Model

Naming convertion follows camel style in struct and underscore style in tables. For instance:

```go
Obj.Id => `id`
Obj.id => `id`
Obj.AvgTime => `avg_time`
Obj.avgTime => `avg_time`
Obj.avg_time => `avg_time` // bad style
```

An example:

```go
type Demo struct {
    // Primary and auto_increament key
    Id          int64 `dao:"primary;auto_increment"`
    Name, Value string
    // Omitted because it doesn't existed in table though it's not a good design actually.
    Extra       string `dao:"omit"`
    // The column name is count actually. Also it's not a good design.
    Cnt         int    `dao:"column=count"`
    CreateTime  int64
}
```

## Condition

All conditions are specified by `Query{}`.
It supports:

* Nil / NotNil
* Equal / NotEqual
* Less / LessOrEqual / Greater / GreaterOrEqual
* In / NotIn
* Like / StartsWith / EndsWith
* Expr (Use carefully)
* Order by
* Page / Limit
* Sub query

## Creation

```go
// create db object
db, _ := sql.Open("mysql", "root@tcp(127.0.0.1:3306)/test?charset=utf8mb4,utf8")

// create dao objects
demoDao := NewDao(Demo{}, db)
userDao := NewDao(User{}, db)
// etc.
```

More examples you can refer to [dao_test.go](dao_test.go).

## Insert

```go
demo := Demo{
    Name:    "n1",
    Value:   "v1",
    Created: time.Now().Unix(),
}

// the object can be passed into by value, reference, reference of reference, etc.
// it will be processed correctly internally.
affected, id, err := dao.Insert(context.Background(), demo)
affected, id, err := dao.Insert(context.Background(), &demo)

// for batch(single transaction) through which can achieve better performance
affected, ids, err := dao.BatchInsert(context.Background(), []interface{}{demo, demo1, demo2})

// insert ignore and replace can be supported by options
```

## Query

Get single object from table by primary key:

```go
obj, err := dao.SelectOne(context.Background(), 1)
// err should be nil when success
demo = obj.(*Demo) // do additional type convertion
```

For a collection of objects:

```go
list, err := dao.Select(context.Background(), (&Query{}).
    StartsWith("Name", "key-").
    Greater("Id", 321).
    Order("Id", false).
    Page(2, 10). // page 2, pagesize 10
    Data(),
)
for _, obj := range list {
    demo := obj.(*Demo)
    // etc.
}
```

You can find more examples in `dao_test.go` including `SelectOneBy`, `SelectOneByCondition`, `SelectBy`.

## Update

Update an object after getting and modifying:

```go
affected, err := dao.Update(context.Background(), demo)
```

For batch updating:

```go
affected, err := dao.BatchUpdate(context.Background(), []interface{}{demo1, demo2, demo3})
```

In some special scenarios maybe you just want to update one or more fields, thus part of fields, you can make it by:

```go
affected, err := dao.UpdateBy(context.Background(), (&Query{}).
    Equal("Id", 3).
    Data(),
    &types.UpdateEntry{
        Field: "value", // Better to use field name. But column name can be recognized too in most of methods.
        Value: "test update",
    },
    types.NewIncrease("Cnt", 1),
    &types.UpdateEntry{
        Field: "Name", 
        Expr: "concat(?, @Value@)",
        Args: []interface{}{"key-"}
    },
)
```

## Delete

For single primary key objects deleting by primary key is quite simple:

```go
// one
affected, err := dao.Delete(context.Background(), id0)

// multiple
affected, err := dao.Delete(context.Background(), id1, id2, id3)
```

For range deleting or union primary keys structure:

```go
// id >= 33, type = -1
affected, err := dao.DeleteRange(context.Background(), (&Query{}).
    Equal("Type", -1).
    GreaterOrEqual("Id", 33).
    Data(),
)
```

if you were a very careful guy and you know the accurate count or upper bound of rows when deleting, for some databases, you can spedify the `limit`:

```go
// id >= 33, type = -1, 100 records at most
affected, err := dao.DeleteRange(context.Background(), (&Query{}).
    Equal("Type", -1).
    GreaterOrEqual("Id", 33).
    Limit(100).
    Data(),
)
```

## Other Features

There are other features you can expirence:

* Aggregation: Count/CountBy/Sum/Avg
* Transaction: You can refer to `TestTxnCommit` and `TestTxnRollback` in `dao_test.go`.

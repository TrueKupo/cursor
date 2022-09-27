

1. import library in your project 

```go
   import "github.com/truekupo/cursor"
```

2. add `cursor` tag to each field in structure which you want to be used as a cursor
```go
type Object struct {
 	ID int64
	Name string `cursor:""`
	CreatedAt time.Time `cursor:"default"`
}
```

3. initialize the cursor 
```go
cursor, err := cursor.DefaultCursor(Object{}).
	WithLimit(10).
	WithDirection(common.PageDirForward).
	WithCursorID("Q3JlYXRlZEF0OjE2NjQxNzcyODE0NDU2NzY=")
  ```

4. use cursor in your SQL queries
```go
sql := "SELECT * FROM Objects WHERE 1=1"
params := make(map[string]interface{})
sql, params = spanner.NewBuilder(sql).
	WithCursor(cursor).
	WithParams(params).
	ToSQL()
```


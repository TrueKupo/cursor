## Import cursor library to the service

```shell
$ go get github.com/truekupo/cursor
```

```go
import "github.com/truekupo/cursor"
```


## Model preparation

For a cursor to understand what fields of a model can be used for queries - special tag cursor has to be added to each field: 

```go
package model

import "time"

type Object struct {
    ID string `cursor:""`
    Name string
    CreatedAt time.Time `cursor:"default"`
}
```

Fields that are not tagged with cursor will return an error if used in a cursor.
Tag value cursor:"default" tells cursor builder that this field should be used as a cursor if there’s no cursorId in request.

## Cursor initialisation

Cursor object can be initialised in a few ways, with params or with a method chain.

**Create cursor with params object:**

```go
params := &cursor.Params{
    ID:    "SUQ6MzdlOTNmYmMtNjk3Yy00NmJjLWFhNjMtZDFmMTI3MGNjYzc3",
    Dir:   1,
    Limit: 10,
}

cr, err := cursor.FromParams(Object{}, params)
if err != nil {
    panic(err)
}
```

Where Dir is an integer with 0 = Forward direction, 1 = Backward direction.


**Create cursor with a method chain:**

```go 
cr := cursor.NewDefault(Object{}).
	WithDirection(cursor.Forward).
	WithLimit(10).
	WithCursorID("Q3JlYXRlZEF0OjE2NjQxNzcyODE0NDU2NzY=")
```

In both cases, all parameters/arguments can be omitted for default values to be used. 
The only exception is the model argument which tells the cursor builder what model to use:

```go
cr := cursor.NewDefault(Object{})
```

```go
cr, err := cursor.FromParams(Object{}, &cursor.Params{})
```


## Building SQL

Cursor SQL builder can be instantiated with the following constructor by passing cursor object and database type to it:

```go
builder := cursor.GetBuilder(cr, cursor.Spanner)
```

Cursor sql should be applied to existing query at the end, since it appends `ORDER BY` and `LIMIT` directives to it.

```go
sql := "SELECT * FROM Objects WHERE Kind = @Kind"
params := make(map[string]any)
params["Kind"] = 1

sql, params, err = cursor.GetBuilder(cr, cursor.Spanner).
    WithSQL(sql).
    WithParams(params).
    ToSQL()

// SELECT * FROM Objects WHERE Kind = @Kind AND ID < @ID ORDER BY ID LIMIT 21
```

When no cursor id is provided - only order by and limit will be applied to sql query:

```go
cr = cursor.NewDefault(Object{})
sql, _, err = cr.Builder(cursor.Spanner)
	.WithSQL("SELECT * FROM Objects")
	.ToSQL()

// SELECT * FROM Objects ORDER BY CreatedAt LIMIT 21
```

## Getting results

After executing sql query and retrieving result set, you can get a slice of result and a page info object based on a cursor data:

```go
var out []*Object

// execute query and fetch result
...

res, page, err := cursor.GetResult(cr, out)
```

Resulting slice with be limited to a limit value from the cursor.

package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/truekupo/cursor"
)

type Object struct {
	ID        string    `cursor:""`
	CreatedAt time.Time `cursor:"default,desc"`
}

func main() {

	// create cursor from cursor.Params object
	in := &cursor.Params{
		ID:    "SUQ6MzdlOTNmYmMtNjk3Yy00NmJjLWFhNjMtZDFmMTI3MGNjYzc3",
		Dir:   0,
		Limit: 20,
	}
	cr, err := cursor.FromParams(Object{}, in)
	if err != nil {
		panic(err)
	}
	// and then build sql with created cursor
	sql := "SELECT * FROM Objects WHERE Kind = @Kind"
	params := make(map[string]any)
	params["Kind"] = 1
	sql, params, err = cursor.GetBuilder(cr, cursor.Spanner).
		WithSQL(sql).WithParams(params).
		ToSQL()
	if err != nil {
		panic(err)
	}
	fmt.Println(sql)
	fmt.Println(params)

	// or, initialise cursor and apply sql builder in one chain
	sql = "SELECT * FROM Objects"
	params = make(map[string]any)
	cr = cursor.NewDefault(Object{})
	sql, params, err = cr.WithCursorID("Q3JlYXRlZEF0OjE2NjQxNzcyODE0NDU2NzY=").
		Builder(cursor.Spanner).
		WithSQL(sql).WithParams(params).
		ToSQL()
	if err != nil {
		panic(err)
	}
	fmt.Println(sql)
	fmt.Println(params)

	// after getting result from database, get result slice and page info
	var out []*Object
	for n := 0; n < 31; n++ {
		out = append(out, &Object{ID: strconv.Itoa(n), CreatedAt: time.Now()})
	}
	res, page, err := cursor.GetResult(cr, out)
	fmt.Printf("result length: %d; page info: %+v\n", len(res), page)

	// when no cursor id is provided - only order by and limit will be applied to sql query
	cr = cursor.NewDefault(Object{})
	sql, _, err = cr.Builder(cursor.Spanner).WithSQL("SELECT * FROM Objects").ToSQL()
	if err != nil {
		panic(err)
	}
	fmt.Println(sql)
}

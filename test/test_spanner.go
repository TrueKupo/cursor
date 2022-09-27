package main

import (
	"fmt"
	"time"

	"github.com/truekupo/cursor"
)

type Object struct {
	CreatedAt time.Time `cursor:"default"`
}

func main() {
	cr := cursor.NewDefault(Object{}).
		WithLimit(10).
		WithDirection(cursor.Forward).
		WithCursorID("Q3JlYXRlZEF0OjE2NjQxNzcyODE0NDU2NzY=")

	//goland:noinspection ALL
	sql := "SELECT * FROM Objects"
	params := make(map[string]any)
	sql, params, err := cr.Builder(cursor.Spanner).
		WithSQL(sql).WithParams(params).
		ToSQL()
	if err != nil {
		panic(err)
	}
	fmt.Println(sql)
	fmt.Println(params)

	//goland:noinspection ALL
	sql = "SELECT * FROM Objects WHERE Kind = @Kind"
	params = make(map[string]any)
	params["Kind"] = 1
	sql, params, err = cr.Builder(cursor.Spanner).
		WithSQL(sql).WithParams(params).
		ToSQL()
	if err != nil {
		panic(err)
	}
	fmt.Println(sql)
	fmt.Println(params)
}

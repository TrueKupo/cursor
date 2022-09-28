package main

import (
	"fmt"
	"time"

	"github.com/truekupo/cursor"
)

type Object struct {
	ID        string    `cursor:""`
	CreatedAt time.Time `cursor:"default"`
}

func main() {

	// create cursor and sql builder separately

	cr := cursor.NewDefault(Object{}).
		WithLimit(10).
		WithDirection(cursor.Backward).
		WithCursorID("Q3JlYXRlZEF0OjE2NjQxNzcyODE0NDU2NzY=")

	//goland:noinspection ALL
	sql := "SELECT * FROM Objects"
	params := make(map[string]any)
	sql, params, err := cursor.GetBuilder(cr, cursor.Spanner).
		WithSQL(sql).WithParams(params).
		ToSQL()
	if err != nil {
		panic(err)
	}
	fmt.Println(sql)
	fmt.Println(params)

	// create cursor and builder in one chain

	//goland:noinspection ALL
	sql = "SELECT * FROM Objects WHERE Kind = @Kind"
	params = make(map[string]any)
	params["Kind"] = 1
	sql, params, err = cursor.NewDefault(Object{}).
		WithCursorID("SUQ6MzdlOTNmYmMtNjk3Yy00NmJjLWFhNjMtZDFmMTI3MGNjYzc3").
		Builder(cursor.Spanner).
		WithSQL(sql).WithParams(params).
		ToSQL()
	if err != nil {
		panic(err)
	}
	fmt.Println(sql)
	fmt.Println(params)
}

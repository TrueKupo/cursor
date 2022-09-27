package main

import (
	"fmt"
	"time"

	"github.com/truekupo/cursor/common"
	"github.com/truekupo/cursor/spanner"
)

type Object struct {
	CreatedAt time.Time `cursor:"default"`
}

func main() {
	cr := common.DefaultCursor(Object{}).
		WithLimit(10).
		WithDirection(common.PageDirForward).
		WithCursorID("Q3JlYXRlZEF0OjE2NjQxNzcyODE0NDU2NzY=")

	//goland:noinspection ALL
	sql := "SELECT * FROM Objects WHERE 1=1"
	params := make(map[string]interface{})
	sql, params = spanner.NewBuilder(sql).
		WithCursor(cr).
		WithParams(params).
		ToSQL()
	fmt.Println(sql, params)
}

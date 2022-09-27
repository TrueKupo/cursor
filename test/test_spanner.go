package main

import (
	"fmt"
	"time"

	"github.com/truekupo/cursor/common"
	"github.com/truekupo/cursor/spanner"
)

type Chat struct {
	CreatedAt time.Time `cursor:"default"`
}

func main() {
	cursor, err := common.NewCursor("Q3JlYXRlZEF0OjE2NjQxNzcyODE0NDU2NzY=", 10, common.PageDirForward, Chat{})
	if err != nil {
		panic(err)
	}
	sql := "SELECT * FROM Chats WHERE 1=1"
	params := make(map[string]interface{})
	sql, params = spanner.NewBuilder(sql).
		WithCursor(cursor).
		WithParams(params).
		ToSQL()
	fmt.Println(sql, params)
}

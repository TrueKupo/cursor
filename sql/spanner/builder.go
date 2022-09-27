package spanner

import (
	"fmt"
	"strconv"
)

type BuildData interface {
	Field() string
	Value() any
	IsForward() bool
	IsBackward() bool
	Limit() uint32
}

func BuildSQL(data BuildData) (string, map[string]any) {
	params := make(map[string]any)
	field := data.Field()
	var cond string
	if data.Value() != nil {
		cond += fmt.Sprint(field)
		if data.IsForward() {
			cond += " > "
		} else {
			cond += " < "
		}
		cond += "@" + field
		params[field] = data.Value()
	}

	order := " ORDER BY " + field
	limit := " LIMIT " + strconv.Itoa(int(data.Limit()+1))

	sql := cond + order + limit
	return sql, params
}

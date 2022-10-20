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
	IsAsc() bool
	IsDesc() bool
	Limit() uint32
}

func BuildSQL(data BuildData) (string, map[string]any) {
	params := make(map[string]any)
	field := data.Field()
	var cond, sortOrder string
	if data.Value() != nil {
		cond += fmt.Sprint(field)
		if data.IsForward() {
			if data.IsAsc() {
				cond += " > "
			} else {
				cond += " < "
			}
		} else {
			if data.IsAsc() {
				cond += " < "
			} else {
				cond += " > "
			}
		}
		cond += "@" + field
		params[field] = data.Value()
	}

	if data.IsForward() {
		if data.IsAsc() {
			sortOrder = " ASC"
		} else {
			sortOrder = " DESC"
		}
	} else {
		if data.IsAsc() {
			sortOrder = " DESC"
		} else {
			sortOrder = " ASC"
		}
	}

	order := " ORDER BY " + field + sortOrder
	limit := " LIMIT " + strconv.Itoa(int(data.Limit()+1))

	sql := cond + order + limit
	return sql, params
}

package spanner

import (
	"fmt"
	"strconv"

	"github.com/truekupo/cursor/common"
)

type ParamsMap map[string]interface{}

type sqlBuilder struct {
	sql    string
	params ParamsMap
}

func NewBuilder(sql string) *sqlBuilder {
	return &sqlBuilder{
		sql:    sql,
		params: make(map[string]interface{}),
	}
}

func (b *sqlBuilder) WithCursor(c common.IPageCursor) *sqlBuilder {
	field := c.Field()
	var cond string
	if c.ID() != "" {
		cond += fmt.Sprint(" AND ", field)
		if c.IsForward() {
			cond += " > "
		} else {
			cond += " < "
		}
		cond += "@" + field
		b.params[field] = c.Value()
	}

	order := " ORDER BY " + field
	limit := " LIMIT " + strconv.Itoa(int(c.Limit()+1))

	b.sql += cond + order + limit
	return b
}

func (b *sqlBuilder) WithParams(p ParamsMap) *sqlBuilder {
	for n := range p {
		b.params[n] = p[n]
	}
	return b
}

func (b *sqlBuilder) ToSQL() (string, ParamsMap) {
	return b.sql, b.params
}

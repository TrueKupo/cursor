package cursor

import (
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/truekupo/cursor/sql/spanner"
)

type sqlBuilderKind uint8

const (
	Spanner sqlBuilderKind = iota + 1
)

type BuilderParams map[string]interface{}

type Builder interface {
	WithSQL(sql string) Builder
	WithParams(params BuilderParams) Builder
	ToSQL() (string, BuilderParams, error)
}

type sqlBuilder struct {
	kind   sqlBuilderKind
	cursor Cursor
	sql    string
	params BuilderParams
}

func GetBuilder(c Cursor, kind sqlBuilderKind) Builder {
	return &sqlBuilder{
		kind:   kind,
		cursor: c,
		sql:    "",
		params: make(map[string]interface{}),
	}
}

func (b *sqlBuilder) WithSQL(sql string) Builder {
	b.sql = sql
	return b
}

func (b *sqlBuilder) WithParams(params BuilderParams) Builder {
	b.params = params
	return b
}

func (b *sqlBuilder) ToSQL() (string, BuilderParams, error) {
	var s string
	var p map[string]any
	switch b.kind {
	case Spanner:
		s, p = spanner.BuildSQL(b.cursor)
	default:
		return "", nil, status.Error(codes.InvalidArgument, "invalid builder type")
	}
	sql := b.sql
	if b.cursor.CursorID() != "" {
		if strings.Contains(sql, "WHERE") {
			sql += " AND "
		} else {
			sql += " WHERE "
		}
	}
	sql += s

	params := make(map[string]any)
	for n := range b.params {
		params[n] = b.params[n]
	}
	for n := range p {
		params[n] = p[n]
	}

	return sql, params, nil
}

func (b *sqlBuilder) WithCursor(c Cursor) *sqlBuilder {
	field := c.Field()
	var cond string
	if c.CursorID() != "" {
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

package cursor

import (
	"encoding/base64"
	"reflect"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	tagName  = "cursor"
	defValue = "default"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

// valueType types of cursor values
type valueType uint8

// types of cursor values
const (
	valueTypeInt64 valueType = iota + 1
	valueTypeString
	valueTypeTime
)

// direction page cursor direction type
type direction uint8

// page cursor direction values
const (
	Forward direction = iota
	Backward
)

// NewDefault ...
func NewDefault(obj any) Cursor {
	pc, err := NewCursor(obj, defaultLimit, Forward)
	if err != nil {
		return nil
	}

	return pc
}

// Cursor page cursor interface
type Cursor interface {
	ID() string
	Limit() uint32
	Dir() direction
	IsForward() bool
	IsBackward() bool
	Field() string
	Kind() valueType
	Value() any
	CreateID(obj any) string

	WithLimit(limit uint32) Cursor
	WithDirection(dir direction) Cursor
	WithCursorID(cursorID string) Cursor

	Builder(kind sqlBuilderKind) Builder
}

// IPage page info interface
type IPage interface {
	FirstID() string
	LastID() string
	HasPrev() bool
	HasNext() bool
	Length() uint32
}

type pageCursor struct {
	cursorID string
	limit    uint32
	dir      direction
	field    string
	value    any
	kind     valueType

	model any

	builder Builder
}

func (c *pageCursor) WithLimit(limit uint32) Cursor {
	c.limit = limit
	return c
}

func (c *pageCursor) WithDirection(dir direction) Cursor {
	c.dir = dir
	return c
}

func (c *pageCursor) WithCursorID(cursorID string) Cursor {

	if cursorID == "" {
		err := c.initEmptyCursor()
		if err != nil {
			// TODO
			// may be use some isValid field
			// add metrics/alerts
			return c
		}
	}

	if cursorID != "" {
		b, err := base64.StdEncoding.DecodeString(cursorID)
		if err != nil {
			return c
			//return nil, status.Errorf(codes.InvalidArgument, "failed to decode base64 id to cursor value: %s", cursorID)
		}
		parts := strings.Split(string(b), ":")
		if len(parts) != 2 {
			return c
			//return nil, status.Errorf(codes.InvalidArgument, "invalid cursor id")
		}
		c.field, c.value = parts[0], parts[1]

		t := reflect.TypeOf(c.model)
		sf, ok := t.FieldByName(c.field)
		if !ok {
			return c
			//return nil, status.Errorf(codes.InvalidArgument, "not supported cursor field")
		}
		_, ok = sf.Tag.Lookup(tagName)
		if !ok {
			return c
			//return nil, status.Errorf(codes.InvalidArgument, "not supported cursor field")
		}

		kind := mapFieldType(c.model, c.field)
		if kind == 0 {
			return c
			//return nil, status.Errorf(codes.InvalidArgument, "not supported cursor type")
		}

		c.kind = kind

		c.value, err = decodeFieldValue(parts[1], c.kind)
		if err != nil {
			return c
			//return nil, err
		}
	}

	c.cursorID = cursorID
	return c
}

type pageInfo struct {
	firstID string
	lastID  string
	hasPrev bool
	hasNext bool
	length  uint32
}

// ID get last id
func (c *pageCursor) ID() string {
	return c.cursorID
}

// Limit get query limit
func (c *pageCursor) Limit() uint32 {
	return c.limit
}

// Dir get cursor direction
func (c *pageCursor) Dir() direction {
	return c.dir
}

// IsForward true if cursor direction is forward
func (c *pageCursor) IsForward() bool {
	return c.dir == Forward
}

// IsBackward true if cursor direction is backward
func (c *pageCursor) IsBackward() bool {
	return c.dir == Backward
}

// Field name of field to be used as a cursor
func (c *pageCursor) Field() string {
	return c.field
}

// Value raw string value of cursor
func (c *pageCursor) Value() any {
	return c.value
}

func (c *pageCursor) Kind() valueType {
	return c.kind
}

func (c *pageCursor) CreateID(obj any) string {

	kind := mapFieldType(obj, c.Field())
	if kind == 0 {
		return "INVALID"
	}

	v := reflect.ValueOf(obj)
	f := reflect.Indirect(v).FieldByName(c.Field())

	var value string
	switch c.Kind() {
	case valueTypeInt64:
		value = strconv.FormatInt(f.Int(), 10)
	case valueTypeString:
		value = f.String()
	case valueTypeTime:
		tm := f.Interface().(time.Time)
		m := tm.UnixMicro()
		value = strconv.FormatInt(m, 10)

	default:
		return "INVALID"
	}
	return base64.StdEncoding.EncodeToString([]byte(c.Field() + ":" + value))
}

// FirstID get first cursor id of result page
func (p *pageInfo) FirstID() string {
	return p.firstID
}

// LastID get last cursor id of result page
func (p *pageInfo) LastID() string {
	return p.lastID
}

// HasPrev result has previous page
func (p *pageInfo) HasPrev() bool {
	return p.hasPrev
}

// HasNext result has next page
func (p *pageInfo) HasNext() bool {
	return p.hasNext
}

// Length of result page
func (p *pageInfo) Length() uint32 {
	return p.length
}

func (c *pageCursor) initEmptyCursor() error {
	t := reflect.TypeOf(c.model)

	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get(tagName)
		if tag == defValue {
			kind := mapFieldType(c.model, t.Field(i).Name)
			if kind == 0 {
				return status.Errorf(codes.InvalidArgument, "not supported cursor type")
			}

			c.field = t.Field(i).Name
			c.kind = kind
			break
		}
	}

	if c.field == "" {
		return status.Errorf(codes.InvalidArgument, "default field for cursor not found")
	}

	return nil
}

// NewCursor creates new page cursor object
func NewCursor(obj interface{}, limit uint32, dir direction) (Cursor, error) {

	c := &pageCursor{
		limit: limit,
		dir:   dir,

		model: obj,
	}

	err := c.initEmptyCursor()
	if err != nil {
		return nil, err
	}

	if c.limit == 0 || c.limit > maxLimit {
		c.limit = defaultLimit
	}

	return c, nil
}

// GetResult gets slice of result models
func GetResult[R any](c Cursor, in []*R) ([]*R, IPage, error) {
	return getCursorSlice(c, in), getPageInfo(c, in), nil
}

func getCursorSlice[R any](c Cursor, in []*R) []*R {
	l := len(in)
	if l > int(c.Limit()) {
		l = int(c.Limit())
	}
	return in[:l]
}

func getPageInfo[R any](c Cursor, in []*R) IPage {
	if len(in) == 0 {
		return &pageInfo{hasNext: false}
	}
	resultLen := len(in)

	res := &pageInfo{}

	length := c.Limit()
	if uint32(resultLen) < c.Limit() {
		length = uint32(resultLen)
	}
	res.hasPrev = c.ID() != ""
	res.hasNext = resultLen > int(c.Limit())
	res.length = length

	if length > 0 {
		res.firstID = c.CreateID(in[0])
		res.lastID = c.CreateID(in[length-1])
	}

	return res
}

func mapFieldType(obj any, name string) valueType {

	field := reflect.Indirect(reflect.ValueOf(obj)).FieldByName(name)
	switch field.Kind() {
	case reflect.Int, reflect.Int32:
		return valueTypeInt64
	case reflect.String:
		return valueTypeString
	case reflect.Struct:
		if field.Type() == reflect.TypeOf(time.Time{}) {
			return valueTypeTime
		}
	}

	return 0
}

func decodeFieldValue(raw string, kind valueType) (any, error) {
	switch kind {
	case valueTypeInt64:
		return strconv.ParseInt(raw, 10, 64)
	case valueTypeString:
		return raw, nil
	case valueTypeTime:
		usec, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, err
		}
		return time.UnixMicro(usec), nil
	}
	return nil, status.Error(codes.InvalidArgument, "invalid value type")
}

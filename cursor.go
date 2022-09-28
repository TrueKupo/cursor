package cursor

import (
	"encoding/base64"
	"fmt"
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
	invalidDirection
)

// NewDefault ...
func NewDefault(obj any) Cursor {
	pc, err := NewCursor(obj, defaultLimit, Forward)
	if err != nil {
		return nil
	}

	return pc
}

// Cursor descriptor providing information about cursor and method to modify it
type Cursor interface {
	// CursorID is base64 encoded string containing field name and value to be used as a cursor
	CursorID() string
	// Limit result length with this number
	Limit() uint32
	// Direction of the search, can be either Forward or Backward
	Direction() direction
	// IsForward is true if the direction is Forward
	IsForward() bool
	// IsBackward is true if the direction is Backward
	IsBackward() bool
	// Field name in the model to be used as a cursor
	Field() string
	// Kind is a data type of cursor field
	Kind() valueType
	// Value of cursor field which will be used for search
	Value() any
	// CreateID generates new base64 cursor id pointing to the passed object
	CreateID(obj any) string

	// WithLimit sets result length limit
	WithLimit(limit uint32) Cursor
	// WithDirection sets search direction
	WithDirection(dir direction) Cursor
	// WithCursorID applies new cursor id to be used in a search
	WithCursorID(cursorID string) Cursor

	// Builder helper method which returns sql builder with this cursor
	Builder(kind sqlBuilderKind) Builder
}

// Page provides information about result page
type Page interface {
	// FirstID is id of first item on a page
	FirstID() string
	// LastID is id of last item on a page
	LastID() string
	// HasPrev indicates that there's items on a previous page
	HasPrev() bool
	// HasNext indicates that there's items on a next page
	HasNext() bool
	// Length returns length of result set
	Length() uint32
}

type Params struct {
	ID    string
	Dir   int
	Limit uint32
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

func (c *pageCursor) CursorID() string {
	return c.cursorID
}

func (c *pageCursor) Limit() uint32 {
	return c.limit
}

func (c *pageCursor) Direction() direction {
	return c.dir
}

func (c *pageCursor) IsForward() bool {
	return c.dir == Forward
}

func (c *pageCursor) IsBackward() bool {
	return c.dir == Backward
}

func (c *pageCursor) Field() string {
	return c.field
}

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

func (c *pageCursor) Builder(kind sqlBuilderKind) Builder {
	return GetBuilder(c, kind)
}

func (p *pageInfo) FirstID() string {
	return p.firstID
}

func (p *pageInfo) LastID() string {
	return p.lastID
}

func (p *pageInfo) HasPrev() bool {
	return p.hasPrev
}

func (p *pageInfo) HasNext() bool {
	return p.hasNext
}

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
func NewCursor(obj any, limit uint32, dir direction) (Cursor, error) {
	return createCursor(obj, limit, dir)
}

// FromParams creates new page cursor from Params object
func FromParams(obj any, params *Params) (Cursor, error) {
	if params == nil {
		return nil, status.Error(codes.InvalidArgument, "nil value passed as params")
	}
	dir, err := decodeDirection(params.Dir)
	if err != nil {
		return nil, err
	}
	cr, err := createCursor(obj, params.Limit, dir)
	if err != nil {
		return nil, err
	}
	if params.ID != "" {
		return applyCursorID(cr, params.ID)
	}
	return cr, nil
}

// GetResult returns slice of result models
func GetResult[R any](c Cursor, in []*R) ([]*R, Page, error) {
	return getCursorSlice(c, in), getPageInfo(c, in), nil
}

func getCursorSlice[R any](c Cursor, in []*R) []*R {
	l := len(in)
	if l > int(c.Limit()) {
		l = int(c.Limit())
	}
	return in[:l]
}

func getPageInfo[R any](c Cursor, in []*R) Page {
	if len(in) == 0 {
		return &pageInfo{hasNext: false}
	}
	resultLen := len(in)

	res := &pageInfo{}

	length := c.Limit()
	if uint32(resultLen) < c.Limit() {
		length = uint32(resultLen)
	}
	res.hasPrev = c.CursorID() != ""
	res.hasNext = resultLen > int(c.Limit())
	res.length = length

	if length > 0 {
		res.firstID = c.CreateID(in[0])
		res.lastID = c.CreateID(in[length-1])
	}

	return res
}

func createCursor(obj any, limit uint32, dir direction) (*pageCursor, error) {
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

func applyCursorID(c *pageCursor, cursorID string) (Cursor, error) {

	if cursorID == "" {
		err := c.initEmptyCursor()
		if err != nil {
			// TODO
			// may be use some isValid field
			// add metrics/alerts
			return c, err
		}
	}

	if cursorID != "" {
		b, err := base64.StdEncoding.DecodeString(cursorID)
		if err != nil {
			//return c
			return c, status.Errorf(codes.InvalidArgument, "failed to decode base64 id to cursor value: %s", cursorID)
		}
		parts := strings.Split(string(b), ":")
		if len(parts) != 2 {
			//return c
			return c, status.Errorf(codes.InvalidArgument, "invalid cursor id")
		}
		c.field, c.value = parts[0], parts[1]

		t := reflect.TypeOf(c.model)
		sf, ok := t.FieldByName(c.field)
		if !ok {
			//return c
			return c, status.Errorf(codes.InvalidArgument, "not supported cursor field")
		}
		_, ok = sf.Tag.Lookup(tagName)
		if !ok {
			//return c
			return c, status.Errorf(codes.InvalidArgument, "not supported cursor field")
		}

		kind := mapFieldType(c.model, c.field)
		if kind == 0 {
			//return c
			return c, status.Errorf(codes.InvalidArgument, "not supported cursor type")
		}

		c.kind = kind

		c.value, err = decodeFieldValue(parts[1], c.kind)
		if err != nil {
			//return c
			return c, err
		}
	}

	c.cursorID = cursorID
	return c, nil
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

func decodeDirection(dir int) (direction, error) {
	switch dir {
	case 0:
		return Forward, nil
	case 1:
		return Backward, nil
	}
	return invalidDirection, status.Errorf(codes.InvalidArgument, fmt.Sprintf("invalid direction value %d", dir))
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

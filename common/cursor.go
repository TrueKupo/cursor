package common

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

// PageDir page cursor direction type
type PageDir uint8

// page cursor direction values
const (
	PageDirForward PageDir = iota
	PageDirBackward
)

const (
	tagName  = "cursor"
	defValue = "default"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

// CursorValueType types of cursor values
type CursorValueType uint8

// types of cursor values
const (
	CursorTypeInt64 CursorValueType = iota + 1
	CursorTypeString
	CursorTypeTime
)

// IPageCursor page cursor interface
type IPageCursor interface {
	ID() string
	Limit() uint32
	Dir() PageDir
	IsForward() bool
	IsBackward() bool
	Field() string
	Kind() CursorValueType
	Value() any
	CreateID(obj any) string
}

// IPageInfo page info interface
type IPageInfo interface {
	FirstID() string
	LastID() string
	HasPrev() bool
	HasNext() bool
	Length() uint32
}

type pageCursor struct {
	id    string
	limit uint32
	dir   PageDir
	field string
	value any
	kind  CursorValueType
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
	return c.id
}

// Limit get query limit
func (c *pageCursor) Limit() uint32 {
	return c.limit
}

// Dir get cursor direction
func (c *pageCursor) Dir() PageDir {
	return c.dir
}

// IsForward true if cursor direction is forward
func (c *pageCursor) IsForward() bool {
	return c.dir == PageDirForward
}

// IsBackward true if cursor direction is backward
func (c *pageCursor) IsBackward() bool {
	return c.dir == PageDirBackward
}

// Field name of field to be used as a cursor
func (c *pageCursor) Field() string {
	return c.field
}

// Value raw string value of cursor
func (c *pageCursor) Value() any {
	return c.value
}

func (c *pageCursor) Kind() CursorValueType {
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
	case CursorTypeInt64:
		value = strconv.FormatInt(f.Int(), 10)
	case CursorTypeString:
		value = f.String()
	case CursorTypeTime:
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

// NewCursor creates new page cursor object
func NewCursor(cursorID string, limit uint32, dir PageDir, obj interface{}) (IPageCursor, error) {

	c := &pageCursor{
		id:    cursorID,
		limit: limit,
		dir:   dir,
	}

	if cursorID == "" {
		t := reflect.TypeOf(obj)

		for i := 0; i < t.NumField(); i++ {
			tag := t.Field(i).Tag.Get(tagName)
			if tag == defValue {
				kind := mapFieldType(obj, t.Field(i).Name)
				if kind == 0 {
					return nil, status.Errorf(codes.InvalidArgument, "not supported cursor type")
				}

				c.field = t.Field(i).Name
				c.kind = kind
				break
			}
		}

		if c.field == "" {
			return nil, status.Errorf(codes.InvalidArgument, "default field for cursor not found")
		}
	}

	if cursorID != "" {
		b, err := base64.StdEncoding.DecodeString(cursorID)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to decode base64 id to cursor value: %s", cursorID)
		}
		parts := strings.Split(string(b), ":")
		if len(parts) != 2 {
			return nil, status.Errorf(codes.InvalidArgument, "invalid cursor id")
		}
		c.field, c.value = parts[0], parts[1]

		t := reflect.TypeOf(obj)
		sf, ok := t.FieldByName(c.field)
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "not supported cursor field")
		}
		_, ok = sf.Tag.Lookup(tagName)
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "not supported cursor field")
		}

		kind := mapFieldType(obj, c.field)
		if kind == 0 {
			return nil, status.Errorf(codes.InvalidArgument, "not supported cursor type")
		}

		c.kind = kind

		c.value, err = decodeFieldValue(parts[1], c.kind)
		if err != nil {
			return nil, err
		}
	}

	if limit == 0 {
		c.limit = defaultLimit
	} else if limit > maxLimit {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprint("max limit value is ", maxLimit))
	}

	return c, nil
}

// GetCursorResult gets slice of result models
func GetCursorResult[R any](c IPageCursor, in []*R) ([]*R, IPageInfo, error) {
	return getCursorSlice(c, in), getPageInfo(c, in), nil
}

func getCursorSlice[R any](c IPageCursor, in []*R) []*R {
	l := len(in)
	if l > int(c.Limit()) {
		l = int(c.Limit())
	}
	return in[:l]
}

func getPageInfo[R any](c IPageCursor, in []*R) IPageInfo {
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

func mapFieldType(obj any, name string) CursorValueType {

	field := reflect.Indirect(reflect.ValueOf(obj)).FieldByName(name)
	switch field.Kind() {
	case reflect.Int, reflect.Int32:
		return CursorTypeInt64
	case reflect.String:
		return CursorTypeString
	case reflect.Struct:
		if field.Type() == reflect.TypeOf(time.Time{}) {
			return CursorTypeTime
		}
	}

	return 0
}

func decodeFieldValue(raw string, kind CursorValueType) (any, error) {
	switch kind {
	case CursorTypeInt64:
		return strconv.ParseInt(raw, 10, 64)
	case CursorTypeString:
		return raw, nil
	case CursorTypeTime:
		usec, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, err
		}
		return time.UnixMicro(usec), nil
	}
	return nil, status.Error(codes.InvalidArgument, "invalid value type")
}

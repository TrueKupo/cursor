package cursor

// PageDir page cursor direction type
type PageDir uint8

// page cursor direction values
const (
	PageDirForward PageDir = iota
	PageDirBackward
)

type defaultCursor struct {
}

func DefaultCursor(obj any) *defaultCursor {
}

func (d *defaultCursor) WithLimit(limit int64) *defaultCursor {
  
  return d
}

func (d *defaultCursor) WithDirection(limit PageDir) *defaultCursor {
  
  return d
}

func (d *defaultCursor) WithCursorID(cursorID string) *defaultCursor {
  return d
}

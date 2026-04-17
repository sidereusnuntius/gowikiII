package wikierr

import "errors"

var (
	ErrNotFound = errors.New("not found")

	ErrMissing = errors.New("missing")
)

// type WikierrCode uint8

// const (
// 	ErrNotFound WikierrCode = iota
// 	ErrInternal
// )

// type Wikierr struct {
// 	Code    WikierrCode
// 	Message string
// }

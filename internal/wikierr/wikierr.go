package wikierr

// var (
// 	ErrNotFound = errors.New("not found")

// 	ErrMissing = errors.New("missing")
// )

type WikierrCode uint8

const (
	ErrNotFound WikierrCode = iota
	ErrInternal
	ErrMissing
)

var (
	Missing = New(ErrMissing, "missing")
)

// TODO: enrich this error struct, allowing callers to define an error with a message code (which can be translated into multiple languages) and with context args.
// For instance, for a not found error, they could pass as message something like ErrArticleNotFound, and as args the article slug.
type Wikierr struct {
	Code    WikierrCode
	Message string
}

func (err Wikierr) Error() string {
	return err.Message
}

func New(code WikierrCode, message string, args ...any) error {
	return Wikierr{
		Code:    code,
		Message: message,
	}
}

func Is(err error, code WikierrCode) bool {
	if wikierr, ok := err.(Wikierr); ok {
		return wikierr.Code == code
	}
	return false
}

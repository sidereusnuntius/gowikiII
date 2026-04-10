package wikierr

type ValidationError struct {
	Fields map[string]error
}

func NewValidationError() ValidationError {
	return ValidationError{
		Fields: map[string]error{},
	}
}

func (ve ValidationError) AppendIfNonNil(fieldname string, err error) {
	if err != nil {
		ve.Fields[fieldname] = err
	}
}

func (ve ValidationError) Append(fieldname string, err error) {
	ve.Fields[fieldname] = err
}

func (ve ValidationError) Invalid() bool {
	return len(ve.Fields) > 0
}

func (ve ValidationError) Error() string {
	return "invalid fields"
}

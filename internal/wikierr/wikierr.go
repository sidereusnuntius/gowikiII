package wikierr

type ValidationError struct {
	Fields map[string]error
}

func NewValidationError() ValidationError {
	return ValidationError{
		Fields: map[string]error{},
	}
}

func (ve ValidationError) Append(fieldname string, err error) {
	ve.Fields[fieldname] = err
}

func (ve ValidationError) Error() string {
	return "invalid fields"
}

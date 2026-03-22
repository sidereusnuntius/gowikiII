package validation

import (
	"fmt"
	"regexp"
	"sync"
)

type Rule struct {
	Field string
	// Min length of the field.
	Min int
	// Max length.
	Max         int
	CustomFunc  func(string) error
	Pattern     string
	regex       *regexp.Regexp
	onceCompile sync.Once
}

func (r *Rule) compile() {
	r.regex = regexp.MustCompile("^" + r.Pattern + "$")
}

func (r *Rule) Apply(in string) error {
	if len(r.Pattern) > 0 {
		r.onceCompile.Do(r.compile)
	}

	switch {
	case len(in) < r.Min:
		if len(in) == 0 {
			return fmt.Errorf("%s is empty", r.Field)
		}
		return fmt.Errorf("%s is too short (min: %d)", r.Field, r.Min)
	case len(in) > r.Max:
		return fmt.Errorf("%s is too long (max: %d)", r.Field, r.Max)
	}

	if r.regex != nil && !r.regex.MatchString(in) {
		return fmt.Errorf("%s contains invalid characters", r.Field)
	}

	if r.CustomFunc != nil {
		if err := r.CustomFunc(in); err != nil {
			return err
		}
	}

	return nil
}

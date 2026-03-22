package validation

import (
	"fmt"
	"strings"
	"testing"
)

type testcase struct {
	input string
	errs  bool
	err   string
}

func testRules(t *testing.T, fieldname string, cases []testcase, rule *Rule) {
	for _, c := range cases {
		t.Run(
			fmt.Sprintf("%s %s", fieldname, c.input),
			func(t *testing.T) {
				err := rule.Apply(c.input)
				if err != nil {
					if !c.errs {
						t.Errorf("unexpected error: '%s'", err.Error())
					} else if !strings.Contains(err.Error(), c.err) {
						t.Errorf("error does not contain '%s': '%s'", c.err, err.Error())
					}
				} else if c.errs {
					t.Errorf("expected an error")
				}
			},
		)
	}
}

func TestRules_Username(t *testing.T) {
	cases := []testcase{
		{input: "", errs: true, err: "empty"},
		{input: "ab", errs: true, err: "too short"},
		{input: "abcdefghijklmnopk", errs: true, err: "too long"},
		{input: "u/original/", errs: true, err: "invalid"},
		{input: "jorge amado", errs: true, err: "invalid"},
		{input: "alice", errs: false, err: ""},
	}

	testRules(t, "username", cases, &Username)
}

func TestRules_Email(t *testing.T) {
	cases := []testcase{
		{input: "", errs: true, err: "empty"},
		{input: "jorge", errs: true, err: "invalid"},
		{input: "jorge@", errs: true, err: "invalid"},
		{input: "jorge@gmail@com", errs: true, err: "invalid"},
		{input: "jorge@uol.com.br", errs: false, err: ""},
	}

	testRules(t, "email", cases, &Email)
}

func TestRules_Password(t *testing.T) {
	cases := []testcase{
		{input: "", errs: true, err: "empty"},
		{input: "asdsd", errs: true, err: "short"},
		{input: "coxinha123", errs: false, err: ""},
	}

	testRules(t, "password", cases, &Password)
}

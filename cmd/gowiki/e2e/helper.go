package e2e

import "testing"

func fatalErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

package activitystreams

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sidereusnuntius/gowiki/internal/model/streams"
)

func TestActor(t *testing.T) {
	dmp := diffmatchpatch.New()

	text1 := "a monad is a monoid in te category of edofuctors"
	text2 := "A monad is a monoid in the category of endofunctors."

	diff := dmp.DiffMain(text1, text2, true)

	patches := dmp.PatchMake(diff)
	text := dmp.PatchToText(patches)
	published, _ := time.Parse(streams.Format, "2026-04-08T15:04:05Z")

	patch := Patch{
		Actor:     "https://bio.wiki/u/sally",
		Object:    "http://comp.wiki/a/monad",
		Diff:      text,
		Prev:      "https://comp.wiki/a/monad/edits/0",
		Published: published,
		Summary:   "Fixed typo",
	}

	cases := []struct {
		title    string
		input    string
		expected Patch
		errs     bool
	}{
		{
			title: "patch with string id",
			input: fmt.Sprintf(`{
				"@context": "https://www.w3.org/ns/activitystreams",
				"summary": "Fixed typo",
				"type": "Patch",
				"actor": "https://bio.wiki/u/sally",
				"object": "http://comp.wiki/a/monad",
				"diff": "%s",
				"prev": "https://comp.wiki/a/monad/edits/0",
				"published": "%s"
			}`, text, published.Format(streams.Format)),
			expected: patch,
			errs:     false,
		},
	}

	for _, c := range cases {
		t.Run(c.title, func(t *testing.T) {
			activity, err := ReadActivity([]byte(c.input))
			if err != nil {
				t.Fatalf("ReadActivity(): unexpected error: %v", err)
			}

			patch, err := activity.AsPatch()
			if err != nil {
				if !c.errs {
					t.Errorf("activity.AsPatch(): unexpected: %v", err)
				}
				return
			}

			if diff := cmp.Diff(c.expected, patch); len(diff) > 0 {
				t.Error(diff)
			}
		})
	}
}

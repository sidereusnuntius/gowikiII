package activitystreams

import (
	"time"

	"github.com/sidereusnuntius/gowiki/internal/wikierr"
)

type Patch struct {
	IRI       string
	Actor     string
	Object    string
	Diff      string
	Prev      string
	Published time.Time
	Summary   string
}

// TODO: gather errors into one.
func (a *Activity) AsPatch() (Patch, error) {
	errs := wikierr.NewValidationError()
	id, err := a.Id()
	if err != nil {
		errs.Append("id", err)
	}

	actor, err := a.Actor()
	if err != nil {
		errs.Append("actor", err)
	}

	object, err := a.ObjectId()
	if err != nil {
		errs.Append("object", err)
	}

	diff := string(a.json.GetStringBytes("diff"))
	if len(diff) == 0 {
		errs.Append("diff", wikierr.ErrMissing)
	}

	published, ok, err := a.Published()
	if err != nil {
		errs.Append("published", err)
	} else if !ok {
		errs.Append("published", wikierr.ErrMissing)
	}

	patch := Patch{
		IRI:    id,
		Actor:  actor,
		Object: object,
		Diff:   diff,
		// Prev:      prev,
		Published: published,
	}

	// Can be empty.
	summary := string(a.json.GetStringBytes("summary"))
	if len(summary) > 0 {
		patch.Summary = summary
	}

	prev := string(a.json.GetStringBytes("prev"))
	if len(prev) > 0 {
		patch.Prev = prev
	}

	return patch, nil
}

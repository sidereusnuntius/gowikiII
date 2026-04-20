package activitystreams

import (
	"github.com/sidereusnuntius/gowiki/internal/model"
	"github.com/sidereusnuntius/gowiki/internal/wikierr"
)

func (a *Activity) AsFollow() (model.Follow, error) {
	errs := wikierr.NewValidationError()
	id, err := a.Id()
	errs.AppendIfNonNil("id", err)

	actor, err := a.Actor()
	errs.AppendIfNonNil("actor", err)

	object, err := a.ObjectId()
	errs.AppendIfNonNil("object", err)

	published, _, err := a.Published()
	errs.AppendIfNonNil("published", err)

	if errs.Invalid() {
		return model.Follow{}, errs
	}

	follow := model.Follow{
		IRI:         id,
		FollowerIRI: actor,
		FolloweeIRI: object,
		Published:   published,
	}

	return follow, nil
}

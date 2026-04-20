package federation

import (
	"context"

	"github.com/sidereusnuntius/gowiki/internal/model/activitystreams"
	"github.com/sidereusnuntius/gowiki/internal/wikilog"
)

func (f *Federation) ProcessActivity(ctx context.Context, activity activitystreams.Activity) error {
	var err error
	switch activity.Type {
	case ActivityPatch:
		err = f.ProcessPatch(ctx, activity)
	case ActivityFollow:
		err = f.ProcessFollow(ctx, activity)
	case ActivityAccept:
		err = f.ProcessAccept(ctx, activity)
	}

	return err
}

func (f *Federation) ProcessPatch(ctx context.Context, activity activitystreams.Activity) error {
	patch, err := activity.AsPatch()
	if err != nil {
		return err
	}

	if err = f.EnsureActor(ctx, patch.Actor); err != nil {
		return err
	}

	return f.Articles.RemotePatch(ctx, patch)
}

func (f *Federation) ProcessFollow(ctx context.Context, activity activitystreams.Activity) error {
	wikilog.Logger.Info().Msg("processing follow activity")
	follow, err := activity.AsFollow()
	if err != nil {
		return err
	}

	if err = f.EnsureActor(ctx, follow.FollowerIRI); err != nil {
		return err
	}

	return f.Actors.Follow(ctx, &follow)
}

// This can get more complex if we allow accepting other kinds of objects.
func (f *Federation) ProcessAccept(ctx context.Context, activity activitystreams.Activity) error {
	object, err := activity.ObjectId()
	if err != nil {
		return err
	}

	actor, err := activity.Actor()
	if err != nil {
		return err
	}

	return f.Actors.AcceptFollow(ctx, actor, object)
}

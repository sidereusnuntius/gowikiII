package federation

import (
	"context"

	"github.com/sidereusnuntius/gowiki/internal/model/activitystreams"
)

func (f *Federation) ProcessActivity(ctx context.Context, activity activitystreams.Activity) error {
	var err error
	switch activity.Type {
	case ActivityPatch:
		err = f.ProcessPatch(ctx, activity)
	}

	return err
}

func (f *Federation) ProcessPatch(ctx context.Context, activity activitystreams.Activity) error {
	patch, err := activity.AsPatch()
	if err != nil {
		return err
	}

}

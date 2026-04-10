package federation

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/sidereusnuntius/gowiki/internal/model/activitystreams"
)

type Federation struct {
}

const (
	ActivityPatch = "Patch"
)

func (f *Federation) InstanceInbox(r *http.Request) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("io.ReadAll(): %w", err)
	}

	activity, err := activitystreams.ReadActivity(body)
	if err != nil {
		return err
	}

	return f.ProcessActivity(r.Context(), activity)
}

func (f *Federation) EnsureActor(ctx context.Context, actorId string) error {

}

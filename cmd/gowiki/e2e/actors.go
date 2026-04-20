package e2e

import (
	"slices"
	"testing"
)

// hasActor asserts that the instance contains a local copy of the actor.
func (r *TestRig) hasActor(t *testing.T, actorID string) func(t *testing.T) {
	return func(t *testing.T) {
		actor, err := r.Wiki.Actors.GetActorByIRI(t.Context(), actorID)
		fatalErr(t, err)

		if actor.URI != actorID {
			t.Errorf("expected actor to have IRI %s, got %s", actorID, actor.URI)
		}
	}
}

// assertFollowedBy asserts that actor is followed by the follower.
func (r *TestRig) assertFollowedBy(actor, follower string) func(t *testing.T) {
	return func(t *testing.T) {
		followers, err := r.Wiki.Actors.GetFollowers(t.Context(), actor)
		fatalErr(t, err)

		if len(followers) == 0 {
			t.Fatalf("actor %s has no followers", actor)
		}

		if i := slices.Index(followers, follower); i != -1 {
			t.Logf("found: %s does follow %s", follower, actor)
			return
		}
		t.Fatalf("actor %s does not follow %s", follower, actor)
	}
}

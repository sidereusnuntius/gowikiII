package e2e

import "testing"

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

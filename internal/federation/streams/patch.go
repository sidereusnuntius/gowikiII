package streams

import (
	"github.com/sidereusnuntius/gowiki/internal/model"
)

type Patch struct {
	Activity
	Diff    string `json:"diff,omitempty"`
	Summary string `json:"summary,omitempty"`
}

func PatchAS(articleIRI, patches string, revision *model.Revision) Patch {
	return Patch{
		Activity: Activity{
			Base: Base{
				Context:   context,
				Type:      "Patch",
				Id:        revision.IRI,
				Published: revision.Published.UTC().Format(Format),
			},
			Actor:  revision.ActorIRI,
			Object: articleIRI,
		},
		Diff:    patches,
		Summary: revision.Summary,
	}
}

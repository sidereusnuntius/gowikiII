package streams

import (
	"github.com/sidereusnuntius/gowiki/internal/model"
)

func FollowAS(follow *model.Follow) Activity {
	return Activity{
		Base: Base{
			Context:   context,
			Type:      "Follow",
			Id:        follow.IRI,
			Published: follow.Published.Format(Format),
		},
		Actor:  follow.FollowerIRI,
		Object: follow.FolloweeIRI,
	}
}

func Accept(followID, actorID string) Activity {
	return Activity{
		Base: Base{
			Context: context,
			Type:    "Accept",
		},
		Actor:  actorID,
		Object: followID,
	}
}

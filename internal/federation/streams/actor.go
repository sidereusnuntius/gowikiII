package streams

import "github.com/sidereusnuntius/gowiki/internal/model"

type Actor struct {
	Base
	Username  string    `json:"preferredUsername,omitzero"`
	Name      string    `json:"name,omitempty"`
	Summary   string    `json:"summary,omitempty"`
	Inbox     string    `json:"inbox,omitempty"`
	Outbox    string    `json:"outbox,omitempty"`
	Followers string    `json:"followers,omitempty"`
	Following string    `json:"following,omitempty"`
	PublicKey PublicKey `json:"publicKey,omitzero"`
	Endpoints Endpoints `json:"endpoints,omitzero"`
}

type PublicKey struct {
	Id    string `json:"id"`
	Owner string `json:"owner"`
	Pem   string `json:"publicKeyPem"`
}

func ActorAS(actor *model.Actor) Actor {
	return Actor{
		Base: Base{
			Context:   context,
			Type:      actor.Type,
			Id:        actor.URI,
			Published: actor.Published.Format(Format),
			Updated:   actor.Updated.Format(Format),
			Url:       actor.URL,
		},
		Username:  actor.Username,
		Name:      actor.DisplayName,
		Summary:   actor.Summary,
		Inbox:     actor.Inbox,
		Outbox:    actor.Outbox,
		Followers: actor.Followers,
		Following: actor.Following,
		PublicKey: PublicKey{
			Id:    actor.PublicKey.URI,
			Owner: actor.URI,
			Pem:   string(actor.PublicKey.Pem),
		},
		Endpoints: Endpoints{
			SharedInbox: actor.SharedInbox,
		},
	}
}

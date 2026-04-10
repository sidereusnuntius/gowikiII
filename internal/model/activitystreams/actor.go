package activitystreams

import (
	"time"

	"github.com/sidereusnuntius/gowiki/internal/model"
	"github.com/sidereusnuntius/gowiki/internal/wikierr"
)

type Actor struct {
	Id        string
	Type      string
	Username  string
	Name      string
	Summary   string
	Inbox     string
	Outbox    string
	Followers string
	Following string
	PublicKey model.PublicKey
	Endpoints string
	Published time.Time
	Updated   time.Time
}

func (o *Object) Username() (string, error) {
	username := string(o.json.GetStringBytes("preferredUsername"))
	if len(username) == 0 { // TODO: perform complete validation.
		return "", wikierr.ErrMissing
	}

	return username, nil
}

func (o *Object) PublicKey() (model.PublicKey, error) {
	errs := wikierr.NewValidationError()
	publicKey := o.json.Get("publicKey")
	if publicKey == nil {
		return model.PublicKey{}, wikierr.ErrMissing
	}

	id, err := GetIRI("id", publicKey)
	errs.AppendIfNonNil("id", err)

	owner, err := GetIRI("owner", publicKey)
	errs.AppendIfNonNil("owner", err)

	publicKeyPem := publicKey.GetStringBytes("publicKeyPem")
	if len(publicKeyPem) == 0 {
		errs.AppendIfNonNil("publicKeyPem", err)
	}

	if errs.Invalid() {
		return model.PublicKey{}, errs
	}

	return model.PublicKey{
		URI:      id,
		OwnerIRI: owner,
		Pem:      publicKeyPem,
	}, nil
}

func (o *Object) AsActor() (Actor, error) {
	errs := wikierr.NewValidationError()
	id, err := o.Id()
	errs.AppendIfNonNil("id", err)

	username, err := o.Username()
	errs.AppendIfNonNil("preferredUsername", err)

	inbox, err := GetIRI("inbox", o.json)
	errs.AppendIfNonNil("inbox", err)

	outbox, err := GetIRI("outbox", o.json)
	errs.AppendIfNonNil("outbox", err)

	publicKey, err := o.PublicKey()
	errs.AppendIfNonNil("publicKey", err)

	actor := Actor{
		Id:        id,
		Type:      o.Type,
		Username:  username,
		Inbox:     inbox,
		Outbox:    outbox,
		PublicKey: publicKey,
	}

	// Optional fields.
	displayName := string(o.json.GetStringBytes("name"))
	if len(displayName) > 0 {
		actor.Name = displayName
	}

	followers, err := GetIRI("followers", o.json)
	if err == nil {
		actor.Followers = followers
	}

	following, err := GetIRI("following", o.json)
	if err == nil {
		actor.Following = following
	}

	published, ok, err := o.Published()
	errs.Append("published", err)
	if ok && err == nil {
		actor.Published = published
	}

	return actor, nil
}

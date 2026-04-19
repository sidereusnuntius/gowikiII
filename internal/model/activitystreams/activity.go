package activitystreams

import (
	"github.com/sidereusnuntius/gowiki/internal/wikierr"
	"github.com/valyala/fastjson"
)

type Activity struct {
	Object
}

func ReadActivity(raw []byte) (Activity, error) {
	obj, err := ReadObject(raw)
	if err != nil {
		return Activity{}, err
	}

	return Activity{
		Object: obj,
	}, nil
}

// Actor returns the id of the actor of the activity, if this property is set.
func (a *Activity) Actor() (string, error) {
	actor := a.json.Get("actor")
	if actor == nil {
		return "", wikierr.Missing
	}

	var actorId []byte
	switch actor.Type() {
	case fastjson.TypeString:
		actorId = actor.GetStringBytes()

	case fastjson.TypeObject:
		actorId = actor.GetStringBytes("id")
	}

	if len(actorId) == 0 {
		return "", wikierr.Missing
	}

	return string(actorId), wikierr.Missing
}

func (a *Activity) ObjectId() (string, error) {
	object := a.json.Get("object")
	if object == nil {
		return "", wikierr.Missing
	}

	var objectId []byte
	switch object.Type() {
	case fastjson.TypeString:
		objectId = object.GetStringBytes()
	case fastjson.TypeObject:
		objectId = object.GetStringBytes("id")
	}

	if len(objectId) == 0 {
		return "", wikierr.Missing
	}

	return string(objectId), nil
}

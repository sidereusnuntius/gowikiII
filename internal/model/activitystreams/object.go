package activitystreams

import (
	"errors"
	"fmt"
	"time"

	"github.com/sidereusnuntius/gowiki/internal/federation/streams"
	"github.com/sidereusnuntius/gowiki/internal/wikierr"
	"github.com/valyala/fastjson"
)

type Object struct {
	json *fastjson.Value
	Type string
}

func ReadObject(raw []byte) (Object, error) {
	obj, err := fastjson.ParseBytes(raw)
	if err != nil {
		return Object{}, fmt.Errorf("fastjson.ParseBytes(): failed to parse request body: %w", err)
	}

	typ := obj.GetStringBytes("type")
	if len(typ) == 0 {
		return Object{}, errors.New("activity does not have a type")
	}

	return Object{
		json: obj,
		Type: string(typ),
	}, nil
}

// GetIRI returns the value of the property with the provided name,
// and checks if it is a valid IRI.
func GetIRI(name string, json *fastjson.Value) (string, error) {
	iri := string(json.GetStringBytes(name))
	if len(iri) == 0 {
		return "", wikierr.ErrMissing
	}

	// TODO: Perform IRI validation
	return iri, nil
}

func (o *Object) Id() (string, error) {
	return GetIRI("id", o.json)
}

func (o *Object) Published() (time.Time, bool, error) {
	publishedStr := string(o.json.GetStringBytes("published"))
	if len(publishedStr) == 0 {
		var zero time.Time
		return zero, false, nil
	}

	published, err := time.Parse(streams.Format, publishedStr)
	// TODO invalid time, expected format: <Format>

	return published, err == nil, err
}

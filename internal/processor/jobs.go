package processor

type FetchActorArgs struct {
	IRI           string `json:"iri"`
	InstanceActor bool   `json:"instanceActor"`
}

func (FetchActorArgs) Kind() string {
	return "fetch_actor"
}

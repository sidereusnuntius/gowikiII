package streams

var context = []string{
	"https://www.w3.org/ns/activitystreams",
}

var Format = "2006-01-02T15:04:05Z"

type Base struct {
	Context   any    `json:"@context"`
	Type      string `json:"type"`
	Id        string `json:"id"`
	Published string `json:"published,omitempty"`
	Updated   string `json:"updated,omitempty"`
	Url       string `json:"url,omitempty"`
}

type Activity struct {
	Base
	Actor  string `json:"actor,omitempty"`
	Object any    `json:"object,omitempty"`
}

type Endpoints struct {
	SharedInbox string `json:"sharedInbox,omitzero"`
}

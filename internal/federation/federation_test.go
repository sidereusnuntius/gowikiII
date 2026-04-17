package federation

import (
	"fmt"
	"testing"

	"github.com/sidereusnuntius/gowiki/internal/mocks"
	"github.com/sidereusnuntius/gowiki/internal/model"
	"github.com/sidereusnuntius/gowiki/internal/processor"
	"github.com/sidereusnuntius/gowiki/internal/tests"
	"github.com/sidereusnuntius/gowiki/internal/wikierr"
	"github.com/stretchr/testify/mock"
)

type state struct {
	store      *mocks.MockFedStore
	client     *mocks.MockClient
	actors     *mocks.MockActors
	articles   *mocks.MockArticles
	publisher  *mocks.MockPublisher
	federation *Federation
}

func initialize() state {
	store := new(mocks.MockFedStore)
	client := new(mocks.MockClient)
	actors := new(mocks.MockActors)
	articles := new(mocks.MockArticles)
	publisher := new(mocks.MockPublisher)
	federation := New(tests.TestConfig("test.wiki"), store, client, actors, articles, publisher)

	return state{
		store:      store,
		client:     client,
		actors:     actors,
		articles:   articles,
		publisher:  publisher,
		federation: federation,
	}
}

func TestCheckOriginHost(t *testing.T) {
	cases := []struct {
		title string
		host  string

		// Whether there is a record for the host in the local store.
		cached       bool
		cachedHost   model.Host
		storeErr     error
		storeSaveErr error
		publisherErr error

		allowed bool
		errs    bool
	}{
		{
			title:  "cached host, allowed",
			host:   "bio.wiki",
			cached: true,
			cachedHost: model.Host{
				ID:     1,
				Host:   "bio.wiki",
				Status: model.Peer,
			},
			storeErr:     nil,
			storeSaveErr: nil,
			publisherErr: nil,

			allowed: true,
			errs:    false,
		},
		{
			title:  "cached host, blocked",
			host:   "bio.wiki",
			cached: true,
			cachedHost: model.Host{
				ID:     1,
				Host:   "bio.wiki",
				Status: model.Blocked,
			},
			storeErr:     nil,
			storeSaveErr: nil,
			publisherErr: nil,

			allowed: false,
			errs:    false,
		},
		{
			title:        "not cached host; must fetch",
			host:         "bio.wiki",
			cached:       false,
			cachedHost:   model.Host{},
			storeErr:     wikierr.ErrNotFound,
			storeSaveErr: nil,
			publisherErr: nil,

			allowed: true,
			errs:    false,
		},
	}

	for _, c := range cases {
		t.Run(c.title, func(t *testing.T) {
			state := initialize()

			if c.cached {
				// Host is already known.
				state.store.On("GetHost", mock.Anything, c.host).Return(c.cachedHost, nil)
			} else {
				// Host is unkown; should save and issue a fetch request.
				state.store.On("GetHost", mock.Anything, c.host).Return(model.Host{}, c.storeErr)
				savedhost := model.Host{
					Host:   c.host,
					Status: model.HostUnknown,
				}
				state.store.On("SaveHost", mock.Anything, &savedhost).Return(c.storeSaveErr)

				args := processor.FetchActorArgs{
					IRI:           fmt.Sprintf("http://%s", c.host),
					InstanceActor: true,
				}
				state.publisher.On("Publish", mock.Anything, args).Return(c.publisherErr)
			}

			allowed, err := state.federation.CheckOriginHost(t.Context(), c.host)
			if allowed != c.allowed {
				t.Errorf("state.federation.CheckOriginHost(): expected %v, got %v", c.allowed, allowed)
			}

			if !c.errs && err != nil {
				t.Errorf("state.federation.CheckOriginHost(): unexpected error: %v", err)
			}

			state.store.AssertExpectations(t)
			state.publisher.AssertExpectations(t)
		})
	}
}

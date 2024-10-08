package schema

import (
	"context"
	"errors"

	"github.com/dagger/dagger/core"
	"github.com/dagger/dagger/dagql"
)

type cacheSchema struct {
	srv *dagql.Server
}

var _ SchemaResolvers = &cacheSchema{}

func (s *cacheSchema) Name() string {
	return "cache"
}

func (s *cacheSchema) Install() {
	dagql.Fields[*core.Query]{
		dagql.Func("cacheVolume", s.cacheVolume).
			Doc("Constructs a cache volume for a given cache key.").
			ArgDoc("key", `A string identifier to target this cache volume (e.g., "modules-cache").`).
			Impure("evaluate each time"),
		dagql.Func("cache", s.cache).
			Doc("Constructs a cache volume for a given cache key.").
			ArgDoc("key", `A string identifier to target this cache volume (e.g., "modules-cache").`).
			Impure("evaluate each time"),
	}.Install(s.srv)

	dagql.Fields[*core.CacheVolume]{}.Install(s.srv)
}

func (s *cacheSchema) Dependencies() []SchemaResolvers {
	return nil
}

type cacheArgs struct {
	Key string
}

func (s *cacheSchema) cache(ctx context.Context, parent *core.Query, args cacheArgs) (*core.CacheVolume, error) {
	return core.NewCache(args.Key), nil
}

func (s *cacheSchema) cacheVolume(ctx context.Context, parent *core.Query, args cacheArgs) (*core.CacheVolume, error) {
	// TODO(vito): inject some sort of scope/session/project/user derived value
	// here instead of a static value
	//
	// we have to inject something so we can tell it's a valid ID
	m, err := parent.Server.CurrentModule(ctx)
	if err != nil && !errors.Is(err, core.ErrNoCurrentModule) {
		return nil, err
	}

	key := args.Key
	if m != nil {
		// return nil, fmt.Errorf("INSIDE IF DIGEST IS -> %q == %q", m.NameField, m.Source.ID().Digest())
		key = m.Source.ID().Digest().String() + key
	}

	// test module used is here: https://github.com/rajatjindal/dagger-same-cache-volume-id
	// I am trying to ensure that when I try to create a cache-volume with same name, across two diff modules (foo and bar),
	// they should return unpredictable and different CacheVolumeID.
	// right now, my tests show that they always return the same ID:
	// ChV4eGgzOjAzOWZlMzM0Y2YzN2ZiM2EScQoVeHhoMzowMzlmZTMzNGNmMzdmYjNhElgSDwoLQ2FjaGVWb2x1bWUYARoLY2FjaGVWb2x1bWUiHwoDa2V5Ehg6FnZvbHVtZS1uYW1lLWNoZWNrLWVsc2UoAUoVeHhoMzowMzlmZTMzNGNmMzdmYjNh

	// To try without an additional select, uncomment following line
	// and comment out rest of the code in this function
	// example traces:
	// foo -> https://dagger.cloud/rajatjindal/traces/10551b3b41758a6cb3ed2f69491de610
	// bar -> https://dagger.cloud/rajatjindal/traces/4ff8f9a58ffc427bb6dadfeaa2650096
	// return core.NewCache(key), nil

	// The following code tries to make a select call
	// with assumption that the id is generated during `cacheSelect` function
	// and with this additional select, we are sending the key as original-key + module digest
	// example traces:
	// foo -> https://dagger.cloud/rajatjindal/traces/3b3512c0f0db9aec8573b4dfed412f26
	// bar -> https://dagger.cloud/rajatjindal/traces/f49e73c419e89aec04aa5b6a0db75d98
	var svc dagql.Instance[*core.CacheVolume]
	err = s.srv.Select(ctx, s.srv.Root(), &svc,
		dagql.Selector{
			Field: "cache",
			Args: []dagql.NamedInput{
				{
					Name:  "key",
					Value: dagql.NewString(key),
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return svc.Self, nil
}

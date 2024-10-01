package schema

import (
	"context"
	"fmt"

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
			ArgDoc("key", `A string identifier to target this cache volume (e.g., "modules-cache").`),
		dagql.Func("cacheVolume2", s.cacheVolume2).
			Doc("Constructs a cache volume for a given cache key.").
			ArgDoc("key", `A string identifier to target this cache volume (e.g., "modules-cache").`),
	}.Install(s.srv)

	dagql.Fields[*core.CacheVolume]{}.Install(s.srv)
}

func (s *cacheSchema) Dependencies() []SchemaResolvers {
	return nil
}

type cacheArgs2 struct {
	Key string

	// Accessor is the scoped per-module name, which should guarantee uniqueness.
	// It is used to ensure the dagql ID digest is unique per module; the digest is what's
	// used as the actual key for the cache volume store.
	Accessor dagql.Optional[dagql.String]
}

func (s *cacheSchema) cacheVolume2(ctx context.Context, parent *core.Query, args cacheArgs2) (*core.CacheVolume, error) {
	return &core.CacheVolume{
		Keys:     []string{args.Key},
		Query:    parent,
		IDDigest: dagql.CurrentID(ctx).Digest(),
	}, nil
}

type cacheArgs struct {
	Key string
}

func (s *cacheSchema) cacheVolume(ctx context.Context, parent *core.Query, args cacheArgs) (i dagql.Instance[*core.CacheVolume], err error) {
	cacheVolumeStore, err := parent.CacheVolumes(ctx)
	if err != nil {
		return i, fmt.Errorf("failed to get secret store: %w", err)
	}

	accessor, err := core.GetClientResourceAccessor(ctx, parent, args.Key)
	if err != nil {
		return i, fmt.Errorf("failed to get client resource name: %w", err)
	}

	// NB: to avoid putting the plaintext value in the graph, return a freshly
	// minted Object that just gets the secret by name
	if err := s.srv.Select(ctx, s.srv.Root(), &i, dagql.Selector{
		Field: "cacheVolume2", // TODO: need some bikeshedding for the name here
		Args: []dagql.NamedInput{
			{
				Name:  "key",
				Value: dagql.NewString(args.Key),
			},
			{
				Name:  "accessor",
				Value: dagql.Opt(dagql.NewString(accessor)),
			},
		},
	}); err != nil {
		return i, fmt.Errorf("failed to select cache volume: %w", err)
	}

	if err := cacheVolumeStore.AddCacheVolume(i.Self, args.Key); err != nil {
		return i, fmt.Errorf("failed to add cache volume: %w", err)
	}

	return i, nil
}

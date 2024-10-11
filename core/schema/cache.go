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
		dagql.NodeFunc("cacheVolume", s.cacheVolume).
			Doc("Constructs a cache volume for a given cache key.").
			ArgDoc("key", `A string identifier to target this cache volume (e.g., "modules-cache").`).
			Impure("evaluate each time 1"),
		dagql.Func("cache", s.cache).
			Doc("Constructs a cache volume for a given cache key.").
			ArgDoc("key", `A string identifier to target this cache volume (e.g., "modules-cache").`),
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

func (s *cacheSchema) cacheVolume(ctx context.Context, parent dagql.Instance[*core.Query], args cacheArgs) (dagql.Instance[*core.CacheVolume], error) {
	// TODO(vito): inject some sort of scope/session/project/user derived value
	// here instead of a static value
	//
	// we have to inject something so we can tell it's a valid ID
	var inst dagql.Instance[*core.CacheVolume]

	m, err := parent.Self.Server.CurrentModule(ctx)
	if err != nil && !errors.Is(err, core.ErrNoCurrentModule) {
		return inst, err
	}

	key := args.Key
	if m != nil {
		key = m.Source.ID().Digest().String() + key
	}

	err = s.srv.Select(ctx, s.srv.Root(), &inst,
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
		return inst, err
	}

	return inst, nil
}

package schema

import (
	"context"
	"errors"
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
		dagql.NodeFunc("cacheVolume", s.cacheVolume).
			Doc("Constructs a cache volume for a given cache key.").
			ArgDoc("key", `A string identifier to target this cache volume (e.g., "modules-cache").`),
	}.Install(s.srv)

	dagql.Fields[*core.CacheVolume]{}.Install(s.srv)
}

func (s *cacheSchema) Dependencies() []SchemaResolvers {
	return nil
}

type cacheArgs struct {
	Key       string
	Namespace string `default:""`
}

func (s *cacheSchema) cacheVolume(ctx context.Context, parent dagql.Instance[*core.Query], args cacheArgs) (dagql.Instance[*core.CacheVolume], error) {
	if args.Namespace != "" {
		return dagql.NewInstanceForCurrentID(ctx, s.srv, parent, core.NewCache(args.Key+args.Namespace))
	}

	var inst dagql.Instance[*core.CacheVolume]
	m, err := parent.Self.Server.CurrentModule(ctx)
	if err != nil && !errors.Is(err, core.ErrNoCurrentModule) {
		return inst, err
	}

	id := ""
	if m != nil {
		id = m.Source.ID().Digest().String()
	}

	// if no source id digest, just return the NewCache based on key
	if id == "" {
		return dagql.NewInstanceForCurrentID(ctx, s.srv, parent, core.NewCache(args.Key))
	}

	// otherwise append namespace and call again

	// taint it if no namespace is provided
	dagql.Taint(ctx)
	// return dagql.NewInstanceForCurrentID(ctx, s.srv, parent, core.NewCache(args.Key))

	err = s.srv.Select(ctx, s.srv.Root(), &inst, dagql.Selector{
		Field: "cacheVolume",
		Args: []dagql.NamedInput{
			{
				Name:  "key",
				Value: dagql.NewString(args.Key),
			},
			{
				Name:  "namespace",
				Value: dagql.NewString(id),
			},
		},
	})
	if err != nil {
		return inst, fmt.Errorf("TODO A GOOD ERROR: %#v", err)
	}

	return inst, nil
}

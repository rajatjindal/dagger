package schema

import (
	"context"

	"github.com/dagger/dagger/core"
	"github.com/dagger/dagger/dagql"
)

type socketSchema struct {
	srv *dagql.Server
}

var _ SchemaResolvers = &socketSchema{}

func (s *socketSchema) Install(ctx context.Context) {
	dagql.Fields[*core.Socket]{}.Install(ctx, s.srv)
}

package secretprovider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/moby/buildkit/session/secrets"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SecretResolver interface {
	GetSecret(context.Context, string) ([]byte, error)
}

var resolvers = map[string]SecretResolver{
	"env":   envProvider{},
	"file":  fileProvider{},
	"cmd":   cmdProvider{},
	"op":    opProvider{},
	"vault": vaultProvider{},
	// "libsecret": libsecretProvider,
}

func ResolverForID(uri string) (SecretResolver, string, error) {
	scheme, _, ok := strings.Cut(uri, "://")
	if !ok {
		return nil, "", fmt.Errorf("parse %q: malformed id", uri)
	}

	resolver, ok := resolvers[scheme]
	if !ok {
		return nil, "", fmt.Errorf("unsupported secret provider: %q", scheme)
	}
	return resolver, uri, nil
}

type SecretProvider struct {
}

func NewSecretProvider() SecretProvider {
	return SecretProvider{}
}

func (sp SecretProvider) Register(server *grpc.Server) {
	secrets.RegisterSecretsServer(server, sp)
}

func (sp SecretProvider) GetSecret(ctx context.Context, req *secrets.GetSecretRequest) (*secrets.GetSecretResponse, error) {
	resolver, uri, err := ResolverForID(req.ID)
	if err != nil {
		return nil, err
	}

	plaintext, err := resolver.GetSecret(ctx, uri)
	if err != nil {
		if errors.Is(err, secrets.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, err
	}

	return &secrets.GetSecretResponse{
		Data: plaintext,
	}, nil
}

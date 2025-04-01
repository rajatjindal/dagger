package secretprovider

import (
	"context"
	"fmt"
	"os"

	vault "github.com/hashicorp/vault/api"
	auth "github.com/hashicorp/vault/api/auth/approle"
)

var (
	vaultClient *vault.Client
	vaultCache  = make(map[string]map[string]any)
)

type vaultProvider struct{}

func (v vaultProvider) CacheProvider() CacheProvider {
	return ObjectCacheProvider{}
}

// HashiCorp Vault provider for SecretProvider
func (v vaultProvider) GetSecret(ctx context.Context, uri string) ([]byte, error) {
	cached, err := v.CacheProvider().Get(uri)
	if err != nil {
		return nil, err
	}

	if cached != "" {
		return []byte(cached), nil
	}

	opts, err := v.CacheProvider().GetOpts(uri)
	if err != nil {
		return nil, err
	}

	// KVv2 mount path. Default "secret"
	mount := os.Getenv("VAULT_PATH_PREFIX")
	if mount == "" {
		mount = "secret"
	}

	// read the secret
	s, err := vaultClient.KVv2(mount).Get(ctx, opts.secretPath)
	if err != nil {
		return nil, err
	}

	// store in cache
	err = v.CacheProvider().Put(uri, s.Data)
	if err != nil {
		return nil, err
	}

	// now lets just get from cache itself
	data, err := v.CacheProvider().Get(uri)
	if err != nil {
		return nil, err
	}

	return []byte(data), nil
}

// Load configuration from environment and create a new vault client
func vaultConfigureClient(ctx context.Context) error {
	config := vault.DefaultConfig()

	// Load configuration from environment
	err := config.ReadEnvironment()
	if err != nil {
		return err
	}

	// Create client. Auths with VAULT_TOKEN by default
	client, err := vault.NewClient(config)
	if err != nil {
		return err
	}

	// Use AppRole if provided
	roleID := os.Getenv("VAULT_APPROLE_ROLE_ID")
	if roleID != "" {
		secretID := &auth.SecretID{FromEnv: "VAULT_APPROLE_SECRET_ID"}
		// Authenticate
		appRoleAuth, err := auth.NewAppRoleAuth(
			roleID,
			secretID,
		)
		if err != nil {
			return fmt.Errorf("unable to initialize Vault AppRole auth method: %w", err)
		}

		authInfo, err := client.Auth().Login(ctx, appRoleAuth)
		if err != nil {
			return fmt.Errorf("unable to login to Vault AppRole auth method: %w", err)
		}
		if authInfo == nil {
			return fmt.Errorf("no auth info was returned after Vault AppRole login")
		}
	}

	// Set client
	vaultClient = client
	return nil
}

package secretprovider

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

var (
	secretCache map[string]dataWithTTL
)

var ErrorNotFoundInCache = errors.New("valid entry not found in secrets cache")

type dataWithTTL struct {
	data      any
	expiresAt time.Time
}

type Opts struct {
	cachekey    string
	secretPath  string
	secretField string
	ttl         time.Duration
}

type CacheProvider interface {
	Put(input string, data any) error
	Get(input string) (string, error)
	GetOpts(input string) (Opts, error)
}

// ObjectCacheProvider accepts a key of format path/to/some-secret.specific-field. It
// then caches the complete object retrieved with key `provider://path/to/some-secret`
// and when a user requests path/to/some-secret.foo or path/to/some-secret.bar,
// it only makes one request and returns the second from the cached data.
type ObjectCacheProvider struct{}

func (o ObjectCacheProvider) Put(input string, data any) error {
	opts, err := o.GetOpts(input)
	if err != nil {
		return err
	}

	// todo: add mutex
	secretCache[opts.cachekey] = dataWithTTL{
		data:      data,
		expiresAt: time.Now().Add(opts.ttl),
	}

	return nil
}

func (o ObjectCacheProvider) Get(input string) (string, error) {
	opts, err := o.GetOpts(input)
	if err != nil {
		return "", err
	}

	existing, ok := secretCache[opts.cachekey]
	if !ok || hasExpired(existing) {
		return "", ErrorNotFoundInCache
	}

	data := existing.data.(map[string]any)
	value, ok := data[opts.secretField]
	if !ok {
		return "", fmt.Errorf("field %q not found in cached secret data", opts.secretField)
	}

	return value.(string), nil
}

func (o ObjectCacheProvider) GetOpts(input string) (Opts, error) {
	parsed, err := url.Parse(input)
	if err != nil {
		return Opts{}, err
	}

	path := parsed.Path
	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return Opts{}, fmt.Errorf("no key found to be retrieved. expected path to be of format some/path/object.key")
	}

	if len(parts) > 2 {
		return Opts{}, fmt.Errorf("too many dots found in the path. expected path to be of format some/path/object.key")
	}

	opts := Opts{
		cachekey:    fmt.Sprintf("%s://%s", parsed.Scheme, parts[0]),
		secretPath:  parts[0],
		secretField: parts[1],
	}

	ttlInput := parsed.Query().Get("ttl")
	if ttlInput != "" {
		ttl, err := time.ParseDuration(ttlInput)
		if err != nil {
			return Opts{}, err
		}

		opts.ttl = ttl
	}

	return opts, nil
}

type RawCacheProvider struct{}

func (r RawCacheProvider) Put(input string, data any) error {
	opts, err := r.GetOpts(input)
	if err != nil {
		return err
	}

	secretCache[opts.cachekey] = dataWithTTL{
		data:      data,
		expiresAt: time.Now().Add(opts.ttl),
	}
	return nil
}

func (r RawCacheProvider) Get(input string) (string, error) {
	opts, err := r.GetOpts(input)
	if err != nil {
		return "", err
	}

	existing, ok := secretCache[opts.cachekey]
	if !ok || hasExpired(existing) {
		return "", ErrorNotFoundInCache
	}

	return existing.data.(string), nil
}

func (r RawCacheProvider) GetOpts(input string) (Opts, error) {
	parsed, err := url.Parse(input)
	if err != nil {
		return Opts{}, err
	}

	opts := Opts{
		cachekey:    fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Path),
		secretPath:  parsed.Path,
		secretField: "",
	}

	ttlInput := parsed.Query().Get("ttl")
	if ttlInput != "" {
		ttl, err := time.ParseDuration(ttlInput)
		if err != nil {
			return Opts{}, err
		}

		opts.ttl = ttl
	}

	return opts, nil
}

func hasExpired(data dataWithTTL) bool {
	// if no ttl set, assume no ttl required
	if data.expiresAt.IsZero() {
		return false
	}

	if data.expiresAt.After(time.Now()) {
		return false
	}

	return true
}

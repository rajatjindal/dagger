package core

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/opencontainers/go-digest"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/dagger/dagger/dagql"
	"github.com/dagger/dagger/dagql/call"
)

// CacheVolume is a persistent volume with a globally scoped identifier.
type CacheVolume struct {
	Keys []string `json:"keys"`

	Query *Query

	// The digest of the DagQL ID that accessed this cache volume, used as its identifier
	// in cache volume store.
	IDDigest digest.Digest
}

func (*CacheVolume) Type() *ast.Type {
	return &ast.Type{
		NamedType: "CacheVolume",
		NonNull:   true,
	}
}

func (*CacheVolume) TypeDescription() string {
	return "A directory whose contents persist across runs."
}

func (cache *CacheVolume) LLBID() string {
	return string(cache.IDDigest)
}

func NewCache(keys ...string) *CacheVolume {
	return &CacheVolume{Keys: keys}
}

func (cache *CacheVolume) Clone() *CacheVolume {
	cp := *cache
	cp.Keys = cloneSlice(cp.Keys)

	return &cp
}

// Sum returns a checksum of the cache tokens suitable for use as a cache key.
func (cache *CacheVolume) Sum() string {
	hash := sha256.New()
	for _, tok := range cache.Keys {
		_, _ = hash.Write([]byte(tok + "\x00"))
	}

	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

type CacheSharingMode string

var CacheSharingModes = dagql.NewEnum[CacheSharingMode]()

var (
	CacheSharingModeShared = CacheSharingModes.Register("SHARED",
		"Shares the cache volume amongst many build pipelines")
	CacheSharingModePrivate = CacheSharingModes.Register("PRIVATE",
		"Keeps a cache volume for a single build pipeline")
	CacheSharingModeLocked = CacheSharingModes.Register("LOCKED",
		"Shares the cache volume amongst many build pipelines, but will serialize the writes")
)

func (mode CacheSharingMode) Type() *ast.Type {
	return &ast.Type{
		NamedType: "CacheSharingMode",
		NonNull:   true,
	}
}

func (mode CacheSharingMode) TypeDescription() string {
	return "Sharing mode of the cache volume."
}

func (mode CacheSharingMode) Decoder() dagql.InputDecoder {
	return CacheSharingModes
}

func (mode CacheSharingMode) ToLiteral() call.Literal {
	return CacheSharingModes.Literal(mode)
}

// CacheSharingMode marshals to its lowercased value.
//
// NB: as far as I can recall this is purely for ~*aesthetic*~. GraphQL consts
// are so shouty!
func (mode CacheSharingMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(strings.ToLower(string(mode)))
}

// CacheSharingMode marshals to its lowercased value.
//
// NB: as far as I can recall this is purely for ~*aesthetic*~. GraphQL consts
// are so shouty!
func (mode *CacheSharingMode) UnmarshalJSON(payload []byte) error {
	var str string
	if err := json.Unmarshal(payload, &str); err != nil {
		return err
	}

	*mode = CacheSharingMode(strings.ToUpper(str))

	return nil
}

type CacheVolumeStore struct {
	cacheVolumes map[digest.Digest]*storedCacheVolume
	mu           sync.RWMutex
}

type storedCacheVolume struct {
	CacheVolume *CacheVolume
	Name        string
}

func NewCacheVolumeStore() *CacheVolumeStore {
	return &CacheVolumeStore{
		cacheVolumes: map[digest.Digest]*storedCacheVolume{},
	}
}

func (store *CacheVolumeStore) AddCacheVolume(cacheVolume *CacheVolume, name string) error {
	if cacheVolume == nil {
		return fmt.Errorf("cacheVolume must not be nil")
	}
	if cacheVolume.Query == nil {
		return fmt.Errorf("cacheVolume must have a query")
	}
	if cacheVolume.IDDigest == "" {
		return fmt.Errorf("cacheVolume must have an ID digest")
	}

	store.mu.Lock()
	defer store.mu.Unlock()
	store.cacheVolumes[cacheVolume.IDDigest] = &storedCacheVolume{
		CacheVolume: cacheVolume,
		Name:        name,
	}
	return nil
}

func (store *CacheVolumeStore) AddCacheVolumeFromOtherStore(cacheVolume *CacheVolume, otherStore *CacheVolumeStore) error {
	otherStore.mu.RLock()
	cacheVolumeVals, ok := otherStore.cacheVolumes[cacheVolume.IDDigest]
	otherStore.mu.RUnlock()
	if !ok {
		return fmt.Errorf("secret %s not found in other store", cacheVolume.IDDigest)
	}
	//TODO(rajatjindal): if volume is marked as PRIVATE, return error?
	return store.AddCacheVolume(cacheVolume, cacheVolumeVals.Name)
}

func (store *CacheVolumeStore) HasSecret(idDgst digest.Digest) bool {
	store.mu.RLock()
	defer store.mu.RUnlock()
	_, ok := store.cacheVolumes[idDgst]
	return ok
}

func (store *CacheVolumeStore) GetCacheName(idDgst digest.Digest) (string, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	cacheVolume, ok := store.cacheVolumes[idDgst]
	if !ok {
		return "", false
	}
	return cacheVolume.Name, true
}

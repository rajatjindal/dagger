package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/moby/buildkit/identity"
	"github.com/stretchr/testify/require"

	"dagger.io/dagger"
	"github.com/dagger/dagger/testctx"
)

type CacheSuite struct{}

func TestCache(t *testing.T) {
	testctx.Run(testCtx, t, CacheSuite{}, Middleware()...)
}

func (CacheSuite) TestVolume(ctx context.Context, t *testctx.T) {
	c := connect(ctx, t)

	volID1, err := c.CacheVolume("ab").ID(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, volID1)

	volID2, err := c.CacheVolume("ab").ID(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, volID2)

	volID3, err := c.CacheVolume("ac").ID(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, volID3)

	require.Equal(t, volID1, volID2)
	require.NotEqual(t, volID1, volID3)
}

func (CacheSuite) TestVolumeWithSubmount(ctx context.Context, t *testctx.T) {
	c := connect(ctx, t)

	t.Run("file mount", func(ctx context.Context, t *testctx.T) {
		subfile := c.Directory().WithNewFile("foo", "bar").File("foo")
		ctr := c.Container().From(alpineImage).
			WithMountedCache("/cache", c.CacheVolume(identity.NewID())).
			WithMountedFile("/cache/subfile", subfile)

		out, err := ctr.WithExec([]string{"cat", "/cache/subfile"}).Stdout(ctx)
		require.NoError(t, err)
		require.Equal(t, "bar", strings.TrimSpace(out))

		contents, err := ctr.File("/cache/subfile").Contents(ctx)
		require.NoError(t, err)
		require.Equal(t, "bar", strings.TrimSpace(contents))
	})

	t.Run("dir mount", func(ctx context.Context, t *testctx.T) {
		subdir := c.Directory().WithNewFile("foo", "bar").WithNewFile("baz", "qux")
		ctr := c.Container().From(alpineImage).
			WithMountedCache("/cache", c.CacheVolume(identity.NewID())).
			WithMountedDirectory("/cache/subdir", subdir)

		for fileName, expectedContents := range map[string]string{
			"foo": "bar",
			"baz": "qux",
		} {
			subpath := filepath.Join("/cache/subdir", fileName)
			out, err := ctr.WithExec([]string{"cat", subpath}).Stdout(ctx)
			require.NoError(t, err)
			require.Equal(t, expectedContents, strings.TrimSpace(out))

			contents, err := ctr.File(subpath).Contents(ctx)
			require.NoError(t, err)
			require.Equal(t, expectedContents, strings.TrimSpace(contents))

			dir := ctr.Directory("/cache/subdir")
			contents, err = dir.File(fileName).Contents(ctx)
			require.NoError(t, err)
			require.Equal(t, expectedContents, strings.TrimSpace(contents))
		}
	})
}

func (CacheSuite) TestLocalImportCacheReuse(ctx context.Context, t *testctx.T) {
	hostDirPath := t.TempDir()
	err := os.WriteFile(filepath.Join(hostDirPath, "foo"), []byte("bar"), 0o644)
	require.NoError(t, err)

	runExec := func(c *dagger.Client) string {
		out, err := c.Container().From(alpineImage).
			WithDirectory("/fromhost", c.Host().Directory(hostDirPath)).
			WithExec([]string{"stat", "/fromhost/foo"}).
			WithExec([]string{"sh", "-c", "head -c 128 /dev/random | sha256sum"}).
			Stdout(ctx)
		require.NoError(t, err)
		return out
	}

	c1 := connect(ctx, t)
	out1 := runExec(c1)

	c2 := connect(ctx, t)
	out2 := runExec(c2)

	require.Equal(t, out1, out2)
}

// 1. in same session, same module, same key, same id
// 2. in same session, diff module, same key, diff id
// 3. in diff session, same module, same key, diff id
// 4. in diff session, diff module, same key, diff id
// 5. cache referenced by id, can be reused across modules
func (CacheSuite) TestCacheIsNamespaced(ctx context.Context, t *testctx.T) {
	c := connect(ctx, t)

	fooTmpl := `package main
	import (
		"context"
	)
	
	type Foo struct {}

	func (f *Foo) GetCacheVolumeId(ctx context.Context) (string, error) {
		id, err := dag.CacheVolume("volume-name").ID(ctx)
		return "foo -> " + string(id), err
	}
	`

	barTmpl := `package main
	import (
		"context"
	)
	
	type Bar struct {}

	func (b *Bar) GetCacheVolumeId(ctx context.Context) (string, error) {
		id, err := dag.CacheVolume("volume-name").ID(ctx)
		return "bar -> " + string(id), err
	}
	`

	ctrbar := c.Container().
		From(golangImage).
		WithMountedFile(testCLIBinPath, daggerCliFile(t, c)).
		WithWorkdir("/work/bar").
		With(daggerExec("init", "--name=bar", "--source=.", "--sdk=go")).
		WithNewFile("main.go", barTmpl)

	c2 := connect(ctx, t)

	ctrfoo := c2.Container().
		From(golangImage).
		WithMountedFile(testCLIBinPath, daggerCliFile(t, c2)).
		WithWorkdir("/work/foo").
		With(daggerExec("init", "--name=foo", "--source=.", "--sdk=go")).
		WithNewFile("main.go", fooTmpl)

	fooId, err := ctrbar.
		WithWorkdir("/work/bar").
		With(daggerExec("call", "get-cache-volume-id")).
		Stdout(ctx)

	require.NoError(t, err)

	barId, err := ctrfoo.
		WithWorkdir("/work/foo").
		With(daggerExec("call", "get-cache-volume-id")).
		Stdout(ctx)
	require.NoError(t, err)

	require.Equal(t, fooId, barId)
}

// func (CacheSuite) TestCachePersistWithinSession(ctx context.Context, t *testctx.T) {
// 	c := connect(ctx, t)

// 	tmpl := `package main
// 	import (
// 		"context"
// 	)

// 	type Foo struct {}

// 	func (f *Foo) TestCacheVolumePersistenceInSession(ctx context.Context) *dagger.Container {
// 		return dag.
// 			Container().
// 			From("alpine:latest").
// 			WithMountedCache("/foo-cache", dag.CacheVolume("volume-name")).
// 			WithNewFile("/foo-cache/bar.txt", "/foo-cache/bar.txt file contents")
// 	}

// 	func (f *Foo) GetCacheVolumeId(ctx context.Context) (string, error) {
// 		entries, err := dag.
// 			Container().
// 			From("alpine:latest").
// 			WithMountedCache("/foo-cache", dag.CacheVolume("volume-name")).
// 			WithNewFile("/foo-cache/bar.txt", "/foo-cache/bar.txt file contents").
// 			Entries(ctx)

// 		entries2, err := dag.
// 			Container().
// 			From("alpine:latest").
// 			WithMountedCache("/foo-cache", dag.CacheVolume("volume-name")).
// 			Directory("/foo-cache").
// 			Entries(ctx)
// 	}
// 	`

// 	ctr := c.Container().
// 		From(golangImage).
// 		WithMountedFile(testCLIBinPath, daggerCliFile(t, c)).
// 		WithWorkdir("/work").
// 		With(daggerExec("init", "--name=foo", "--source=.", "--sdk=go")).
// 		WithNewFile("main.go", tmpl).

// 	fooId, err := ctr.
// 		With(daggerExec("call", "get-cache-volume-id")).
// 		Stdout(ctx)
// 	require.NoError(t, err)

// 	barId, err := ctr.
// 		With(daggerExec("call", "get-bar-cache-volume-id")).
// 		Stdout(ctx)
// 	require.NoError(t, err)

// 	require.Equal(t, fooId, barId+"2")
// }

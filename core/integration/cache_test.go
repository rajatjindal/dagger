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

// func (CacheSuite) TestCacheIsNamespaced(ctx context.Context, t *testctx.T) {
// 	c := connect(ctx, t)

// 	fooTmpl := `package main
// 	import (
// 		"context"
// 	)

// 	type Foo struct {}
// 	func (f *Foo) GetCacheVolumeId(ctx context.Context) (string, error) {
// 		id, err := dag.CacheVolume("volume-name").ID(ctx)
// 		return string(id), err
// 	}
// 	`
// 	barTmpl := `package main
// 	import (
// 		"context"
// 	)

// 	type Bar struct {}
// 	func (b *Bar) GetCacheVolumeId(ctx context.Context) (string, error) {
// 		id, err := dag.CacheVolume("volume-name").ID(ctx)
// 		return string(id), err
// 	}
// 	`
// 	ctr := c.Container().
// 		From(golangImage).
// 		WithMountedFile(testCLIBinPath, daggerCliFile(t, c)).
// 		WithWorkdir("/work/bar").
// 		With(daggerExec("init", "--name=bar", "--source=.", "--sdk=go")).
// 		WithNewFile("main.go", barTmpl).
// 		WithWorkdir("/work/foo").
// 		With(daggerExec("init", "--name=foo", "--source=.", "--sdk=go")).
// 		WithNewFile("main.go", fooTmpl)

// 	fooId, err := ctr.
// 		WithWorkdir("/work/foo").
// 		With(daggerExec("call", "get-cache-volume-id")).
// 		Stdout(ctx)
// 	require.NoError(t, err)

// 	barId, err := ctr.
// 		WithWorkdir("/work/bar").
// 		With(daggerExec("call", "get-cache-volume-id")).
// 		Stdout(ctx)
// 	require.NoError(t, err)
// 	require.NotEqual(t, fooId, barId)
// }

// func (CacheSuite) TestCacheVolumeCanBePassedById(ctx context.Context, t *testctx.T) {
// 	c := connect(ctx, t)

// 	fooTmpl := `package main

// import (
// 	"context"
// )

// type Foo struct{}

// func (f *Foo) GetCacheVolumeId(ctx context.Context) (string, error) {
// 	cacheVolume := dag.CacheVolume("volume-name2")
// 	_, err := dag.Container().
// 		From("alpine:latest").
// 		WithMountedCache("/foo", cacheVolume).
// 		WithExec([]string{"sh", "-c", "echo -n 'hello foo' > /foo/bar.txt"}).
// 		Sync(ctx)
// 	if err != nil {
// 		return "", err
// 	}

// 	id, err := cacheVolume.ID(ctx)
// 	return string(id), err
// }

// func (f *Foo) UseCacheVolumeAcrossModuleUsingId(ctx context.Context, id string) (string, error) {
// 	return dag.Bar().UseCacheVolumeID(ctx, id)
// }

// func (f *Foo) UseCacheVolumeAcrossModuleUsingName(ctx context.Context, name string) (string, error) {
// 	return dag.Bar().UseCacheVolumeByName(ctx, name)
// }
// 	`

// 	barTmpl := `package main

// import (
// 	"context"
// 	"dagger/bar/internal/dagger"
// )

// type Bar struct{}

// func (f *Bar) UseCacheVolumeId(ctx context.Context, id string) (string, error) {
// 	return dag.Container().
// 		From("alpine:latest").
// 		WithMountedCache("/bar", dag.LoadCacheVolumeFromID(dagger.CacheVolumeID(id))).
// 		WithExec([]string{"sh", "-c", "ls /bar/bar.txt"}).
// 		Stdout(ctx)
// }

// func (f *Bar) UseCacheVolumeByName(ctx context.Context, name string) (string, error) {
// 	return dag.Container().
// 		From("alpine:latest").
// 		WithMountedCache("/bar", dag.CacheVolume(name)).
// 		WithExec([]string{"sh", "-c", "ls /bar/bar.txt"}).
// 		Stdout(ctx)
// }
// `

// 	ctr := c.Container().
// 		From(golangImage).
// 		WithMountedFile(testCLIBinPath, daggerCliFile(t, c)).
// 		WithWorkdir("/work/bar").
// 		With(daggerExec("init", "--name=bar", "--source=.", "--sdk=go")).
// 		WithNewFile("main.go", barTmpl).
// 		WithWorkdir("/work").
// 		With(daggerExec("init", "--name=foo", "--source=.", "--sdk=go")).
// 		WithNewFile("main.go", fooTmpl).
// 		With(daggerExec("use", "./bar"))

// 	fooCacheVolId, err := ctr.
// 		WithWorkdir("/work").
// 		With(daggerExec("call", "get-cache-volume-id")).
// 		Stdout(ctx)
// 	require.NoError(t, err)

// 	usingIdOutput, err := ctr.
// 		WithWorkdir("/work").
// 		With(daggerExec("call", "use-cache-volume-across-module-using-id", fmt.Sprintf("--id=%s", fooCacheVolId))).
// 		Stdout(ctx)
// 	require.NoError(t, err)
// 	require.Equal(t, "/bar/bar.txt\n", usingIdOutput)

// 	usingNameOutput, err := ctr.
// 		WithWorkdir("/work").
// 		With(daggerExec("call", "use-cache-volume-across-module-using-name", "--name=volume-name2")).
// 		Stdout(ctx)
// 	require.NoError(t, err)
// 	require.Equal(t, "/bar/bar.txt\n", usingNameOutput)

// 	//verify bar module cannot directly consume cache volume by name
// 	usingNameOutputDiffModule, err := dag.Container().
// 		From("alpine:latest").
// 		WithMountedFile(testCLIBinPath, daggerCliFile(t, c)).
// 		WithWorkdir("/work/bar").
// 		With(daggerExec("init", "--name=bar", "--source=.", "--sdk=go")).
// 		WithNewFile("main.go", barTmpl).
// 		With(daggerExec("call", "use-cache-volume-by-name", "--name=volume-name2")).
// 		Stdout(ctx)

// 	require.NoError(t, err)
// 	require.Equal(t, "/bar/bar2.txt\n", usingNameOutputDiffModule)
// }

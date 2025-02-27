package core

import (
	"context"
	"testing"

	"github.com/dagger/testctx"
	"github.com/stretchr/testify/require"
)

type PrivateDepsSuite struct{}

func TestPrivateDeps(t *testing.T) {
	testctx.New(t, Middleware()...).RunTests(PrivateDepsSuite{})
}

func (PrivateDepsSuite) TestPrivateDeps(ctx context.Context, t *testctx.T) {
	t.Run("golang", func(ctx context.Context, t *testctx.T) {
		privateDepCode := `package main

import (
        "context"
        "dagger/foo/internal/dagger"

        "github.com/rajatjindal/bkpapi/backend/pkg/api"
)

type Foo struct{}

// Returns a container that echoes whatever string argument is provided
func (m *Foo) ContainerEcho(stringArg string) *dagger.Container {
        // this forces the private dep
        _ = api.Server{}
        return dag.Container().From("alpine:latest").WithExec([]string{stringArg, stringArg})
}`

		daggerjson := `{
  "name": "foo",
  "engineVersion": "v0.16.2",
  "sdk": {
    "source": "go",
    "config": {
      "goprivate": "github.com/rajatjindal"
    }
  }
}`

		c := connect(ctx, t)
		sockPath, cleanup := setupPrivateRepoSSHAgent(t)
		defer cleanup()

		socket := c.Host().UnixSocket(sockPath)

		modGen := c.Container().From(golangImage).
			WithMountedFile(testCLIBinPath, daggerCliFile(t, c)).
			WithExec([]string{"apk", "add", "git", "openssh"}).
			WithUnixSocket("/sock/unix-socket", socket).
			WithEnvVariable("SSH_AUTH_SOCK", "/sock/unix-socket").
			WithWorkdir("/work").
			WithNewFile("/root/.gitconfig", `
[url "ssh://git@github.com/"]
	insteadOf = https://github.com/
`).
			With(daggerExec("init", "--name=foo", "--sdk=go")).
			WithNewFile("main.go", privateDepCode).
			WithNewFile("dagger.json", daggerjson)

		_, err := modGen.
			With(daggerExec("develop", "-vvv")).
			Stdout(ctx)
		require.NoError(t, err)
	})
}

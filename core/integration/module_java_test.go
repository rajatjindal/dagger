package core

import (
	"path/filepath"
	"testing"

	"dagger.io/dagger"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"

	"github.com/dagger/testctx"
)

type JavaSuite struct{}

func TestJava(t *testing.T) {
	testctx.New(t, Middleware()...).RunTests(JavaSuite{})
}

func (JavaSuite) TestInit(_ context.Context, t *testctx.T) {
	t.Run("from upstream", func(ctx context.Context, t *testctx.T) {
		c := connect(ctx, t)

		modGen := c.Container().From(golangImage).
			WithMountedFile(testCLIBinPath, daggerCliFile(t, c)).
			WithWorkdir("/work").
			With(daggerExec("init", "--name=bare", "--sdk=github.com/dagger/dagger/sdk/java"))

		out, err := modGen.
			With(daggerQuery(`{bare{containerEcho(stringArg:"hello"){stdout}}}`)).
			Stdout(ctx)
		require.NoError(t, err)
		require.JSONEq(t, `{"bare":{"containerEcho":{"stdout":"hello\n"}}}`, out)
	})

	t.Run("from alias", func(ctx context.Context, t *testctx.T) {
		c := connect(ctx, t)

		modGen := c.Container().From(golangImage).
			WithMountedFile(testCLIBinPath, daggerCliFile(t, c)).
			WithWorkdir("/work").
			With(daggerExec("init", "--name=bare", "--sdk=java"))

		out, err := modGen.
			With(daggerQuery(`{bare{containerEcho(stringArg:"hello"){stdout}}}`)).
			Stdout(ctx)
		require.NoError(t, err)
		require.JSONEq(t, `{"bare":{"containerEcho":{"stdout":"hello\n"}}}`, out)
	})

	t.Run("from alias with ref", func(ctx context.Context, t *testctx.T) {
		c := connect(ctx, t)

		modGen := c.Container().From(golangImage).
			WithMountedFile(testCLIBinPath, daggerCliFile(t, c)).
			WithWorkdir("/work").
			With(daggerExec("init", "--name=bare", "--sdk=java@main"))

		out, err := modGen.
			With(daggerQuery(`{bare{containerEcho(stringArg:"hello"){stdout}}}`)).
			Stdout(ctx)
		require.NoError(t, err)
		require.JSONEq(t, `{"bare":{"containerEcho":{"stdout":"hello\n"}}}`, out)
	})
}

func (JavaSuite) TestFields(_ context.Context, t *testctx.T) {
	t.Run("can set and retrieve field using custom function", func(ctx context.Context, t *testctx.T) {
		c := connect(ctx, t)

		out, err := javaModule(t, c, "fields").
			With(daggerShell("with-version a.b.c | get-version")).
			Stdout(ctx)

		require.NoError(t, err)
		require.Contains(t, out, "a.b.c")
	})

	t.Run("can set and retrieve field using direct access to the field when decorated", func(ctx context.Context, t *testctx.T) {
		c := connect(ctx, t)

		out, err := javaModule(t, c, "fields").
			With(daggerShell("with-version a.b.c | version")).
			Stdout(ctx)

		require.NoError(t, err)
		require.Contains(t, out, "a.b.c")
	})

	t.Run("can set and retrieve public field using direct access to the field", func(ctx context.Context, t *testctx.T) {
		c := connect(ctx, t)

		out, err := javaModule(t, c, "fields").
			With(daggerShell("with-version a.b.c | public-version")).
			Stdout(ctx)

		require.NoError(t, err)
		require.Contains(t, out, "a.b.c")
	})

	t.Run("can set and retrieve non exposed field using custom function", func(ctx context.Context, t *testctx.T) {
		c := connect(ctx, t)

		out, err := javaModule(t, c, "fields").
			With(daggerShell("with-version a.b.c | get-internal-version")).
			Stdout(ctx)

		require.NoError(t, err)
		require.Contains(t, out, "a.b.c")
	})

	t.Run("can set but not retrieve non exposed field using direct access to the field", func(ctx context.Context, t *testctx.T) {
		c := connect(ctx, t)

		_, err := javaModule(t, c, "fields").
			With(daggerShell("with-version a.b.c | internal-version")).
			Stdout(ctx)

		require.Error(t, err)
	})
}

func (JavaSuite) TestDefaultValue(_ context.Context, t *testctx.T) {
	t.Run("can set a value", func(ctx context.Context, t *testctx.T) {
		c := connect(ctx, t)

		out, err := javaModule(t, c, "defaults").
			With(daggerCall("echo", "--value=hello")).
			Stdout(ctx)

		require.NoError(t, err)
		require.Equal(t, "hello", out)
	})

	t.Run("can use a default value", func(ctx context.Context, t *testctx.T) {
		c := connect(ctx, t)

		out, err := javaModule(t, c, "defaults").
			With(daggerCall("echo")).
			Stdout(ctx)

		require.NoError(t, err)
		require.Equal(t, "default value", out)
	})
}

func (JavaSuite) TestOptionalValue(_ context.Context, t *testctx.T) {
	t.Run("can run without a value", func(ctx context.Context, t *testctx.T) {
		c := connect(ctx, t)

		out, err := javaModule(t, c, "defaults").
			With(daggerCall("echo-else")).
			Stdout(ctx)

		require.NoError(t, err)
		require.Equal(t, "default value if null", out)
	})

	t.Run("can set a value", func(ctx context.Context, t *testctx.T) {
		c := connect(ctx, t)

		out, err := javaModule(t, c, "defaults").
			With(daggerCall("echo-else", "--value", "foo")).
			Stdout(ctx)

		require.NoError(t, err)
		require.Equal(t, "foo", out)
	})

	t.Run("ensure Optional and @Default work together", func(ctx context.Context, t *testctx.T) {
		c := connect(ctx, t)

		out, err := javaModule(t, c, "defaults").
			With(daggerCall("echo-opt-default")).
			Stdout(ctx)

		require.NoError(t, err)
		require.Equal(t, "default value", out)
	})
}

func (JavaSuite) TestDefaultPath(_ context.Context, t *testctx.T) {
	t.Run("can set a path for a file", func(ctx context.Context, t *testctx.T) {
		c := connect(ctx, t)

		out, err := javaModule(t, c, "defaults").
			With(daggerCall("file-name", "--file=./pom.xml")).
			Stdout(ctx)

		require.NoError(t, err)
		require.Equal(t, "pom.xml", out)
	})

	t.Run("can use a default path for a file", func(ctx context.Context, t *testctx.T) {
		c := connect(ctx, t)

		out, err := javaModule(t, c, "defaults").
			With(daggerCall("file-name")).
			Stdout(ctx)

		require.NoError(t, err)
		require.Equal(t, "dagger.json", out)
	})

	t.Run("can set a path for a dir", func(ctx context.Context, t *testctx.T) {
		c := connect(ctx, t)

		out, err := javaModule(t, c, "defaults").
			With(daggerCall("file-names", "--dir", ".")).
			Stdout(ctx)

		require.NoError(t, err)
		require.Contains(t, out, "pom.xml")
	})

	t.Run("can use a default path for a dir", func(ctx context.Context, t *testctx.T) {
		c := connect(ctx, t)

		out, err := javaModule(t, c, "defaults").
			With(daggerCall("file-names")).
			Stdout(ctx)

		require.NoError(t, err)
		require.Equal(t, "Defaults.java", out)
	})
}

func (JavaSuite) TestIgnore(_ context.Context, t *testctx.T) {
	t.Run("without ignore", func(ctx context.Context, t *testctx.T) {
		c := connect(ctx, t)

		out, err := javaModule(t, c, "defaults").
			With(daggerCall("files-no-ignore")).
			Stdout(ctx)

		require.NoError(t, err)
		require.Contains(t, out, "dagger.json")
		require.Contains(t, out, "pom.xml")
	})

	t.Run("with ignore", func(ctx context.Context, t *testctx.T) {
		c := connect(ctx, t)

		out, err := javaModule(t, c, "defaults").
			With(daggerCall("files-ignore")).
			Stdout(ctx)

		require.NoError(t, err)
		require.Contains(t, out, "dagger.json")
		require.NotContains(t, out, "pom.xml")
	})

	t.Run("with negated ignore", func(ctx context.Context, t *testctx.T) {
		c := connect(ctx, t)

		out, err := javaModule(t, c, "defaults").
			With(daggerCall("files-neg-ignore")).
			Stdout(ctx)

		require.NoError(t, err)
		require.NotContains(t, out, "dagger.json")
		require.NotContains(t, out, "pom.xml")
		require.Contains(t, out, "src")
	})
}

func (JavaSuite) TestConstructor(_ context.Context, t *testctx.T) {
	t.Run("value set", func(ctx context.Context, t *testctx.T) {
		c := connect(ctx, t)

		out, err := javaModule(t, c, "construct").
			With(daggerCall("--value", "from cli", "echo")).
			Stdout(ctx)

		require.NoError(t, err)
		require.Equal(t, "from cli", out)
	})

	t.Run("default value from constructor", func(ctx context.Context, t *testctx.T) {
		c := connect(ctx, t)

		out, err := javaModule(t, c, "construct").
			With(daggerCall("echo")).
			Stdout(ctx)

		require.NoError(t, err)
		require.Equal(t, "from constructor", out)
	})
}

func (JavaSuite) TestEnum(_ context.Context, t *testctx.T) {
	t.Run("can use an enum value", func(ctx context.Context, t *testctx.T) {
		c := connect(ctx, t)

		out, err := javaModule(t, c, "enums").
			With(daggerCall("print", "--severity=LOW")).
			Stdout(ctx)

		require.NoError(t, err)
		require.Equal(t, "LOW", out)
	})

	t.Run("can not use a value not defined in the enum", func(ctx context.Context, t *testctx.T) {
		c := connect(ctx, t)

		_, err := javaModule(t, c, "enums").
			With(daggerCall("print", "--severity=FOO")).
			Stdout(ctx)

		require.Error(t, err)
		requireErrOut(t, err, "value should be one of LOW,MEDIUM,HIGH")
	})

	t.Run("can return an enum value", func(ctx context.Context, t *testctx.T) {
		c := connect(ctx, t)

		out, err := javaModule(t, c, "enums").
			With(daggerCall("from-string", "--severity=MEDIUM")).
			Stdout(ctx)

		require.NoError(t, err)
		require.Equal(t, "MEDIUM", out)
	})

	t.Run("can return a list of enum values", func(ctx context.Context, t *testctx.T) {
		c := connect(ctx, t)

		out, err := javaModule(t, c, "enums").
			With(daggerCall("get-severities-list")).
			Stdout(ctx)

		require.NoError(t, err)
		require.Equal(t, "LOW\nMEDIUM\nHIGH\n", out)
	})

	t.Run("can return an array of enum values", func(ctx context.Context, t *testctx.T) {
		c := connect(ctx, t)

		out, err := javaModule(t, c, "enums").
			With(daggerCall("get-severities-array")).
			Stdout(ctx)

		require.NoError(t, err)
		require.Equal(t, "LOW\nMEDIUM\nHIGH\n", out)
	})

	t.Run("can read list of enum values", func(ctx context.Context, t *testctx.T) {
		c := connect(ctx, t)

		out, err := javaModule(t, c, "enums").
			With(daggerCall("list-to-string", "--severities=MEDIUM,LOW")).
			Stdout(ctx)

		require.NoError(t, err)
		require.Equal(t, "MEDIUM,LOW", out)
	})

	t.Run("can read array of enum values", func(ctx context.Context, t *testctx.T) {
		c := connect(ctx, t)

		out, err := javaModule(t, c, "enums").
			With(daggerCall("array-to-string", "--severities=HIGH,LOW")).
			Stdout(ctx)

		require.NoError(t, err)
		require.Equal(t, "HIGH,LOW", out)
	})
}

func javaModule(t *testctx.T, c *dagger.Client, moduleName string) *dagger.Container {
	t.Helper()
	modSrc, err := filepath.Abs(filepath.Join("./testdata/modules/java", moduleName))
	require.NoError(t, err)

	sdkSrc, err := filepath.Abs("../../sdk/java")
	require.NoError(t, err)

	return goGitBase(t, c).
		WithDirectory("modules/"+moduleName, c.Host().Directory(modSrc)).
		WithDirectory("sdk/java", c.Host().Directory(sdkSrc)).
		WithWorkdir("/work/modules/" + moduleName)
}

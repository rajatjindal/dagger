package impl

// import (
// 	"context"
// 	"fmt"
// 	"os"

// 	"github.com/dagger/dagger/dagql/introspection"
// 	"github.com/dagger/dagger/testctx"
// 	"github.com/rajatjindal/daggerverse/compat/internal/dagger"
// )

// // getSchemaForModuleForEngineVersion for given module using specified engine version
// // if engineVersion == "dev", use the engine built during the integration tests
// func getSchemaForModuleForEngineVersion(ctx context.Context, t *testctx.T, c *dagger.Client, module, engineVersion string) (string, error) {
// 	var engineSvc *dagger.Service
// 	var client *dagger.Container
// 	var err error

// 	if engineVersion == "dev" {
// 		engineSvc = devEngineContainer(c).AsService()
// 		client, err = engineClientContainer(ctx, t, c, engineSvc)
// 		require.NoError(t, err)
// 	} else {
// 		engineSvc = devEngineContainerWithVersion(c, engineVersion).AsService()
// 		client, err = engineClientContainerWithVersion(ctx, c, engineSvc, engineVersion)
// 		require.NoError(t, err)
// 	}

// 	return client.WithNewFile("/schema-query.graphql", introspection.Query).
// 		WithExec([]string{"dagger", "query", "-m", module, "--doc", "/schema-query.graphql"}).
// 		Stdout(ctx)
// }

// // devEngineContainer returns a nested dev engine.
// func devEngineContainer(c *dagger.Client, withs ...func(*dagger.Container) *dagger.Container) *dagger.Container {
// 	// This loads the engine.tar file from the host into the container, that
// 	// was set up by the test caller. This is used to spin up additional dev
// 	// engines.
// 	var tarPath string
// 	if v, ok := os.LookupEnv("_DAGGER_TESTS_ENGINE_TAR"); ok {
// 		tarPath = v
// 	} else {
// 		tarPath = "./bin/engine.tar"
// 	}
// 	devEngineTar := c.Host().File(tarPath)

// 	ctr := c.Container().Import(devEngineTar)
// 	for _, with := range withs {
// 		ctr = with(ctr)
// 	}

// 	deviceName, cidr := testutil.GetUniqueNestedEngineNetwork()
// 	return ctr.
// 		WithMountedCache("/var/lib/dagger", c.CacheVolume("dagger-dev-engine-state-"+identity.NewID())).
// 		WithExposedPort(1234, dagger.ContainerWithExposedPortOpts{Protocol: dagger.Tcp}).
// 		WithExec([]string{
// 			"--addr", "tcp://0.0.0.0:1234",
// 			"--addr", "unix:///var/run/buildkit/buildkitd.sock",
// 			// avoid network conflicts with other tests
// 			"--network-name", deviceName,
// 			"--network-cidr", cidr,
// 		}, dagger.ContainerWithExecOpts{
// 			UseEntrypoint:            true,
// 			InsecureRootCapabilities: true,
// 		})
// }

// // engineClientContainerWithVersion returns a container with specific version of dagger cli
// // and connected to the given devEngine service
// func engineClientContainerWithVersion(ctx context.Context, c *dagger.Client, devEngine *dagger.Service, version string) (*dagger.Container, error) {
// 	endpoint, err := devEngine.Endpoint(ctx, dagger.ServiceEndpointOpts{Port: 1234, Scheme: "tcp"})
// 	if err != nil {
// 		return nil, err
// 	}
// 	return c.Container().From(fmt.Sprintf("ghcr.io/dagger/engine:%s", version)).
// 		WithServiceBinding("dev-engine", devEngine).
// 		WithEnvVariable("_EXPERIMENTAL_DAGGER_RUNNER_HOST", endpoint), nil
// }

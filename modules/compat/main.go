package main

import (
	"context"
	"dagger/compat/internal/dagger"
	"dagger/compat/pkg/impl"
	"fmt"

	"github.com/dagger/dagger/dagql/introspection"
	"github.com/moby/buildkit/identity"
	"github.com/tidwall/gjson"
	"github.com/wI2L/jsondiff"
)

type Compat struct{}

// Returns a container that echoes whatever string argument is provided
func (m *Compat) ContainerEcho(stringArg string) *dagger.Container {
	return dag.Container().From("alpine:latest").WithExec([]string{"echo", stringArg})
}

// Returns lines that match a pattern in the files of the provided Directory
func (m *Compat) GrepDir(ctx context.Context, directoryArg *dagger.Directory, pattern string) (string, error) {
	return dag.Container().
		From("alpine:latest").
		WithMountedDirectory("/mnt", directoryArg).
		WithWorkdir("/mnt").
		WithExec([]string{"grep", "-R", pattern, "."}).
		Stdout(ctx)
}

func (m *Compat) Check(ctx context.Context, module, versionA, versionB string) error {
	schemaA, err := getSchemaForModuleForEngineVersion(ctx, module, versionA)
	if err != nil {
		return err
	}

	schemaB, err := getSchemaForModuleForEngineVersion(ctx, module, versionB)
	if err != nil {
		return err
	}

	patch, err := jsondiff.CompareJSON([]byte(schemaA), []byte(schemaB))
	if err != nil {
		return err
	}

	fmt.Println(patch.String())
	return nil
}

func getSchemaForModuleForEngineVersion(ctx context.Context, module, engineVersion string) (string, error) {
	var engineSvc *dagger.Service
	var client *dagger.Container
	var err error

	engineSvc = engineServiceWithVersion(engineVersion)
	client, err = engineClientContainerWithVersion(ctx, engineSvc, engineVersion)
	if err != nil {
		return "", err
	}

	rawIntrospection, err := client.WithNewFile("/schema-query.graphql", introspection.Query).
		WithExec([]string{"dagger", "query", "-m", module, "--doc", "/schema-query.graphql"}).
		Stdout(ctx)

	if err != nil {
		return "", err
	}

	return gjson.Get(rawIntrospection, "__schema").String(), nil
}

func engineClientContainerWithVersion(ctx context.Context, devEngine *dagger.Service, version string) (*dagger.Container, error) {
	endpoint, err := devEngine.Endpoint(ctx, dagger.ServiceEndpointOpts{Port: 1234, Scheme: "tcp"})
	if err != nil {
		return nil, err
	}

	return dag.Container().From(fmt.Sprintf("ghcr.io/dagger/engine:%s", version)).
		WithServiceBinding("dev-engine", devEngine).
		WithEnvVariable("_EXPERIMENTAL_DAGGER_RUNNER_HOST", endpoint), nil
}

func engineServiceWithVersion(version string, withs ...func(*dagger.Container) *dagger.Container) *dagger.Service {
	ctr := dag.Container().From(fmt.Sprintf("ghcr.io/dagger/engine:%s", version))
	for _, with := range withs {
		ctr = with(ctr)
	}

	deviceName, cidr := impl.GetUniqueNestedEngineNetwork()
	return ctr.
		WithMountedCache("/var/lib/dagger", dag.CacheVolume("dagger-dev-engine-state-"+identity.NewID())).
		WithExposedPort(1234, dagger.ContainerWithExposedPortOpts{Protocol: dagger.Tcp}).
		WithExec([]string{
			"--addr", "tcp://0.0.0.0:1234",
			"--addr", "unix:///var/run/buildkit/buildkitd.sock",
			// // avoid network conflicts with other tests
			"--network-name", deviceName,
			"--network-cidr", cidr,
		}, dagger.ContainerWithExecOpts{
			UseEntrypoint:            true,
			InsecureRootCapabilities: true,
		}).AsService()
}

package main

import (
	"context"
	"dagger/compat/internal/dagger"
	"dagger/compat/pkg/impl"
	"fmt"

	"github.com/dagger/dagger/dagql/introspection"
	"github.com/josephburnett/jd/v2"
	"github.com/moby/buildkit/identity"
	"github.com/tidwall/gjson"
)

type Compat struct {
	source *dagger.Directory
}

func (m *Compat) Check(ctx context.Context,
	module, versionA, versionB string,
	//+optional
	source *dagger.Directory) error {
	schemaA, err := m.getSchemaForModuleForEngineVersion(ctx, module, versionA, source)
	if err != nil {
		return err
	}

	schemaB, err := m.getSchemaForModuleForEngineVersion(ctx, module, versionB, source)
	if err != nil {
		return err
	}

	a, _ := jd.ReadJsonString(schemaA)
	b, _ := jd.ReadJsonString(schemaB)

	diff := a.Diff(b).Render()
	if diff != "" {
		return fmt.Errorf("%s", diff)
	}

	return nil
}

func (m *Compat) getSchemaForModuleForEngineVersion(ctx context.Context, module, engineVersion string, source *dagger.Directory) (string, error) {
	var engineSvc *dagger.Service
	var client *dagger.Container
	var err error

	if engineVersion == "dev" {
		client = devEngineAndClient(ctx, source)
	} else {
		engineSvc = engineServiceWithVersion(engineVersion)
		client, err = engineClientContainerWithVersion(ctx, engineSvc, engineVersion)
		if err != nil {
			return "", err
		}
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

func devEngineAndClient(ctx context.Context, source *dagger.Directory) *dagger.Container {
	return dag.DaggerDev(source).Dev()
}

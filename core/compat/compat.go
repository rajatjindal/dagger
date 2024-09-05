package compat

import (
	"context"

	"github.com/dagger/dagger/engine/strcase"
	"golang.org/x/mod/semver"
)

// the new strcase implementation is used if the version
// is greater than this cutoff version
const strcaseVersionCutOff = "v0.12.7"

type CompatCtxKey struct{}

type Compat struct {
	EngineVersion string
	LegacyOrNew   string
	Strcase       strcase.Caser
}

func GetCompatFromContext(ctx context.Context) *Compat {
	okval, ok := ctx.Value(CompatCtxKey{}).(*Compat)
	if !ok {
		return &Compat{
			Strcase: strcase.NewCaser(),
		}
	}

	return okval
}

func AddCompatToContext(ctx context.Context, engineVersion string) context.Context {
	compat := GetCompatFromContext(ctx)

	// if engineVersion is empty OR not a valid semver, treat it as newer version
	if !semver.IsValid(engineVersion) || semver.Compare(engineVersion, strcaseVersionCutOff) > 0 {
		compat.Strcase = strcase.NewCaser()
		compat.LegacyOrNew = "new"
		compat.EngineVersion = engineVersion
	} else {
		compat.Strcase = strcase.NewLegacyCaser()
		compat.LegacyOrNew = "legacy"
		compat.EngineVersion = engineVersion
	}

	return context.WithValue(ctx, CompatCtxKey{}, compat)
}

func Strcase(ctx context.Context) strcase.Caser {
	return GetCompatFromContext(ctx).Strcase
}

package compat

import (
	"context"

	"github.com/dagger/dagger/engine/strcase"
	"golang.org/x/mod/semver"
)

const strcaseVersionCutOff = "v0.12.8"

type CompatCtxKey struct{}

type Compat struct {
	Strcase strcase.Caser
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
	if compat == nil {
		compat = &Compat{}
	}

	// if engineVersion is empty OR not a valid semver, treat it as newer version
	if !semver.IsValid(engineVersion) || semver.Compare(engineVersion, strcaseVersionCutOff) > 0 {
		compat.Strcase = strcase.NewCaser()
	} else {
		compat.Strcase = strcase.NewLegacyCaser()
	}

	return context.WithValue(ctx, CompatCtxKey{}, compat)
}

func Strcase(ctx context.Context) strcase.Caser {
	return GetCompatFromContext(ctx).Strcase
}

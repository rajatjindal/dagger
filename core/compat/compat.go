package compat

import (
	"context"
	"fmt"

	"github.com/dagger/dagger/engine/strcase"
	"golang.org/x/mod/semver"
)

const strcaseVersionCutOff = "v0.12.5"

type CompatCtxKey struct{}

type Compat struct {
	Strcase strcase.Caser
}

func MustGetCompatFromContext(ctx context.Context) *Compat {
	okval, ok := ctx.Value(CompatCtxKey{}).(*Compat)
	if !ok {
		panic("compat context is not set")
	}

	return okval
}

func getCompatFromContext(ctx context.Context) *Compat {
	okval, ok := ctx.Value(CompatCtxKey{}).(*Compat)
	if !ok {
		return &Compat{
			Strcase: strcase.NewCaser(),
		}
	}

	return okval
}

func MustAddCompatToContext(ctx context.Context, engineVersion string) context.Context {
	compat := getCompatFromContext(ctx)
	if compat == nil {
		compat = &Compat{}
	}

	if !semver.IsValid(engineVersion) {
		panic(fmt.Sprintf("INVALID ENGINE VERSION %q", engineVersion))
	}

	if semver.Compare(engineVersion, strcaseVersionCutOff) > 0 {
		compat.Strcase = strcase.NewCaser()
	} else {
		compat.Strcase = strcase.NewLegacyCaser()
	}

	return context.WithValue(ctx, CompatCtxKey{}, compat)
}

func AddCompatToContext(ctx context.Context, engineVersion string) context.Context {
	compat := getCompatFromContext(ctx)
	if compat == nil {
		compat = &Compat{}
	}

	if semver.Compare(engineVersion, strcaseVersionCutOff) > 0 {
		compat.Strcase = strcase.NewCaser()
	} else {
		compat.Strcase = strcase.NewLegacyCaser()
	}

	return context.WithValue(ctx, CompatCtxKey{}, compat)
}

func Strcase(ctx context.Context) strcase.Caser {
	return MustGetCompatFromContext(ctx).Strcase
}

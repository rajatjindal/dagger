package strcase

import (
	"github.com/ettle/strcase"
)

func ToCamel(inp string) string {
	return strcase.ToCamel(inp)
}

func ToPascal(inp string) string {
	return strcase.ToPascal(inp)
}

func ToKebab(inp string) string {
	return strcase.ToKebab(inp)
}

func ToScreamingSnake(inp string) string {
	return strcase.ToSNAKE(inp)
}

func ConfigureAcronym(key, val string) {
	// for github.com/ettle/strcase, it is a noop,
	// however keeping it here to ensure this case
	// is handled if we change the underlying lib
	// in future.
}

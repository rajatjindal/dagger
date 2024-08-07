package strcase

import (
	"github.com/ettle/strcase"
)

var overrides = map[string]bool{}
var caser = strcase.NewCaser(false, nil, nil)

func ToPascal(inp string) string {
	return caser.ToPascal(inp)
}

func ToCamel(inp string) string {
	return caser.ToCamel(inp)
}

func ToKebab(inp string) string {
	return caser.ToKebab(inp)
}

func ToScreamingSnake(inp string) string {
	return caser.ToSNAKE(inp)
}

func ConfigureAcronym(key, val string) {
	overrides[val] = true
	caser = strcase.NewCaser(false, overrides, nil)
}

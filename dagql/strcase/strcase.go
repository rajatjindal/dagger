package strcase

import (
	"github.com/ettle/strcase"
)

var splitFn = strcase.NewSplitFn(
	[]rune{'*', '.', ',', '-', '_'},
	strcase.SplitCase,
	strcase.SplitAcronym,
	strcase.PreserveNumberFormatting,
	strcase.SplitBeforeNumber,
	strcase.SplitAfterNumber,
)
var overrides = map[string]bool{}
var caser = strcase.NewCaser(false, nil, splitFn)

// ToPascal returns words in PascalCase (capitalized words concatenated together).
func ToPascal(inp string) string {
	return caser.ToPascal(inp)
}

// ToCamel returns words in camelCase (capitalized words concatenated together, with first word lower case).
func ToCamel(inp string) string {
	return caser.ToCamel(inp)
}

// ToKebab returns words in kebab-case (lower case words with dashes).
func ToKebab(inp string) string {
	return caser.ToKebab(inp)
}

// ToScreamingSnake returns words in SNAKE_CASE (upper case words with underscores).
func ToScreamingSnake(inp string) string {
	return caser.ToSNAKE(inp)
}

func ConfigureAcronym(key, val string) {
	// overrides[val] = true
	// caser = strcase.NewCaser(false, overrides, nil)
}

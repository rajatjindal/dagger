package strcase

import (
	"github.com/ettle/strcase"
)

var caser *strcase.Caser

func init() {
	var splitFn = strcase.NewSplitFn(
		[]rune{'*', '.', ',', '-', '_'},
		strcase.SplitCase,
		strcase.SplitAcronym,
		strcase.PreserveNumberFormatting,
		strcase.SplitBeforeNumber,
		strcase.SplitAfterNumber,
	)

	overrides := map[string]bool{
		"JSON": true,
	}

	caser = strcase.NewCaser(false, overrides, splitFn)
}

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

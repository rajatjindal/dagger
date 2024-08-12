package strcase

import (
	"github.com/ettle/strcase"
)

var caser *strcase.Caser

func init() {
	caser = newCaser()
}

func newCaser() *strcase.Caser {
	var splitFn = strcase.NewSplitFn(
		[]rune{'*', '.', ',', '-', '_'},
		strcase.SplitCase,
		strcase.SplitAcronym,
		strcase.PreserveNumberFormatting,
		strcase.SplitBeforeNumber,
		strcase.SplitAfterNumber,
	)

	return strcase.NewCaser(false, nil, splitFn)
}

// ToPascal returns words in PascalCase (capitalized words concatenated together).
func ToPascal(inp string) string {
	return caser.ToCase(inp, strcase.TitleCase|strcase.PreserveInitialism, '\u0000')
}

// ToCamel returns words in camelCase (capitalized words concatenated together, with first word lower case).
func ToCamel(inp string) string {
	return caser.ToCase(inp, strcase.CamelCase|strcase.PreserveInitialism, '\u0000')
}

// ToKebab returns words in kebab-case (lower case words with dashes).
func ToKebab(inp string) string {
	return caser.ToKebab(inp)
}

// ToScreamingSnake returns words in SNAKE_CASE (upper case words with underscores).
func ToScreamingSnake(inp string) string {
	return caser.ToSNAKE(inp)
}

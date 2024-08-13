package strcase

import (
	"github.com/ettle/strcase"
)

var activeCaser *strcase.Caser

func init() {
	activeCaser = newCaser()
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
	return activeCaser.ToCase(inp, strcase.TitleCase|strcase.PreserveInitialism, '\u0000')
}

// ToCamel returns words in camelCase (capitalized words concatenated together, with first word lower case).
func ToCamel(inp string) string {
	return activeCaser.ToCamel(inp)
}

// ToKebab returns words in kebab-case (lower case words with dashes).
func ToKebab(inp string) string {
	return activeCaser.ToKebab(inp)
}

// ToScreamingSnake returns words in SNAKE_CASE (upper case words with underscores).
func ToScreamingSnake(inp string) string {
	return activeCaser.ToSNAKE(inp)
}

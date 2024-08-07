package strcase

import (
	"sync"

	"github.com/ettle/strcase"
)

type Caser struct {
	caser *strcase.Caser

	sync.Mutex
}

var caser = &Caser{}
var overrides = sync.Map{}

func init() {
	updateCaser()
}

func updateCaser() {
	caser.Lock()
	defer caser.Unlock()

	var splitFn = strcase.NewSplitFn(
		[]rune{'*', '.', ',', '-', '_'},
		strcase.SplitCase,
		strcase.SplitAcronym,
		strcase.PreserveNumberFormatting,
		strcase.SplitBeforeNumber,
		strcase.SplitAfterNumber,
	)

	ioverrides := map[string]bool{}
	overrides.Range(func(key, _ any) bool {
		ioverrides[key.(string)] = true
		return true
	})

	caser.caser = strcase.NewCaser(false, ioverrides, splitFn)
}

// ToPascal returns words in PascalCase (capitalized words concatenated together).
func ToPascal(inp string) string {
	return caser.caser.ToPascal(inp)
}

// ToCamel returns words in camelCase (capitalized words concatenated together, with first word lower case).
func ToCamel(inp string) string {
	return caser.caser.ToCamel(inp)
}

// ToKebab returns words in kebab-case (lower case words with dashes).
func ToKebab(inp string) string {
	return caser.caser.ToKebab(inp)
}

// ToScreamingSnake returns words in SNAKE_CASE (upper case words with underscores).
func ToScreamingSnake(inp string) string {
	return caser.caser.ToSNAKE(inp)
}

func ConfigureAcronym(key, val string) {
	overrides.Store(val, true)
	updateCaser()
}

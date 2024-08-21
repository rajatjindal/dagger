package strcaselegacy

import (
	"github.com/iancoleman/strcase"
)

// ToPascal returns words in PascalCase (capitalized words concatenated together).
func ToPascal(inp string) string {
	return strcase.ToCamel(inp)
}

// ToCamel returns words in camelCase (capitalized words concatenated together, with first word lower case).
func ToCamel(inp string) string {
	return strcase.ToLowerCamel(inp)
}

// ToKebab returns words in kebab-case (lower case words with dashes).
func ToKebab(inp string) string {
	return strcase.ToKebab(inp)
}

// ToScreamingSnake returns words in SNAKE_CASE (upper case words with underscores).
func ToScreamingSnake(inp string) string {
	return strcase.ToScreamingSnake(inp)
}

// ToSnake returns words in snake_case (lower case words with underscores).
func ToSnake(inp string) string {
	return strcase.ToSnake(inp)
}

func ConfigureAcronyms(key, value string) {
	strcase.ConfigureAcronym(key, value)
}

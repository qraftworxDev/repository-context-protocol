package ast

import "repository-context-protocol/internal/models"

// Language-agnostic parser interface
type LanguageParser interface {
	ParseFile(path string, content []byte) (*models.FileContext, error)
	GetSupportedExtensions() []string
	GetLanguageName() string
}

type ParserRegistry struct {
	parsers map[string]LanguageParser
}

func NewParserRegistry() *ParserRegistry {
	return &ParserRegistry{
		parsers: make(map[string]LanguageParser),
	}
}

func (r *ParserRegistry) Register(parser LanguageParser) {
	for _, ext := range parser.GetSupportedExtensions() {
		r.parsers[ext] = parser
	}
}

func (r *ParserRegistry) GetParser(fileExt string) (LanguageParser, bool) {
	parser, exists := r.parsers[fileExt]
	return parser, exists
}

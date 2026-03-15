package parser

// Symbol represents an extracted code symbol from AST parsing.
type Symbol struct {
	Name      string // symbol name (e.g., "HandleRequest", "UserModel")
	Type      string // function, method, type, class, interface, import, export, test
	Language  string // go, python, etc.
	Signature string // full signature (e.g., "func HandleRequest(w http.ResponseWriter, r *http.Request) error")
	Docstring string // leading comment/docstring
	StartLine int    // 1-based line number
	EndLine   int    // 1-based line number
	FilePath  string // relative to repo root
	BodyHash  string // SHA256 of the symbol body text
}

// Parser interface — one implementation per language.
type Parser interface {
	Language() string
	Extensions() []string // e.g., [".go"] or [".py", ".pyi"]
	Parse(filePath string, source []byte) ([]Symbol, error)
}

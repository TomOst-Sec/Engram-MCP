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

// Reference represents a usage/call of a symbol found inside a function body
// or initializer. This enables "where is X called" queries including indirect
// calls through function pointers and dispatch tables.
type Reference struct {
	ToName    string // the symbol name being referenced (e.g., "f")
	Kind      string // call, pointer_assign, callback_arg, initializer, address_of, param_alias, indirect_call
	FromFunc  string // enclosing function name, or "<top-level>" for file scope
	FilePath  string // relative to repo root
	Line      int    // 1-based
	Context   string // the source line for display
	ParamName string // for param_alias: the parameter name in the callee
	TableName string // for initializer: the table variable name (e.g., "dispatch_table")
}

// Parser interface — one implementation per language.
type Parser interface {
	Language() string
	Extensions() []string // e.g., [".go"] or [".py", ".pyi"]
	Parse(filePath string, source []byte) ([]Symbol, error)
}

// ReferenceParser is an optional interface that parsers can implement
// to extract function references (calls, pointer assignments, dispatch tables).
type ReferenceParser interface {
	ParseReferences(filePath string, source []byte) ([]Reference, error)
}

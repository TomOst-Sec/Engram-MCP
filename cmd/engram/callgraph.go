package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/TomOst-Sec/colony-project/internal/cli"
	"github.com/spf13/cobra"
)

var (
	cgDepth    int
	cgReverse  bool
	cgAllPaths bool
)

func newCallgraphCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "callgraph <function>",
		Short: "Show call graph for a function",
		Long: `Build and display a call graph showing how a function is reached.

By default shows callers (who calls this function, including via pointers and dispatch tables).
Use --reverse to show callees (what this function calls).

Examples:
  engram callgraph ff            # who calls ff? (direct, pointer, table)
  engram callgraph ff --depth 3  # trace 3 levels deep
  engram callgraph main --reverse # what does main call?`,
		Args: cobra.ExactArgs(1),
		RunE: runCallgraph,
	}
	cmd.Flags().IntVarP(&cgDepth, "depth", "d", 3, "Maximum depth to traverse")
	cmd.Flags().BoolVarP(&cgReverse, "reverse", "r", false, "Show callees instead of callers")
	cmd.Flags().BoolVar(&cgAllPaths, "all", false, "Show all paths, not just unique edges")
	return cmd
}

// edge represents one call graph edge.
type edge struct {
	from     string // caller function
	to       string // callee function
	kind     string // call, pointer_assign, callback_arg, initializer
	file     string
	line     int
	context  string
}

// node tracks a function's definition location.
type node struct {
	name string
	file string
	line int
}

func runCallgraph(cmd *cobra.Command, args []string) error {
	target := args[0]

	store, _, _, err := openDatabase()
	if err != nil {
		return err
	}
	defer store.Close()

	db := store.DB()

	// Check if target exists as a defined symbol
	var defFile string
	var defLine int
	err = db.QueryRow(
		"SELECT file_path, start_line FROM code_index WHERE symbol_name = ? AND symbol_type = 'function' LIMIT 1",
		target,
	).Scan(&defFile, &defLine)
	if err == sql.ErrNoRows {
		fmt.Fprintf(os.Stderr, "Warning: no function definition found for %q\n", target)
	} else if err == nil {
		fmt.Fprintf(os.Stdout, "%s %s\n",
			cli.SymbolName.Render(target),
			cli.Label.Render(fmt.Sprintf("defined at %s:%d", defFile, defLine)))
	}

	if cgReverse {
		return showCallees(db, target)
	}

	// Show table membership summary (improvement 4)
	showTableMembership(db, target)

	return showCallers(db, target)
}

// showCallers traces backwards: who calls `target`? Handles indirect paths
// through tables and pointer assigns.
func showCallers(db *sql.DB, target string) error {
	fmt.Fprintf(os.Stdout, "\n%s\n\n", cli.Label.Render(fmt.Sprintf("Call graph — who calls %s?", target)))

	visited := map[string]bool{}
	return traceCallers(db, target, 0, "", visited)
}

func traceCallers(db *sql.DB, name string, depth int, prefix string, visited map[string]bool) error {
	if depth >= cgDepth {
		return nil
	}
	if visited[name] {
		return nil
	}
	visited[name] = true

	// 1. Direct references (call, indirect_call, pointer_assign, param_alias)
	rows, err := db.Query(
		`SELECT file_path, to_name, kind, from_func, line, context
		 FROM code_references WHERE to_name = ?
		 AND kind IN ('call', 'indirect_call', 'pointer_assign', 'param_alias', 'callback_arg', 'address_of')
		 ORDER BY file_path, line`, name)
	if err != nil {
		return err
	}
	defer rows.Close()

	var edges []edge
	for rows.Next() {
		var e edge
		if err := rows.Scan(&e.file, &e.to, &e.kind, &e.from, &e.line, &e.context); err != nil {
			return err
		}
		edges = append(edges, e)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// 2. Table membership: find all tables that contain this function
	type tableInfo struct {
		tableVar string
		filePath string
		occupied int
		total    int
	}
	var tables []tableInfo
	tableRows, err := db.Query(`
		SELECT tm.table_var, tm.file_path,
			(SELECT COUNT(*) FROM table_membership t2 WHERE t2.table_var = tm.table_var AND t2.func_name = ?),
			(SELECT COUNT(*) FROM table_membership t2 WHERE t2.table_var = tm.table_var)
		FROM table_membership tm WHERE tm.func_name = ?
		GROUP BY tm.table_var, tm.file_path`, name, name)
	if err == nil {
		defer tableRows.Close()
		for tableRows.Next() {
			var ti tableInfo
			if err := tableRows.Scan(&ti.tableVar, &ti.filePath, &ti.occupied, &ti.total); err == nil {
				tables = append(tables, ti)
			}
		}
	}

	totalItems := len(edges) + len(tables)
	printed := 0

	// Print table entries with probability
	for _, ti := range tables {
		printed++
		connector := "├── "
		childPrefix := prefix + "│   "
		if printed == totalItems {
			connector = "└── "
			childPrefix = prefix + "    "
		}

		pct := 0.0
		if ti.total > 0 {
			pct = 100.0 * float64(ti.occupied) / float64(ti.total)
		}
		fmt.Fprintf(os.Stdout, "%s%s %s  %s  %s\n",
			prefix, connector,
			cli.Label.Render("[tbl]"),
			cli.SymbolName.Render(ti.tableVar),
			cli.Label.Render(fmt.Sprintf("%s — %d/%d slots (%.0f%%)", ti.filePath, ti.occupied, ti.total, pct)))

		// Trace who calls through this table
		if err := traceCallers(db, ti.tableVar, depth+1, childPrefix, visited); err != nil {
			return err
		}
	}

	// Print direct references
	for _, e := range edges {
		printed++
		connector := "├── "
		childPrefix := prefix + "│   "
		if printed == totalItems {
			connector = "└── "
			childPrefix = prefix + "    "
		}

		kindTag := kindSymbol(e.kind)
		loc := fmt.Sprintf("%s:%d", e.file, e.line)

		fmt.Fprintf(os.Stdout, "%s%s %s  %s  %s  %s\n",
			prefix, connector,
			cli.Label.Render(kindTag),
			cli.SymbolName.Render(e.from),
			cli.FilePath.Render(loc),
			e.context)

		// Recurse: who calls this caller?
		if err := traceCallers(db, e.from, depth+1, childPrefix, visited); err != nil {
			return err
		}
	}

	return nil
}

// findTableName finds the variable name of a dispatch table given a file and
// line where a function appears in an initializer_list.
func findTableName(db *sql.DB, file string, line int) string {
	// Look for a symbol (variable) defined near this line.
	// The table declaration is usually a few lines above the initializer entries.
	// We search code_index for declarations in the same file near this line.
	// Fallback: scan references for call-kind refs to identifiers in the same file
	// that are subscript calls (dispatch_table[idx](..)).

	// Strategy: find all "call" kind references in the same file whose context
	// contains "[" (subscript call pattern like `table[idx](...)`)
	// and whose from_func is in the neighborhood.

	// Simpler approach: look at the DB for identifiers referenced as "call" kind
	// in the same file that appear as table names.
	// Actually best: just look for the preceding declaration. The initializer
	// entries for a table like `static math_op table[] = { ff, ... }` mean
	// the table name is declared on a line before `line`. We can get it from
	// the source directly, but we can also look at code_references for all
	// "call" entries in the same file with subscript pattern.

	// Find the table variable name: look for "call" refs where context shows
	// subscript dispatch pattern: `name[expr](` — e.g. `dispatch_table[idx](a, b)`
	// These are NOT defined functions, they are array variables.
	rows, err := db.Query(
		`SELECT DISTINCT cr.to_name FROM code_references cr
		 WHERE cr.file_path = ? AND cr.kind = 'call'
		 AND cr.context LIKE '%' || cr.to_name || '[%](%'
		 AND NOT EXISTS (
		   SELECT 1 FROM code_index ci
		   WHERE ci.symbol_name = cr.to_name AND ci.symbol_type = 'function'
		 )`, file)
	if err != nil {
		return ""
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		return name
	}
	return ""
}

// showCallees traces forward: what does `target` call?
func showCallees(db *sql.DB, target string) error {
	fmt.Fprintf(os.Stdout, "\n%s\n\n", cli.Label.Render(fmt.Sprintf("Call graph — what does %s call?", target)))

	visited := map[string]bool{}
	return traceCallees(db, target, 0, "", visited)
}

func traceCallees(db *sql.DB, funcName string, depth int, prefix string, visited map[string]bool) error {
	if depth >= cgDepth {
		return nil
	}
	if visited[funcName] {
		return nil
	}
	visited[funcName] = true

	// Find all references FROM this function
	rows, err := db.Query(
		`SELECT file_path, to_name, kind, from_func, line, context
		 FROM code_references WHERE from_func = ? ORDER BY line`, funcName)
	if err != nil {
		return err
	}
	defer rows.Close()

	var edges []edge
	for rows.Next() {
		var e edge
		if err := rows.Scan(&e.file, &e.to, &e.kind, &e.from, &e.line, &e.context); err != nil {
			return err
		}
		edges = append(edges, e)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// Deduplicate by target name (skip variables like 'result', 'a', 'b')
	seen := map[string]bool{}
	var unique []edge
	for _, e := range edges {
		key := e.to + ":" + e.kind
		if seen[key] {
			continue
		}
		seen[key] = true
		unique = append(unique, e)
	}

	for i, e := range unique {
		connector := "├── "
		childPrefix := prefix + "│   "
		if i == len(unique)-1 {
			connector = "└── "
			childPrefix = prefix + "    "
		}

		kindTag := kindSymbol(e.kind)
		loc := fmt.Sprintf("%s:%d", e.file, e.line)

		// Check if this is a known function
		isFunc := isDefinedFunction(db, e.to)
		nameStyle := e.to
		if isFunc {
			nameStyle = e.to + "()"
		}

		fmt.Fprintf(os.Stdout, "%s%s %s  %s  %s\n",
			prefix, connector,
			cli.Label.Render(kindTag),
			cli.SymbolName.Render(nameStyle),
			cli.FilePath.Render(loc))

		// If it's a table variable (not a defined function), show table contents
		if e.kind == "call" && !isFunc && isTableVariable(db, e.to, e.file) {
			tableContents := getTableContents(db, e.to, e.file)
			for j, tc := range tableContents {
				tConn := "├── "
				if j == len(tableContents)-1 {
					tConn = "└── "
				}
				fmt.Fprintf(os.Stdout, "%s%s %s  %s  %s\n",
					childPrefix, tConn,
					cli.Label.Render("[table entry]"),
					cli.SymbolName.Render(tc.to),
					cli.FilePath.Render(fmt.Sprintf("%s:%d", tc.file, tc.line)))
			}
		}

		if isFunc {
			if err := traceCallees(db, e.to, depth+1, childPrefix, visited); err != nil {
				return err
			}
		}
	}

	return nil
}

func isDefinedFunction(db *sql.DB, name string) bool {
	var count int
	err := db.QueryRow(
		"SELECT COUNT(*) FROM code_index WHERE symbol_name = ? AND symbol_type = 'function'", name,
	).Scan(&count)
	return err == nil && count > 0
}

// isTableVariable checks if `name` is used as a subscript-dispatch variable
// (i.e., `name[expr](...)`) in the given file, confirming it's a function pointer table.
func isTableVariable(db *sql.DB, name, file string) bool {
	var count int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM code_references
		 WHERE file_path = ? AND to_name = ? AND kind = 'call'
		 AND context LIKE '%' || ? || '[%](%'`,
		file, name, name).Scan(&count)
	return err == nil && count > 0
}

// getTableContents returns the initializer entries for a specific table variable.
// It finds the line range of the table declaration, then returns initializer refs in that range.
func getTableContents(db *sql.DB, tableName, file string) []edge {
	// Find the approximate line range: look for initializer refs that are near each other
	// and close to where the table variable is declared.
	// Strategy: find all initializer refs in the file, group them by proximity,
	// and pick the group that's near the table's usage.

	// First, find the table's call site line to know roughly where the table is
	var useLine int
	_ = db.QueryRow(
		`SELECT line FROM code_references WHERE file_path = ? AND to_name = ? AND kind = 'call' LIMIT 1`,
		file, tableName).Scan(&useLine)

	// Get all initializer entries in the file
	rows, err := db.Query(
		`SELECT file_path, to_name, kind, from_func, line, context
		 FROM code_references WHERE file_path = ? AND kind = 'initializer'
		 AND from_func = '<top-level>' ORDER BY line`, file)
	if err != nil {
		return nil
	}
	defer rows.Close()

	// Group consecutive initializer entries (gap > 3 lines = new group)
	type group struct {
		entries  []edge
		minLine  int
		maxLine  int
	}
	var groups []group
	var cur *group

	for rows.Next() {
		var e edge
		if err := rows.Scan(&e.file, &e.to, &e.kind, &e.from, &e.line, &e.context); err != nil {
			continue
		}
		if cur == nil || e.line-cur.maxLine > 3 {
			groups = append(groups, group{minLine: e.line, maxLine: e.line})
			cur = &groups[len(groups)-1]
		}
		cur.entries = append(cur.entries, e)
		cur.maxLine = e.line
	}

	// Pick the group closest to (and before) the usage line
	var best *group
	for i := range groups {
		g := &groups[i]
		if useLine == 0 || g.maxLine < useLine {
			if best == nil || g.maxLine > best.maxLine {
				best = g
			}
		}
	}
	// Fallback: closest group overall
	if best == nil && len(groups) > 0 {
		best = &groups[0]
		for i := range groups {
			g := &groups[i]
			d1 := useLine - g.minLine
			d2 := useLine - best.minLine
			if d1 < 0 { d1 = -d1 }
			if d2 < 0 { d2 = -d2 }
			if d1 < d2 {
				best = g
			}
		}
	}

	if best == nil {
		return nil
	}
	return best.entries
}

func kindSymbol(kind string) string {
	switch kind {
	case "call":
		return "[call]"
	case "indirect_call":
		return "[->()]"
	case "pointer_assign":
		return "[ptr=]"
	case "callback_arg":
		return "[cb()]"
	case "param_alias":
		return "[arg=]"
	case "initializer":
		return "[tbl]"
	case "address_of":
		return "[&ref]"
	default:
		return "[" + kind + "]"
	}
}

// showTableMembership prints a summary of which dispatch tables contain (and don't contain)
// the target function, with slot counts and probability.
func showTableMembership(db *sql.DB, target string) {
	// Get all known tables
	rows, err := db.Query(`
		SELECT table_var, file_path, COUNT(*) as total,
			SUM(CASE WHEN func_name = ? THEN 1 ELSE 0 END) as occupied
		FROM table_membership
		GROUP BY table_var, file_path
		ORDER BY table_var`, target)
	if err != nil {
		return
	}
	defer rows.Close()

	type tbl struct {
		name     string
		file     string
		total    int
		occupied int
	}
	var present, absent []tbl

	for rows.Next() {
		var t tbl
		if err := rows.Scan(&t.name, &t.file, &t.total, &t.occupied); err != nil {
			continue
		}
		if t.occupied > 0 {
			present = append(present, t)
		} else {
			absent = append(absent, t)
		}
	}

	if len(present) == 0 && len(absent) == 0 {
		return
	}

	fmt.Fprintf(os.Stdout, "\n%s\n", cli.Label.Render("Table membership:"))
	for _, t := range present {
		pct := 100.0 * float64(t.occupied) / float64(t.total)
		fmt.Fprintf(os.Stdout, "  %s %s  %s\n",
			cli.SymbolName.Render("IN"),
			cli.SymbolName.Render(t.name),
			cli.Label.Render(fmt.Sprintf("%d/%d slots (%.0f%%) — %s", t.occupied, t.total, pct, t.file)))
	}
	for _, t := range absent {
		fmt.Fprintf(os.Stdout, "  %s %s  %s\n",
			cli.FilePath.Render("--"),
			cli.FilePath.Render(t.name),
			cli.Label.Render(fmt.Sprintf("NOT present — %s (%d slots)", t.file, t.total)))
	}
	fmt.Fprintln(os.Stdout)
}

// printDOT outputs the call graph in Graphviz DOT format.
func printDOT(db *sql.DB, target string, reverse bool) error {
	fmt.Println("digraph callgraph {")
	fmt.Println("  rankdir=LR;")
	fmt.Println("  node [shape=box, fontname=\"monospace\"];")
	fmt.Printf("  %q [style=filled, fillcolor=lightblue];\n", target)

	var rows *sql.Rows
	var err error
	if reverse {
		rows, err = db.Query(
			`SELECT DISTINCT from_func, to_name, kind FROM code_references
			 WHERE from_func = ? AND kind IN ('call','pointer_assign','callback_arg')`,
			target)
	} else {
		rows, err = db.Query(
			`SELECT DISTINCT from_func, to_name, kind FROM code_references
			 WHERE to_name = ?`, target)
	}
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var from, to, kind string
		if err := rows.Scan(&from, &to, &kind); err != nil {
			continue
		}
		style := ""
		switch kind {
		case "pointer_assign":
			style = " [style=dashed, label=\"ptr\"]"
		case "callback_arg":
			style = " [style=dotted, label=\"cb\"]"
		case "initializer":
			style = " [style=bold, label=\"table\"]"
		}
		if reverse {
			fmt.Printf("  %q -> %q%s;\n", from, to, style)
		} else {
			fmt.Printf("  %q -> %q%s;\n", from, to, style)
		}
	}

	// Also show table entries
	tblRows, err := db.Query(
		`SELECT DISTINCT to_name, file_path FROM code_references
		 WHERE to_name = ? AND kind = 'initializer'`, target)
	if err == nil {
		defer tblRows.Close()
		for tblRows.Next() {
			var toName, file string
			if err := tblRows.Scan(&toName, &file); err != nil {
				continue
			}
			// Find which table contains this
			tableName := findTableName(db, file, 0)
			if tableName != "" {
				fmt.Printf("  %q -> %q [style=bold, label=\"in table\"];\n", tableName, toName)
				// Who calls through the table?
				callerRows, err := db.Query(
					`SELECT DISTINCT from_func FROM code_references
					 WHERE to_name = ? AND kind = 'call'`, tableName)
				if err == nil {
					for callerRows.Next() {
						var caller string
						if err := callerRows.Scan(&caller); err != nil {
							continue
						}
						fmt.Printf("  %q -> %q [label=\"[idx]\"];\n", caller, tableName)
					}
					callerRows.Close()
				}
			}
		}
	}

	fmt.Println("}")
	return nil
}

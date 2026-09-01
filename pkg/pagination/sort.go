package pagination

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// MaxSortFields caps the length of the ORDER BY chain.
const MaxSortFields = 3

// ErrInvalidSort: the sort parameter failed the whitelist.
var ErrInvalidSort = errors.New("invalid sort parameter")

// Whitelist maps public names in the query string to database column names.
// Values must be developer-written constants, never user input.
type Whitelist map[string]string

// Sortable: the sort rules for one resource.
type Sortable struct {
	Allowed    Whitelist // only keys listed here are accepted
	Default    string    // used when the sort param is empty, same syntax as the query
	TieBreaker string    // unique column that keeps ordering stable across pages
}

// OrderBy turns "-created_at,name" into a safe ORDER BY clause.
// A "-" prefix means DESC, no prefix means ASC.
func (s Sortable) OrderBy(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		raw = s.Default
	}

	clauses := make([]string, 0, MaxSortFields+1)
	seen := make(map[string]bool, MaxSortFields+1)

	for _, field := range strings.Split(raw, ",") {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}

		direction := "ASC"
		if trimmed, isDesc := strings.CutPrefix(field, "-"); isDesc {
			direction, field = "DESC", trimmed
		}

		column, ok := s.Allowed[field]
		if !ok {
			return "", fmt.Errorf("%w: unknown column %q, allowed: %s", ErrInvalidSort, field, strings.Join(s.Fields(), ", "))
		}
		if seen[column] {
			return "", fmt.Errorf("%w: column %q listed more than once", ErrInvalidSort, field)
		}
		if len(clauses) == MaxSortFields {
			return "", fmt.Errorf("%w: at most %d columns", ErrInvalidSort, MaxSortFields)
		}

		seen[column] = true
		clauses = append(clauses, column+" "+direction)
	}

	// Without a unique trailing column, tied rows can duplicate or vanish across pages.
	if s.TieBreaker != "" && !seen[s.TieBreaker] {
		clauses = append(clauses, s.TieBreaker+" ASC")
	}
	return strings.Join(clauses, ", "), nil
}

// Fields: the ordered list of allowed public names, for error messages & docs.
func (s Sortable) Fields() []string {
	fields := make([]string, 0, len(s.Allowed))
	for name := range s.Allowed {
		fields = append(fields, name)
	}
	slices.Sort(fields)
	return fields
}

// Package domain holds what every layer agrees on: the errors a rule can
// raise and the words that name a broken rule. Each part of the business lives
// in a package of its own below this one, free of HTTP, SQL and the clock.
package domain

import (
	"errors"
	"strings"
)

// The causes the transport tells apart. Adapters wrap them with %w and the
// edge maps them with errors.Is, so a new cause never needs a new case there.
var (
	ErrNotFound        = errors.New("not found")
	ErrConflict        = errors.New("conflict")
	ErrUnauthenticated = errors.New("unauthenticated")
)

// Code names a broken rule in words a program can switch on. The web app turns
// each one into a sentence, so the set is part of the API contract.
type Code string

// The rules an input can break.
const (
	CodeRequired   Code = "required"
	CodeTooShort   Code = "too_short"
	CodeTooLong    Code = "too_long"
	CodeOutOfRange Code = "out_of_range"
	CodeInvalid    Code = "invalid"
	CodeTaken      Code = "taken"
)

// Violation is one broken rule on one field. Field is the name the API
// contract uses, on purpose: one vocabulary from the table to the form.
type Violation struct {
	Field string
	Code  Code
}

// ValidationError lists every rule an input broke, not only the first one, so
// a form can point at all of them at once.
type ValidationError struct {
	Violations []Violation
}

func (e *ValidationError) Error() string {
	parts := make([]string, len(e.Violations))
	for i, v := range e.Violations {
		parts[i] = v.Field + " " + string(v.Code)
	}
	return "invalid input: " + strings.Join(parts, ", ")
}

// Violations collects broken rules while an input is checked.
type Violations []Violation

// Add records that field broke the rule named by code.
func (vs *Violations) Add(field string, code Code) {
	*vs = append(*vs, Violation{Field: field, Code: code})
}

// Err is nil when no rule was broken, and a *ValidationError otherwise.
func (vs Violations) Err() error {
	if len(vs) == 0 {
		return nil
	}
	return &ValidationError{Violations: vs}
}

// ConflictError is a rule about other records, such as a name that is already
// used. It matches ErrConflict, and Field says which input clashed.
type ConflictError struct {
	Field string
	Code  Code
}

func (e *ConflictError) Error() string {
	return "conflict: " + e.Field + " " + string(e.Code)
}

// Is makes errors.Is(err, ErrConflict) hold for every *ConflictError.
func (e *ConflictError) Is(target error) bool {
	return target == ErrConflict
}

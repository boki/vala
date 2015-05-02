/*
Package vala is a simple, extensible, library to make argument
validation in Go palatable.

This package uses the fluent programming style to provide
simultaneously more robust and more terse parameter validation.

	Begin().Validate(
		NotNil(a, "a"),
		NotNil(b, "b"),
		NotNil(c, "c"),
	).CheckAndPanic().Validate( // Panic will occur here if a, b, or c are nil.
		Rng(a.Items, 50, 70, "a.Items"),
		Gt(b.UserCount, 0, "b.UserCount"),
		Eq(c.Name, "Vala", "c.name"),
		Not(Eq(c.FriendlyName, "Foo", "c.FriendlyName"), "!Eq"),
	).Check()

Notice how checks can be tiered.

Vala is also extensible. As long as a function conforms to the Checker
specification, you can pass it into the Validate method:

	func ReportFitsRepository(report *Report, repository *Repository) Checker {
		return func() *CheckerError {
			if repository.Type != report.Type {
				return fmt.Errorf("A %s report does not belong in a %s repository.", report.Type, repository.Type)
			}
			return nil
		}
	}

	func AuthorCanUpload(authorName string, repository *Repository) Checker {
		return func() *CheckerError {
			if !repository.AuthorCanUpload(authorName) {
				return fmt.Errorf("%s does not have access to this repository.", authorName)
			}
			return nil
		}
	}

	func AuthorIsCollaborator(authorName string, report *Report) Checker {
		return func() *CheckerError {
			for _, collaboratorName := range report.Collaborators() {
				if collaboratorName == authorName {
					return nil
				}
			}
			return fmt.Errorf("The given author was not one of the collaborators for this report.")
		}
	}

	func HandleReport(authorName string, report *Report, repository *Repository) {
		Begin().Validate(
			AuthorIsCollaborator(authorName, report),
			AuthorCanUpload(authorName, repository),
			ReportFitsRepository(report, repository),
		).CheckAndPanic()
	}
*/
package vala

import (
	"errors"
	"fmt"
	"reflect"
)

// A single validation error
type CheckerError struct {
	// Optional parameter name
	Name string
	Err  error
}

func (err *CheckerError) Error() string {
	if err.Name != "" {
		return fmt.Sprintf("%s: %s", err.Name, err.Err)
	}
	return err.Err.Error()
}

var (
	ErrNot      = errors.New("Not")
	ErrEq       = errors.New("arg1 == arg2")
	ErrNe       = errors.New("arg1 != arg2")
	ErrNotNil   = errors.New("arg != nil")
	ErrRng      = errors.New("min <= arg <= max")
	ErrLt       = errors.New("arg < value")
	ErrLe       = errors.New("arg <= value")
	ErrGt       = errors.New("arg > value")
	ErrGe       = errors.New("arg >= value")
	ErrNotEmpty = errors.New("arg != \"\"")
)

// Validation contains all the errors from performing Checkers, and is
// the fluent type off which all Validation methods hang.
type Validation struct {
	Errors []*CheckerError
}

func validationFactory() *Validation {
	return &Validation{[]*CheckerError{}}
}

func (err *Validation) Error() string {
	if len(err.Errors) > 0 {
		msg := "Parameter validation failed:"
		for _, e := range err.Errors {
			msg += "\n\t" + e.Error()
		}
		return msg
	}
	return ""
}

// Begin begins a validation check.
func Begin() *Validation {
	return nil
}

// Check aggregates all checker errors into a single error and returns
// this error.
func (val *Validation) Check() error {
	if val == nil || len(val.Errors) <= 0 {
		return nil
	}
	return val
}

// CheckAndPanic aggregates all checker errors into a single error and
// panics with this error.
func (val *Validation) CheckAndPanic() *Validation {
	if val == nil || len(val.Errors) <= 0 {
		return val
	}
	panic(val)
}

// CheckSetErrorAndPanic aggregates any Errors produced by the
// Checkers into a single error, and sets the address of retError to
// this, and panics. The canonical use-case of this is to pass in the
// address of an error you would like to return, and then to catch the
// panic and do nothing.
func (val *Validation) CheckSetErrorAndPanic(retError *error) *Validation {
	if val == nil || len(val.Errors) <= 0 {
		return val
	}
	*retError = val
	panic(*retError)
}

// Validate runs all of the checkers passed in and collects errors
// into an internal collection. To take action on these errors, call
// one of the Check* methods.
func (val *Validation) Validate(checkers ...Checker) *Validation {
	for _, checker := range checkers {
		if err := checker(); err != nil {
			if val == nil {
				val = validationFactory()
			}
			val.Errors = append(val.Errors, err)
		}
	}
	return val
}

//
// Checker functions
//

// Checker defines the type of function which can represent a Vala checker.
type Checker func() *CheckerError

func newCheckerError(nameOrErr interface{}, def error) *CheckerError {
	if name, ok := nameOrErr.(string); ok {
		return &CheckerError{Name: name, Err: def}
	} else if ce, ok := nameOrErr.(*CheckerError); ok {
		return ce
	}
	return &CheckerError{Err: nameOrErr.(error)}
}

// Not returns the inverse of any Checker passed in. nameOrErr specifies the name
// of the parameter or a custom error.
func Not(checker Checker, nameOrErr interface{}) Checker {
	return func() *CheckerError {
		if err := checker(); err == nil {
			return newCheckerError(nameOrErr, ErrNot)
		}
		return nil
	}
}

// Eq checks that the arguments pass arg1 == arg2. nameOrErr specifies the name
// of the parameter or a custom error.
func Eq(arg1, arg2 interface{}, nameOrErr interface{}) Checker {
	return func() *CheckerError {
		if arg1 == arg2 {
			return nil
		}
		return newCheckerError(nameOrErr, ErrEq)
	}
}

// Ne checks that the arguments pass arg1 != arg2. nameOrErr specifies the name
// of the parameter or a custom error.
func Ne(arg1, arg2 interface{}, nameOrErr interface{}) Checker {
	return func() *CheckerError {
		if arg1 != arg2 {
			return nil
		}
		return newCheckerError(nameOrErr, ErrNe)
	}
}

// NotNil checks to see if the value passed in is nil. This Checker
// attempts to check the most performant things first, and then
// degrade into the less-performant, but accurate checks for nil. nameOrErr
// specifies the name of the parameter or a custom error.
func NotNil(arg interface{}, nameOrErr interface{}) Checker {
	return func() *CheckerError {
		isNotNil := true
		if arg == nil {
			isNotNil = false
		} else if str, ok := arg.(string); ok {
			isNotNil = str != ""
		} else {
			switch v := reflect.ValueOf(arg); v.Kind() {
			case
				reflect.Chan,
				reflect.Func,
				reflect.Interface,
				reflect.Map,
				reflect.Ptr,
				reflect.Slice:
				isNotNil = !v.IsNil()
			default:
				panic("Vala is unable to check this type for nilability at this time.")
			}
		}
		if !isNotNil {
			return newCheckerError(nameOrErr, ErrNotNil)
		}
		return nil
	}
}

// Rng checks that the given argument is in the desired range. nameOrErr
// specifies the name of the parameter or a custom error.
func Rng(arg int, min, max int, nameOrErr interface{}) Checker {
	return func() *CheckerError {
		len := arg
		if len < min || len > max {
			return newCheckerError(nameOrErr, ErrRng)
		}
		return nil
	}
}

// Lt checks that the given argument is less than the given value. nameOrErr
// specifies the name of the parameter or a custom error.
func Lt(arg int, value int, nameOrErr interface{}) Checker {
	return func() *CheckerError {
		if arg >= value {
			return newCheckerError(nameOrErr, ErrLt)
		}
		return nil
	}
}

// Le checks that the given argument is less than or equal to the given value.
// nameOrErr specifies the name of the parameter or a custom error.
func Le(arg int, value int, nameOrErr interface{}) Checker {
	return func() *CheckerError {
		if arg > value {
			return newCheckerError(nameOrErr, ErrLe)
		}
		return nil
	}
}

// Gt checks that the given argument is greater than the given value.
// nameOrErr specifies the name of the parameter or a custom error.
func Gt(arg int, value int, nameOrErr interface{}) Checker {
	return func() *CheckerError {
		if arg <= value {
			return newCheckerError(nameOrErr, ErrGt)
		}
		return nil
	}
}

// Ge checks that the given argument is greater than the given value.
// nameOrErr specifies the name of the parameter or a custom error.
func Ge(arg int, value int, nameOrErr interface{}) Checker {
	return func() *CheckerError {
		if arg < value {
			return newCheckerError(nameOrErr, ErrGe)
		}
		return nil
	}
}

// NotEmpty checks that the given string is not empty.
// nameOrErr specifies the name of the parameter or a custom error.
func NotEmpty(arg, nameOrErr interface{}) Checker {
	return func() *CheckerError {
		if arg == "" {
			return newCheckerError(nameOrErr, ErrNotEmpty)
		}
		return nil
	}
}

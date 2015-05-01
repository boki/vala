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
		Len(a.Items, 50, 70, "a.Items"),
		Gt(b.UserCount, 0, "b.UserCount"),
		Eq(c.Name, "Vala", "c.name"),
		Not(Eq(c.FriendlyName, "Foo", "c.FriendlyName")),
	).Check()

Notice how checks can be tiered.

Vala is also extensible. As long as a function conforms to the Checker
specification, you can pass it into the Validate method:

	func ReportFitsRepository(report *Report, repository *Repository) Checker {
		return func() (passes bool, err error) {

			err = fmt.Errorf("A %s report does not belong in a %s repository.", report.Type, repository.Type)
			passes = (repository.Type == report.Type)
			return passes, err
		}
	}

	func AuthorCanUpload(authorName string, repository *Repository) Checker {
		return func() (passes bool, err error) {
			err = fmt.Errorf("%s does not have access to this repository.", authorName)
			passes = !repository.AuthorCanUpload(authorName)
			return passes, err
		}
	}

	func AuthorIsCollaborator(authorName string, report *Report) Checker {
		return func() (passes bool, err error) {

			err = fmt.Errorf("The given author was not one of the collaborators for this report.")
			for _, collaboratorName := range report.Collaborators() {
				if collaboratorName == authorName {
					passes = true
					break
				}
			}

			return passes, err
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
	"fmt"
	"reflect"
	"strings"
)

func validationFactory(numErrors int) *Validation {
	return &Validation{make([]string, numErrors)}
}

// Validation contains all the errors from performing Checkers, and is
// the fluent type off which all Validation methods hang.
type Validation struct {
	Errors []string
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

	return val.constructErrorMessage()
}

// CheckAndPanic aggregates all checker errors into a single error and
// panics with this error.
func (val *Validation) CheckAndPanic() *Validation {
	if val == nil || len(val.Errors) <= 0 {
		return val
	}

	panic(val.constructErrorMessage())
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

	*retError = val.constructErrorMessage()
	panic(*retError)
}

// Validate runs all of the checkers passed in and collects errors
// into an internal collection. To take action on these errors, call
// one of the Check* methods.
func (val *Validation) Validate(checkers ...Checker) *Validation {

	for _, checker := range checkers {
		if pass, msg := checker(); !pass {
			if val == nil {
				val = validationFactory(1)
			}

			val.Errors = append(val.Errors, msg)
		}
	}

	return val
}

func (val *Validation) constructErrorMessage() error {
	return fmt.Errorf(
		"parameter validation failed:\t%s",
		strings.Join(val.Errors, "\n\t"),
	)
}

//
// Checker functions
//

// Checker defines the type of function which can represent a Vala
// checker.  If the Checker fails, returns false with a corresponding
// error message. If the Checker succeeds, returns true, but _also_
// returns an error message. This helps to support the Not function.
type Checker func() (checkerIsTrue bool, errorMessage string)

// Not returns the inverse of any Checker passed in.
func Not(checker Checker) Checker {

	return func() (passed bool, errorMessage string) {
		if passed, errorMessage = checker(); passed {
			return false, fmt.Sprintf("Not(%s)", errorMessage)
		}

		return true, ""
	}
}

// Eq performs a basic == on the given parameters and fails if
// they are not equal.
func Eq(lhs, rhs interface{}, paramName string) Checker {

	return func() (pass bool, errMsg string) {
		return (lhs == rhs), fmt.Sprintf("Parameters were not equal: %s %v, %v", paramName, lhs, rhs)
	}
}

// Ne performs a basic != on the given parameters and fails if they are equal.
func Ne(lhs, rhs interface{}, paramName string) Checker {

	return func() (isNe bool, errMsg string) {
		return lhs != rhs, fmt.Sprintf("Parameters were equal: %s %v == %v", paramName, lhs, rhs)
	}
}

// NotNil checks to see if the value passed in is nil. This Checker
// attempts to check the most performant things first, and then
// degrade into the less-performant, but accurate checks for nil.
func NotNil(obtained interface{}, paramName string) Checker {
	return func() (isNotNil bool, errMsg string) {

		if obtained == nil {
			isNotNil = false
		} else if str, ok := obtained.(string); ok {
			isNotNil = str != ""
		} else {
			switch v := reflect.ValueOf(obtained); v.Kind() {
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

		return isNotNil, "Parameter was nil: " + paramName
	}
}

// Len checks to ensure the given argument is in the desired length.
func Len(param interface{}, minLength, maxLength int, paramName string) Checker {

	return func() (hasLen bool, errMsg string) {
		len := reflect.ValueOf(param).Len()
		hasLen = minLength <= len && len <= maxLength
		return hasLen, "Parameter did not contain the correct number of elements: " + paramName
	}
}

// Lt checks to ensure the given argument is less than the given value.
func Lt(param int, comparativeVal int, paramName string) Checker {

	return func() (isLt bool, errMsg string) {
		if isLt = param < comparativeVal; !isLt {
			errMsg = fmt.Sprintf(
				"Parameter was not less than:  %s(%d) >= %d",
				paramName,
				param,
				comparativeVal)
		}

		return isLt, errMsg
	}
}

// Le checks to ensure the given argument is less than or equal to the given value.
func Le(param int, comparativeVal int, paramName string) Checker {

	return func() (isLe bool, errMsg string) {
		if isLe = param <= comparativeVal; !isLe {
			errMsg = fmt.Sprintf(
				"Parameter was not less than or equal to:  %s(%d) > %d",
				paramName,
				param,
				comparativeVal)
		}

		return isLe, errMsg
	}
}

// Gt checks to ensure the given argument is greater than the
// given value.
func Gt(param int, comparativeVal int, paramName string) Checker {

	return func() (isGt bool, errMsg string) {
		if isGt = param > comparativeVal; !isGt {
			errMsg = fmt.Sprintf(
				"Parameter was not greater than:  %s(%d) <= %d",
				paramName,
				param,
				comparativeVal)
		}

		return isGt, errMsg
	}
}

// Ge checks to ensure the given argument is greater than the
// given value.
func Ge(param int, comparativeVal int, paramName string) Checker {

	return func() (isGe bool, errMsg string) {
		if isGe = param >= comparativeVal; !isGe {
			errMsg = fmt.Sprintf(
				"Parameter was not greater than or equal to:  %s(%d) < %d",
				paramName,
				param,
				comparativeVal)
		}

		return isGe, errMsg
	}
}

// NotEmpty checks to ensure the given string is not empty.
func NotEmpty(obtained, paramName string) Checker {
	return func() (isNotEmpty bool, errMsg string) {
		isNotEmpty = obtained != ""
		errMsg = fmt.Sprintf("Parameter is an empty string: %s", paramName)
		return
	}
}

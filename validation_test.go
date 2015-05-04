package vala

import (
	"errors"
	"strconv"
	"testing"
)

var (
	myErr = errors.New("My custom error")
)

func TestError(t *testing.T) {
	v := &Validation{[]error{}}

	err := v.Error()
	if err != "" {
		t.Errorf("Received an unexpected error message: %v", err)
	}

	cerr := &CheckerError{"Name", ErrNe}
	v.Errors = append(v.Errors, cerr)
	err = v.Error()
	if err == "" {
		t.Errorf("Expected an error message.")
	}
}

func TestNewCheckerError(t *testing.T) {
	def := ErrNotEmpty
	name := "Test"
	err := newCheckerError(name, def)
	if got := err.(*CheckerError).Err; got != ErrNotEmpty {
		t.Errorf("Expected %v; got %v", ErrNotEmpty, got)
	}
	if got, expected := err.Error(), "Test: arg != \"\""; got != expected {
		t.Errorf("Expected %v; got %v", expected, got)
	}

	customErr := ErrNotNil
	err = newCheckerError(customErr, def)
	if got := err.(*CheckerError).Err; got != ErrNotNil {
		t.Errorf("Expected %v; got %v", ErrNotNil, got)
	}
	if got, expected := err.Error(), "arg != nil"; got != expected {
		t.Errorf("Expected %v; got %v", expected, got)
	}

	customErr2 := &CheckerError{"Test", ErrNe}
	err = newCheckerError(customErr2, def)
	if err != customErr2 {
		t.Errorf("Expected %v; got %v", customErr2, err)
	}
}

func TestCheckerErrorError(t *testing.T) {
	err := &CheckerError{"Test", ErrNotNil}
	if got, expected := err.Error(), "Test: arg != nil"; got != expected {
		t.Errorf("Expected %v; got %v", expected, got)
	}

	err = &CheckerError{Err: ErrNotNil}
	if got, expected := err.Error(), "arg != nil"; got != expected {
		t.Errorf("Expected %v; got %v", expected, got)
	}
}

func TestPanicIsIssued(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.FailNow()
		}
	}()

	Begin().Validate(
		Eq("foo", "bar", "foo"),
	).CheckAndPanic()
}

func TestErrorReturns(t *testing.T) {
	err := Begin().Validate(
		Eq("foo", "bar", "foo"),
	).Check()
	if err == nil {
		t.FailNow()
	}
}

func TestSetError(t *testing.T) {
	var returnErr error
	defer func() {
		if r := recover(); r == nil {
			t.FailNow()
		}

		if returnErr == nil {
			t.FailNow()
		}
	}()

	Begin().Validate(
		Eq("foo", "bar", "foo"),
	).CheckSetErrorAndPanic(&returnErr)

	t.Error("We should have never reached this.")
	t.FailNow()
}

func TestNot(t *testing.T) {
	err := Begin().Validate(
		Not(Eq("foo", "bar", "foo"), "!Eq"),
	).Check()
	if err != nil {
		t.Error("Received an unexpected error.")
		t.FailNow()
	}

	err = Begin().Validate(
		Not(Eq("foo", "foo", "varName"), "!Eq"),
	).Check()
	if err == nil {
		t.Error("Expected an error.")
		t.Fail()
	}
}

func TestEq(t *testing.T) {
	err := Begin().Validate(
		Eq("foo", "bar", "foo"),
	).Check()
	if err == nil {
		t.Errorf("Expected error")
		t.FailNow()
	}

	err = Begin().Validate(
		Eq("foo", "foo", "foo"),
	).Check()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		t.FailNow()
	}
}

func TestNe(t *testing.T) {
	err := Begin().Validate(
		Ne("foo", "foo", "foo"),
	).Check()
	if err == nil {
		t.Errorf("Unexpected error: %v", err)
		t.FailNow()
	}

	err = Begin().Validate(
		Ne("foo", "bar", "foo"),
	).Check()
	if err != nil {
		t.Errorf("Expected error")
		t.FailNow()
	}
}

func TestNotNil(t *testing.T) {
	defer func() {
		recover()
	}()

	err := Begin().Validate(
		NotNil("foo", "foo"),
		NotNil(t, "t"),
	).Check()
	if err != nil {
		t.Errorf("Received an unexpected error: %v", err)
		t.FailNow()
	}

	var nilSlice []string
	err = Begin().Validate(
		Not(NotNil(nil, "foo"), "!NotNil"),
		Not(NotNil(nilSlice, "nilSlice"), "!NotNil"),
	).Check()
	if err != nil {
		t.Errorf("Received an unexpected error: %v", err)
		t.FailNow()
	}

	var zeroInt int
	err = Begin().Validate(
		Not(NotNil(zeroInt, "zeroInt"), "!NotNil"),
	).Check()
	t.Fatal("Should have paniced")
}

func TestLen(t *testing.T) {
	err := Begin().Validate(
		Rng(len([]int{1, 2}), 2, 2, "tmpA"),
		Rng(len([]int{}), 0, 0, "tmpB"),
		Rng(len("1"), 1, 1, "tmpC"),
	).Check()
	if err != nil {
		t.Fatal("Received an unexpected error: %v", err)
	}

	err = Begin().Validate(
		Rng(0, 1, 1, "tmpC"),
	).Check()
	if err == nil {
		t.Fatal("Expected an error.")
	}

	err = Begin().Validate(
		Rng(3, 2, 5, "tmpC"),
	).Check()
	if err != nil {
		t.Fatal("Received an unexpected error: %v", err)
	}
}

func TestLt(t *testing.T) {
	err := Begin().Validate(
		Lt(0, 1, "tmpA"),
	).Check()
	if err != nil {
		t.Fatalf("Received an unexpected error: %v", err)
	}

	err = Begin().Validate(
		Lt(1, 0, "tmpC"),
	).Check()
	if err == nil {
		t.Fatal("Expected an error.")
	}
}

func TestLe(t *testing.T) {
	err := Begin().Validate(
		Le(0, 1, "tmpA"),
		Le(1, 1, "tmpA"),
	).Check()
	if err != nil {
		t.Fatalf("Received an unexpected error: %v", err)
	}

	err = Begin().Validate(
		Le(2, 1, "tmpC"),
	).Check()
	if err == nil {
		t.Fatalf("Expected an error.")
	}
}

func TestGt(t *testing.T) {
	err := Begin().Validate(
		Gt(1, 0, "tmpA"),
	).Check()
	if err != nil {
		t.Fatalf("Received an unexpected error: %v", err)
	}

	err = Begin().Validate(
		Gt(0, 1, "tmpC"),
	).Check()
	if err == nil {
		t.Fatalf("Expected an error.")
	}
}

func TestGe(t *testing.T) {
	err := Begin().Validate(
		Ge(1, 1, "tmpA"),
		Ge(2, 1, "tmpA"),
	).Check()
	if err != nil {
		t.Fatalf("Received an unexpected error: %v", err)
	}

	err = Begin().Validate(
		Ge(0, 1, "tmpC"),
	).Check()
	if err == nil {
		t.Fatalf("Expected an error.")
	}
}

func TestBool(t *testing.T) {
	err := Begin().Validate(
		Bool("", "empty"),
		Bool("a", "syntax"),
	).Check()
	if err == nil {
		t.Fatalf("Expected an error.")
	}
	if got, expected := len(err.(*Validation).Errors), 2; got != expected {
		t.Fatalf("Expected %v errors; got %v", expected, got)
	}
	if got, expected := err.(*Validation).Errors[0].(*CheckerError).Err, ErrBool; got != expected {
		t.Errorf("Expected %v; got %v", expected, got)
	}
	if got, expected := err.(*Validation).Errors[1].(*CheckerError).Err, ErrBool; got != expected {
		t.Errorf("Expected %v; got %v", expected, got)
	}

	err = Begin().Validate(
		Bool("1", "1"),
		Bool("True", "True"),
		Bool("f", "f"),
	).Check()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestInt(t *testing.T) {
	err := Begin().Validate(
		Int("", 8, "empty"),
		Int("128", 8, "out of range"),
		Int("0xabcd", 32, "syntax"),
	).Check()
	if err == nil {
		t.Fatalf("Expected an error.")
	}
	if got, expected := len(err.(*Validation).Errors), 3; got != expected {
		t.Fatalf("Expected %v errors; got %v", expected, got)
	}
	if got, expected := err.(*Validation).Errors[0].(*CheckerError).Err, strconv.ErrSyntax; got != expected {
		t.Errorf("Expected %v; got %v", expected, got)
	}
	if got, expected := err.(*Validation).Errors[1].(*CheckerError).Err, strconv.ErrRange; got != expected {
		t.Errorf("Expected %v; got %v", expected, got)
	}
	if got, expected := err.(*Validation).Errors[2].(*CheckerError).Err, strconv.ErrSyntax; got != expected {
		t.Errorf("Expected %v; got %v", expected, got)
	}

	err = Begin().Validate(
		Int("127", 8, "a"),
		Int("-127", 8, "b"),
		Int("128", 32, "c"),
	).Check()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestUint(t *testing.T) {
	err := Begin().Validate(
		Uint("", 8, "empty"),
		Uint("-127", 8, "syntax"),
		Uint("256", 8, "range"),
	).Check()
	if err == nil {
		t.Fatalf("Expected an error.")
	}
	if got, expected := len(err.(*Validation).Errors), 3; got != expected {
		t.Fatalf("Expected %v errors; got %v", expected, got)
	}
	if got, expected := err.(*Validation).Errors[0].(*CheckerError).Err, strconv.ErrSyntax; got != expected {
		t.Errorf("Expected %v; got %v", expected, got)
	}
	if got, expected := err.(*Validation).Errors[1].(*CheckerError).Err, strconv.ErrSyntax; got != expected {
		t.Errorf("Expected %v; got %v", expected, got)
	}
	if got, expected := err.(*Validation).Errors[2].(*CheckerError).Err, strconv.ErrRange; got != expected {
		t.Errorf("Expected %v; got %v", expected, got)
	}

	err = Begin().Validate(
		Uint("127", 8, "a"),
		Uint("255", 8, "b"),
		Uint("128", 32, "c"),
	).Check()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestFloat(t *testing.T) {
	err := Begin().Validate(
		Float("", 32, "empty"),
		Float("1.2.", 32, "syntax"),
		Float("1.7976931348623159e308", 64, "range"),
		Float("1,2345", 64, "syntax"),
	).Check()
	if err == nil {
		t.Fatalf("Expected an error.")
	}
	if got, expected := len(err.(*Validation).Errors), 4; got != expected {
		t.Fatalf("Expected %v errors; got %v", expected, got)
	}
	if got, expected := err.(*Validation).Errors[0].(*CheckerError).Err, strconv.ErrSyntax; got != expected {
		t.Errorf("Expected %v; got %v", expected, got)
	}
	if got, expected := err.(*Validation).Errors[1].(*CheckerError).Err, strconv.ErrSyntax; got != expected {
		t.Errorf("Expected %v; got %v", expected, got)
	}
	if got, expected := err.(*Validation).Errors[2].(*CheckerError).Err, strconv.ErrRange; got != expected {
		t.Errorf("Expected %v; got %v", expected, got)
	}
	if got, expected := err.(*Validation).Errors[3].(*CheckerError).Err, strconv.ErrSyntax; got != expected {
		t.Errorf("Expected %v; got %v", expected, got)
	}

	err = Begin().Validate(
		Float("1.2", 32, "a"),
		Float("3", 32, "b"),
		Float("1234567890.123", 32, "c"),
	).Check()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestNotEmpty(t *testing.T) {
	err := Begin().Validate(
		NotEmpty("", "tmpA"),
	).Check()
	if err == nil {
		t.Fatalf("Expected an error.")
	}

	err = Begin().Validate(
		NotEmpty("abc", "tmpA"),
	).Check()
	if err != nil {
		t.Fatalf("Received an unexpected error: %v", err)
	}

	err = Begin().Validate(
		NotEmpty("", myErr),
	).Check()
	if err == nil {
		t.Fatalf("Expected an error")
	}
	if got := err.(*Validation).Errors[0].(*CheckerError).Err; got != myErr {
		t.Fatalf("Expected %v; got %v", myErr, got)
	}
}

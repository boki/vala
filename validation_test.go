package vala

import (
	"errors"
	"testing"
)

var (
	myErr = errors.New("My custom error")
)

func TestError(t *testing.T) {
	v := &Validation{[]*CheckerError{}}

	err := v.Error()
	if err != "" {
		t.Errorf("Received an unexpected error message: %v", err)
	}

	v.Errors = append(v.Errors, &CheckerError{"Name", ErrNe})
	err = v.Error()
	if err == "" {
		t.Errorf("Expected an error message.")
	}
}

func TestNewCheckerError(t *testing.T) {
	def := ErrNotEmpty
	name := "Test"
	err := newCheckerError(name, def)
	if got := err.Err; got != ErrNotEmpty {
		t.Errorf("Expected %v; got %v", ErrNotEmpty, got)
	}
	if got, expected := err.Error(), "Test: arg != \"\""; got != expected {
		t.Errorf("Expected %v; got %v", expected, got)
	}

	customErr := ErrNotNil
	err = newCheckerError(customErr, def)
	if got := err.Err; got != ErrNotNil {
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
		Len([]int{1, 2}, 2, 2, "tmpA"),
		Len([]int{}, 0, 0, "tmpB"),
		Len("1", 1, 1, "tmpC"),
	).Check()
	if err != nil {
		t.Fatal("Received an unexpected error: %v", err)
	}

	err = Begin().Validate(
		Len("", 1, 1, "tmpC"),
	).Check()
	if err == nil {
		t.Fatal("Expected an error.")
	}

	err = Begin().Validate(
		Len("abc", 2, 5, "tmpC"),
	).Check()
	if err != nil {
		t.Fatal("Received an unexpected error: %v", err)
	}

	err = Begin().Validate(
		Len("abc", 4, 7, "tmpC"),
	).Check()
	if err == nil {
		t.Fatal("Expected an error.")
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
	if got := err.(*Validation).Errors[0].Err; got != myErr {
		t.Fatalf("Expected %v; got %v", myErr, got)
	}
}

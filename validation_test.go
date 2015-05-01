package vala

import (
	"testing"
)

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
		Not(Eq("foo", "bar", "foo")),
	).Check()

	if err != nil {
		t.Error("Received an unexpected error.")
		t.FailNow()
	}

	err = Begin().Validate(
		Not(Eq("foo", "foo", "varName")),
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
		t.FailNow()
	}

	err = Begin().Validate(
		Eq("foo", "foo", "foo"),
	).Check()

	if err != nil {
		t.FailNow()
	}
}

func TestIsNil(t *testing.T) {

	err := Begin().Validate(
		NotNil("foo", "foo"),
		NotNil(t, "t"),
	).Check()

	if err != nil {
		t.Error("Received an unexpected error.")
		t.FailNow()
	}

	var nilSlice []string

	err = Begin().Validate(
		Not(NotNil(nil, "foo")),
		Not(NotNil(nilSlice, "nilSlice")),
	).Check()

	if err != nil {
		t.Errorf("Received an unexpected error: %v", err)
		t.FailNow()
	}
}

func TestLen(t *testing.T) {

	err := Begin().Validate(
		Len([]int{1, 2}, 2, 2, "tmpA"),
		Len([]int{}, 0, 0, "tmpB"),
		Len("1", 1, 1, "tmpC"),
	).Check()

	if err != nil {
		t.Errorf("Received an unexpected error: %v", err)
		t.FailNow()
	}

	err = Begin().Validate(
		Len("", 1, 1, "tmpC"),
	).Check()

	if err == nil {
		t.Errorf("Expected an error.")
		t.FailNow()
	}

	err = Begin().Validate(
		Len("abc", 2, 5, "tmpC"),
	).Check()

	if err != nil {
		t.Errorf("Received an unexpected error: %v", err)
		t.FailNow()
	}

	err = Begin().Validate(
		Len("abc", 4, 7, "tmpC"),
	).Check()

	if err == nil {
		t.Errorf("Expected an error.")
		t.FailNow()
	}
}

func TestGt(t *testing.T) {

	err := Begin().Validate(
		Gt(1, 0, "tmpA"),
	).Check()

	if err != nil {
		t.Errorf("Received an unexpected error: %v", err)
		t.FailNow()
	}

	err = Begin().Validate(
		Gt(0, 1, "tmpC"),
	).Check()

	if err == nil {
		t.Errorf("Expected an error.")
		t.FailNow()
	}
}

func TestNotEmpty(t *testing.T) {

	err := Begin().Validate(
		NotEmpty("", "tmpA"),
	).Check()

	if err == nil {
		t.Errorf("Expected an error.")
		t.FailNow()
	}
}

vala [![GoDoc](https://godoc.org/github.com/boki/vala?status.svg)](https://godoc.org/github.com/boki/vala)
====

A simple, extensible, library to make argument validation in Go palatable.

Instead of this:

```go
func BoringValidation(a, b, c, d, e, f, g MyType) {
  if (a == nil)
    panic("a is nil")
  if (b == nil)
    panic("b is nil")
  if (c == nil)
    panic("c is nil")
  if (d == nil)
    panic("d is nil")
  if (e == nil)
    panic("e is nil")
  if (f == nil)
    panic("f is nil")
  if (g == nil)
    panic("g is nil")
}
```

Do this:

```go
func ClearValidation(a, b, c, d, e, f, g MyType) {
  Begin().Validate(
    NotNil(a, "a"),
    NotNil(b, "b"),
    NotNil(c, "c"),
    NotNil(d, "d"),
    NotNil(e, "e"),
    NotNil(f, "f"),
    NotNil(g, "g"),
  ).CheckAndPanic() // All values will get checked before an error is thrown!
}
```

Instead of this:

```go
func BoringValidation(a, b, c, d, e, f, g MyType) error {
  if (a == nil)
    return fmt.Errorf("a is nil")
  if (b == nil)
    return fmt.Errorf("b is nil")
  if (c == nil)
    return fmt.Errorf("c is nil")
  if (d == nil)
    return fmt.Errorf("d is nil")
  if (e == nil)
    return fmt.Errorf("e is nil")
  if (f == nil)
    return fmt.Errorf("f is nil")
  if (g == nil)
    return fmt.Errorf("g is nil")
}
```

Do this:

```go
func ClearValidation(a, b, c, d, e, f, g MyType) (err error) {
  defer func() { recover() }
  Begin().Validate(
    NotNil(a, "a"),
    NotNil(b, "b"),
    NotNil(c, "c"),
    NotNil(d, "d"),
    NotNil(e, "e"),
    NotNil(f, "f"),
    NotNil(g, "g"),
  ).CheckSetErrorAndPanic(&err) // Return error will get set, and the function will return.

  // ...

  VeryExpensiveFunction(c, d)
}
```

Tier your validation:

```go
func ClearValidation(a, b, c MyType) (err error) {
  err = Begin().Validate(
    NotNil(a, "a"),
    NotNil(b, "b"),
    NotNil(c, "c"),
  ).CheckAndPanic().Validate( // Panic will occur here if a, b, or c are nil.
    Rng(len(a.Items), 50, 50, "a.Items"),
    Gt(b.UserCount, 0, "b.UserCount"),
    Eq(c.Name, "Vala", "c.name"),
    Not(Eq(c.FriendlyName, "Foo", "c.FriendlyName"), "!Eq"),
  ).Check()

  if err != nil {
    return err
  }

  // ...

  VeryExpensiveFunction(c, d)
}
```

The `nameOrErr` parameter to the default checker functions allows you to either specify the parameters name or to provide a custom error to be used instead of the default error values:

```go
a, b := nil, nil
err := Begin().Validate(
  NotNil(a, ErrANotNil),
  NotNil(b, "b"),
).Check()
```

`err` will have a value of:

```go
Validation{
  Errors: []*CheckerError{
    &CheckerError{Name: "", Err: ErrANotNil},
    &CheckerError{Name: "b", Err: ErrNotNil},
  }
}
```

Extend with your own validators for readability. Note that an error should always be returned so that the Not function can return a message if it passes. Unlike idiomatic Go, use the boolean to check for success.

```go
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
```

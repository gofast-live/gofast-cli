# Coding Conventions

## Error Handling

- Always declare the `err` variable on a separate line before the `if` statement that checks for an error. This improves readability and simplifies debugging.
- When handling errors, wrap the original error with additional context using `fmt.Errorf` and the `%w` verb. If you are creating a new error from scratch, use `errors.New`.

**Good (wrapping):**
```go
err := doSomething()
if err != nil {
    return fmt.Errorf("error doing something: %w", err)
}
```

**Good (new error):**
```go
if somethingIsWrong {
    return errors.New("something is wrong")
}
```

**Bad:**
```go
if err := doSomething(); err != nil {
    // handle error
}
```

## TypeScript/JavaScript Function Style

- Always use function declarations instead of function expressions. This ensures consistency and improves readability.

**Good:**
```typescript
function myFunction() {
  // ...
}
```

**Bad:**
```typescript
const myFunction = () => {
  // ...
};
```

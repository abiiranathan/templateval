# templateval - Go Template Validation Library

![Go](https://github.com/abiiranathan/templateval/workflows/Go/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/abiiranathan/templateval)](https://goreportcard.com/report/github.com/abiiranathan/templateval)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`templateval` is a comprehensive Go library for validating template variables with strong typing support. 
It provides flexible validation for all Go types including basic types, structs, interfaces, functions, slices, maps, and more.

## Features

- **Type-safe validation** for all Go types
- **Custom validators** for complex validation logic
- **Required/Optional** field specification
- **Partial validation** for incomplete data sets
- **Template registry** for reusable validation rules
- **Concurrent-safe** implementation
- **Detailed error reporting** with field-specific messages
- **Comprehensive type compatibility** checks

## Installation

```bash
go get github.com/abiiranathan/templateval
```

## Quick Start

```go
package main

import (
	"fmt"
	"github.com/abiiranathan/templateval"
)

func main() {
	// Create a new validator
	validator := templateval.NewValidator().
		RequiredString("name", func(v any) error {
			if s, ok := v.(string); ok && len(s) < 3 {
				return fmt.Errorf("must be at least 3 characters")
			}
			return nil
		}).
		RequiredInt("age", func(v any) error {
			if i, ok := v.(int); ok && i < 18 {
				return fmt.Errorf("must be at least 18")
			}
			return nil
		}).
		OptionalString("email")

	// Validate data
	err := validator.Validate(map[string]any{
		"name": "John Doe",
		"age":  25,
	})

	if err != nil {
		fmt.Println("Validation failed:", err)
	} else {
		fmt.Println("Validation passed!")
	}
}
```

## Core Concepts

### Validator

The `Validator` is the main component that holds validation rules. You can:

- Add rules for different types of fields
- Validate complete or partial data sets
- Manage rules dynamically

### Validation Rules

Each rule specifies:

- Field name
- Expected type
- Required/Optional status
- Optional custom validation function

### Registry

The `Registry` allows storing and reusing validators under named templates.

## API Documentation

### Validator Methods

#### Basic Type Validators

- `RequiredString(name string, validators ...func(any) error) *Validator`
- `OptionalString(name string, validators ...func(any) error) *Validator`
- `RequiredInt(name string, validators ...func(any) error) *Validator`
- `OptionalInt(name string, validators ...func(any) error) *Validator`
- Similar methods for other basic types: Bool, Float, etc.

#### Complex Type Validators

- `RequiredOf(name string, sample any) *Validator`
- `OptionalOf(name string, sample any) *Validator`
- `RequiredSlice(name string, elemType reflect.Type) *Validator`
- `RequiredMap(name string, keyType, valueType reflect.Type) *Validator`
- `RequiredStruct(name string, typ reflect.Type) *Validator`
- `RequiredInterface(name string, typ reflect.Type) *Validator`
- `RequiredFunc(name string, typ reflect.Type) *Validator`
- `RequiredPointer(name string, typ reflect.Type) *Validator`
- `RequiredChan(name string, typ reflect.Type) *Validator`
- Optional variants for all above

#### Validation

- `Validate(variables map[string]any) error` - Full validation
- `ValidatePartial(variables map[string]any) error` - Partial validation

#### Rule Management

- `HasRule(name string) bool`
- `GetRule(name string) (Rule, bool)`
- `RemoveRule(name string)`
- `ListFields() []string`
- `Clone() *Validator`

### Registry Methods

- `Register(name string, validator *Validator)`
- `Unregister(name string)`
- `HasTemplate(name string) bool`
- `GetValidator(name string) (*Validator, bool)`
- `ListTemplates() []string`
- `Clear()`
- `Validate(name string, variables map[string]any) error`
- `ValidatePartial(name string, variables map[string]any) error`

## Advanced Usage

### Custom Validators

```go
validator := NewValidator().
	RequiredString("username", func(v any) error {
		username := v.(string)
		if len(username) < 5 {
			return fmt.Errorf("must be at least 5 characters")
		}
		if !strings.Contains(username, "_") {
			return fmt.Errorf("must contain an underscore")
		}
		return nil
	})
```

### Interface Validation

```go
type Stringer interface {
	String() string
}

validator := NewValidator().
	RequiredInterface("stringer", reflect.TypeOf((*Stringer)(nil)).Elem())

err := validator.Validate(map[string]any{
	"stringer": myStringerImpl,
})
```

### Function Validation

```go
type Processor func(string) int

validator := NewValidator().
	RequiredFunc("processor", reflect.TypeOf(Processor(nil)))

err := validator.Validate(map[string]any{
	"processor": func(s string) int { return len(s) },
})
```

### Template Registry

```go
registry := NewRegistry()

// Register a user template
userValidator := NewValidator().
	RequiredString("name").
	RequiredInt("age")
registry.Register("user", userValidator)

// Validate using the template
err := registry.Validate("user", map[string]any{
	"name": "Alice",
	"age": 30,
})
```

## Error Handling

Validation errors provide detailed information:

```go
err := validator.Validate(data)
if err != nil {
	if verr, ok := err.(ValidationErrors); ok {
		for _, e := range verr {
			fmt.Printf("Field %s: %s\n", e.Field, e.Message)
		}
	}
}
```

## Performance Considerations

- The library uses reflection but caches type information
- Validators are safe for concurrent use
- Consider reusing validators via the Registry

## Contributing

Contributions are welcome! Please open issues or pull requests on GitHub.

## License

MIT License
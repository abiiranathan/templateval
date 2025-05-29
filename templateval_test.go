package templateval

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

// Test types
type TestUser struct {
	Name  string
	Age   int
	Email string
}

type TestProduct struct {
	ID    int
	Name  string
	Price float64
}

type TestInterface interface {
	TestMethod() string
}

type TestImplementer struct {
	Value string
}

func (t TestImplementer) TestMethod() string {
	return t.Value
}

type TestStringer interface {
	String() string
}

type TestStringImplementer struct {
	Data string
}

func (t TestStringImplementer) String() string {
	return t.Data
}

// Test function types
type TestFunc func(string) int

type TestVarFunc func(string, ...int) bool

// Custom validators
func validatePositiveInt(v any) error {
	if i, ok := v.(int); ok && i <= 0 {
		return errors.New("must be positive")
	}
	return nil
}

func validateNonEmptyString(v any) error {
	if s, ok := v.(string); ok && s == "" {
		return errors.New("must not be empty")
	}
	return nil
}

func TestValidator_AnyType(t *testing.T) {
	validator := NewValidator().
		RequiredOf("User", TestUser{}).
		OptionalOf("Product", TestProduct{}).
		RequiredOf("Time", time.Now()).
		RequiredSlice("StringSlice", reflect.TypeOf("")).
		RequiredMap("StringMap", reflect.TypeOf(""), reflect.TypeOf(0))

	tests := []struct {
		name      string
		variables map[string]any
		wantError bool
	}{
		{
			name: "valid custom types",
			variables: map[string]any{
				"User": TestUser{
					Name:  "John",
					Age:   30,
					Email: "john@test.com",
				},
				"Product": TestProduct{
					ID:    1,
					Name:  "Test Product",
					Price: 99.99,
				},
				"Time":        time.Now(),
				"StringSlice": []string{"a", "b", "c"},
				"StringMap":   map[string]int{"count": 5},
			},
			wantError: false,
		},
		{
			name: "wrong struct type",
			variables: map[string]any{
				"User":        "not a struct", // should be TestUser
				"Time":        time.Now(),
				"StringSlice": []string{"a", "b", "c"},
				"StringMap":   map[string]int{"count": 5},
			},
			wantError: true,
		},
		{
			name: "missing required custom type",
			variables: map[string]any{
				// "User" is missing
				"Time":        time.Now(),
				"StringSlice": []string{"a", "b", "c"},
				"StringMap":   map[string]int{"count": 5},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.variables)
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidator_BasicTypes(t *testing.T) {
	validator := NewValidator().
		RequiredString("name", validateNonEmptyString).
		OptionalString("nickname").
		RequiredInt("age", validatePositiveInt).
		OptionalInt("score").
		RequiredBool("active").
		OptionalBool("verified")

	tests := []struct {
		name      string
		variables map[string]any
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid basic types",
			variables: map[string]any{
				"name":     "John",
				"nickname": "Johnny",
				"age":      25,
				"score":    100,
				"active":   true,
				"verified": false,
			},
			wantError: false,
		},
		{
			name: "missing required fields",
			variables: map[string]any{
				"nickname": "Johnny",
			},
			wantError: true,
		},
		{
			name: "wrong types",
			variables: map[string]any{
				"name":   123,    // should be string
				"age":    "25",   // should be int
				"active": "true", // should be bool
			},
			wantError: true,
		},
		{
			name: "custom validation failure",
			variables: map[string]any{
				"name":   "", // empty string should fail validation
				"age":    -5, // negative should fail validation
				"active": true,
			},
			wantError: true,
		},
		{
			name: "nil values for optional fields",
			variables: map[string]any{
				"name":     "John",
				"nickname": nil, // nil is ok for optional
				"age":      25,
				"score":    nil, // nil is ok for optional
				"active":   true,
				"verified": nil, // nil is ok for optional
			},
			wantError: false,
		},
		{
			name: "nil values for required fields",
			variables: map[string]any{
				"name":   nil, // nil not ok for required
				"age":    25,
				"active": true,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.variables)
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidator_ComplexTypes(t *testing.T) {
	validator := NewValidator().
		RequiredSlice("tags", reflect.TypeOf("")).
		OptionalSlice("categories", reflect.TypeOf(0)).
		RequiredMap("metadata", reflect.TypeOf(""), reflect.TypeOf("")).
		OptionalMap("counters", reflect.TypeOf(""), reflect.TypeOf(0)).
		RequiredPointer("user", reflect.TypeOf(TestUser{})).
		OptionalPointer("product", reflect.TypeOf(TestProduct{})).
		RequiredChan("events", reflect.TypeOf("")).
		OptionalChan("signals", reflect.TypeOf(0))

	userPtr := &TestUser{Name: "John", Age: 30}
	productPtr := &TestProduct{ID: 1, Name: "Test"}
	eventChan := make(chan string)
	signalChan := make(chan int)

	tests := []struct {
		name      string
		variables map[string]any
		wantError bool
	}{
		{
			name: "valid complex types",
			variables: map[string]any{
				"tags":       []string{"go", "test"},
				"categories": []int{1, 2, 3},
				"metadata":   map[string]string{"key": "value"},
				"counters":   map[string]int{"views": 100},
				"user":       userPtr,
				"product":    productPtr,
				"events":     eventChan,
				"signals":    signalChan,
			},
			wantError: false,
		},
		{
			name: "wrong slice element type",
			variables: map[string]any{
				"tags":     []int{1, 2, 3}, // should be []string
				"metadata": map[string]string{"key": "value"},
				"user":     userPtr,
				"events":   eventChan,
			},
			wantError: true,
		},
		{
			name: "wrong map types",
			variables: map[string]any{
				"tags":     []string{"go", "test"},
				"metadata": map[int]string{1: "value"}, // wrong key type
				"user":     userPtr,
				"events":   eventChan,
			},
			wantError: true,
		},
		{
			name: "wrong pointer type",
			variables: map[string]any{
				"tags":     []string{"go", "test"},
				"metadata": map[string]string{"key": "value"},
				"user":     "not a pointer", // should be *TestUser
				"events":   eventChan,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.variables)
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidator_InterfaceTypes(t *testing.T) {
	stringerType := reflect.TypeOf((*TestStringer)(nil)).Elem()
	testInterfaceType := reflect.TypeOf((*TestInterface)(nil)).Elem()

	validator := NewValidator().
		RequiredInterface("stringer", stringerType).
		OptionalInterface("tester", testInterfaceType)

	tests := []struct {
		name      string
		variables map[string]any
		wantError bool
	}{
		{
			name: "valid interface implementations",
			variables: map[string]any{
				"stringer": TestStringImplementer{Data: "test"},
				"tester":   TestImplementer{Value: "test"},
			},
			wantError: false,
		},
		{
			name: "invalid interface implementation",
			variables: map[string]any{
				"stringer": "string doesn't implement TestStringer",
			},
			wantError: true,
		},
		{
			name: "missing required interface",
			variables: map[string]any{
				"tester": TestImplementer{Value: "test"},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.variables)
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidator_FunctionTypes(t *testing.T) {
	testFuncType := reflect.TypeOf(TestFunc(nil))
	varFuncType := reflect.TypeOf(TestVarFunc(nil))

	validator := NewValidator().
		RequiredFunc("processor", testFuncType).
		OptionalFunc("validator", varFuncType)

	testFunc := func(s string) int { return len(s) }
	varFunc := func(s string, nums ...int) bool { return len(nums) > 0 }

	tests := []struct {
		name      string
		variables map[string]any
		wantError bool
	}{
		{
			name: "valid function types",
			variables: map[string]any{
				"processor": testFunc,
				"validator": varFunc,
			},
			wantError: false,
		},
		{
			name: "wrong function signature",
			variables: map[string]any{
				"processor": func(i int) string { return "" }, // wrong signature
			},
			wantError: true,
		},
		{
			name: "missing required function",
			variables: map[string]any{
				"validator": varFunc,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.variables)
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidator_StructTypes(t *testing.T) {
	userType := reflect.TypeOf(TestUser{})
	productType := reflect.TypeOf(TestProduct{})

	validator := NewValidator().
		RequiredStruct("user", userType).
		OptionalStruct("product", productType)

	tests := []struct {
		name      string
		variables map[string]any
		wantError bool
	}{
		{
			name: "valid struct types",
			variables: map[string]any{
				"user":    TestUser{Name: "John", Age: 30},
				"product": TestProduct{ID: 1, Name: "Test", Price: 99.99},
			},
			wantError: false,
		},
		{
			name: "wrong struct type",
			variables: map[string]any{
				"user": TestProduct{ID: 1}, // should be TestUser
			},
			wantError: true,
		},
		{
			name: "missing required struct",
			variables: map[string]any{
				"product": TestProduct{ID: 1, Name: "Test", Price: 99.99},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.variables)
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidator_PartialValidation(t *testing.T) {
	validator := NewValidator().
		RequiredString("name").
		RequiredInt("age").
		OptionalString("email")

	tests := []struct {
		name      string
		variables map[string]any
		wantError bool
	}{
		{
			name: "partial validation with some fields",
			variables: map[string]any{
				"name": "John",
				// age is missing but should not cause error in partial validation
			},
			wantError: false,
		},
		{
			name: "partial validation with wrong type",
			variables: map[string]any{
				"name": 123, // wrong type should still cause error
			},
			wantError: true,
		},
		{
			name: "partial validation with custom validation failure",
			variables: map[string]any{
				"email": "invalid-email",
			},
			wantError: false, // No custom validator, so should pass
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePartial(tt.variables)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidatePartial() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidator_RuleManagement(t *testing.T) {
	validator := NewValidator()

	// Test adding rules
	validator.RequiredString("name").OptionalInt("age")

	if !validator.HasRule("name") {
		t.Error("Expected HasRule('name') to be true")
	}

	if !validator.HasRule("age") {
		t.Error("Expected HasRule('age') to be true")
	}

	if validator.HasRule("nonexistent") {
		t.Error("Expected HasRule('nonexistent') to be false")
	}

	// Test getting rules
	rule, exists := validator.GetRule("name")
	if !exists {
		t.Error("Expected GetRule('name') to exist")
	}
	if !rule.Required {
		t.Error("Expected name rule to be required")
	}

	rule, exists = validator.GetRule("age")
	if !exists {
		t.Error("Expected GetRule('age') to exist")
	}
	if rule.Required {
		t.Error("Expected age rule to be optional")
	}

	// Test listing fields
	fields := validator.ListFields()
	if len(fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(fields))
	}

	// Test removing rules
	validator.RemoveRule("age")
	if validator.HasRule("age") {
		t.Error("Expected HasRule('age') to be false after removal")
	}

	fields = validator.ListFields()
	if len(fields) != 1 {
		t.Errorf("Expected 1 field after removal, got %d", len(fields))
	}
}

func TestValidator_Clone(t *testing.T) {
	original := NewValidator().
		RequiredString("name").
		OptionalInt("age")

	clone := original.Clone()

	// Modify original
	original.RequiredBool("active")

	// Clone should not have the new rule
	if clone.HasRule("active") {
		t.Error("Clone should not have rule added to original after cloning")
	}

	// Original should have the new rule
	if !original.HasRule("active") {
		t.Error("Original should have the new rule")
	}

	// Both should have original rules
	if !clone.HasRule("name") || !clone.HasRule("age") {
		t.Error("Clone should have original rules")
	}
}

func TestValidationErrors(t *testing.T) {
	errors := ValidationErrors{
		{Field: "name", Message: "required"},
		{Field: "age", Message: "must be positive"},
		{Field: "name", Message: "too short"},
	}

	// Test Error() method
	errorStr := errors.Error()
	if !strings.Contains(errorStr, "name") || !strings.Contains(errorStr, "age") {
		t.Error("Error string should contain field names")
	}

	// Test HasField
	if !errors.HasField("name") {
		t.Error("Expected HasField('name') to be true")
	}
	if !errors.HasField("age") {
		t.Error("Expected HasField('age') to be true")
	}
	if errors.HasField("nonexistent") {
		t.Error("Expected HasField('nonexistent') to be false")
	}

	// Test GetFieldErrors
	nameErrors := errors.GetFieldErrors("name")
	if len(nameErrors) != 2 {
		t.Errorf("Expected 2 errors for 'name', got %d", len(nameErrors))
	}

	ageErrors := errors.GetFieldErrors("age")
	if len(ageErrors) != 1 {
		t.Errorf("Expected 1 error for 'age', got %d", len(ageErrors))
	}

	// Test single error
	singleError := ValidationErrors{{Field: "test", Message: "error"}}
	if singleError.Error() != "field 'test': error" {
		t.Errorf("Single error format wrong: %s", singleError.Error())
	}

	// Test empty errors
	emptyErrors := ValidationErrors{}
	if emptyErrors.Error() != "" {
		t.Error("Empty errors should return empty string")
	}
}

func TestValidator_PanicConditions(t *testing.T) {
	validator := NewValidator()

	// Test panic conditions
	tests := []struct {
		name     string
		testFunc func()
	}{
		{
			name: "empty field name",
			testFunc: func() {
				validator.AddRule("", reflect.TypeOf(""), true)
			},
		},
		{
			name: "nil field type",
			testFunc: func() {
				validator.AddRule("test", nil, true)
			},
		},
		{
			name: "RequiredOf with nil sample",
			testFunc: func() {
				validator.RequiredOf("test", nil)
			},
		},
		{
			name: "OptionalOf with nil sample",
			testFunc: func() {
				validator.OptionalOf("test", nil)
			},
		},
		{
			name: "RequiredInterface with non-interface type",
			testFunc: func() {
				validator.RequiredInterface("test", reflect.TypeOf(""))
			},
		},
		{
			name: "RequiredStruct with non-struct type",
			testFunc: func() {
				validator.RequiredStruct("test", reflect.TypeOf(""))
			},
		},
		{
			name: "RequiredFunc with non-func type",
			testFunc: func() {
				validator.RequiredFunc("test", reflect.TypeOf(""))
			},
		},
		{
			name: "RequiredSlice with nil element type",
			testFunc: func() {
				validator.RequiredSlice("test", nil)
			},
		},
		{
			name: "RequiredMap with nil key type",
			testFunc: func() {
				validator.RequiredMap("test", nil, reflect.TypeOf(""))
			},
		},
		{
			name: "RequiredMap with nil value type",
			testFunc: func() {
				validator.RequiredMap("test", reflect.TypeOf(""), nil)
			},
		},
		{
			name: "RequiredPointer with nil element type",
			testFunc: func() {
				validator.RequiredPointer("test", nil)
			},
		},
		{
			name: "RequiredChan with nil element type",
			testFunc: func() {
				validator.RequiredChan("test", nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Error("Expected panic but didn't get one")
				}
			}()
			tt.testFunc()
		})
	}
}

func TestValidator_ConcurrentAccess(t *testing.T) {
	validator := NewValidator()
	var wg sync.WaitGroup

	// Test concurrent rule addition
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			validator.RequiredString(fmt.Sprintf("field%d", i))
		}(i)
	}

	// Test concurrent validation
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			validator.Validate(map[string]any{
				"field0": "test",
				"field1": "test",
			})
		}()
	}

	wg.Wait()

	// Verify all fields were added
	fields := validator.ListFields()
	if len(fields) < 10 {
		t.Errorf("Expected at least 10 fields, got %d", len(fields))
	}
}

// Registry Tests

func TestRegistry_BasicOperations(t *testing.T) {
	registry := NewRegistry()

	// Test empty registry
	if registry.HasTemplate("nonexistent") {
		t.Error("Empty registry should not have any templates")
	}

	templates := registry.ListTemplates()
	if len(templates) != 0 {
		t.Errorf("Empty registry should have 0 templates, got %d", len(templates))
	}

	// Test registration
	validator1 := NewValidator().RequiredString("name")
	validator2 := NewValidator().RequiredInt("age")

	registry.Register("template1", validator1)
	registry.Register("template2", validator2)

	if !registry.HasTemplate("template1") {
		t.Error("Registry should have template1")
	}

	if !registry.HasTemplate("template2") {
		t.Error("Registry should have template2")
	}

	templates = registry.ListTemplates()
	if len(templates) != 2 {
		t.Errorf("Registry should have 2 templates, got %d", len(templates))
	}

	// Test GetValidator
	v1, exists := registry.GetValidator("template1")
	if !exists || v1 != validator1 {
		t.Error("GetValidator should return the registered validator")
	}

	_, exists = registry.GetValidator("nonexistent")
	if exists {
		t.Error("GetValidator should return false for nonexistent template")
	}
}

func TestRegistry_Validation(t *testing.T) {
	registry := NewRegistry()

	validator := NewValidator().
		RequiredString("name").
		RequiredInt("age")

	registry.Register("user", validator)

	tests := []struct {
		name      string
		template  string
		variables map[string]any
		wantError bool
	}{
		{
			name:     "valid validation",
			template: "user",
			variables: map[string]any{
				"name": "John",
				"age":  30,
			},
			wantError: false,
		},
		{
			name:     "invalid validation",
			template: "user",
			variables: map[string]any{
				"name": "John",
				// missing age
			},
			wantError: true,
		},
		{
			name:      "nonexistent template",
			template:  "nonexistent",
			variables: map[string]any{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.Validate(tt.template, tt.variables)
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestRegistry_PartialValidation(t *testing.T) {
	registry := NewRegistry()

	validator := NewValidator().
		RequiredString("name").
		RequiredInt("age")

	registry.Register("user", validator)

	// Test partial validation
	err := registry.ValidatePartial("user", map[string]any{
		"name": "John",
		// age is missing but should not cause error in partial validation
	})
	if err != nil {
		t.Errorf("ValidatePartial should not error for missing fields: %v", err)
	}

	// Test with nonexistent template
	err = registry.ValidatePartial("nonexistent", map[string]any{})
	if err == nil {
		t.Error("ValidatePartial should error for nonexistent template")
	}
}

func TestRegistry_Management(t *testing.T) {
	registry := NewRegistry()

	validator := NewValidator().RequiredString("name")
	registry.Register("test", validator)

	// Test unregister
	registry.Unregister("test")
	if registry.HasTemplate("test") {
		t.Error("Template should be unregistered")
	}

	// Re-register for clear test
	registry.Register("test1", validator)
	registry.Register("test2", validator)

	if len(registry.ListTemplates()) != 2 {
		t.Error("Should have 2 templates before clear")
	}

	// Test clear
	registry.Clear()
	if len(registry.ListTemplates()) != 0 {
		t.Error("Should have 0 templates after clear")
	}
}

func TestRegistry_PanicConditions(t *testing.T) {
	registry := NewRegistry()

	tests := []struct {
		name     string
		testFunc func()
	}{
		{
			name: "empty template name",
			testFunc: func() {
				registry.Register("", NewValidator())
			},
		},
		{
			name: "nil validator",
			testFunc: func() {
				registry.Register("test", nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Error("Expected panic but didn't get one")
				}
			}()
			tt.testFunc()
		})
	}
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewRegistry()
	var wg sync.WaitGroup

	// Test concurrent registration
	for i := range 10 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			validator := NewValidator().RequiredString("name")
			registry.Register(fmt.Sprintf("template%d", i), validator)
		}(i)
	}

	// Test concurrent validation
	validator := NewValidator().RequiredString("test")
	registry.Register("concurrent", validator)

	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			registry.Validate("concurrent", map[string]any{"test": "value"})
		}()
	}

	wg.Wait()

	// Verify all templates were registered
	templates := registry.ListTemplates()
	if len(templates) < 11 { // 10 + 1 concurrent
		t.Errorf("Expected at least 11 templates, got %d", len(templates))
	}
}

// Type compatibility tests
func TestTypeCompatibility(t *testing.T) {
	tests := []struct {
		name     string
		actual   reflect.Type
		expected reflect.Type
		want     bool
	}{
		{
			name:     "same types",
			actual:   reflect.TypeOf(""),
			expected: reflect.TypeOf(""),
			want:     true,
		},
		{
			name:     "any type compatibility",
			actual:   reflect.TypeOf(""),
			expected: reflect.TypeOf((*any)(nil)).Elem(),
			want:     true,
		},
		{
			name:     "numeric compatibility int to int64",
			actual:   reflect.TypeOf(int(0)),
			expected: reflect.TypeOf(int64(0)),
			want:     true,
		},
		{
			name:     "numeric compatibility float32 to float64",
			actual:   reflect.TypeOf(float32(0)),
			expected: reflect.TypeOf(float64(0)),
			want:     true,
		},
		{
			name:     "interface compatibility",
			actual:   reflect.TypeOf(TestImplementer{}),
			expected: reflect.TypeOf((*TestInterface)(nil)).Elem(),
			want:     true,
		},
		{
			name:     "slice compatibility",
			actual:   reflect.TypeOf([]string{}),
			expected: reflect.TypeOf([]string{}),
			want:     true,
		},
		{
			name:     "map compatibility",
			actual:   reflect.TypeOf(map[string]int{}),
			expected: reflect.TypeOf(map[string]int{}),
			want:     true,
		},
		{
			name:     "pointer compatibility",
			actual:   reflect.TypeOf(&TestUser{}),
			expected: reflect.TypeOf(&TestUser{}),
			want:     true,
		},
		{
			name:     "incompatible types",
			actual:   reflect.TypeOf(""),
			expected: reflect.TypeOf(0),
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isTypeCompatible(tt.actual, tt.expected)
			if got != tt.want {
				t.Errorf("isTypeCompatible(%v, %v) = %v; want %v", tt.actual, tt.expected, got, tt.want)
			}
		})
	}
}

func TestInterfaceType(t *testing.T) {
	var name string

	interfaceType := InterfaceType(name)
	if interfaceType.String() != reflect.String.String() {
		t.Fail()
	}
}

/*
templateval is a comprehensive Go library for validating template
variables with strong typing support.

It provides flexible validation for all Go types including basic types,
structs, interfaces, functions, slices, maps, and more.
*/
package templateval

import (
	"fmt"
	"maps"
	"reflect"
	"strings"
	"sync"
)

// FieldRule defines validation rules for a template variable
type FieldRule struct {
	Type      reflect.Type    // Type of the field.
	Required  bool            // Whether this field is required or not
	Validator func(any) error // Optional custom validator function
}

// ValidationError represents a validation failure
type ValidationError struct {
	Field   string // Field name failing the calidation
	Message string // Error message
}

// Implements the error interface.
func (e ValidationError) Error() string {
	return fmt.Sprintf("Field '%s': %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

// Implements the error interface.
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}
	if len(ve) == 1 {
		return ve[0].Error()
	}

	var sb strings.Builder
	sb.WriteString(ve[0].Error())
	for i := 1; i < len(ve); i++ {
		sb.WriteString("; ")
		sb.WriteString(ve[i].Error())
	}
	return sb.String()
}

// HasField checks if a specific field has validation errors
func (ve ValidationErrors) HasField(field string) bool {
	for _, err := range ve {
		if err.Field == field {
			return true
		}
	}
	return false
}

// GetFieldErrors returns all errors for a specific field
func (ve ValidationErrors) GetFieldErrors(field string) []ValidationError {
	var errors []ValidationError
	for _, err := range ve {
		if err.Field == field {
			errors = append(errors, err)
		}
	}
	return errors
}

// TemplateValidator validates template variables against defined rules
type TemplateValidator struct {
	rules map[string]FieldRule // Map of field rules for a given template.
	mu    sync.RWMutex         // Thread safety for concurrent access
}

// NewValidator creates a new template validator
func NewValidator() *TemplateValidator {
	return &TemplateValidator{
		rules: make(map[string]FieldRule),
	}
}

// Clone creates a deep copy of the validator
func (v *TemplateValidator) Clone() *TemplateValidator {
	v.mu.RLock()
	defer v.mu.RUnlock()
	clone := NewValidator()
	maps.Copy(clone.rules, v.rules)
	return clone
}

// AddRule adds a validation rule for a field with optional custom validator
func (v *TemplateValidator) AddRule(fieldName string, fieldType reflect.Type, required bool, customValidator ...func(any) error) *TemplateValidator {
	if fieldName == "" {
		panic("field name cannot be empty")
	}
	if fieldType == nil {
		panic("field type cannot be nil")
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	rule := FieldRule{
		Type:     fieldType,
		Required: required,
	}
	if len(customValidator) > 0 && customValidator[0] != nil {
		rule.Validator = customValidator[0]
	}
	v.rules[fieldName] = rule
	return v
}

// RemoveRule removes a validation rule for a field
func (v *TemplateValidator) RemoveRule(fieldName string) *TemplateValidator {
	v.mu.Lock()
	defer v.mu.Unlock()
	delete(v.rules, fieldName)
	return v
}

// HasRule checks if a rule exists for the given field
func (v *TemplateValidator) HasRule(fieldName string) bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	_, exists := v.rules[fieldName]
	return exists
}

// GetRule returns the rule for a field
func (v *TemplateValidator) GetRule(fieldName string) (FieldRule, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	rule, exists := v.rules[fieldName]
	return rule, exists
}

// ListFields returns all field names that have rules
func (v *TemplateValidator) ListFields() []string {
	v.mu.RLock()
	defer v.mu.RUnlock()

	fields := make([]string, 0, len(v.rules))
	for name := range v.rules {
		fields = append(fields, name)
	}
	return fields
}

// Required adds a required field rule for any type
func (v *TemplateValidator) Required(fieldName string, fieldType reflect.Type, customValidator ...func(any) error) *TemplateValidator {
	return v.AddRule(fieldName, fieldType, true, customValidator...)
}

// Optional adds an optional field rule for any type
func (v *TemplateValidator) Optional(fieldName string, fieldType reflect.Type, customValidator ...func(any) error) *TemplateValidator {
	return v.AddRule(fieldName, fieldType, false, customValidator...)
}

// RequiredOf adds a required field rule using a sample value to infer type
func (v *TemplateValidator) RequiredOf(fieldName string, sample any, customValidator ...func(any) error) *TemplateValidator {
	if sample == nil {
		panic("sample cannot be nil for type inference")
	}
	return v.Required(fieldName, reflect.TypeOf(sample), customValidator...)
}

// OptionalOf adds an optional field rule using a sample value to infer type
func (v *TemplateValidator) OptionalOf(fieldName string, sample any, customValidator ...func(any) error) *TemplateValidator {
	if sample == nil {
		panic("sample cannot be nil for type inference")
	}
	return v.Optional(fieldName, reflect.TypeOf(sample), customValidator...)
}

// Type-specific methods with validation
func (v *TemplateValidator) RequiredInterface(fieldName string, interfaceType reflect.Type, customValidator ...func(any) error) *TemplateValidator {
	validateTypeKind(interfaceType, reflect.Interface, "RequiredInterface")
	return v.Required(fieldName, interfaceType, customValidator...)
}

func (v *TemplateValidator) OptionalInterface(fieldName string, interfaceType reflect.Type, customValidator ...func(any) error) *TemplateValidator {
	validateTypeKind(interfaceType, reflect.Interface, "OptionalInterface")
	return v.Optional(fieldName, interfaceType, customValidator...)
}

func (v *TemplateValidator) RequiredStruct(fieldName string, structType reflect.Type, customValidator ...func(any) error) *TemplateValidator {
	validateTypeKind(structType, reflect.Struct, "RequiredStruct")
	return v.Required(fieldName, structType, customValidator...)
}

func (v *TemplateValidator) OptionalStruct(fieldName string, structType reflect.Type, customValidator ...func(any) error) *TemplateValidator {
	validateTypeKind(structType, reflect.Struct, "OptionalStruct")
	return v.Optional(fieldName, structType, customValidator...)
}

func (v *TemplateValidator) RequiredFunc(fieldName string, funcType reflect.Type, customValidator ...func(any) error) *TemplateValidator {
	validateTypeKind(funcType, reflect.Func, "RequiredFunc")
	return v.Required(fieldName, funcType, customValidator...)
}

func (v *TemplateValidator) OptionalFunc(fieldName string, funcType reflect.Type, customValidator ...func(any) error) *TemplateValidator {
	validateTypeKind(funcType, reflect.Func, "OptionalFunc")
	return v.Optional(fieldName, funcType, customValidator...)
}

// validateTypeKind is a helper to validate type kinds
func validateTypeKind(t reflect.Type, expectedKind reflect.Kind, methodName string) {
	if t == nil {
		panic(fmt.Sprintf("%s: type cannot be nil", methodName))
	}
	if t.Kind() != expectedKind {
		panic(fmt.Sprintf("%s expects %s type, got %s", methodName, expectedKind, t.Kind()))
	}
}

// Convenience methods for common types
func (v *TemplateValidator) RequiredString(fieldName string, customValidator ...func(any) error) *TemplateValidator {
	return v.Required(fieldName, stringType, customValidator...)
}

func (v *TemplateValidator) OptionalString(fieldName string, customValidator ...func(any) error) *TemplateValidator {
	return v.Optional(fieldName, stringType, customValidator...)
}

func (v *TemplateValidator) RequiredInt(fieldName string, customValidator ...func(any) error) *TemplateValidator {
	return v.Required(fieldName, intType, customValidator...)
}

func (v *TemplateValidator) OptionalInt(fieldName string, customValidator ...func(any) error) *TemplateValidator {
	return v.Optional(fieldName, intType, customValidator...)
}

func (v *TemplateValidator) RequiredBool(fieldName string, customValidator ...func(any) error) *TemplateValidator {
	return v.Required(fieldName, boolType, customValidator...)
}

func (v *TemplateValidator) OptionalBool(fieldName string, customValidator ...func(any) error) *TemplateValidator {
	return v.Optional(fieldName, boolType, customValidator...)
}

func (v *TemplateValidator) RequiredSlice(fieldName string, elementType reflect.Type, customValidator ...func(any) error) *TemplateValidator {
	if elementType == nil {
		panic("element type cannot be nil for slice")
	}
	return v.Required(fieldName, reflect.SliceOf(elementType), customValidator...)
}

func (v *TemplateValidator) OptionalSlice(fieldName string, elementType reflect.Type, customValidator ...func(any) error) *TemplateValidator {
	if elementType == nil {
		panic("element type cannot be nil for slice")
	}
	return v.Optional(fieldName, reflect.SliceOf(elementType), customValidator...)
}

func (v *TemplateValidator) RequiredMap(fieldName string, keyType, valueType reflect.Type, customValidator ...func(any) error) *TemplateValidator {
	if keyType == nil || valueType == nil {
		panic("key and value types cannot be nil for map")
	}
	return v.Required(fieldName, reflect.MapOf(keyType, valueType), customValidator...)
}

func (v *TemplateValidator) OptionalMap(fieldName string, keyType, valueType reflect.Type, customValidator ...func(any) error) *TemplateValidator {
	if keyType == nil || valueType == nil {
		panic("key and value types cannot be nil for map")
	}
	return v.Optional(fieldName, reflect.MapOf(keyType, valueType), customValidator...)
}

func (v *TemplateValidator) RequiredPointer(fieldName string, elementType reflect.Type, customValidator ...func(any) error) *TemplateValidator {
	if elementType == nil {
		panic("element type cannot be nil for pointer")
	}
	return v.Required(fieldName, reflect.PointerTo(elementType), customValidator...)
}

func (v *TemplateValidator) OptionalPointer(fieldName string, elementType reflect.Type, customValidator ...func(any) error) *TemplateValidator {
	if elementType == nil {
		panic("element type cannot be nil for pointer")
	}
	return v.Optional(fieldName, reflect.PointerTo(elementType), customValidator...)
}

func (v *TemplateValidator) RequiredChan(fieldName string, elementType reflect.Type, customValidator ...func(any) error) *TemplateValidator {
	if elementType == nil {
		panic("element type cannot be nil for channel")
	}
	return v.Required(fieldName, reflect.ChanOf(reflect.BothDir, elementType), customValidator...)
}

func (v *TemplateValidator) OptionalChan(fieldName string, elementType reflect.Type, customValidator ...func(any) error) *TemplateValidator {
	if elementType == nil {
		panic("element type cannot be nil for channel")
	}
	return v.Optional(fieldName, reflect.ChanOf(reflect.BothDir, elementType), customValidator...)
}

// Pre-computed common types for performance
var (
	stringType = reflect.TypeOf("")
	intType    = reflect.TypeOf(0)
	boolType   = reflect.TypeOf(true)
	anyType    = reflect.TypeOf((*any)(nil)).Elem()
)

// ValidatePartial validates only the provided variables, ignoring missing required fields
// Useful for partial updates or progressive validation
func (v *TemplateValidator) ValidatePartial(variables map[string]any) error {
	v.mu.RLock()
	defer v.mu.RUnlock()

	var errors ValidationErrors

	for fieldName, value := range variables {
		rule, exists := v.rules[fieldName]
		if !exists {
			continue // Skip unknown fields in partial validation
		}

		if err := v.validateFieldValue(fieldName, value, rule, false); err != nil {
			if ve, ok := err.(ValidationErrors); ok {
				errors = append(errors, ve...)
			} else {
				errors = append(errors, ValidationError{Field: fieldName, Message: err.Error()})
			}
		}
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// Validate checks the provided variables against the defined rules
func (v *TemplateValidator) Validate(variables map[string]any) error {
	v.mu.RLock()
	defer v.mu.RUnlock()

	var errors ValidationErrors

	// Check all rules
	for fieldName, rule := range v.rules {
		value, exists := variables[fieldName]

		if !exists {
			if rule.Required {
				errors = append(errors, ValidationError{
					Field:   fieldName,
					Message: "required field is missing",
				})
			}
			continue
		}

		if err := v.validateFieldValue(fieldName, value, rule, true); err != nil {
			if ve, ok := err.(ValidationErrors); ok {
				errors = append(errors, ve...)
			} else {
				errors = append(errors, ValidationError{Field: fieldName, Message: err.Error()})
			}
		}
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// validateFieldValue validates a single field value
func (v *TemplateValidator) validateFieldValue(fieldName string, value any, rule FieldRule, checkRequired bool) error {
	// Check for nil values
	if value == nil {
		if checkRequired && rule.Required {
			return ValidationError{Field: fieldName, Message: "required field cannot be nil"}
		}
		return nil // nil is ok for optional fields
	}

	// Type validation
	valueType := reflect.TypeOf(value)
	if !isTypeCompatible(valueType, rule.Type) {
		return ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("expected type %s, got %s", rule.Type, valueType),
		}
	}

	// Custom validation
	if rule.Validator != nil {
		if err := rule.Validator(value); err != nil {
			return ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("custom validation failed: %v", err),
			}
		}
	}

	return nil
}

// Type compatibility cache for performance
var typeCompatibilityCache = struct {
	mu    sync.RWMutex
	cache map[typePair]bool
}{
	cache: make(map[typePair]bool),
}

type typePair struct {
	actual   reflect.Type
	expected reflect.Type
}

// isTypeCompatible checks if the actual type is compatible with the expected type
func isTypeCompatible(actual, expected reflect.Type) bool {
	// Check cache first
	pair := typePair{actual, expected}
	typeCompatibilityCache.mu.RLock()
	if result, exists := typeCompatibilityCache.cache[pair]; exists {
		typeCompatibilityCache.mu.RUnlock()
		return result
	}

	typeCompatibilityCache.mu.RUnlock()

	result := isTypeCompatibleImpl(actual, expected)

	// Cache the result
	typeCompatibilityCache.mu.Lock()
	typeCompatibilityCache.cache[pair] = result
	typeCompatibilityCache.mu.Unlock()

	return result
}

func isTypeCompatibleImpl(actual, expected reflect.Type) bool {
	// Direct type match
	if actual == expected {
		return true
	}

	// Handle any/interface{} - everything is compatible with interface{}
	if expected == anyType {
		return true
	}

	// Handle interface compatibility
	if expected.Kind() == reflect.Interface {
		return actual.Implements(expected)
	}

	// Handle pointer compatibility
	if actual.Kind() == reflect.Ptr && expected.Kind() == reflect.Ptr {
		return isTypeCompatible(actual.Elem(), expected.Elem())
	}
	if actual.Kind() == reflect.Ptr && expected.Kind() != reflect.Ptr {
		return isTypeCompatible(actual.Elem(), expected)
	}

	// Handle numeric type conversions
	if isNumericType(actual) && isNumericType(expected) {
		return true
	}

	// Handle slice compatibility
	if actual.Kind() == reflect.Slice && expected.Kind() == reflect.Slice {
		return isTypeCompatible(actual.Elem(), expected.Elem())
	}

	// Handle map compatibility
	if actual.Kind() == reflect.Map && expected.Kind() == reflect.Map {
		return isTypeCompatible(actual.Key(), expected.Key()) &&
			isTypeCompatible(actual.Elem(), expected.Elem())
	}

	// Handle chan compatibility
	if actual.Kind() == reflect.Chan && expected.Kind() == reflect.Chan {
		return isTypeCompatible(actual.Elem(), expected.Elem())
	}

	// Handle struct embedding/composition
	if actual.Kind() == reflect.Struct && expected.Kind() == reflect.Struct {
		return isStructCompatible(actual, expected)
	}

	// Handle function compatibility
	if actual.Kind() == reflect.Func && expected.Kind() == reflect.Func {
		return isFuncCompatible(actual, expected)
	}

	return false
}

// isNumericType checks if a type is numeric
func isNumericType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return true
	}
	return false
}

// isStructCompatible checks if two struct types are compatible
func isStructCompatible(actual, expected reflect.Type) bool {
	// Cache expected fields
	expectedFields := make(map[string]reflect.Type)
	for i := 0; i < expected.NumField(); i++ {
		field := expected.Field(i)
		if field.IsExported() {
			expectedFields[field.Name] = field.Type
		}
	}

	// Check if actual has all expected fields with compatible types
	for i := range actual.NumField() {
		field := actual.Field(i)
		if !field.IsExported() {
			continue
		}

		if expectedType, exists := expectedFields[field.Name]; exists {
			if !isTypeCompatible(field.Type, expectedType) {
				return false
			}
			delete(expectedFields, field.Name)
		}
	}

	// All expected fields must be found
	return len(expectedFields) == 0
}

// isFuncCompatible checks if two function types are compatible
func isFuncCompatible(actual, expected reflect.Type) bool {
	// Check basic signature compatibility
	if actual.NumIn() != expected.NumIn() ||
		actual.NumOut() != expected.NumOut() ||
		actual.IsVariadic() != expected.IsVariadic() {
		return false
	}

	// Check input parameter types
	for i := range actual.NumIn() {
		if !isTypeCompatible(actual.In(i), expected.In(i)) {
			return false
		}
	}

	// Check output parameter types
	for i := range actual.NumOut() {
		if !isTypeCompatible(actual.Out(i), expected.Out(i)) {
			return false
		}
	}

	return true
}

// TemplateRegistry manages validators for multiple templates
type TemplateRegistry struct {
	validators map[string]*TemplateValidator
	mu         sync.RWMutex
}

// NewRegistry creates a new template registry
func NewRegistry() *TemplateRegistry {
	return &TemplateRegistry{
		validators: make(map[string]*TemplateValidator),
	}
}

// Register adds a validator for a template. If template is an empty string, it will panic.
// if validator is nil, no registration will happen.
func (r *TemplateRegistry) Register(name string, validator *TemplateValidator) {
	if name == "" {
		panic("template name cannot be empty")
	}

	if validator == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.validators[name] = validator
}

// Unregister removes a validator for a template
func (r *TemplateRegistry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.validators, name)
}

// Validate validates variables for a specific template
func (r *TemplateRegistry) Validate(name string, variables map[string]any) error {
	r.mu.RLock()
	validator, exists := r.validators[name]
	r.mu.RUnlock()

	if !exists {
		return nil // No validator registered
	}
	return validator.Validate(variables)
}

// ValidatePartial validates variables partially for a specific template
func (r *TemplateRegistry) ValidatePartial(name string, variables map[string]any) error {
	r.mu.RLock()
	validator, exists := r.validators[name]
	r.mu.RUnlock()

	if !exists {
		return nil // No validator registered
	}
	return validator.ValidatePartial(variables)
}

// GetValidator returns the validator for a template
func (r *TemplateRegistry) GetValidator(name string) (*TemplateValidator, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	validator, exists := r.validators[name]
	return validator, exists
}

// ListTemplates returns all registered template names
func (r *TemplateRegistry) ListTemplates() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	templates := make([]string, 0, len(r.validators))
	for name := range r.validators {
		templates = append(templates, name)
	}
	return templates
}

// HasTemplate checks if a template is registered
func (r *TemplateRegistry) HasTemplate(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.validators[name]
	return exists
}

// Clear removes all registered validators
func (r *TemplateRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.validators = make(map[string]*TemplateValidator)
}

// Helper function to get type from sample value
func TypeOf[T any](sample T) reflect.Type {
	return reflect.TypeOf(sample)
}

// Helper function to create interface type
func InterfaceType[T any](sample T) reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

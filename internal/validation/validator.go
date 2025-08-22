package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"

	"discord-mcp/pkg/types"
)

// Validator handles parameter validation for MCP tools
type Validator struct {
	// Cache compiled regex patterns for performance
	patterns map[string]*regexp.Regexp
}

// NewValidator creates a new parameter validator
func NewValidator() *Validator {
	return &Validator{
		patterns: make(map[string]*regexp.Regexp),
	}
}

// ValidateToolParams validates parameters against a tool's JSON schema
func (v *Validator) ValidateToolParams(toolName string, params map[string]interface{}) error {
	schema, exists := GetToolSchema(toolName)
	if !exists {
		return fmt.Errorf("no schema found for tool: %s", toolName)
	}

	schemaMap, ok := schema.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid schema format for tool: %s", toolName)
	}

	// Validate required fields
	if err := v.validateRequired(schemaMap, params); err != nil {
		return NewValidationError("missing required parameter", err.Error(), nil)
	}

	// Validate each parameter
	properties, hasProps := schemaMap["properties"].(map[string]interface{})
	if hasProps {
		for paramName, value := range params {
			propSchema, hasProp := properties[paramName]
			if !hasProp {
				return NewValidationError("unknown parameter", fmt.Sprintf("parameter '%s' is not defined", paramName), paramName)
			}

			if err := v.validateParameter(paramName, value, propSchema); err != nil {
				return err
			}
		}
	}

	// Validate conditional requirements (anyOf, oneOf, etc.)
	if err := v.validateConditionals(schemaMap, params); err != nil {
		return err
	}

	return nil
}

// validateRequired checks if all required parameters are present
func (v *Validator) validateRequired(schema map[string]interface{}, params map[string]interface{}) error {
	required, hasRequired := schema["required"].([]string)
	if !hasRequired {
		return nil
	}

	for _, reqParam := range required {
		if _, exists := params[reqParam]; !exists {
			return fmt.Errorf("required parameter '%s' is missing", reqParam)
		}
	}

	return nil
}

// validateParameter validates a single parameter against its schema
func (v *Validator) validateParameter(paramName string, value interface{}, propSchema interface{}) error {
	schema, ok := propSchema.(map[string]interface{})
	if !ok {
		return NewValidationError("invalid schema", fmt.Sprintf("invalid schema for parameter '%s'", paramName), paramName)
	}

	// Type validation
	if err := v.validateType(paramName, value, schema); err != nil {
		return err
	}

	// String validations
	if value != nil && reflect.TypeOf(value).Kind() == reflect.String {
		if err := v.validateString(paramName, value.(string), schema); err != nil {
			return err
		}
	}

	// Numeric validations
	if value != nil && v.isNumeric(value) {
		if err := v.validateNumeric(paramName, value, schema); err != nil {
			return err
		}
	}

	// Array validations
	if value != nil && reflect.TypeOf(value).Kind() == reflect.Slice {
		if err := v.validateArray(paramName, value, schema); err != nil {
			return err
		}
	}

	// Object validations (recursive)
	if value != nil && reflect.TypeOf(value).Kind() == reflect.Map {
		if err := v.validateObject(paramName, value, schema); err != nil {
			return err
		}
	}

	return nil
}

// validateType checks if the value matches the expected type
func (v *Validator) validateType(paramName string, value interface{}, schema map[string]interface{}) error {
	expectedType, hasType := schema["type"].(string)
	if !hasType {
		return nil // No type constraint
	}

	if value == nil {
		return NewValidationError("null value", fmt.Sprintf("parameter '%s' cannot be null", paramName), paramName)
	}

	actualKind := reflect.TypeOf(value).Kind()

	switch expectedType {
	case "string":
		if actualKind != reflect.String {
			return NewValidationError("type mismatch", fmt.Sprintf("parameter '%s' must be a string, got %T", paramName, value), paramName)
		}
	case "integer":
		if !v.isInteger(value) {
			return NewValidationError("type mismatch", fmt.Sprintf("parameter '%s' must be an integer, got %T", paramName, value), paramName)
		}
	case "number":
		if !v.isNumeric(value) {
			return NewValidationError("type mismatch", fmt.Sprintf("parameter '%s' must be a number, got %T", paramName, value), paramName)
		}
	case "boolean":
		if actualKind != reflect.Bool {
			return NewValidationError("type mismatch", fmt.Sprintf("parameter '%s' must be a boolean, got %T", paramName, value), paramName)
		}
	case "array":
		if actualKind != reflect.Slice && actualKind != reflect.Array {
			return NewValidationError("type mismatch", fmt.Sprintf("parameter '%s' must be an array, got %T", paramName, value), paramName)
		}
	case "object":
		if actualKind != reflect.Map {
			return NewValidationError("type mismatch", fmt.Sprintf("parameter '%s' must be an object, got %T", paramName, value), paramName)
		}
	}

	return nil
}

// validateString validates string constraints
func (v *Validator) validateString(paramName, value string, schema map[string]interface{}) error {
	// MinLength validation
	if minLen, hasMin := schema["minLength"]; hasMin {
		if minLenInt, ok := minLen.(int); ok {
			if len(value) < minLenInt {
				return NewValidationError("length constraint", fmt.Sprintf("parameter '%s' must be at least %d characters, got %d", paramName, minLenInt, len(value)), paramName)
			}
		}
	}

	// MaxLength validation
	if maxLen, hasMax := schema["maxLength"]; hasMax {
		if maxLenInt, ok := maxLen.(int); ok {
			if len(value) > maxLenInt {
				return NewValidationError("length constraint", fmt.Sprintf("parameter '%s' must be at most %d characters, got %d", paramName, maxLenInt, len(value)), paramName)
			}
		}
	}

	// Pattern validation
	if pattern, hasPattern := schema["pattern"]; hasPattern {
		if patternStr, ok := pattern.(string); ok {
			regex, err := v.getCompiledRegex(patternStr)
			if err != nil {
				return NewValidationError("pattern error", fmt.Sprintf("invalid regex pattern for parameter '%s': %v", paramName, err), paramName)
			}
			if !regex.MatchString(value) {
				return NewValidationError("pattern mismatch", fmt.Sprintf("parameter '%s' does not match required pattern: %s", paramName, patternStr), paramName)
			}
		}
	}

	// Enum validation
	if enum, hasEnum := schema["enum"]; hasEnum {
		if enumSlice, ok := enum.([]string); ok {
			found := false
			for _, allowedValue := range enumSlice {
				if value == allowedValue {
					found = true
					break
				}
			}
			if !found {
				return NewValidationError("enum constraint", fmt.Sprintf("parameter '%s' must be one of: %v, got '%s'", paramName, enumSlice, value), paramName)
			}
		}
	}

	return nil
}

// validateNumeric validates numeric constraints
func (v *Validator) validateNumeric(paramName string, value interface{}, schema map[string]interface{}) error {
	numValue := v.getNumericValue(value)

	// Minimum validation
	if min, hasMin := schema["minimum"]; hasMin {
		if minFloat, ok := v.toFloat64(min); ok {
			if numValue < minFloat {
				return NewValidationError("range constraint", fmt.Sprintf("parameter '%s' must be at least %v, got %v", paramName, min, value), paramName)
			}
		}
	}

	// Maximum validation
	if max, hasMax := schema["maximum"]; hasMax {
		if maxFloat, ok := v.toFloat64(max); ok {
			if numValue > maxFloat {
				return NewValidationError("range constraint", fmt.Sprintf("parameter '%s' must be at most %v, got %v", paramName, max, value), paramName)
			}
		}
	}

	return nil
}

// validateArray validates array constraints
func (v *Validator) validateArray(paramName string, value interface{}, schema map[string]interface{}) error {
	rv := reflect.ValueOf(value)
	length := rv.Len()

	// MinItems validation
	if minItems, hasMin := schema["minItems"]; hasMin {
		if minInt, ok := minItems.(int); ok {
			if length < minInt {
				return NewValidationError("array constraint", fmt.Sprintf("parameter '%s' must have at least %d items, got %d", paramName, minInt, length), paramName)
			}
		}
	}

	// MaxItems validation
	if maxItems, hasMax := schema["maxItems"]; hasMax {
		if maxInt, ok := maxItems.(int); ok {
			if length > maxInt {
				return NewValidationError("array constraint", fmt.Sprintf("parameter '%s' must have at most %d items, got %d", paramName, maxInt, length), paramName)
			}
		}
	}

	// UniqueItems validation
	if unique, hasUnique := schema["uniqueItems"]; hasUnique {
		if uniqueBool, ok := unique.(bool); ok && uniqueBool {
			seen := make(map[interface{}]bool)
			for i := 0; i < length; i++ {
				item := rv.Index(i).Interface()
				if seen[item] {
					return NewValidationError("uniqueness constraint", fmt.Sprintf("parameter '%s' contains duplicate items", paramName), paramName)
				}
				seen[item] = true
			}
		}
	}

	// Items validation (validate each item against item schema)
	if itemSchema, hasItems := schema["items"]; hasItems {
		for i := 0; i < length; i++ {
			item := rv.Index(i).Interface()
			itemParamName := fmt.Sprintf("%s[%d]", paramName, i)
			if err := v.validateParameter(itemParamName, item, itemSchema); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateObject validates object constraints (recursive)
func (v *Validator) validateObject(paramName string, value interface{}, schema map[string]interface{}) error {
	// This would handle nested object validation
	// For now, we'll keep it simple and just validate that it's a map
	if reflect.TypeOf(value).Kind() != reflect.Map {
		return NewValidationError("type mismatch", fmt.Sprintf("parameter '%s' must be an object", paramName), paramName)
	}
	return nil
}

// validateConditionals handles anyOf, oneOf, allOf, not constraints
func (v *Validator) validateConditionals(schema map[string]interface{}, params map[string]interface{}) error {
	// Handle "anyOf" - at least one of the conditions must be satisfied
	if anyOf, hasAnyOf := schema["anyOf"]; hasAnyOf {
		if conditions, ok := anyOf.([]map[string]interface{}); ok {
			satisfied := false
			for _, condition := range conditions {
				if err := v.validateRequired(condition, params); err == nil {
					satisfied = true
					break
				}
			}
			if !satisfied {
				return NewValidationError("conditional constraint", "at least one of the specified conditions must be met", nil)
			}
		}
	}

	// Handle "not" - the condition must NOT be satisfied
	if not, hasNot := schema["not"]; hasNot {
		if notSchema, ok := not.(map[string]interface{}); ok {
			if err := v.validateConditionals(notSchema, params); err == nil {
				return NewValidationError("conditional constraint", "the specified condition must not be satisfied", nil)
			}
		}
	}

	return nil
}

// Helper methods

func (v *Validator) getCompiledRegex(pattern string) (*regexp.Regexp, error) {
	if regex, exists := v.patterns[pattern]; exists {
		return regex, nil
	}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	v.patterns[pattern] = regex
	return regex, nil
}

func (v *Validator) isNumeric(value interface{}) bool {
	kind := reflect.TypeOf(value).Kind()
	return kind >= reflect.Int && kind <= reflect.Complex128
}

func (v *Validator) isInteger(value interface{}) bool {
	kind := reflect.TypeOf(value).Kind()
	return kind >= reflect.Int && kind <= reflect.Uint64
}

func (v *Validator) getNumericValue(value interface{}) float64 {
	if floatVal, ok := v.toFloat64(value); ok {
		return floatVal
	}
	return 0
}

func (v *Validator) toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// ValidationError represents a parameter validation error
type ValidationError struct {
	Type    string      `json:"type"`
	Message string      `json:"message"`
	Field   interface{} `json:"field,omitempty"`
}

func (e *ValidationError) Error() string {
	if e.Field != nil {
		return fmt.Sprintf("%s: %s (field: %v)", e.Type, e.Message, e.Field)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(errorType, message string, field interface{}) *ValidationError {
	return &ValidationError{
		Type:    errorType,
		Message: message,
		Field:   field,
	}
}

// FormatValidationError returns a properly formatted MCP tool result for validation errors
func FormatValidationError(err error) types.CallToolResult {
	var validationErr *ValidationError
	if ve, ok := err.(*ValidationError); ok {
		validationErr = ve
	} else {
		validationErr = &ValidationError{
			Type:    "validation",
			Message: err.Error(),
		}
	}

	return types.CallToolResult{
		Content: []types.Content{{
			Type: "text",
			Text: fmt.Sprintf("❌ Validation Error: %s", validationErr.Message),
			Data: map[string]interface{}{
				"error_type": validationErr.Type,
				"message":    validationErr.Message,
				"field":      validationErr.Field,
			},
		}},
		IsError: true,
	}
}

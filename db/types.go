package db

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
)

type Int64Array []int64

func NewInt64ArrayWithValue(values ...int64) Int64Array {
	return Int64Array(values)
}

// Scan implements the sql.Scanner interface for reading from database
func (a *Int64Array) Scan(value interface{}) error {
	if value == nil {
		*a = Int64Array{}
		return nil
	}

	switch v := value.(type) {
	case string:
		return a.scanString(v)
	case []byte:
		return a.scanString(string(v))
	default:
		return fmt.Errorf("cannot scan %T into Int64Array", value)
	}
}

// scanString parses PostgreSQL array string format like "{1,2,3}"
func (a *Int64Array) scanString(s string) error {
	if s == "" || s == "{}" {
		*a = Int64Array{}
		return nil
	}

	// Remove curly braces
	s = strings.Trim(s, "{}")
	if s == "" {
		*a = Int64Array{}
		return nil
	}

	// Split by comma
	parts := strings.Split(s, ",")
	result := make(Int64Array, len(parts))

	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		val, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid integer in array: %s", part)
		}
		result[i] = val
	}

	*a = result
	return nil
}

// Value implements the driver.Valuer interface for writing to database
func (a Int64Array) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "{}", nil
	}

	var builder strings.Builder
	builder.WriteByte('{')

	for i, val := range a {
		if i > 0 {
			builder.WriteByte(',')
		}
		builder.WriteString(strconv.FormatInt(val, 10))
	}

	builder.WriteByte('}')
	return builder.String(), nil
}

// GormDataType tells GORM what data type to use
func (Int64Array) GormDataType() string {
	return "integer[]"
}

// String returns a string representation of the array
func (a Int64Array) String() string {
	if len(a) == 0 {
		return "[]"
	}

	var builder strings.Builder
	builder.WriteByte('[')

	for i, val := range a {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(strconv.FormatInt(val, 10))
	}

	builder.WriteByte(']')
	return builder.String()
}

// Contains checks if the array contains a specific value
func (a Int64Array) Contains(val int64) bool {
	for _, v := range a {
		if v == val {
			return true
		}
	}
	return false
}

// Add appends a value to the array if it doesn't already exist
func (a *Int64Array) Add(val int64) {
	if !a.Contains(val) {
		*a = append(*a, val)
	}
}

// Remove removes a value from the array
func (a *Int64Array) Remove(val int64) {
	for i, v := range *a {
		if v == val {
			*a = append((*a)[:i], (*a)[i+1:]...)
			return
		}
	}
}

// ToSlice returns a regular []int64 slice
func (a Int64Array) ToSlice() []int64 {
	return []int64(a)
}

// FromSlice creates Int64Array from []int64
func FromSlice(slice []int64) Int64Array {
	return Int64Array(slice)
}

// StringArray represents a PostgreSQL text array
type StringArray []string

// Scan implements the sql.Scanner interface for reading from database
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = StringArray{}
		return nil
	}

	switch v := value.(type) {
	case string:
		return a.scanString(v)
	case []byte:
		return a.scanString(string(v))
	default:
		return fmt.Errorf("cannot scan %T into StringArray", value)
	}
}

// scanString parses PostgreSQL array string format like "{\"hello\",\"world\"}"
func (a *StringArray) scanString(s string) error {
	if s == "" || s == "{}" {
		*a = StringArray{}
		return nil
	}

	// Remove outer curly braces
	s = strings.Trim(s, "{}")
	if s == "" {
		*a = StringArray{}
		return nil
	}

	// Parse PostgreSQL array format
	result := make(StringArray, 0)
	var current strings.Builder
	inQuotes := false
	escaped := false

	for _, char := range s {
		switch char {
		case '"':
			if escaped {
				current.WriteRune(char)
				escaped = false
			} else {
				inQuotes = !inQuotes
			}
		case '\\':
			if escaped {
				current.WriteRune(char)
				escaped = false
			} else {
				escaped = true
			}
		case ',':
			if escaped {
				current.WriteRune(char)
				escaped = false
			} else if inQuotes {
				current.WriteRune(char)
			} else {
				// End of current element
				result = append(result, current.String())
				current.Reset()
			}
		default:
			if escaped {
				current.WriteRune('\\')
				escaped = false
			}
			current.WriteRune(char)
		}
	}

	// Add the last element
	if current.Len() > 0 || len(result) > 0 {
		result = append(result, current.String())
	}

	*a = result
	return nil
}

// Value implements the driver.Valuer interface for writing to database
func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "{}", nil
	}

	var builder strings.Builder
	builder.WriteByte('{')

	for i, val := range a {
		if i > 0 {
			builder.WriteByte(',')
		}

		// Escape and quote the string
		builder.WriteByte('"')
		for _, char := range val {
			switch char {
			case '"':
				builder.WriteString(`\"`)
			case '\\':
				builder.WriteString(`\\`)
			default:
				builder.WriteRune(char)
			}
		}
		builder.WriteByte('"')
	}

	builder.WriteByte('}')
	return builder.String(), nil
}

// GormDataType tells GORM what data type to use
func (StringArray) GormDataType() string {
	return "text[]"
}

// String returns a string representation of the array
func (a StringArray) String() string {
	if len(a) == 0 {
		return "[]"
	}

	var builder strings.Builder
	builder.WriteByte('[')

	for i, val := range a {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteByte('"')
		builder.WriteString(val)
		builder.WriteByte('"')
	}

	builder.WriteByte(']')
	return builder.String()
}

// Contains checks if the array contains a specific value
func (a StringArray) Contains(val string) bool {
	for _, v := range a {
		if v == val {
			return true
		}
	}
	return false
}

// ContainsIgnoreCase checks if the array contains a specific value (case insensitive)
func (a StringArray) ContainsIgnoreCase(val string) bool {
	lowerVal := strings.ToLower(val)
	for _, v := range a {
		if strings.ToLower(v) == lowerVal {
			return true
		}
	}
	return false
}

// Add appends a value to the array if it doesn't already exist
func (a *StringArray) Add(val string) {
	if !a.Contains(val) {
		*a = append(*a, val)
	}
}

// Remove removes a value from the array
func (a *StringArray) Remove(val string) {
	for i, v := range *a {
		if v == val {
			*a = append((*a)[:i], (*a)[i+1:]...)
			return
		}
	}
}

// Filter returns a new StringArray with elements that satisfy the predicate
func (a StringArray) Filter(predicate func(string) bool) StringArray {
	result := make(StringArray, 0)
	for _, val := range a {
		if predicate(val) {
			result = append(result, val)
		}
	}
	return result
}

// Join concatenates all elements with the given separator
func (a StringArray) Join(separator string) string {
	return strings.Join([]string(a), separator)
}

// ToSlice returns a regular []string slice
func (a StringArray) ToSlice() []string {
	return []string(a)
}

// FromStringSlice creates StringArray from []string
func FromStringSlice(slice []string) StringArray {
	return StringArray(slice)
}

// IsEmpty returns true if the array is empty
func (a StringArray) IsEmpty() bool {
	return len(a) == 0
}

// Len returns the length of the array
func (a StringArray) Len() int {
	return len(a)
}

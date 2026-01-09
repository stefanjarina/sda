package utils

import (
	"testing"
)

func TestEnum_String(t *testing.T) {
	enum := NewEnum([]string{"a", "b", "c"}, "a")

	result := enum.String()

	if result != "a" {
		t.Errorf("Expected 'a', got '%s'", result)
	}
}

func TestEnum_Set_Valid(t *testing.T) {
	enum := NewEnum([]string{"a", "b", "c"}, "a")

	err := enum.Set("b")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if enum.Value != "b" {
		t.Errorf("Expected 'b', got '%s'", enum.Value)
	}
}

func TestEnum_Set_Invalid(t *testing.T) {
	enum := NewEnum([]string{"a", "b", "c"}, "a")

	err := enum.Set("d")

	if err == nil {
		t.Error("Expected error for invalid value")
	}
}

func TestEnum_Type(t *testing.T) {
	enum := NewEnum([]string{"a", "b", "c"}, "a")

	result := enum.Type()

	if result != "string" {
		t.Errorf("Expected 'string', got '%s'", result)
	}
}

func TestNewEnum(t *testing.T) {
	allowed := []string{"option1", "option2"}
	defaultVal := "option1"

	enum := NewEnum(allowed, defaultVal)

	if enum.Allowed[0] != "option1" {
		t.Error("Expected allowed options to be set")
	}
	if enum.Value != defaultVal {
		t.Errorf("Expected default value '%s', got '%s'", defaultVal, enum.Value)
	}
}

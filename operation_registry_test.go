package calcdate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetOperationRegistry(t *testing.T) {
	registry := GetOperationRegistry()

	// Verify we have categories
	assert.NotEmpty(t, registry, "Registry should not be empty")
	assert.Greater(t, len(registry), 5, "Should have multiple categories")

	// Verify each category has operations
	for _, cat := range registry {
		assert.NotEmpty(t, cat.Name, "Category should have a name")
		assert.NotEmpty(t, cat.Operations, "Category should have operations")

		// Verify each operation has required fields
		for _, op := range cat.Operations {
			assert.NotEmpty(t, op.Name, "Operation should have a name")
			assert.NotEmpty(t, op.Category, "Operation should have a category")
			assert.NotEmpty(t, op.Description, "Operation should have a description")
			assert.NotEmpty(t, op.Example, "Operation should have an example")
		}
	}
}

func TestOperationRegistryCategories(t *testing.T) {
	registry := GetOperationRegistry()

	expectedCategories := []string{
		"Date Values",
		"Arithmetic Operations",
		"Time Units",
		"Boundaries - Day",
		"Boundaries - Week",
		"Boundaries - Month",
		"Boundaries - Year",
		"Boundaries - Quarter",
		"Boundaries - Time",
		"Value Setters",
		"Transform Operations",
		"Range Operations",
		"Pipeline Operations",
	}

	// Build map of actual categories
	actualCategories := make(map[string]bool)
	for _, cat := range registry {
		actualCategories[cat.Name] = true
	}

	// Verify all expected categories exist
	for _, expected := range expectedCategories {
		assert.True(t, actualCategories[expected], "Expected category %s not found", expected)
	}
}

func TestOperationRegistryKeyOperations(t *testing.T) {
	registry := GetOperationRegistry()

	// Flatten all operations into a map for easier lookup
	operations := make(map[string]OperationInfo)
	for _, cat := range registry {
		for _, op := range cat.Operations {
			operations[op.Name] = op
		}
	}

	// Verify key operations exist
	keyOps := []string{
		"today",
		"now",
		"yesterday",
		"tomorrow",
		"start",
		"startofday",
		"end",
		"endofday",
		"startofweek",
		"endofweek",
		"startofmonth",
		"endofmonth",
		"startofyear",
		"endofyear",
		"startofquarter",
		"endofquarter",
	}

	for _, opName := range keyOps {
		op, exists := operations[opName]
		assert.True(t, exists, "Key operation %s should exist", opName)
		assert.NotEmpty(t, op.Description, "Operation %s should have description", opName)
		assert.NotEmpty(t, op.Example, "Operation %s should have example", opName)
	}
}

func TestOperationRegistryAliases(t *testing.T) {
	registry := GetOperationRegistry()

	// Find operations with aliases
	var opsWithAliases []OperationInfo
	for _, cat := range registry {
		for _, op := range cat.Operations {
			if len(op.Aliases) > 0 {
				opsWithAliases = append(opsWithAliases, op)
			}
		}
	}

	// Verify "start" and "startofday" are aliases
	found := false
	for _, op := range opsWithAliases {
		if op.Name == "start" || op.Name == "startofday" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should have start/startofday aliases documented")
}

func TestOperationRegistryExamplesFormat(t *testing.T) {
	registry := GetOperationRegistry()

	for _, cat := range registry {
		for _, op := range cat.Operations {
			// All examples should contain "calcdate"
			assert.Contains(t, op.Example, "calcdate", "Example for %s should contain 'calcdate'", op.Name)
			// All examples should contain "-x" flag
			assert.Contains(t, op.Example, "-x", "Example for %s should contain '-x' flag", op.Name)
		}
	}
}

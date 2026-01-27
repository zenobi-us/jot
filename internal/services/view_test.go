package services

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenobi-us/opennotes/internal/core"
)

func TestViewService_BuiltinViews_Initialization(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")

	// Check all 6 built-in views are initialized
	expectedViews := []string{"today", "recent", "kanban", "untagged", "orphans", "broken-links"}
	for _, viewName := range expectedViews {
		view, err := vs.GetView(viewName)
		assert.NoError(t, err, "view %s should exist", viewName)
		assert.NotNil(t, view, "view %s should not be nil", viewName)
		assert.Equal(t, viewName, view.Name)
		assert.NotEmpty(t, view.Description)
	}
}

func TestViewService_TodayView(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view, err := vs.GetView("today")

	require.NoError(t, err)
	assert.Equal(t, "today", view.Name)
	assert.Contains(t, view.Description, "today")
	assert.Equal(t, 1, len(view.Query.Conditions))
	assert.Equal(t, "{{today}}", view.Query.Conditions[0].Value)
}

func TestViewService_RecentView(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view, err := vs.GetView("recent")

	require.NoError(t, err)
	assert.Equal(t, "recent", view.Name)
	assert.Equal(t, 20, view.Query.Limit)
	assert.Equal(t, "metadata->>'updated_at' DESC", view.Query.OrderBy)
}

func TestViewService_KanbanView_HasParameter(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view, err := vs.GetView("kanban")

	require.NoError(t, err)
	assert.Equal(t, 1, len(view.Parameters))
	assert.Equal(t, "status", view.Parameters[0].Name)
	assert.Equal(t, "list", view.Parameters[0].Type)
	assert.Equal(t, "backlog,todo,in-progress,reviewing,testing,deploying,done", view.Parameters[0].Default)
}

func TestViewService_UntaggedView(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view, err := vs.GetView("untagged")

	require.NoError(t, err)
	assert.Equal(t, "untagged", view.Name)
	assert.Equal(t, 1, len(view.Query.Conditions))
	assert.Equal(t, "metadata->>'tags'", view.Query.Conditions[0].Field)
	assert.Equal(t, "IS NULL", view.Query.Conditions[0].Operator)
}

func TestViewService_OrphansView(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view, err := vs.GetView("orphans")

	require.NoError(t, err)
	assert.Equal(t, "orphans", view.Name)
	assert.Equal(t, 1, len(view.Parameters))
	assert.Equal(t, "definition", view.Parameters[0].Name)
}

func TestViewService_BrokenLinksView(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view, err := vs.GetView("broken-links")

	require.NoError(t, err)
	assert.Equal(t, "broken-links", view.Name)
	assert.Contains(t, view.Description, "link")
}

func TestViewService_ResolveTemplateVariables_Today(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	now := time.Now()
	expected := now.Format("2006-01-02")

	result := vs.ResolveTemplateVariables("{{today}}")
	assert.Equal(t, expected, result)
}

func TestViewService_ResolveTemplateVariables_Yesterday(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	now := time.Now()
	expected := now.AddDate(0, 0, -1).Format("2006-01-02")

	result := vs.ResolveTemplateVariables("{{yesterday}}")
	assert.Equal(t, expected, result)
}

func TestViewService_ResolveTemplateVariables_ThisWeek(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	now := time.Now()
	expected := getStartOfWeek(now).Format("2006-01-02")

	result := vs.ResolveTemplateVariables("{{this_week}}")
	assert.Equal(t, expected, result)
}

func TestViewService_ResolveTemplateVariables_ThisMonth(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	now := time.Now()
	expected := now.Format("2006-01") + "-01"

	result := vs.ResolveTemplateVariables("{{this_month}}")
	assert.Equal(t, expected, result)
}

func TestViewService_ResolveTemplateVariables_Now(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	result := vs.ResolveTemplateVariables("{{now}}")

	// Parse result as RFC3339 to verify format
	_, err = time.Parse(time.RFC3339, result)
	assert.NoError(t, err)
}

func TestViewService_ResolveTemplateVariables_Multiple(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	result := vs.ResolveTemplateVariables("From {{yesterday}} to {{today}}")

	assert.NotContains(t, result, "{{")
	assert.Contains(t, result, "to")
}

func TestViewService_ValidateViewDefinition_ValidView(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name:        "test-view",
		Description: "Test view",
		Query: core.ViewQuery{
			Conditions: []core.ViewCondition{
				{
					Field:    "metadata->>'created_at'",
					Operator: "=",
					Value:    "test",
				},
			},
		},
	}

	err = vs.ValidateViewDefinition(view)
	assert.NoError(t, err)
}

func TestViewService_ValidateViewDefinition_InvalidName(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name:        "test view!",
		Description: "Test view",
	}

	err = vs.ValidateViewDefinition(view)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid view name")
}

func TestViewService_ValidateViewDefinition_TooManyConditions(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name: "test-view",
	}

	// Add 11 conditions
	for i := 0; i < 11; i++ {
		view.Query.Conditions = append(view.Query.Conditions, core.ViewCondition{
			Field:    "data.test",
			Operator: "=",
			Value:    "test",
		})
	}

	err = vs.ValidateViewDefinition(view)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too many conditions")
}

func TestViewService_ValidateViewDefinition_InvalidField(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name: "test-view",
		Query: core.ViewQuery{
			Conditions: []core.ViewCondition{
				{
					Field:    "invalid_field",
					Operator: "=",
					Value:    "test",
				},
			},
		},
	}

	err = vs.ValidateViewDefinition(view)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid field")
}

func TestViewService_ValidateViewDefinition_InvalidOperator(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name: "test-view",
		Query: core.ViewQuery{
			Conditions: []core.ViewCondition{
				{
					Field:    "metadata->>'created_at'",
					Operator: "INVALID",
					Value:    "test",
				},
			},
		},
	}

	err = vs.ValidateViewDefinition(view)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid operator")
}

func TestViewService_ValidateParameters_RequiredParameter(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name: "test-view",
		Parameters: []core.ViewParameter{
			{
				Name:     "author",
				Type:     "string",
				Required: true,
			},
		},
	}

	// Missing required parameter should error
	err = vs.ValidateParameters(view, map[string]string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required parameter")

	// Providing required parameter should succeed
	err = vs.ValidateParameters(view, map[string]string{"author": "John"})
	assert.NoError(t, err)
}

func TestViewService_ValidateParameters_UnknownParameter(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name: "test-view",
		Parameters: []core.ViewParameter{
			{
				Name: "author",
				Type: "string",
			},
		},
	}

	err = vs.ValidateParameters(view, map[string]string{"unknown": "value"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown parameter")
}

func TestViewService_ValidateParamType_String(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	param := &core.ViewParameter{Type: "string"}

	// Valid string
	err = vs.validateParamType(param, "test")
	assert.NoError(t, err)

	// String too long
	err = vs.validateParamType(param, string(make([]byte, 257)))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too long")
}

func TestViewService_ValidateParamType_List(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	param := &core.ViewParameter{Type: "list"}

	// Valid list
	err = vs.validateParamType(param, "item1,item2,item3")
	assert.NoError(t, err)

	// Empty item
	err = vs.validateParamType(param, "item1,,item3")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty list item")
}

func TestViewService_ValidateParamType_Date(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	param := &core.ViewParameter{Type: "date"}

	// Valid date
	err = vs.validateParamType(param, "2026-01-20")
	assert.NoError(t, err)

	// Invalid date
	err = vs.validateParamType(param, "2026/01/20")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid date")
}

func TestViewService_ValidateParamType_Bool(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	param := &core.ViewParameter{Type: "bool"}

	// Valid booleans
	err = vs.validateParamType(param, "true")
	assert.NoError(t, err)

	err = vs.validateParamType(param, "FALSE")
	assert.NoError(t, err)

	// Invalid boolean
	err = vs.validateParamType(param, "maybe")
	assert.Error(t, err)
}

func TestViewService_ApplyParameterDefaults(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Parameters: []core.ViewParameter{
			{
				Name:    "author",
				Default: "unknown",
			},
			{
				Name:    "priority",
				Default: "normal",
			},
		},
	}

	result := vs.ApplyParameterDefaults(view, map[string]string{"author": "John"})

	assert.Equal(t, "John", result["author"])
	assert.Equal(t, "normal", result["priority"])
}

func TestViewService_ParseViewParameters_Valid(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	params, err := vs.ParseViewParameters("author=John,status=todo,priority=high")

	assert.NoError(t, err)
	assert.Equal(t, "John", params["author"])
	assert.Equal(t, "todo", params["status"])
	assert.Equal(t, "high", params["priority"])
}

func TestViewService_ParseViewParameters_Empty(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	params, err := vs.ParseViewParameters("")

	assert.NoError(t, err)
	assert.Empty(t, params)
}

func TestViewService_ParseViewParameters_InvalidFormat(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	_, err = vs.ParseViewParameters("invalid-format")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid parameter format")
}

func TestViewService_FormatQueryValue_IN(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	result := vs.FormatQueryValue("IN", "todo,in-progress,done")

	assert.Equal(t, "('todo','in-progress','done')", result)
}

func TestViewService_FormatQueryValue_LIKE(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	result := vs.FormatQueryValue("LIKE", "test%")

	assert.Equal(t, "'test%'", result)
}

func TestViewService_FormatQueryValue_Number(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	result := vs.FormatQueryValue("=", "42")

	assert.Equal(t, "42", result)
}

func TestViewService_FormatQueryValue_String(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	result := vs.FormatQueryValue("=", "workflow")

	assert.Equal(t, "'workflow'", result)
}

func TestViewService_GetView_NotFound(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	_, err = vs.GetView("nonexistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "view not found")
}

func TestViewService_GetView_GlobalConfig(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	config := map[string]interface{}{
		"views": map[string]interface{}{
			"custom-view": map[string]interface{}{
				"name":        "custom-view",
				"description": "Custom test view",
				"query": map[string]interface{}{
					"order_by": "updated DESC",
				},
			},
		},
	}

	configJSON, err := json.Marshal(config)
	require.NoError(t, err)
	err = os.WriteFile(configPath, configJSON, 0644)
	require.NoError(t, err)

	cfg, err := NewConfigServiceWithPath(configPath)
	require.NoError(t, err)

	vs := NewViewServiceWithConfigPath(cfg, "", configPath)
	view, err := vs.GetView("custom-view")

	assert.NoError(t, err, "expected to find custom-view")
	assert.Equal(t, "custom-view", view.Name)
	assert.Equal(t, "Custom test view", view.Description)
}

func TestViewService_GetView_NotebookConfigOverride(t *testing.T) {
	// Create temporary notebook directory
	tmpDir := t.TempDir()
	notebookConfigPath := filepath.Join(tmpDir, NotebookConfigFile)

	config := map[string]interface{}{
		"views": map[string]interface{}{
			"today": map[string]interface{}{
				"name":        "today",
				"description": "Notebook override",
				"query": map[string]interface{}{
					"order_by": "updated DESC",
				},
			},
		},
	}

	configJSON, err := json.Marshal(config)
	require.NoError(t, err)
	err = os.WriteFile(notebookConfigPath, configJSON, 0644)
	require.NoError(t, err)

	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, tmpDir)
	view, err := vs.GetView("today")

	assert.NoError(t, err)
	assert.Equal(t, "Notebook override", view.Description)
}

func TestViewService_Precedence_NotebookOverGlobal(t *testing.T) {
	// Create temporary config directory
	tmpDir := t.TempDir()
	globalConfigPath := filepath.Join(tmpDir, "global-config.json")

	globalConfig := map[string]interface{}{
		"views": map[string]interface{}{
			"test": map[string]interface{}{
				"description": "Global view",
			},
		},
	}

	globalJSON, err := json.Marshal(globalConfig)
	require.NoError(t, err)
	err = os.WriteFile(globalConfigPath, globalJSON, 0644)
	require.NoError(t, err)

	// Create notebook directory
	notebookDir := filepath.Join(tmpDir, "notebook")
	err = os.Mkdir(notebookDir, 0755)
	require.NoError(t, err)

	notebookConfigPath := filepath.Join(notebookDir, NotebookConfigFile)
	notebookConfig := map[string]interface{}{
		"views": map[string]interface{}{
			"test": map[string]interface{}{
				"description": "Notebook view",
			},
		},
	}

	notebookJSON, err := json.Marshal(notebookConfig)
	require.NoError(t, err)
	err = os.WriteFile(notebookConfigPath, notebookJSON, 0644)
	require.NoError(t, err)

	cfg, err := NewConfigServiceWithPath(globalConfigPath)
	require.NoError(t, err)

	vs := NewViewService(cfg, notebookDir)
	view, err := vs.GetView("test")

	assert.NoError(t, err)
	assert.Equal(t, "Notebook view", view.Description)
}

func TestViewService_Precedence_GlobalOverBuiltin(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	config := map[string]interface{}{
		"views": map[string]interface{}{
			"today": map[string]interface{}{
				"name":        "today",
				"description": "Global override of built-in",
			},
		},
	}

	configJSON, err := json.Marshal(config)
	require.NoError(t, err)
	err = os.WriteFile(configPath, configJSON, 0644)
	require.NoError(t, err)

	cfg, err := NewConfigServiceWithPath(configPath)
	require.NoError(t, err)

	vs := NewViewServiceWithConfigPath(cfg, "", configPath)
	view, err := vs.GetView("today")

	assert.NoError(t, err)
	assert.Equal(t, "Global override of built-in", view.Description)
}

func TestViewService_GenerateSQL_SimpleCondition(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name: "test",
		Query: core.ViewQuery{
			Conditions: []core.ViewCondition{
				{
					Field:    "created",
					Operator: "=",
					Value:    "2026-01-20",
				},
			},
		},
	}

	sql, args, err := vs.GenerateSQL(view, nil)
	assert.NoError(t, err)
	assert.Contains(t, sql, "read_markdown(?, include_filepath:=true)")
	assert.Contains(t, sql, "WHERE created = ?")
	// Only the condition value, glob is added by caller
	assert.Equal(t, []interface{}{"2026-01-20"}, args)
}

func TestViewService_GenerateSQL_WithTemplateVariables(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name: "today-test",
		Query: core.ViewQuery{
			Conditions: []core.ViewCondition{
				{
					Field:    "created",
					Operator: ">=",
					Value:    "{{today}}",
				},
			},
		},
	}

	sql, args, err := vs.GenerateSQL(view, nil)
	assert.NoError(t, err)
	assert.Contains(t, sql, "WHERE created >= ?")

	// Today should be resolved to a date string
	assert.Len(t, args, 1)
	today := time.Now().Format("2006-01-02")
	assert.Equal(t, today, args[0])
}

func TestViewService_GenerateSQL_INOperator(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name: "kanban",
		Query: core.ViewQuery{
			Conditions: []core.ViewCondition{
				{
					Field:    "data.status",
					Operator: "IN",
					Value:    "{{status}}",
				},
			},
		},
		Parameters: []core.ViewParameter{
			{
				Name:    "status",
				Type:    "list",
				Default: "todo,in-progress,done",
			},
		},
	}

	sql, args, err := vs.GenerateSQL(view, nil)
	assert.NoError(t, err)
	assert.Contains(t, sql, "WHERE data.status IN (?,?,?)")
	assert.Equal(t, 3, len(args))
	assert.Equal(t, "todo", args[0])
	assert.Equal(t, "in-progress", args[1])
	assert.Equal(t, "done", args[2])
}

func TestViewService_GenerateSQL_ISNULLOperator(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name: "untagged",
		Query: core.ViewQuery{
			Conditions: []core.ViewCondition{
				{
					Field:    "data.tags",
					Operator: "IS NULL",
					Value:    "",
				},
			},
		},
	}

	sql, args, err := vs.GenerateSQL(view, nil)
	assert.NoError(t, err)
	assert.Contains(t, sql, "WHERE data.tags IS NULL")
	assert.Empty(t, args)
}

func TestViewService_GenerateSQL_MultipleConditions(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name: "test",
		Query: core.ViewQuery{
			Conditions: []core.ViewCondition{
				{
					Field:    "data.status",
					Operator: "!=",
					Value:    "archived",
				},
				{
					Field:    "data.tag",
					Operator: "=",
					Value:    "workflow",
				},
			},
			OrderBy: "updated DESC",
			Limit:   50,
		},
	}

	sql, args, err := vs.GenerateSQL(view, nil)
	assert.NoError(t, err)
	assert.Contains(t, sql, "WHERE data.status != ? AND data.tag = ?")
	assert.Contains(t, sql, "ORDER BY updated DESC")
	assert.Contains(t, sql, "LIMIT 50")
	assert.Equal(t, []interface{}{"archived", "workflow"}, args)
}

func TestViewService_GenerateSQL_WithUserParameters(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name: "kanban",
		Query: core.ViewQuery{
			Conditions: []core.ViewCondition{
				{
					Field:    "data.status",
					Operator: "IN",
					Value:    "{{status}}",
				},
			},
		},
		Parameters: []core.ViewParameter{
			{
				Name:    "status",
				Type:    "list",
				Default: "todo,done",
			},
		},
	}

	// User provides custom parameter
	sql, args, err := vs.GenerateSQL(view, map[string]string{"status": "backlog,in-progress"})
	assert.NoError(t, err)
	assert.Contains(t, sql, "WHERE data.status IN (?,?)")
	assert.Equal(t, []interface{}{"backlog", "in-progress"}, args)
}

// Phase 1: GROUP BY Implementation Tests

func TestViewService_GenerateSQL_GroupBy_ValidField(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name: "count-by-status",
		Query: core.ViewQuery{
			GroupBy: "metadata->>'status'",
		},
	}

	sql, _, err := vs.GenerateSQL(view, map[string]string{})
	assert.NoError(t, err)
	assert.Contains(t, sql, "GROUP BY metadata->>'status'")
}

func TestViewService_GenerateSQL_GroupBy_InvalidField(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name: "invalid-groupby",
		Query: core.ViewQuery{
			GroupBy: "'; DROP TABLE notes; --",
		},
	}

	sql, args, err := vs.GenerateSQL(view, map[string]string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid group by field")
	assert.Equal(t, "", sql)
	assert.Nil(t, args)
}

func TestViewService_GenerateSQL_GroupBy_WithOrderBy(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name: "grouped-and-ordered",
		Query: core.ViewQuery{
			GroupBy: "metadata->>'status'",
			OrderBy: "metadata->>'status' ASC",
		},
	}

	sql, _, err := vs.GenerateSQL(view, map[string]string{})
	assert.NoError(t, err)
	assert.Contains(t, sql, "GROUP BY metadata->>'status'")
	assert.Contains(t, sql, "ORDER BY metadata->>'status' ASC")
	// Verify ORDER is after GROUP BY in SQL
	groupByPos := strings.Index(sql, "GROUP BY")
	orderByPos := strings.Index(sql, "ORDER BY")
	assert.True(t, groupByPos > 0 && orderByPos > groupByPos, "ORDER BY should come after GROUP BY")
}

// Phase 1: DISTINCT Support Tests

func TestViewService_GenerateSQL_Distinct_Basic(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name: "unique-notes",
		Query: core.ViewQuery{
			Distinct: true,
		},
	}

	sql, _, err := vs.GenerateSQL(view, map[string]string{})
	assert.NoError(t, err)
	assert.Contains(t, sql, "SELECT DISTINCT *")
}

func TestViewService_GenerateSQL_Distinct_WithWhere(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name: "unique-by-status",
		Query: core.ViewQuery{
			Distinct: true,
			Conditions: []core.ViewCondition{
				{
					Field:    "metadata->>'status'",
					Operator: "=",
					Value:    "done",
				},
			},
		},
	}

	sql, args, err := vs.GenerateSQL(view, map[string]string{})
	assert.NoError(t, err)
	assert.Contains(t, sql, "SELECT DISTINCT *")
	assert.Contains(t, sql, "WHERE metadata->>'status' = ?")
	assert.Equal(t, []interface{}{"done"}, args)
}

// Phase 1: OFFSET Support Tests

func TestViewService_GenerateSQL_Offset_WithLimit(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name: "paginated",
		Query: core.ViewQuery{
			Limit:  10,
			Offset: 20,
		},
	}

	sql, _, err := vs.GenerateSQL(view, map[string]string{})
	assert.NoError(t, err)
	assert.Contains(t, sql, "LIMIT 10")
	assert.Contains(t, sql, "OFFSET 20")
	// Verify OFFSET comes after LIMIT
	limitPos := strings.Index(sql, "LIMIT")
	offsetPos := strings.Index(sql, "OFFSET")
	assert.True(t, limitPos > 0 && offsetPos > limitPos, "OFFSET should come after LIMIT")
}

func TestViewService_GenerateSQL_Offset_Alone(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name: "skipped-results",
		Query: core.ViewQuery{
			Offset: 50,
		},
	}

	sql, _, err := vs.GenerateSQL(view, map[string]string{})
	assert.NoError(t, err)
	assert.Contains(t, sql, "OFFSET 50")
	assert.NotContains(t, sql, "LIMIT")
}

func TestViewService_GenerateSQL_Pagination_Calculation(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")

	// Simulate pagination: Page 3 with 10 items per page = OFFSET 20, LIMIT 10
	pageSize := 10
	pageNum := 3
	offset := (pageNum - 1) * pageSize

	view := &core.ViewDefinition{
		Name: "page-3",
		Query: core.ViewQuery{
			Limit:  pageSize,
			Offset: offset,
		},
	}

	sql, _, err := vs.GenerateSQL(view, map[string]string{})
	assert.NoError(t, err)
	assert.Contains(t, sql, "LIMIT 10")
	assert.Contains(t, sql, "OFFSET 20")
}

// Phase 2: HAVING Clause Tests

func TestViewService_GenerateSQL_Having_WithCountAggregate(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name: "status-count-filter",
		Query: core.ViewQuery{
			GroupBy: "metadata->>'status'",
			Having: []core.ViewCondition{
				{
					Field:    "COUNT(*)",
					Operator: ">",
					Value:    "5",
				},
			},
		},
	}

	sql, args, err := vs.GenerateSQL(view, map[string]string{})
	assert.NoError(t, err)
	assert.Contains(t, sql, "GROUP BY metadata->>'status'")
	assert.Contains(t, sql, "HAVING COUNT(*) > ?")
	assert.Equal(t, []interface{}{"5"}, args)
}

func TestViewService_GenerateSQL_Having_WithSumAggregate(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name: "sum-filter",
		Query: core.ViewQuery{
			GroupBy: "metadata->>'category'",
			Having: []core.ViewCondition{
				{
					Field:    "SUM(metadata->>'priority')",
					Operator: ">=",
					Value:    "100",
				},
			},
		},
	}

	sql, args, err := vs.GenerateSQL(view, map[string]string{})
	assert.NoError(t, err)
	assert.Contains(t, sql, "GROUP BY metadata->>'category'")
	assert.Contains(t, sql, "HAVING SUM(metadata->>'priority') >= ?")
	assert.Equal(t, []interface{}{"100"}, args)
}

func TestViewService_GenerateSQL_Having_WithMultipleConditions(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name: "multi-condition-having",
		Query: core.ViewQuery{
			GroupBy: "metadata->>'status'",
			Having: []core.ViewCondition{
				{
					Field:    "COUNT(*)",
					Operator: ">",
					Value:    "3",
				},
				{
					Field:    "AVG(metadata->>'priority')",
					Operator: "<",
					Value:    "7",
				},
			},
		},
	}

	sql, args, err := vs.GenerateSQL(view, map[string]string{})
	assert.NoError(t, err)
	assert.Contains(t, sql, "GROUP BY metadata->>'status'")
	assert.Contains(t, sql, "HAVING COUNT(*) > ?")
	assert.Contains(t, sql, "AVG(metadata->>'priority') < ?")
	assert.Equal(t, []interface{}{"3", "7"}, args)
	// Verify conditions are joined with AND
	assert.Contains(t, sql, "HAVING COUNT(*) > ? AND AVG(metadata->>'priority') < ?")
}

func TestViewService_GenerateSQL_Having_InvalidCondition(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name: "invalid-having",
		Query: core.ViewQuery{
			GroupBy: "metadata->>'status'",
			Having: []core.ViewCondition{
				{
					Field:    "'; DROP TABLE notes; --",
					Operator: ">",
					Value:    "5",
				},
			},
		},
	}

	sql, args, err := vs.GenerateSQL(view, map[string]string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid having condition")
	assert.Equal(t, "", sql)
	assert.Nil(t, args)
}

// Phase 2: Aggregate Functions Tests

func TestViewService_GenerateSQL_SelectColumns_Explicit(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name: "select-explicit",
		Query: core.ViewQuery{
			SelectColumns: []string{
				"metadata->>'title'",
				"metadata->>'status'",
			},
		},
	}

	sql, _, err := vs.GenerateSQL(view, map[string]string{})
	assert.NoError(t, err)
	assert.Contains(t, sql, "SELECT metadata->>'title', metadata->>'status'")
}

func TestViewService_GenerateSQL_AggregateColumns_Count(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name: "count-aggregate",
		Query: core.ViewQuery{
			AggregateColumns: map[string]string{
				"total_notes": "COUNT(*)",
			},
		},
	}

	sql, _, err := vs.GenerateSQL(view, map[string]string{})
	assert.NoError(t, err)
	assert.Contains(t, sql, "SELECT COUNT(*) AS total_notes")
}

func TestViewService_GenerateSQL_AggregateColumns_Sum(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name: "sum-aggregate",
		Query: core.ViewQuery{
			AggregateColumns: map[string]string{
				"total_priority": "SUM(metadata->>'priority')",
			},
		},
	}

	sql, _, err := vs.GenerateSQL(view, map[string]string{})
	assert.NoError(t, err)
	assert.Contains(t, sql, "SELECT SUM(metadata->>'priority') AS total_priority")
}

func TestViewService_GenerateSQL_AggregateColumns_AvgWithCasting(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name: "avg-aggregate",
		Query: core.ViewQuery{
			AggregateColumns: map[string]string{
				"avg_priority": "AVG((metadata->>'priority')::INTEGER)",
			},
		},
	}

	sql, _, err := vs.GenerateSQL(view, map[string]string{})
	assert.NoError(t, err)
	assert.Contains(t, sql, "SELECT AVG((metadata->>'priority')::INTEGER) AS avg_priority")
}

func TestViewService_GenerateSQL_MixingSelectAndAggregate(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name: "mixed-select",
		Query: core.ViewQuery{
			SelectColumns: []string{
				"metadata->>'status'",
			},
			AggregateColumns: map[string]string{
				"count":        "COUNT(*)",
				"avg_priority": "AVG((metadata->>'priority')::INTEGER)",
			},
		},
	}

	sql, _, err := vs.GenerateSQL(view, map[string]string{})
	assert.NoError(t, err)
	assert.Contains(t, sql, "SELECT")
	assert.Contains(t, sql, "metadata->>'status'")
	assert.Contains(t, sql, "COUNT(*) AS count")
	assert.Contains(t, sql, "AVG((metadata->>'priority')::INTEGER) AS avg_priority")
}

func TestViewService_GenerateSQL_AggregateColumns_InvalidFunction(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name: "invalid-aggregate",
		Query: core.ViewQuery{
			AggregateColumns: map[string]string{
				"malicious": "'; DROP TABLE notes; --",
			},
		},
	}

	sql, args, err := vs.GenerateSQL(view, map[string]string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid aggregate function")
	assert.Equal(t, "", sql)
	assert.Nil(t, args)
}

// Phase 2: Integration Tests

func TestViewService_GenerateSQL_Having_WithOrderBy(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name: "having-with-order",
		Query: core.ViewQuery{
			GroupBy: "metadata->>'status'",
			Having: []core.ViewCondition{
				{
					Field:    "COUNT(*)",
					Operator: ">",
					Value:    "2",
				},
			},
			OrderBy: "COUNT(*) DESC",
		},
	}

	sql, _, err := vs.GenerateSQL(view, map[string]string{})
	assert.NoError(t, err)
	// Verify clause order: GROUP BY before HAVING before ORDER BY
	groupByPos := strings.Index(sql, "GROUP BY")
	havingPos := strings.Index(sql, "HAVING")
	orderByPos := strings.Index(sql, "ORDER BY")
	assert.True(t, groupByPos > 0 && havingPos > groupByPos && orderByPos > havingPos,
		"Clause order should be: GROUP BY, HAVING, ORDER BY")
}

func TestViewService_GenerateSQL_HavingWithLimitAndOffset(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name: "complete-aggregation",
		Query: core.ViewQuery{
			GroupBy: "metadata->>'status'",
			Having: []core.ViewCondition{
				{
					Field:    "COUNT(*)",
					Operator: ">=",
					Value:    "1",
				},
			},
			Limit:  5,
			Offset: 10,
		},
	}

	sql, _, err := vs.GenerateSQL(view, map[string]string{})
	assert.NoError(t, err)
	assert.Contains(t, sql, "GROUP BY metadata->>'status'")
	assert.Contains(t, sql, "HAVING COUNT(*) >= ?")
	assert.Contains(t, sql, "LIMIT 5")
	assert.Contains(t, sql, "OFFSET 10")
	// Verify order: HAVING before LIMIT before OFFSET
	havingPos := strings.Index(sql, "HAVING")
	limitPos := strings.Index(sql, "LIMIT")
	offsetPos := strings.Index(sql, "OFFSET")
	assert.True(t, havingPos > 0 && limitPos > havingPos && offsetPos > limitPos,
		"Clause order should be: HAVING, LIMIT, OFFSET")
}

func TestViewService_GenerateSQL_AggregateWithGroupByAndHaving(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name: "full-aggregation",
		Query: core.ViewQuery{
			SelectColumns: []string{
				"metadata->>'status'",
			},
			AggregateColumns: map[string]string{
				"count":        "COUNT(*)",
				"max_priority": "MAX((metadata->>'priority')::INTEGER)",
			},
			GroupBy: "metadata->>'status'",
			Having: []core.ViewCondition{
				{
					Field:    "COUNT(*)",
					Operator: ">",
					Value:    "0",
				},
			},
			OrderBy: "count DESC",
			Limit:   10,
		},
	}

	sql, args, err := vs.GenerateSQL(view, map[string]string{})
	assert.NoError(t, err)

	// Verify SELECT clause has all columns
	assert.Contains(t, sql, "SELECT")
	assert.Contains(t, sql, "metadata->>'status'")
	assert.Contains(t, sql, "COUNT(*) AS count")
	assert.Contains(t, sql, "MAX((metadata->>'priority')::INTEGER) AS max_priority")

	// Verify GROUP BY
	assert.Contains(t, sql, "GROUP BY metadata->>'status'")

	// Verify HAVING
	assert.Contains(t, sql, "HAVING COUNT(*) > ?")

	// Verify ORDER BY and LIMIT
	assert.Contains(t, sql, "ORDER BY count DESC")
	assert.Contains(t, sql, "LIMIT 10")

	// Verify argument values
	assert.Equal(t, []interface{}{"0"}, args)
}

// Phase 3: Time Arithmetic Tests

func TestViewService_ResolveTemplateVariables_TodayPlusN(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	now := time.Now()
	expected := now.AddDate(0, 0, 7).Format("2006-01-02")

	result := vs.ResolveTemplateVariables("{{today+7}}")
	assert.Equal(t, expected, result)
}

func TestViewService_ResolveTemplateVariables_TodayMinusN(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	now := time.Now()
	expected := now.AddDate(0, 0, -30).Format("2006-01-02")

	result := vs.ResolveTemplateVariables("{{today-30}}")
	assert.Equal(t, expected, result)
}

func TestViewService_ResolveTemplateVariables_ThisWeekPlusN(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	now := time.Now()
	targetDate := now.AddDate(0, 0, 14)
	expected := getStartOfWeek(targetDate).Format("2006-01-02")

	result := vs.ResolveTemplateVariables("{{this_week+2}}")
	assert.Equal(t, expected, result)
}

func TestViewService_ResolveTemplateVariables_ThisMonthPlusN(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	now := time.Now()
	targetDate := now.AddDate(0, 3, 0)
	expected := getFirstOfMonth(targetDate).Format("2006-01-02")

	result := vs.ResolveTemplateVariables("{{this_month+3}}")
	assert.Equal(t, expected, result)
}

func TestViewService_ResolveTemplateVariables_EndOfMonth(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	now := time.Now()
	expected := getEndOfMonth(now).Format("2006-01-02")

	result := vs.ResolveTemplateVariables("{{end_of_month}}")
	assert.Equal(t, expected, result)
}

func TestViewService_ResolveTemplateVariables_StartOfMonth(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	now := time.Now()
	expected := getFirstOfMonth(now).Format("2006-01-02")

	result := vs.ResolveTemplateVariables("{{start_of_month}}")
	assert.Equal(t, expected, result)
}

func TestViewService_ResolveTemplateVariables_NextWeek(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	now := time.Now()
	nextWeekDate := now.AddDate(0, 0, 7)
	expected := getStartOfWeek(nextWeekDate).Format("2006-01-02")

	result := vs.ResolveTemplateVariables("{{next_week}}")
	assert.Equal(t, expected, result)
}

func TestViewService_ResolveTemplateVariables_NextMonth(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	now := time.Now()
	nextMonthDate := now.AddDate(0, 1, 0)
	expected := getFirstOfMonth(nextMonthDate).Format("2006-01-02")

	result := vs.ResolveTemplateVariables("{{next_month}}")
	assert.Equal(t, expected, result)
}

func TestViewService_ResolveTemplateVariables_LastWeek(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	now := time.Now()
	lastWeekDate := now.AddDate(0, 0, -7)
	expected := getStartOfWeek(lastWeekDate).Format("2006-01-02")

	result := vs.ResolveTemplateVariables("{{last_week}}")
	assert.Equal(t, expected, result)
}

func TestViewService_ResolveTemplateVariables_LastMonth(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	now := time.Now()
	lastMonthDate := now.AddDate(0, -1, 0)
	expected := getFirstOfMonth(lastMonthDate).Format("2006-01-02")

	result := vs.ResolveTemplateVariables("{{last_month}}")
	assert.Equal(t, expected, result)
}

func TestViewService_ResolveTemplateVariables_Quarter(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	result := vs.ResolveTemplateVariables("{{quarter}}")

	// Should be Q1, Q2, Q3, or Q4
	assert.True(t, strings.HasPrefix(result, "Q"), "Quarter should start with Q")
	assert.True(t, len(result) == 2, "Quarter should be 2 characters (e.g., Q1)")
}

func TestViewService_ResolveTemplateVariables_Year(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	result := vs.ResolveTemplateVariables("{{year}}")
	expected := time.Now().Format("2006")

	assert.Equal(t, expected, result)
}

func TestViewService_ResolveTemplateVariables_StartOfQuarter(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	expected := getStartOfQuarter(time.Now()).Format("2006-01-02")

	result := vs.ResolveTemplateVariables("{{start_of_quarter}}")
	assert.Equal(t, expected, result)
}

func TestViewService_ResolveTemplateVariables_EndOfQuarter(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	expected := getEndOfQuarter(time.Now()).Format("2006-01-02")

	result := vs.ResolveTemplateVariables("{{end_of_quarter}}")
	assert.Equal(t, expected, result)
}

// Phase 3: Environment Variables Tests

func TestViewService_ResolveTemplateVariables_EnvironmentVariable(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")

	// Set an environment variable for this test
	envVar := "TEST_OPENNOTES_VAR"
	envValue := "test_value_123"
	t.Setenv(envVar, envValue)

	result := vs.ResolveTemplateVariables("{{env:TEST_OPENNOTES_VAR}}")
	assert.Equal(t, envValue, result)
}

func TestViewService_ResolveTemplateVariables_EnvironmentVariableWithDefault(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")

	// Don't set the env var, should use default
	result := vs.ResolveTemplateVariables("{{env:default_value:NONEXISTENT_VAR_XYZ}}")
	assert.Equal(t, "default_value", result)
}

func TestViewService_ResolveTemplateVariables_EnvironmentVariableNotSet(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")

	// Don't set env var, should return empty string (but log warning)
	result := vs.ResolveTemplateVariables("{{env:NONEXISTENT_VAR_ABC}}")
	assert.Equal(t, "", result)
}

func TestViewService_ResolveTemplateVariables_EnvironmentVariableWithDefaultOverride(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")

	// Set the env var, should override default
	envVar := "TEST_OVERRIDE_VAR"
	envValue := "actual_value"
	t.Setenv(envVar, envValue)

	result := vs.ResolveTemplateVariables("{{env:default_value:TEST_OVERRIDE_VAR}}")
	assert.Equal(t, envValue, result)
}

// Phase 3: Integration Tests (Multiple Patterns)

func TestViewService_ResolveTemplateVariables_TimeArithmeticMonthBoundary(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	now := time.Now()

	// Test month arithmetic crossing year boundary (Dec -> Jan)
	targetDate := now.AddDate(0, 6, 0) // 6 months forward
	expected := getFirstOfMonth(targetDate).Format("2006-01-02")

	result := vs.ResolveTemplateVariables("{{this_month+6}}")
	assert.Equal(t, expected, result)
}

func TestViewService_ResolveTemplateVariables_MultiplePatterns(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	now := time.Now()

	// Test multiple patterns in one string
	today := now.Format("2006-01-02")
	tomorrow := now.AddDate(0, 0, 1).Format("2006-01-02")

	result := vs.ResolveTemplateVariables("From {{today}} to {{today+1}}")
	assert.Contains(t, result, today)
	assert.Contains(t, result, tomorrow)
	assert.NotContains(t, result, "{{")
}

func TestViewService_ResolveTemplateVariables_EnvironmentAndTimePatterns(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")

	t.Setenv("TEST_STATUS", "active")
	now := time.Now()
	today := now.Format("2006-01-02")

	result := vs.ResolveTemplateVariables("status={{env:TEST_STATUS}} since={{today}}")
	assert.Contains(t, result, "status=active")
	assert.Contains(t, result, "since="+today)
	assert.NotContains(t, result, "{{")
}

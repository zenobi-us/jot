package services

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenobi-us/jot/internal/core"
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

// TestBuiltinViews_DSLFormat validates that builtin views use the new DSL pipe syntax.
// DSL format: "filter DSL | directives" where filter is parsed by DSL parser
// and directives control sorting, limiting, grouping.
func TestBuiltinViews_DSLFormat(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")

	tests := []struct {
		name        string
		viewName    string
		expectQuery string
		expectType  string
	}{
		{
			name:        "today view uses DSL",
			viewName:    "today",
			expectQuery: "modified:>=today | sort:modified:desc",
		},
		{
			name:        "recent view uses DSL",
			viewName:    "recent",
			expectQuery: "| sort:modified:desc limit:20",
		},
		{
			name:        "kanban view uses DSL",
			viewName:    "kanban",
			expectQuery: "has:status | group:status sort:title:asc",
		},
		{
			name:        "untagged view uses DSL",
			viewName:    "untagged",
			expectQuery: "missing:tag | sort:created:desc",
		},
		{
			name:       "orphans is special view",
			viewName:   "orphans",
			expectType: "special",
		},
		{
			name:       "broken-links is special view",
			viewName:   "broken-links",
			expectType: "special",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view, err := vs.GetView(tt.viewName)
			require.NoError(t, err, "builtin view %s should exist", tt.viewName)
			require.NotNil(t, view, "builtin view %s not found", tt.viewName)

			if tt.expectQuery != "" {
				assert.Equal(t, tt.expectQuery, view.Query, "view %s query mismatch", tt.viewName)
			}
			if tt.expectType != "" {
				assert.Equal(t, tt.expectType, view.Type, "view %s type mismatch", tt.viewName)
			}
		})
	}
}

// TestBuiltinViews_DSLFormat_SpecialViewsHaveNoQuery verifies that special views
// don't have a Query string (they're executed by special handlers).
func TestBuiltinViews_DSLFormat_SpecialViewsHaveNoQuery(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")

	specialViews := []string{"orphans", "broken-links"}
	for _, viewName := range specialViews {
		view, err := vs.GetView(viewName)
		require.NoError(t, err)
		assert.True(t, view.IsSpecialView(), "view %s should be special", viewName)
		assert.Empty(t, view.Query, "special view %s should have empty query", viewName)
	}
}

// TestBuiltinViews_DSLFormat_QueryViewsHavePipeSyntax verifies that query-based views
// use the pipe syntax for separating filter from directives.
func TestBuiltinViews_DSLFormat_QueryViewsHavePipeSyntax(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")

	queryViews := []string{"today", "recent", "kanban", "untagged"}
	for _, viewName := range queryViews {
		view, err := vs.GetView(viewName)
		require.NoError(t, err)
		assert.False(t, view.IsSpecialView(), "view %s should not be special", viewName)
		assert.Contains(t, view.Query, "|", "query view %s should have pipe syntax", viewName)
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
	// New DSL format validation
	assert.Contains(t, view.Query, "modified:>=today")
	assert.Contains(t, view.Query, "sort:modified:desc")
}

func TestViewService_RecentView(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view, err := vs.GetView("recent")

	require.NoError(t, err)
	assert.Equal(t, "recent", view.Name)
	// New DSL format validation
	assert.Contains(t, view.Query, "sort:modified:desc")
	assert.Contains(t, view.Query, "limit:20")
}

func TestViewService_KanbanView(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view, err := vs.GetView("kanban")

	require.NoError(t, err)
	assert.Equal(t, "kanban", view.Name)
	// New DSL format validation
	assert.Contains(t, view.Query, "has:status")
	assert.Contains(t, view.Query, "group:status")
}

func TestViewService_UntaggedView(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view, err := vs.GetView("untagged")

	require.NoError(t, err)
	assert.Equal(t, "untagged", view.Name)
	// New DSL format validation
	assert.Contains(t, view.Query, "missing:tag")
	assert.Contains(t, view.Query, "sort:created:desc")
}

func TestViewService_OrphansView(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view, err := vs.GetView("orphans")

	require.NoError(t, err)
	assert.Equal(t, "orphans", view.Name)
	// Orphans is a special view
	assert.True(t, view.IsSpecialView())
	assert.Equal(t, "special", view.Type)
}

func TestViewService_BrokenLinksView(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view, err := vs.GetView("broken-links")

	require.NoError(t, err)
	assert.Equal(t, "broken-links", view.Name)
	assert.Contains(t, view.Description, "link")
	assert.True(t, view.IsSpecialView())
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
		Query:       "tag:work | sort:modified:desc",
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

func TestViewService_ValidateViewDefinition_SpecialView(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")
	view := &core.ViewDefinition{
		Name:        "orphans-custom",
		Description: "Custom orphans view",
		Type:        "special",
	}

	err = vs.ValidateViewDefinition(view)
	assert.NoError(t, err)
}

// TestViewService_ValidateViewDefinition_DSLErrors validates that the DSL-aware
// validation correctly rejects invalid filter DSL and directive syntax.
func TestViewService_ValidateViewDefinition_DSLErrors(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")

	t.Run("invalid filter DSL syntax", func(t *testing.T) {
		view := &core.ViewDefinition{
			Name:  "bad-filter",
			Query: "tag::invalid | sort:modified:desc", // double colon is invalid
		}
		err := vs.ValidateViewDefinition(view)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid filter DSL")
	})

	t.Run("invalid directive", func(t *testing.T) {
		view := &core.ViewDefinition{
			Name:  "bad-directive",
			Query: "tag:work | baddir:value", // unknown directive
		}
		err := vs.ValidateViewDefinition(view)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid directives")
	})

	t.Run("invalid limit directive", func(t *testing.T) {
		view := &core.ViewDefinition{
			Name:  "bad-limit",
			Query: "tag:work | limit:abc", // limit must be numeric
		}
		err := vs.ValidateViewDefinition(view)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid directives")
	})

	t.Run("valid DSL view passes", func(t *testing.T) {
		view := &core.ViewDefinition{
			Name:  "valid-dsl",
			Query: "tag:work status:todo | sort:modified:desc limit:20",
		}
		err := vs.ValidateViewDefinition(view)
		assert.NoError(t, err)
	})

	t.Run("empty query is valid", func(t *testing.T) {
		view := &core.ViewDefinition{
			Name:  "empty-query",
			Query: "",
		}
		err := vs.ValidateViewDefinition(view)
		assert.NoError(t, err)
	})

	t.Run("directives-only query is valid", func(t *testing.T) {
		view := &core.ViewDefinition{
			Name:  "directives-only",
			Query: "| sort:modified:desc limit:10",
		}
		err := vs.ValidateViewDefinition(view)
		assert.NoError(t, err)
	})
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
				"query":       "tag:custom | sort:modified:desc",
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
				"query":       "modified:>=today | sort:title:asc",
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
				"name":        "test",
				"description": "Global view",
				"query":       "tag:global",
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
				"name":        "test",
				"description": "Notebook view",
				"query":       "tag:notebook",
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
				"query":       "modified:>=today | sort:title:asc",
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

// Time arithmetic tests

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
	assert.True(t, len(result) == 2 && result[0] == 'Q', "Quarter should be Q1-Q4")
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

// Environment variable tests

func TestViewService_ResolveTemplateVariables_EnvironmentVariable(t *testing.T) {
	cfg, err := NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := NewViewService(cfg, "")

	// Set an environment variable for this test
	envVar := "TEST_JOT_VAR"
	envValue := "test_value_123"
	t.Setenv(envVar, envValue)

	result := vs.ResolveTemplateVariables("{{env:TEST_JOT_VAR}}")
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

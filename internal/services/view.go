package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/zenobi-us/opennotes/internal/core"
)

// ViewService manages named reusable query presets
type ViewService struct {
	configService    *ConfigService
	notebookPath     string
	globalConfigPath string
	builtinViews     map[string]*core.ViewDefinition
	log              zerolog.Logger
}

// NewViewService creates a new ViewService
func NewViewService(cfg *ConfigService, notebookPath string) *ViewService {
	return NewViewServiceWithConfigPath(cfg, notebookPath, GlobalConfigFile())
}

// NewViewServiceWithConfigPath creates a new ViewService with a custom config path (for testing)
func NewViewServiceWithConfigPath(cfg *ConfigService, notebookPath string, globalConfigPath string) *ViewService {
	vs := &ViewService{
		configService:    cfg,
		notebookPath:     notebookPath,
		globalConfigPath: globalConfigPath,
		builtinViews:     make(map[string]*core.ViewDefinition),
		log:              Log("ViewService"),
	}

	// Initialize built-in views
	vs.initializeBuiltinViews()

	return vs
}

// initializeBuiltinViews creates all 6 built-in view definitions
func (vs *ViewService) initializeBuiltinViews() {
	// Today view: Notes created or updated today
	vs.builtinViews["today"] = &core.ViewDefinition{
		Name:        "today",
		Description: "Notes created or updated today",
		Query: core.ViewQuery{
			Conditions: []core.ViewCondition{
				{
					Logic:    "AND",
					Field:    "metadata->>'created_at'",
					Operator: ">=",
					Value:    "{{today}}",
				},
			},
			OrderBy: "metadata->>'updated_at' DESC",
			Limit:   50,
		},
	}

	// Recent view: Recently modified notes (last 20)
	vs.builtinViews["recent"] = &core.ViewDefinition{
		Name:        "recent",
		Description: "Recently modified notes (last 20)",
		Query: core.ViewQuery{
			OrderBy: "metadata->>'updated_at' DESC",
			Limit:   20,
		},
	}

	// Kanban view: Notes grouped by status
	vs.builtinViews["kanban"] = &core.ViewDefinition{
		Name:        "kanban",
		Description: "Notes grouped by status column",
		Parameters: []core.ViewParameter{
			{
				Name:        "status",
				Type:        "list",
				Required:    false,
				Default:     "backlog,todo,in-progress,reviewing,testing,deploying,done",
				Description: "Comma-separated list of status values",
			},
		},
		Query: core.ViewQuery{
			Conditions: []core.ViewCondition{
				{
					Logic:    "AND",
					Field:    "metadata->>'status'",
					Operator: "IN",
					Value:    "{{status}}",
				},
			},
			OrderBy: "(metadata->>'priority')::INTEGER DESC, metadata->>'updated_at' DESC",
		},
	}

	// Untagged view: Notes without any tags
	vs.builtinViews["untagged"] = &core.ViewDefinition{
		Name:        "untagged",
		Description: "Notes without any tags",
		Query: core.ViewQuery{
			Conditions: []core.ViewCondition{
				{
					Logic:    "AND",
					Field:    "metadata->>'tags'",
					Operator: "IS NULL",
					Value:    "",
				},
			},
			OrderBy: "metadata->>'created_at' DESC",
		},
	}

	// Orphans view: Notes with no incoming links
	vs.builtinViews["orphans"] = &core.ViewDefinition{
		Name:        "orphans",
		Description: "Notes with no incoming links (no other notes reference them)",
		Parameters: []core.ViewParameter{
			{
				Name:        "definition",
				Type:        "string",
				Required:    false,
				Default:     "no-incoming",
				Description: "Definition type: no-incoming, no-links, or isolated",
			},
		},
		Query: core.ViewQuery{
			OrderBy: "metadata->>'created_at' DESC",
		},
	}

	// Broken links view: Notes with broken links
	vs.builtinViews["broken-links"] = &core.ViewDefinition{
		Name:        "broken-links",
		Description: "Notes containing links to non-existent files",
		Query: core.ViewQuery{
			OrderBy: "metadata->>'updated_at' DESC",
		},
	}
}

// GetView retrieves a view by name, checking hierarchy: notebook > global > built-in
func (vs *ViewService) GetView(name string) (*core.ViewDefinition, error) {
	// 1. Check notebook-specific views (if in notebook context)
	if vs.notebookPath != "" {
		if view, err := vs.loadNotebookView(name); err == nil && view != nil {
			vs.log.Debug().Str("name", name).Msg("Found view in notebook config")
			return view, nil
		}
	}

	// 2. Check global config views
	if view, err := vs.loadGlobalView(name); err == nil && view != nil {
		vs.log.Debug().Str("name", name).Msg("Found view in global config")
		return view, nil
	}

	// 3. Check built-in views
	if view, ok := vs.builtinViews[name]; ok {
		vs.log.Debug().Str("name", name).Msg("Found built-in view")
		return view, nil
	}

	return nil, fmt.Errorf("view not found: %s", name)
}

// loadNotebookView loads a view from notebook .opennotes.json
func (vs *ViewService) loadNotebookView(name string) (*core.ViewDefinition, error) {
	if vs.notebookPath == "" {
		return nil, nil
	}

	configPath := filepath.Join(vs.notebookPath, NotebookConfigFile)
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read notebook config: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse notebook config: %w", err)
	}

	views, ok := config["views"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	viewData, ok := views[name]
	if !ok {
		return nil, nil
	}

	// Re-marshal to raw JSON for parsing
	rawData, err := json.Marshal(viewData)
	if err != nil {
		return nil, err
	}

	return core.ParseViewDefinition(rawData)
}

// loadGlobalView loads a view from global config
func (vs *ViewService) loadGlobalView(name string) (*core.ViewDefinition, error) {
	configPath := vs.globalConfigPath
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read global config: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse global config: %w", err)
	}

	views, ok := config["views"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	viewData, ok := views[name]
	if !ok {
		return nil, nil
	}

	// Re-marshal to raw JSON for parsing
	rawData, err := json.Marshal(viewData)
	if err != nil {
		return nil, err
	}

	return core.ParseViewDefinition(rawData)
}

// ResolveTemplateVariables resolves template variables in a string
func (vs *ViewService) ResolveTemplateVariables(value string) string {
	now := time.Now()

	replacements := map[string]string{
		"{{today}}":      now.Format("2006-01-02"),
		"{{yesterday}}":  now.AddDate(0, 0, -1).Format("2006-01-02"),
		"{{this_week}}":  getStartOfWeek(now).Format("2006-01-02"),
		"{{this_month}}": now.Format("2006-01") + "-01",
		"{{now}}":        now.Format(time.RFC3339),
	}

	for placeholder, replacement := range replacements {
		value = strings.ReplaceAll(value, placeholder, replacement)
	}

	return value
}

// getStartOfWeek returns the start of the week (Monday)
func getStartOfWeek(t time.Time) time.Time {
	// Monday as start of week
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday is 0, make it 7
	}
	offset := 1 - weekday
	return t.AddDate(0, 0, offset)
}

// ValidateViewDefinition validates a view definition for security and correctness
func (vs *ViewService) ValidateViewDefinition(view *core.ViewDefinition) error {
	// Validate view name
	if !isValidViewName(view.Name) {
		return fmt.Errorf("invalid view name: %s (must be alphanumeric with hyphens)", view.Name)
	}

	// Validate conditions count
	if len(view.Query.Conditions) > 10 {
		return fmt.Errorf("too many conditions (max 10)")
	}

	// Validate each condition
	for _, cond := range view.Query.Conditions {
		if err := vs.validateViewCondition(cond); err != nil {
			return err
		}
	}

	// Validate parameters count
	if len(view.Parameters) > 5 {
		return fmt.Errorf("too many parameters (max 5)")
	}

	// Validate each parameter
	for _, param := range view.Parameters {
		if err := vs.validateViewParameter(param); err != nil {
			return err
		}
	}

	return nil
}

// validateViewCondition validates a single condition
func (vs *ViewService) validateViewCondition(cond core.ViewCondition) error {
	// Validate field name
	if err := validateField(cond.Field); err != nil {
		return fmt.Errorf("invalid field in condition: %w", err)
	}

	// Validate operator
	if err := validateOperator(cond.Operator); err != nil {
		return fmt.Errorf("invalid operator in condition: %w", err)
	}

	return nil
}

// validateViewParameter validates a single parameter
func (vs *ViewService) validateViewParameter(param core.ViewParameter) error {
	// Validate parameter type
	validTypes := map[string]bool{
		"string": true,
		"list":   true,
		"date":   true,
		"bool":   true,
	}

	if !validTypes[param.Type] {
		return fmt.Errorf("invalid parameter type: %s (must be string, list, date, or bool)", param.Type)
	}

	return nil
}

// ValidateParameters validates user-provided parameters against view definition
func (vs *ViewService) ValidateParameters(view *core.ViewDefinition, params map[string]string) error {
	// Check required parameters
	for _, param := range view.Parameters {
		if param.Required {
			if _, ok := params[param.Name]; !ok {
				return fmt.Errorf("missing required parameter: %s", param.Name)
			}
		}
	}

	// Validate parameter types
	paramMap := make(map[string]core.ViewParameter)
	for _, param := range view.Parameters {
		paramMap[param.Name] = param
	}

	for name, value := range params {
		param, ok := paramMap[name]
		if !ok {
			return fmt.Errorf("unknown parameter: %s", name)
		}

		if err := vs.validateParamType(&param, value); err != nil {
			return fmt.Errorf("invalid parameter %s: %w", name, err)
		}
	}

	return nil
}

// validateParamType validates a parameter value against its type
func (vs *ViewService) validateParamType(param *core.ViewParameter, value string) error {
	switch param.Type {
	case "string":
		if len(value) > 256 {
			return fmt.Errorf("string too long (max 256 chars)")
		}
	case "list":
		items := strings.Split(value, ",")
		for _, item := range items {
			if len(strings.TrimSpace(item)) == 0 {
				return fmt.Errorf("empty list item")
			}
		}
	case "date":
		if _, err := time.Parse("2006-01-02", value); err != nil {
			return fmt.Errorf("invalid date format (expected YYYY-MM-DD)")
		}
	case "bool":
		lower := strings.ToLower(value)
		if lower != "true" && lower != "false" {
			return fmt.Errorf("invalid boolean (expected true or false)")
		}
	}
	return nil
}

// isValidViewName checks if a view name is valid
func isValidViewName(name string) bool {
	// Allow alphanumeric characters, hyphens, underscores
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, name)
	return matched && len(name) > 0 && len(name) <= 64
}

// validateField checks if a field name is whitelisted
func validateField(field string) error {
	// Whitelist of allowed field prefixes
	allowedPrefixes := []string{
		"metadata->>", // JSON field extraction (primary access pattern)
		"metadata->",  // JSON object access
		"path",
		"file_path",
		"content",
		"stats->", // File statistics JSON
		"stats->>",
	}

	// Remove quotes if present
	cleanField := strings.Trim(field, "\"'")

	// Special case: allow casting syntax like (metadata->>'priority')::INTEGER
	if strings.Contains(cleanField, "::") {
		parts := strings.Split(cleanField, "::")
		if len(parts) == 2 {
			cleanField = strings.TrimSpace(strings.Trim(parts[0], "()"))
		}
	}

	for _, prefix := range allowedPrefixes {
		if cleanField == prefix || strings.HasPrefix(cleanField, prefix) {
			return nil
		}
	}

	return fmt.Errorf("field not allowed: %s", field)
}

// validateOperator checks if an operator is whitelisted
func validateOperator(operator string) error {
	allowedOperators := map[string]bool{
		"=":       true,
		"!=":      true,
		"<":       true,
		">":       true,
		"<=":      true,
		">=":      true,
		"LIKE":    true,
		"IN":      true,
		"IS NULL": true,
	}

	if !allowedOperators[operator] {
		return fmt.Errorf("operator not allowed: %s", operator)
	}

	return nil
}

// ApplyParameterDefaults applies default values to parameters
func (vs *ViewService) ApplyParameterDefaults(view *core.ViewDefinition, params map[string]string) map[string]string {
	result := make(map[string]string)

	// Copy provided parameters
	for k, v := range params {
		result[k] = v
	}

	// Apply defaults for missing parameters
	for _, param := range view.Parameters {
		if _, exists := result[param.Name]; !exists && param.Default != "" {
			result[param.Name] = param.Default
		}
	}

	return result
}

// ParseViewParameters parses view parameters from string flag format (key=value,key2=value2)
func (vs *ViewService) ParseViewParameters(paramStr string) (map[string]string, error) {
	params := make(map[string]string)

	if paramStr == "" {
		return params, nil
	}

	parts := strings.Split(paramStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid parameter format: %s (expected key=value)", part)
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		if key == "" {
			return nil, fmt.Errorf("empty parameter name in: %s", part)
		}

		params[key] = value
	}

	return params, nil
}

// FormatQueryValue formats a value for SQL based on operator type
func (vs *ViewService) FormatQueryValue(operator string, value string) string {
	switch operator {
	case "IN":
		// For IN operator, format as list of strings
		items := strings.Split(value, ",")
		formatted := make([]string, 0, len(items))
		for _, item := range items {
			item = strings.TrimSpace(item)
			formatted = append(formatted, fmt.Sprintf("'%s'", escapeSQL(item)))
		}
		return "(" + strings.Join(formatted, ",") + ")"
	case "LIKE":
		return fmt.Sprintf("'%s'", escapeSQL(value))
	case "IS NULL":
		return ""
	default:
		// For other operators, try to parse as number or string
		if _, err := strconv.ParseFloat(value, 64); err == nil {
			return value
		}
		return fmt.Sprintf("'%s'", escapeSQL(value))
	}
}

// escapeSQL escapes single quotes in SQL strings
func escapeSQL(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// GenerateSQL generates SQL from a view definition with parameters
func (vs *ViewService) GenerateSQL(view *core.ViewDefinition, params map[string]string) (string, []interface{}, error) {
	// Validate parameters
	if err := vs.ValidateParameters(view, params); err != nil {
		return "", nil, err
	}

	// Apply parameter defaults
	resolvedParams := vs.ApplyParameterDefaults(view, params)

	// Resolve template variables
	for key, value := range resolvedParams {
		resolvedParams[key] = vs.ResolveTemplateVariables(value)
	}

	// Build WHERE clause
	var conditions []string
	var args []interface{}

	for _, cond := range view.Query.Conditions {
		// Resolve parameter placeholders in value
		value := cond.Value
		if strings.HasPrefix(value, "{{") && strings.HasSuffix(value, "}}") {
			paramName := strings.Trim(value, "{}")
			if paramValue, ok := resolvedParams[paramName]; ok {
				value = paramValue
			}
		}

		// Resolve template variables
		value = vs.ResolveTemplateVariables(value)

		// Build condition SQL based on operator
		var condSQL string
		switch cond.Operator {
		case "IS NULL":
			condSQL = fmt.Sprintf("%s IS NULL", cond.Field)
		case "IN":
			// Parse comma-separated values
			items := strings.Split(value, ",")
			placeholders := make([]string, len(items))
			for i, item := range items {
				placeholders[i] = "?"
				args = append(args, strings.TrimSpace(item))
			}
			condSQL = fmt.Sprintf("%s IN (%s)", cond.Field, strings.Join(placeholders, ","))
		case "LIKE":
			condSQL = fmt.Sprintf("%s LIKE ?", cond.Field)
			args = append(args, value)
		default:
			// Standard operators: =, !=, <, >, <=, >=
			condSQL = fmt.Sprintf("%s %s ?", cond.Field, cond.Operator)
			args = append(args, value)
		}

		conditions = append(conditions, condSQL)
	}

	// Build query using read_markdown for notebook-relative file access
	// Note: The glob pattern (first parameter) is added by the caller
	selectClause := "SELECT *"
	if view.Query.Distinct {
		selectClause = "SELECT DISTINCT *"
	}
	query := selectClause + " FROM read_markdown(?, include_filepath:=true)"

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	if view.Query.GroupBy != "" {
		if err := validateField(view.Query.GroupBy); err != nil {
			return "", nil, fmt.Errorf("invalid group by field: %w", err)
		}
		query += " GROUP BY " + view.Query.GroupBy
	}

	if view.Query.OrderBy != "" {
		query += " ORDER BY " + view.Query.OrderBy
	}

	if view.Query.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", view.Query.Limit)
	}

	if view.Query.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", view.Query.Offset)
	}

	return query, args, nil
}

// ListAllViews returns all available views across all sources (built-in, global, notebook)
// sorted by origin: built-in first, then global, then notebook
func (vs *ViewService) ListAllViews() ([]core.ViewInfo, error) {
	var allViews []core.ViewInfo
	seenViews := make(map[string]bool) // Track seen view names to avoid duplicates

	// 1. Add built-in views
	builtinViews := vs.ListBuiltinViews()
	for _, view := range builtinViews {
		allViews = append(allViews, view)
		seenViews[view.Name] = true
	}

	// 2. Add global views (if not already in built-in)
	globalViews, err := vs.LoadAllGlobalViews()
	if err != nil {
		vs.log.Warn().Err(err).Msg("Failed to load global views")
		// Don't fail - continue with what we have
	} else {
		for _, view := range globalViews {
			if !seenViews[view.Name] {
				allViews = append(allViews, view)
				seenViews[view.Name] = true
			}
		}
	}

	// 3. Add notebook-specific views (if not already in built-in or global)
	if vs.notebookPath != "" {
		notebookViews, err := vs.LoadAllNotebookViews()
		if err != nil {
			vs.log.Warn().Err(err).Msg("Failed to load notebook views")
			// Don't fail - continue with what we have
		} else {
			for _, view := range notebookViews {
				if !seenViews[view.Name] {
					allViews = append(allViews, view)
					seenViews[view.Name] = true
				}
			}
		}
	}

	return allViews, nil
}

// ListBuiltinViews returns all built-in views as ViewInfo structs
func (vs *ViewService) ListBuiltinViews() []core.ViewInfo {
	var views []core.ViewInfo

	for _, view := range vs.builtinViews {
		views = append(views, core.ViewInfo{
			Name:        view.Name,
			Origin:      "built-in",
			Description: view.Description,
			Parameters:  view.Parameters,
		})
	}

	return views
}

// LoadAllGlobalViews loads all views from global config as ViewInfo structs
func (vs *ViewService) LoadAllGlobalViews() ([]core.ViewInfo, error) {
	var views []core.ViewInfo

	configPath := vs.globalConfigPath
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return views, nil
		}
		return nil, fmt.Errorf("failed to read global config: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse global config: %w", err)
	}

	viewsData, ok := config["views"].(map[string]interface{})
	if !ok {
		return views, nil
	}

	for name, viewData := range viewsData {
		rawData, err := json.Marshal(viewData)
		if err != nil {
			vs.log.Warn().Str("name", name).Err(err).Msg("Failed to marshal global view")
			continue
		}

		view, err := core.ParseViewDefinition(rawData)
		if err != nil {
			vs.log.Warn().Str("name", name).Err(err).Msg("Failed to parse global view")
			continue
		}

		views = append(views, core.ViewInfo{
			Name:        view.Name,
			Origin:      "global",
			Description: view.Description,
			Parameters:  view.Parameters,
		})
	}

	return views, nil
}

// LoadAllNotebookViews loads all views from notebook config as ViewInfo structs
func (vs *ViewService) LoadAllNotebookViews() ([]core.ViewInfo, error) {
	var views []core.ViewInfo

	if vs.notebookPath == "" {
		return views, nil
	}

	configPath := filepath.Join(vs.notebookPath, NotebookConfigFile)
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return views, nil
		}
		return nil, fmt.Errorf("failed to read notebook config: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse notebook config: %w", err)
	}

	viewsData, ok := config["views"].(map[string]interface{})
	if !ok {
		return views, nil
	}

	for name, viewData := range viewsData {
		rawData, err := json.Marshal(viewData)
		if err != nil {
			vs.log.Warn().Str("name", name).Err(err).Msg("Failed to marshal notebook view")
			continue
		}

		view, err := core.ParseViewDefinition(rawData)
		if err != nil {
			vs.log.Warn().Str("name", name).Err(err).Msg("Failed to parse notebook view")
			continue
		}

		views = append(views, core.ViewInfo{
			Name:        view.Name,
			Origin:      "notebook",
			Description: view.Description,
			Parameters:  view.Parameters,
		})
	}

	return views, nil
}

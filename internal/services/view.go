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
	"github.com/zenobi-us/jot/internal/core"
	"github.com/zenobi-us/jot/internal/search/parser"
)

// ViewService manages named reusable query presets
type ViewService struct {
	configService    *ConfigService
	notebookPath     string
	globalConfigPath string
	builtinViews     map[string]*core.ViewDefinition
	log              zerolog.Logger
	executor         *ViewExecutor // executor for view query execution (set via SetExecutionContext)
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

// initializeBuiltinViews creates all 6 built-in view definitions using DSL query strings.
// Views use pipe syntax: "filter DSL | directives"
// Special views (orphans, broken-links) use Type: "special" for custom execution.
func (vs *ViewService) initializeBuiltinViews() {
	// Today view: Notes created or updated today
	vs.builtinViews["today"] = &core.ViewDefinition{
		Name:        "today",
		Description: "Notes created or updated today",
		Query:       "modified:>=today | sort:modified:desc",
	}

	// Recent view: Recently modified notes (last 20)
	vs.builtinViews["recent"] = &core.ViewDefinition{
		Name:        "recent",
		Description: "Recently modified notes (last 20)",
		Query:       "| sort:modified:desc limit:20",
	}

	// Kanban view: Notes grouped by status
	vs.builtinViews["kanban"] = &core.ViewDefinition{
		Name:        "kanban",
		Description: "Notes grouped by status",
		Query:       "has:status | group:status sort:title:asc",
	}

	// Untagged view: Notes without any tags
	vs.builtinViews["untagged"] = &core.ViewDefinition{
		Name:        "untagged",
		Description: "Notes without any tags",
		Query:       "missing:tag | sort:created:desc",
	}

	// Orphans view: Notes with no incoming links (special view)
	vs.builtinViews["orphans"] = &core.ViewDefinition{
		Name:        "orphans",
		Description: "Notes with no incoming links (no other notes reference them)",
		Type:        "special",
	}

	// Broken links view: Notes with broken links (special view)
	vs.builtinViews["broken-links"] = &core.ViewDefinition{
		Name:        "broken-links",
		Description: "Notes containing links to non-existent files",
		Type:        "special",
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

// loadNotebookView loads a view from notebook .jot.json
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

// SaveNotebookView validates and persists a view definition to notebook .jot.json.
// Returns true when an existing view was overwritten.
func (vs *ViewService) SaveNotebookView(view *core.ViewDefinition) (bool, error) {
	if vs.notebookPath == "" {
		return false, fmt.Errorf("notebook context required")
	}

	if strings.TrimSpace(view.Query) == "" {
		return false, fmt.Errorf("view query cannot be empty")
	}

	if err := vs.ValidateViewDefinition(view); err != nil {
		return false, err
	}

	configPath := filepath.Join(vs.notebookPath, NotebookConfigFile)
	config, err := vs.loadOrCreateNotebookConfig(configPath)
	if err != nil {
		return false, err
	}

	views, ok := config["views"].(map[string]interface{})
	if !ok || views == nil {
		views = make(map[string]interface{})
	}

	overwritten := false
	if _, exists := views[view.Name]; exists {
		overwritten = true
	}

	viewJSON, err := json.Marshal(view)
	if err != nil {
		return false, fmt.Errorf("failed to marshal view: %w", err)
	}

	var viewData map[string]interface{}
	if err := json.Unmarshal(viewJSON, &viewData); err != nil {
		return false, fmt.Errorf("failed to parse marshaled view: %w", err)
	}

	views[view.Name] = viewData
	config["views"] = views

	if err := vs.writeNotebookConfig(configPath, config); err != nil {
		return false, err
	}

	return overwritten, nil
}

// DeleteNotebookView removes a view from notebook .jot.json.
// Returns true when a view existed and was deleted.
func (vs *ViewService) DeleteNotebookView(name string) (bool, error) {
	if vs.notebookPath == "" {
		return false, fmt.Errorf("notebook context required")
	}

	if !isValidViewName(name) {
		return false, fmt.Errorf("invalid view name: %s (must be alphanumeric with hyphens)", name)
	}

	configPath := filepath.Join(vs.notebookPath, NotebookConfigFile)
	config, err := vs.loadOrCreateNotebookConfig(configPath)
	if err != nil {
		return false, err
	}

	views, ok := config["views"].(map[string]interface{})
	if !ok || views == nil {
		return false, nil
	}

	if _, exists := views[name]; !exists {
		return false, nil
	}

	delete(views, name)
	if len(views) == 0 {
		delete(config, "views")
	} else {
		config["views"] = views
	}

	if err := vs.writeNotebookConfig(configPath, config); err != nil {
		return false, err
	}

	return true, nil
}

func (vs *ViewService) loadOrCreateNotebookConfig(configPath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]interface{}), nil
		}
		return nil, fmt.Errorf("failed to read notebook config: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse notebook config: %w", err)
	}

	if config == nil {
		config = make(map[string]interface{})
	}

	return config, nil
}

func (vs *ViewService) writeNotebookConfig(configPath string, config map[string]interface{}) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal notebook config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write notebook config: %w", err)
	}

	return nil
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

	// Static replacements (no parsing needed)
	replacements := map[string]string{
		"{{today}}":            now.Format("2006-01-02"),
		"{{yesterday}}":        now.AddDate(0, 0, -1).Format("2006-01-02"),
		"{{this_week}}":        getStartOfWeek(now).Format("2006-01-02"),
		"{{this_month}}":       now.Format("2006-01") + "-01",
		"{{start_of_month}}":   now.Format("2006-01") + "-01",
		"{{end_of_month}}":     getEndOfMonth(now).Format("2006-01-02"),
		"{{now}}":              now.Format(time.RFC3339),
		"{{next_week}}":        getStartOfWeek(now.AddDate(0, 0, 7)).Format("2006-01-02"),
		"{{next_month}}":       getFirstOfMonth(now.AddDate(0, 1, 0)).Format("2006-01-02"),
		"{{last_week}}":        getStartOfWeek(now.AddDate(0, 0, -7)).Format("2006-01-02"),
		"{{last_month}}":       getFirstOfMonth(now.AddDate(0, -1, 0)).Format("2006-01-02"),
		"{{quarter}}":          getCurrentQuarter(now),
		"{{year}}":             now.Format("2006"),
		"{{start_of_quarter}}": getStartOfQuarter(now).Format("2006-01-02"),
		"{{end_of_quarter}}":   getEndOfQuarter(now).Format("2006-01-02"),
	}

	for placeholder, replacement := range replacements {
		value = strings.ReplaceAll(value, placeholder, replacement)
	}

	// Dynamic replacements requiring pattern parsing

	// Handle {{today-N}}, {{today+N}} patterns (time arithmetic by days)
	value = resolveDayArithmetic(value, now)

	// Handle {{this_week-N}}, {{this_month-N}} patterns (time arithmetic by weeks/months)
	value = resolveWeekMonthArithmetic(value, now)

	// Handle {{env:VAR}} and {{env:DEFAULT:VAR}} patterns (environment variables)
	value = resolveEnvironmentVariables(value)

	return value
}

// resolveDayArithmetic handles {{today-N}} and {{today+N}} patterns
func resolveDayArithmetic(value string, now time.Time) string {
	// Match {{today+N}} or {{today-N}} where N is a number
	re := regexp.MustCompile(`\{\{today([+-]\d+)\}\}`)
	return re.ReplaceAllStringFunc(value, func(match string) string {
		// Extract the offset (e.g., "+7" or "-3")
		offsetStr := strings.TrimPrefix(strings.TrimSuffix(match, "}}"), "{{today")
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return match // Return unchanged if parsing fails
		}
		return now.AddDate(0, 0, offset).Format("2006-01-02")
	})
}

// resolveWeekMonthArithmetic handles {{this_week-N}}, {{this_month-N}} patterns
func resolveWeekMonthArithmetic(value string, now time.Time) string {
	// Match {{this_week-N}} or {{this_week+N}}
	reWeek := regexp.MustCompile(`\{\{this_week([+-]\d+)\}\}`)
	value = reWeek.ReplaceAllStringFunc(value, func(match string) string {
		offsetStr := strings.TrimPrefix(strings.TrimSuffix(match, "}}"), "{{this_week")
		offsetWeeks, err := strconv.Atoi(offsetStr)
		if err != nil {
			return match
		}
		targetDate := now.AddDate(0, 0, offsetWeeks*7)
		return getStartOfWeek(targetDate).Format("2006-01-02")
	})

	// Match {{this_month-N}} or {{this_month+N}}
	reMonth := regexp.MustCompile(`\{\{this_month([+-]\d+)\}\}`)
	value = reMonth.ReplaceAllStringFunc(value, func(match string) string {
		offsetStr := strings.TrimPrefix(strings.TrimSuffix(match, "}}"), "{{this_month")
		offsetMonths, err := strconv.Atoi(offsetStr)
		if err != nil {
			return match
		}
		targetDate := now.AddDate(0, offsetMonths, 0)
		return getFirstOfMonth(targetDate).Format("2006-01-02")
	})

	return value
}

// resolveEnvironmentVariables handles {{env:VAR}} and {{env:DEFAULT:VAR}} patterns
func resolveEnvironmentVariables(value string) string {
	// Match {{env:something}} patterns
	re := regexp.MustCompile(`\{\{env:([^}]+)\}\}`)
	return re.ReplaceAllStringFunc(value, func(match string) string {
		// Extract content between env: and }}
		content := strings.TrimPrefix(strings.TrimSuffix(match, "}}"), "{{env:")

		// Check if it has a default value (format: DEFAULT:VAR_NAME)
		if strings.Contains(content, ":") {
			parts := strings.SplitN(content, ":", 2)
			defaultValue := parts[0]
			varName := parts[1]

			val := os.Getenv(varName)
			if val == "" {
				return defaultValue
			}
			return val
		}

		// No default value, just substitute environment variable
		val := os.Getenv(content)
		if val == "" {
			// Log warning if env var not found but don't fail
			log := Log("ViewService")
			log.Warn().Str("var", content).Msg("Environment variable not set, using empty string")
		}
		return val
	})
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

// getFirstOfMonth returns the first day of the month
func getFirstOfMonth(t time.Time) time.Time {
	return t.AddDate(0, 0, 1-t.Day())
}

// getEndOfMonth returns the last day of the month
func getEndOfMonth(t time.Time) time.Time {
	// Get the first day of next month and subtract one day
	nextMonth := t.AddDate(0, 1, 0)
	firstOfNext := getFirstOfMonth(nextMonth)
	return firstOfNext.AddDate(0, 0, -1)
}

// getCurrentQuarter returns the current quarter (Q1, Q2, Q3, Q4)
func getCurrentQuarter(t time.Time) string {
	month := t.Month()
	if month <= 3 {
		return "Q1"
	} else if month <= 6 {
		return "Q2"
	} else if month <= 9 {
		return "Q3"
	}
	return "Q4"
}

// getStartOfQuarter returns the first day of the current quarter
func getStartOfQuarter(t time.Time) time.Time {
	month := t.Month()
	var quarterMonth int
	switch {
	case month <= 3:
		quarterMonth = 1
	case month <= 6:
		quarterMonth = 4
	case month <= 9:
		quarterMonth = 7
	default:
		quarterMonth = 10
	}
	return time.Date(t.Year(), time.Month(quarterMonth), 1, 0, 0, 0, 0, t.Location())
}

// getEndOfQuarter returns the last day of the current quarter
func getEndOfQuarter(t time.Time) time.Time {
	month := t.Month()
	var quarterMonth int
	switch {
	case month <= 3:
		quarterMonth = 3
	case month <= 6:
		quarterMonth = 6
	case month <= 9:
		quarterMonth = 9
	default:
		quarterMonth = 12
	}
	lastDay := time.Date(t.Year(), time.Month(quarterMonth)+1, 1, 0, 0, 0, 0, t.Location())
	return lastDay.AddDate(0, 0, -1)
}

// ValidateViewDefinition validates a view definition for security and correctness.
// For DSL-based views, validates that the query can be parsed.
// For special views, only validates name and parameters.
func (vs *ViewService) ValidateViewDefinition(view *core.ViewDefinition) error {
	// Validate view name
	if !isValidViewName(view.Name) {
		return fmt.Errorf("invalid view name: %s (must be alphanumeric with hyphens)", view.Name)
	}

	// Special views don't need query validation
	if view.IsSpecialView() {
		// Validate parameters count
		if len(view.Parameters) > 5 {
			return fmt.Errorf("too many parameters (max 5)")
		}
		for _, param := range view.Parameters {
			if err := vs.validateViewParameter(param); err != nil {
				return err
			}
		}
		return nil
	}

	// For DSL views, validate the query string can be parsed
	if view.Query != "" {
		// Split query into filter and directives
		filter, directives := SplitViewQuery(view.Query)

		// Validate filter DSL if present
		if filter != "" {
			p := parser.New()
			if _, err := p.Parse(filter); err != nil {
				return fmt.Errorf("invalid filter DSL: %w", err)
			}
		}

		// Validate directives if present
		if directives != "" {
			if _, err := ParseDirectives(directives); err != nil {
				return fmt.Errorf("invalid directives: %w", err)
			}
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
		if strings.HasPrefix(value, "{{") && strings.HasSuffix(value, "}}") {
			return nil
		}
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

// ListAllViews returns all available views across all sources (built-in, global, notebook)
// applying the same precedence contract as GetView: notebook > global > built-in.
func (vs *ViewService) ListAllViews() ([]core.ViewInfo, error) {
	var allViews []core.ViewInfo
	viewIndex := make(map[string]int) // name -> position in allViews

	upsert := func(view core.ViewInfo) {
		if idx, exists := viewIndex[view.Name]; exists {
			allViews[idx] = view
			return
		}
		viewIndex[view.Name] = len(allViews)
		allViews = append(allViews, view)
	}

	// 1. Seed with built-in views
	for _, view := range vs.ListBuiltinViews() {
		upsert(view)
	}

	// 2. Overlay global views
	globalViews, err := vs.LoadAllGlobalViews()
	if err != nil {
		vs.log.Warn().Err(err).Msg("Failed to load global views")
		// Don't fail - continue with built-ins
	} else {
		for _, view := range globalViews {
			upsert(view)
		}
	}

	// 3. Overlay notebook views
	if vs.notebookPath != "" {
		notebookViews, err := vs.LoadAllNotebookViews()
		if err != nil {
			vs.log.Warn().Err(err).Msg("Failed to load notebook views")
			// Don't fail - continue with what we have
		} else {
			for _, view := range notebookViews {
				upsert(view)
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

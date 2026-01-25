package services

import (
	"fmt"
	"reflect"
	"time"

	"github.com/rs/zerolog"
)

// DuckDBConverter handles conversion of DuckDB-specific types to JSON-serializable formats
type DuckDBConverter struct {
	log zerolog.Logger
}

// NewDuckDBConverter creates a new converter instance
func NewDuckDBConverter() *DuckDBConverter {
	return &DuckDBConverter{
		log: Log("DuckDBConverter"),
	}
}

// ConvertRow converts a single row from DuckDB query results to JSON-serializable format
func (c *DuckDBConverter) ConvertRow(row map[string]interface{}) (map[string]interface{}, error) {
	convertedRow := make(map[string]interface{})

	for key, value := range row {
		convertedValue, err := c.convertValue(value)
		if err != nil {
			c.log.Warn().
				Err(err).
				Str("key", key).
				Interface("value", value).
				Msg("Failed to convert column value, using fallback")
			// Use string representation as fallback
			convertedValue = fmt.Sprintf("%v", value)
		}
		convertedRow[key] = convertedValue
	}

	return convertedRow, nil
}

// ConvertResults converts all rows from DuckDB query results to JSON-serializable format
func (c *DuckDBConverter) ConvertResults(results []map[string]interface{}) ([]map[string]interface{}, error) {
	if len(results) == 0 {
		return results, nil
	}

	convertedResults := make([]map[string]interface{}, 0, len(results))

	for i, row := range results {
		convertedRow, err := c.ConvertRow(row)
		if err != nil {
			c.log.Error().
				Err(err).
				Int("rowIndex", i).
				Msg("Failed to convert row, skipping")
			continue
		}
		convertedResults = append(convertedResults, convertedRow)
	}

	return convertedResults, nil
}

// convertValue converts a single value from DuckDB to JSON-serializable format
func (c *DuckDBConverter) convertValue(value interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	// Handle time types
	if timeVal, ok := value.(time.Time); ok {
		return timeVal.Format(time.RFC3339), nil
	}

	// Use reflection to handle unknown map and slice types
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Map:
		return c.convertMapValue(rv)
	case reflect.Slice, reflect.Array:
		return c.convertArrayValue(rv)
	case reflect.Ptr:
		// Handle pointers by dereferencing
		if rv.IsNil() {
			return nil, nil
		}
		return c.convertValue(rv.Elem().Interface())
	default:
		// For primitive types, return as-is
		return value, nil
	}
}

// convertMapValue converts DuckDB MAP types to JSON objects
func (c *DuckDBConverter) convertMapValue(rv reflect.Value) (interface{}, error) {
	if rv.IsNil() {
		return nil, nil
	}

	result := make(map[string]interface{})

	// Iterate over map keys
	for _, key := range rv.MapKeys() {
		// Convert key to string
		var keyStr string
		switch key.Kind() {
		case reflect.String:
			keyStr = key.String()
		default:
			keyStr = fmt.Sprintf("%v", key.Interface())
		}

		// Convert value recursively
		mapValue := rv.MapIndex(key)
		convertedValue, err := c.convertValue(mapValue.Interface())
		if err != nil {
			c.log.Warn().
				Err(err).
				Str("mapKey", keyStr).
				Msg("Failed to convert map value, using string fallback")
			convertedValue = fmt.Sprintf("%v", mapValue.Interface())
		}

		result[keyStr] = convertedValue
	}

	return result, nil
}

// convertArrayValue converts DuckDB ARRAY types to JSON arrays
func (c *DuckDBConverter) convertArrayValue(rv reflect.Value) (interface{}, error) {
	// Check if it's a nil slice (arrays can't be nil, only slices)
	if rv.Kind() == reflect.Slice && rv.IsNil() {
		return nil, nil
	}

	length := rv.Len()
	result := make([]interface{}, 0, length)

	for i := 0; i < length; i++ {
		element := rv.Index(i)
		convertedElement, err := c.convertValue(element.Interface())
		if err != nil {
			c.log.Warn().
				Err(err).
				Int("arrayIndex", i).
				Msg("Failed to convert array element, using string fallback")
			convertedElement = fmt.Sprintf("%v", element.Interface())
		}

		result = append(result, convertedElement)
	}

	return result, nil
}

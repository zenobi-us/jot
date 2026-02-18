package core

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestViewDefinition_JSON(t *testing.T) {
	t.Run("marshals query-based view", func(t *testing.T) {
		v := ViewDefinition{
			Name:        "today",
			Description: "Notes modified today",
			Query:       "modified:>=today | sort:modified:desc",
		}

		data, err := json.Marshal(v)
		require.NoError(t, err)

		// Unmarshal back to verify round-trip
		var decoded ViewDefinition
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)
		assert.Equal(t, v.Name, decoded.Name)
		assert.Equal(t, v.Description, decoded.Description)
		assert.Equal(t, v.Query, decoded.Query)
	})

	t.Run("marshals special view", func(t *testing.T) {
		v := ViewDefinition{
			Name:        "orphans",
			Description: "Notes with no incoming links",
			Type:        "special",
		}

		data, err := json.Marshal(v)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"type":"special"`)
	})

	t.Run("marshals view with parameters", func(t *testing.T) {
		v := ViewDefinition{
			Name:        "by-tag",
			Description: "Notes with a specific tag",
			Query:       "tag:{{tagname}} | sort:title:asc",
			Parameters: []ViewParameter{
				{
					Name:        "tagname",
					Type:        "string",
					Required:    true,
					Description: "Tag to filter by",
				},
			},
		}

		data, err := json.Marshal(v)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"parameters"`)
		assert.Contains(t, string(data), `"tagname"`)
	})

	t.Run("unmarshals query-based view", func(t *testing.T) {
		jsonData := `{"name":"recent","description":"Recent notes","query":"| sort:modified:desc limit:20"}`

		var v ViewDefinition
		err := json.Unmarshal([]byte(jsonData), &v)
		require.NoError(t, err)
		assert.Equal(t, "recent", v.Name)
		assert.Equal(t, "Recent notes", v.Description)
		assert.Equal(t, "| sort:modified:desc limit:20", v.Query)
		assert.Equal(t, "", v.Type) // Default type
	})

	t.Run("unmarshals special view", func(t *testing.T) {
		jsonData := `{"name":"orphans","description":"Orphan notes","type":"special"}`

		var v ViewDefinition
		err := json.Unmarshal([]byte(jsonData), &v)
		require.NoError(t, err)
		assert.Equal(t, "orphans", v.Name)
		assert.Equal(t, "special", v.Type)
		assert.Equal(t, "", v.Query) // No query for special views
	})
}

func TestViewDefinition_IsSpecialView(t *testing.T) {
	t.Run("returns true for special type", func(t *testing.T) {
		v := ViewDefinition{Type: "special"}
		assert.True(t, v.IsSpecialView())
	})

	t.Run("returns false for query type", func(t *testing.T) {
		v := ViewDefinition{Type: "query"}
		assert.False(t, v.IsSpecialView())
	})

	t.Run("returns false for empty type (default)", func(t *testing.T) {
		v := ViewDefinition{}
		assert.False(t, v.IsSpecialView())
	})

	t.Run("returns false for any other type value", func(t *testing.T) {
		v := ViewDefinition{Type: "custom"}
		assert.False(t, v.IsSpecialView())
	})
}

func TestParseViewDefinition(t *testing.T) {
	t.Run("parses valid view JSON", func(t *testing.T) {
		data := json.RawMessage(`{
			"name": "work",
			"description": "Work related notes",
			"query": "tag:work | sort:modified:desc"
		}`)

		v, err := ParseViewDefinition(data)
		require.NoError(t, err)
		assert.Equal(t, "work", v.Name)
		assert.Equal(t, "Work related notes", v.Description)
		assert.Equal(t, "tag:work | sort:modified:desc", v.Query)
	})

	t.Run("parses view with parameters", func(t *testing.T) {
		data := json.RawMessage(`{
			"name": "by-status",
			"description": "Notes by status",
			"query": "status:{{status}} | sort:title:asc",
			"parameters": [
				{
					"name": "status",
					"type": "string",
					"required": true,
					"default": "todo"
				}
			]
		}`)

		v, err := ParseViewDefinition(data)
		require.NoError(t, err)
		assert.Equal(t, "by-status", v.Name)
		require.Len(t, v.Parameters, 1)
		assert.Equal(t, "status", v.Parameters[0].Name)
		assert.Equal(t, "string", v.Parameters[0].Type)
		assert.True(t, v.Parameters[0].Required)
		assert.Equal(t, "todo", v.Parameters[0].Default)
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		data := json.RawMessage(`{invalid json}`)

		_, err := ParseViewDefinition(data)
		assert.Error(t, err)
	})
}

func TestViewParameter(t *testing.T) {
	t.Run("marshals with all fields", func(t *testing.T) {
		p := ViewParameter{
			Name:        "date",
			Type:        "date",
			Required:    true,
			Default:     "today",
			Description: "Date to filter by",
		}

		data, err := json.Marshal(p)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"name":"date"`)
		assert.Contains(t, string(data), `"type":"date"`)
		assert.Contains(t, string(data), `"required":true`)
		assert.Contains(t, string(data), `"default":"today"`)
		assert.Contains(t, string(data), `"description":"Date to filter by"`)
	})

	t.Run("omits empty optional fields", func(t *testing.T) {
		p := ViewParameter{
			Name: "tag",
			Type: "string",
		}

		data, err := json.Marshal(p)
		require.NoError(t, err)
		// required is omitted when false due to omitempty
		assert.NotContains(t, string(data), `"required"`)
		assert.NotContains(t, string(data), `"default"`)
		assert.NotContains(t, string(data), `"description"`)
	})
}

func TestViewInfo(t *testing.T) {
	t.Run("marshals view info with origin", func(t *testing.T) {
		info := ViewInfo{
			Name:        "today",
			Origin:      "built-in",
			Description: "Notes modified today",
		}

		data, err := json.Marshal(info)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"name":"today"`)
		assert.Contains(t, string(data), `"origin":"built-in"`)
	})

	t.Run("supports different origins", func(t *testing.T) {
		origins := []string{"built-in", "global", "notebook"}
		for _, origin := range origins {
			info := ViewInfo{
				Name:   "test",
				Origin: origin,
			}
			data, err := json.Marshal(info)
			require.NoError(t, err)
			assert.Contains(t, string(data), origin)
		}
	})
}

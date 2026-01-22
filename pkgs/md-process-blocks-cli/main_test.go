package main

import (
	"testing"
)

func TestExtractCodeBlock(t *testing.T) {
	tests := []struct {
		name        string
		input       []string
		startIdx    int
		wantLang    string
		wantMeta    map[string]string
		wantContent string
		wantEnd     int
	}{
		{
			name: "simple block no attributes",
			input: []string{
				"```bash",
				"echo hello",
				"```",
			},
			startIdx:    0,
			wantLang:    "bash",
			wantMeta:    map[string]string{},
			wantContent: "echo hello",
			wantEnd:     2,
		},
		{
			name: "block with exec=true",
			input: []string{
				"```bash exec=true",
				"echo hello",
				"```",
			},
			startIdx:    0,
			wantLang:    "bash",
			wantMeta:    map[string]string{"exec": "true"},
			wantContent: "echo hello",
			wantEnd:     2,
		},
		{
			name: "block with quoted exec command",
			input: []string{
				"```d2 exec=\"d2 - -\" replace=true",
				"x -> y",
				"```",
			},
			startIdx: 0,
			wantLang: "d2",
			wantMeta: map[string]string{
				"exec":    "d2 - -",
				"replace": "true",
			},
			wantContent: "x -> y",
			wantEnd:     2,
		},
		{
			name: "multi-line content",
			input: []string{
				"```python",
				"def hello():",
				"    print('world')",
				"```",
			},
			startIdx:    0,
			wantLang:    "python",
			wantMeta:    map[string]string{},
			wantContent: "def hello():\n    print('world')",
			wantEnd:     3,
		},
		{
			name: "multiple attributes",
			input: []string{
				"```bash exec=true replace=true foo=bar",
				"command",
				"```",
			},
			startIdx: 0,
			wantLang: "bash",
			wantMeta: map[string]string{
				"exec":    "true",
				"replace": "true",
				"foo":     "bar",
			},
			wantContent: "command",
			wantEnd:     2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block, endIdx := extractCodeBlock(tt.input, tt.startIdx)

			if block == nil {
				t.Fatal("extractCodeBlock returned nil")
			}

			if block.Language != tt.wantLang {
				t.Errorf("Language = %q, want %q", block.Language, tt.wantLang)
			}

			if len(block.Metadata) != len(tt.wantMeta) {
				t.Errorf("Metadata length = %d, want %d", len(block.Metadata), len(tt.wantMeta))
			}

			for k, v := range tt.wantMeta {
				if got := block.Metadata[k]; got != v {
					t.Errorf("Metadata[%q] = %q, want %q", k, got, v)
				}
			}

			if block.Content != tt.wantContent {
				t.Errorf("Content = %q, want %q", block.Content, tt.wantContent)
			}

			if endIdx != tt.wantEnd {
				t.Errorf("endIdx = %d, want %d", endIdx, tt.wantEnd)
			}
		})
	}
}

func TestShouldProcess(t *testing.T) {
	tests := []struct {
		name  string
		block *CodeBlock
		want  bool
	}{
		{
			name: "exec=true should process",
			block: &CodeBlock{
				Language: "bash",
				Metadata: map[string]string{"exec": "true"},
			},
			want: true,
		},
		{
			name: "exec with command should process",
			block: &CodeBlock{
				Language: "d2",
				Metadata: map[string]string{"exec": "d2 - -"},
			},
			want: true,
		},
		{
			name: "no exec should not process",
			block: &CodeBlock{
				Language: "python",
				Metadata: map[string]string{},
			},
			want: false,
		},
		{
			name: "empty exec should not process",
			block: &CodeBlock{
				Language: "bash",
				Metadata: map[string]string{"exec": ""},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldProcess(tt.block); got != tt.want {
				t.Errorf("shouldProcess() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessMarkdown(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name: "no code blocks",
			input: `# Hello

This is plain text.`,
			want: `# Hello

This is plain text.
`,
			wantErr: false,
		},
		{
			name: "non-executable code block unchanged",
			input: `# Test

` + "```python\nprint('hello')\n```" + `

End`,
			want: `# Test

` + "```python\nprint('hello')\n```" + `

End
`,
			wantErr: false,
		},
		{
			name: "executable block with exec=true and replace=true",
			input: `# Test

` + "```bash exec=true replace=true\necho hello\n```" + `

End`,
			want: `# Test

hello

End
`,
			wantErr: false,
		},
		{
			name: "executable block with exec=true (append mode)",
			input: `# Test

` + "```bash exec=true\necho hello\n```" + `

End`,
			want: `# Test

` + "```bash\necho hello\n```\n\n```\nhello\n\n```" + `

End
`,
			wantErr: false,
		},
		{
			name: "exec with command using stdin",
			input: `# Test

` + "```text exec=\"cat\" replace=true\ntest data\n```" + `

End`,
			want: `# Test

test data
End
`,
			wantErr: false,
		},
		{
			name: "replace with template using {output}",
			input: `# Test

` + "```bash exec=\"echo hello\" replace=\"Result: {output}\"\nignored\n```" + `

End`,
			want: `# Test

Result: hello

End
`,
			wantErr: false,
		},
		{
			name: "multiple blocks mixed",
			input: `# Test

` + "```python\nprint('unchanged')\n```" + `

` + "```bash exec=true replace=true\necho processed\n```" + `

` + "```text\nmore unchanged\n```" + `

End`,
			want: `# Test

` + "```python\nprint('unchanged')\n```" + `

processed

` + "```text\nmore unchanged\n```" + `

End
`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processMarkdown(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("processMarkdown() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("processMarkdown() =\n%q\n\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestProcessBlockModes(t *testing.T) {
	tests := []struct {
		name    string
		block   *CodeBlock
		want    string
		wantErr bool
	}{
		{
			name: "no exec - return unchanged",
			block: &CodeBlock{
				Language: "python",
				Content:  "print('test')",
				Metadata: map[string]string{},
			},
			want:    "```python\nprint('test')\n```\n",
			wantErr: false,
		},
		{
			name: "exec=true - run content as command",
			block: &CodeBlock{
				Language: "bash",
				Content:  "echo test",
				Metadata: map[string]string{"exec": "true"},
			},
			want:    "```bash\necho test\n```\n\n```\ntest\n\n```\n",
			wantErr: false,
		},
		{
			name: "exec=true replace=true - run content, replace block",
			block: &CodeBlock{
				Language: "bash",
				Content:  "echo replaced",
				Metadata: map[string]string{
					"exec":    "true",
					"replace": "true",
				},
			},
			want:    "replaced\n",
			wantErr: false,
		},
		{
			name: "exec=command replace=true - pipe to command, replace",
			block: &CodeBlock{
				Language: "text",
				Content:  "input data",
				Metadata: map[string]string{
					"exec":    "cat",
					"replace": "true",
				},
			},
			want:    "input data",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processBlock(tt.block)
			if (err != nil) != tt.wantErr {
				t.Errorf("processBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("processBlock() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestQuotedValueParsing(t *testing.T) {
	tests := []struct {
		name     string
		fence    string
		wantMeta map[string]string
	}{
		{
			name:  "simple quoted value",
			fence: `bash exec="echo hello"`,
			wantMeta: map[string]string{
				"exec": "echo hello",
			},
		},
		{
			name:  "quoted value with dashes",
			fence: `d2 exec="d2 - -"`,
			wantMeta: map[string]string{
				"exec": "d2 - -",
			},
		},
		{
			name:  "multiple attributes with quotes",
			fence: `bash exec="cat /dev/stdin" replace=true`,
			wantMeta: map[string]string{
				"exec":    "cat /dev/stdin",
				"replace": "true",
			},
		},
		{
			name:  "unquoted and quoted mixed",
			fence: `text exec=cat foo="bar baz" qux=quux`,
			wantMeta: map[string]string{
				"exec": "cat",
				"foo":  "bar baz",
				"qux":  "quux",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := []string{"```" + tt.fence, "content", "```"}
			block, _ := extractCodeBlock(input, 0)

			if block == nil {
				t.Fatal("extractCodeBlock returned nil")
			}

			for k, wantV := range tt.wantMeta {
				gotV, ok := block.Metadata[k]
				if !ok {
					t.Errorf("Metadata missing key %q", k)
					continue
				}
				if gotV != wantV {
					t.Errorf("Metadata[%q] = %q, want %q", k, gotV, wantV)
				}
			}
		})
	}
}

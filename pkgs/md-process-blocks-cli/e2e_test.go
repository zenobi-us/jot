package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestE2E_BasicProcessing(t *testing.T) {
	// Build the binary first
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "md-process-blocks")

	buildCmd := exec.Command("go", "build", "-o", binaryPath)
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name: "stdin to stdout",
			input: `# Test
` + "```bash exec=true replace=true\necho hello\n```",
			want: `# Test
hello
`,
			wantErr: false,
		},
		{
			name: "non-executable blocks unchanged",
			input: `# Code
` + "```python\nprint('test')\n```",
			want: `# Code
` + "```python\nprint('test')\n```\n",
			wantErr: false,
		},
		{
			name: "exec with command via stdin",
			input: `# Test
` + "```text exec=\"cat\" replace=true\ntest data\n```",
			want: `# Test
test data`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath)
			cmd.Stdin = strings.NewReader(tt.input)

			output, err := cmd.Output()
			if (err != nil) != tt.wantErr {
				t.Errorf("Command error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got := string(output)
			if got != tt.want {
				t.Errorf("Output =\n%q\n\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestE2E_FileInputOutput(t *testing.T) {
	// Build the binary
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "md-process-blocks")

	buildCmd := exec.Command("go", "build", "-o", binaryPath)
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	input := `# File Test
` + "```bash exec=true replace=true\necho 'from file'\n```"

	want := `# File Test
'from file'
`

	// Create input file
	inputFile := filepath.Join(tmpDir, "input.md")
	if err := os.WriteFile(inputFile, []byte(input), 0644); err != nil {
		t.Fatalf("Failed to write input file: %v", err)
	}

	// Create output file path
	outputFile := filepath.Join(tmpDir, "output.md")

	// Run command
	cmd := exec.Command(binaryPath, "-i", inputFile, "-o", outputFile)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Read output
	output, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	got := string(output)
	if got != want {
		t.Errorf("Output =\n%q\n\nwant:\n%q", got, want)
	}
}

func TestE2E_AppendMode(t *testing.T) {
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "md-process-blocks")

	buildCmd := exec.Command("go", "build", "-o", binaryPath)
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	input := `# Append Test
` + "```bash exec=true\necho output\n```"

	// Should contain both original block and output
	cmd := exec.Command(binaryPath)
	cmd.Stdin = strings.NewReader(input)

	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	got := string(output)

	// Check that output contains both the original block and the result
	if !strings.Contains(got, "```bash\necho output\n```") {
		t.Error("Output should contain original code block")
	}

	if !strings.Contains(got, "output") {
		t.Error("Output should contain command output")
	}
}

func TestE2E_MultipleBlocks(t *testing.T) {
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "md-process-blocks")

	buildCmd := exec.Command("go", "build", "-o", binaryPath)
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	input := `# Multiple Blocks

` + "```python\nprint('unchanged')\n```" + `

` + "```bash exec=true replace=true\necho first\n```" + `

` + "```bash exec=true replace=true\necho second\n```" + `

` + "```text\nstatic\n```"

	cmd := exec.Command(binaryPath)
	cmd.Stdin = strings.NewReader(input)

	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	got := string(output)

	// Check all expected parts
	checks := []struct {
		desc    string
		want    string
		present bool
	}{
		{"python block unchanged", "```python\nprint('unchanged')\n```", true},
		{"first exec output", "first", true},
		{"second exec output", "second", true},
		{"static block unchanged", "```text\nstatic\n```", true},
		{"first bash block removed", "echo first", false},
		{"second bash block removed", "echo second", false},
	}

	for _, check := range checks {
		contains := strings.Contains(got, check.want)
		if contains != check.present {
			if check.present {
				t.Errorf("%s: expected to contain %q", check.desc, check.want)
			} else {
				t.Errorf("%s: expected NOT to contain %q", check.desc, check.want)
			}
		}
	}
}

func TestE2E_ErrorHandling(t *testing.T) {
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "md-process-blocks")

	buildCmd := exec.Command("go", "build", "-o", binaryPath)
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "non-existent command",
			input: "```bash exec=true replace=true\nnonexistentcommand123\n```",
		},
		{
			name:  "command that fails",
			input: "```bash exec=true replace=true\nfalse\n```",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath)
			cmd.Stdin = strings.NewReader(tt.input)

			_, err := cmd.Output()
			if err == nil {
				t.Error("Expected command to fail but it succeeded")
			}
		})
	}
}

func TestE2E_RealWorldD2(t *testing.T) {
	// Only run if d2 is available
	if _, err := exec.LookPath("d2"); err != nil {
		t.Skip("d2 not found in PATH, skipping")
	}

	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "md-process-blocks")

	buildCmd := exec.Command("go", "build", "-o", binaryPath)
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	input := `# D2 Diagram

` + "```d2 exec=\"d2 - -\" replace=true\nx -> y: Hello\n```"

	cmd := exec.Command(binaryPath)
	cmd.Stdin = strings.NewReader(input)

	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	got := string(output)

	// Check that we got SVG output
	if !strings.Contains(got, "<svg") {
		t.Error("Expected SVG output but didn't find <svg tag")
	}

	if !strings.Contains(got, "xmlns") {
		t.Error("Expected SVG namespace but didn't find it")
	}

	// Make sure the d2 code block is gone
	if strings.Contains(got, "```d2") {
		t.Error("D2 code block should be replaced, not present in output")
	}
}

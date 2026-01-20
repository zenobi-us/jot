package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
)

var (
	inputFile  string
	outputFile string
)

type CodeBlock struct {
	Language string
	Metadata map[string]string
	Content  string
	Start    int
	End      int
}

var rootCmd = &cobra.Command{
	Use:   "md-process-blocks",
	Short: "Process markdown code blocks with custom processors",
	Long: `Process markdown code blocks by executing commands and optionally replacing blocks with output.

Examples:
  # Process d2 blocks, replace with SVG
  md-process-blocks -i input.md -o output.md

  # Extract and execute code blocks
  md-process-blocks -i README.md`,
	RunE: run,
}

func init() {
	rootCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input markdown file (stdin if empty)")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (stdout if empty)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	// Read input
	var input []byte
	var err error

	if inputFile != "" {
		input, err = os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
	} else {
		input, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
	}

	// Process markdown
	output, err := processMarkdown(string(input))
	if err != nil {
		return fmt.Errorf("failed to process markdown: %w", err)
	}

	// Write output
	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(output), 0644); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	} else {
		fmt.Print(output)
	}

	return nil
}

func processMarkdown(content string) (string, error) {
	lines := strings.Split(content, "\n")
	var result strings.Builder
	i := 0

	for i < len(lines) {
		line := lines[i]

		// Check if this is a code block start
		if strings.HasPrefix(line, "```") && len(line) > 3 {
			block, endIdx := extractCodeBlock(lines, i)
			if block != nil {
				// Check if this block should be processed
				if shouldProcess(block) {
					processed, err := processBlock(block)
					if err != nil {
						return "", fmt.Errorf("failed to process block at line %d: %w", i+1, err)
					}
					result.WriteString(processed)
					i = endIdx + 1
					continue
				}
			}
		}

		result.WriteString(line)
		result.WriteString("\n")
		i++
	}

	return result.String(), nil
}

func extractCodeBlock(lines []string, start int) (*CodeBlock, int) {
	firstLine := lines[start]

	// Parse the first line: ```lang key=value key=value
	// Example: ```d2 exec=true replace=true
	// Example: ```bash exec="d2 - -" replace=true
	attrs := firstLine[3:] // Remove ```
	if len(attrs) == 0 {
		return nil, start
	}

	block := &CodeBlock{
		Metadata: make(map[string]string),
		Start:    start,
	}

	// Parse manually to handle quoted values properly
	i := 0
	for i < len(attrs) && !unicode.IsSpace(rune(attrs[i])) {
		i++
	}
	block.Language = attrs[0:i]

	// Skip whitespace
	for i < len(attrs) && unicode.IsSpace(rune(attrs[i])) {
		i++
	}

	// Parse key=value pairs
	for i < len(attrs) {
		// Find key
		keyStart := i
		for i < len(attrs) && attrs[i] != '=' {
			i++
		}
		if i >= len(attrs) {
			break
		}
		key := attrs[keyStart:i]
		i++ // skip '='

		// Find value
		var value string
		if i < len(attrs) && attrs[i] == '"' {
			// Quoted value
			i++ // skip opening quote
			valueStart := i
			for i < len(attrs) && attrs[i] != '"' {
				i++
			}
			value = attrs[valueStart:i]
			if i < len(attrs) {
				i++ // skip closing quote
			}
		} else {
			// Unquoted value
			valueStart := i
			for i < len(attrs) && !unicode.IsSpace(rune(attrs[i])) {
				i++
			}
			value = attrs[valueStart:i]
		}

		block.Metadata[key] = value

		// Skip whitespace
		for i < len(attrs) && unicode.IsSpace(rune(attrs[i])) {
			i++
		}
	}

	// Extract content
	var content strings.Builder
	j := start + 1
	for j < len(lines) {
		if strings.HasPrefix(lines[j], "```") {
			block.End = j
			block.Content = content.String()
			return block, j
		}
		if j > start+1 {
			content.WriteString("\n")
		}
		content.WriteString(lines[j])
		j++
	}

	return nil, start
}

func shouldProcess(block *CodeBlock) bool {
	// Only process blocks that have exec=true or exec="command"
	execVal, hasExec := block.Metadata["exec"]
	return hasExec && execVal != ""
}

func processBlock(block *CodeBlock) (string, error) {
	var result strings.Builder

	// Check if exec is enabled
	execVal, hasExec := block.Metadata["exec"]
	if !hasExec || execVal == "" {
		// Not executable, return original block
		result.WriteString(fmt.Sprintf("```%s\n", block.Language))
		result.WriteString(block.Content)
		result.WriteString("\n```\n")
		return result.String(), nil
	}

	// Determine what to execute
	var cmdStr string
	var cmdParts []string
	var useStdin bool

	// Check if exec has a command value (exec="d2 - -")
	if execVal != "true" {
		// exec="command args" - use block content as stdin
		cmdStr = execVal
		useStdin = true
	} else {
		// exec=true - execute block content as command
		cmdStr = block.Content
		useStdin = false
	}

	// Parse command
	cmdParts = strings.Fields(cmdStr)
	if len(cmdParts) == 0 {
		return "", fmt.Errorf("empty command")
	}

	// Execute command
	command := exec.Command(cmdParts[0], cmdParts[1:]...)
	if useStdin {
		command.Stdin = strings.NewReader(block.Content)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	if err := command.Run(); err != nil {
		return "", fmt.Errorf("command failed: %w\nStderr: %s", err, stderr.String())
	}

	// Check if we should replace the block
	replaceVal := block.Metadata["replace"]
	if replaceVal == "true" {
		// Replace block with output
		return stdout.String(), nil
	}

	// Append output as code block below original
	result.WriteString(fmt.Sprintf("```%s\n", block.Language))
	result.WriteString(block.Content)
	result.WriteString("\n```\n\n")
	result.WriteString("```\n")
	result.WriteString(stdout.String())
	result.WriteString("\n```\n")

	return result.String(), nil
}

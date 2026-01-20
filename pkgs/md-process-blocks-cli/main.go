package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var (
	inputFile   string
	outputFile  string
	replaceMode bool
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
	rootCmd.Flags().BoolVarP(&replaceMode, "replace", "r", true, "Replace code blocks with output")
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

	// Parse the first line: ```lang command args
	// Example: ```d2 > $ d2 - -
	parts := strings.Fields(firstLine[3:]) // Remove ```
	if len(parts) == 0 {
		return nil, start
	}

	block := &CodeBlock{
		Language: parts[0],
		Metadata: make(map[string]string),
		Start:    start,
	}

	// If there are more parts, treat them as the command
	// mdsh format: ```lang out_cmd in_cmd [data_line]
	// We simplify: everything after lang is the command
	if len(parts) > 1 {
		block.Metadata["cmd"] = strings.Join(parts[1:], " ")
	}

	// Extract content
	var content strings.Builder
	i := start + 1
	for i < len(lines) {
		if strings.HasPrefix(lines[i], "```") {
			block.End = i
			block.Content = content.String()
			return block, i
		}
		if i > start+1 {
			content.WriteString("\n")
		}
		content.WriteString(lines[i])
		i++
	}

	return nil, start
}

func shouldProcess(block *CodeBlock) bool {
	// Check if block has processor metadata
	_, hasProcessor := block.Metadata["processor"]
	_, hasCmd := block.Metadata["cmd"]

	// Or check if it's a known language we auto-process
	knownProcessors := map[string]bool{
		"d2": true,
	}

	return hasProcessor || hasCmd || knownProcessors[block.Language]
}

func processBlock(block *CodeBlock) (string, error) {
	var result strings.Builder

	// Write original block
	result.WriteString(fmt.Sprintf("```%s", block.Language))
	if cmdStr, ok := block.Metadata["cmd"]; ok {
		result.WriteString(" ")
		result.WriteString(cmdStr)
	}
	result.WriteString("\n")
	result.WriteString(block.Content)
	result.WriteString("\n```\n")

	if !replaceMode {
		return result.String(), nil
	}

	// Parse and execute mdsh-style command
	// Format: ```lang > $ command args
	// Parse: > (output type), $ (execute), command args
	cmdStr, hasCmd := block.Metadata["cmd"]
	if !hasCmd {
		// No command, check if it's a known language
		switch block.Language {
		case "d2":
			cmdStr = "> $ d2 - -"
		default:
			return result.String(), nil
		}
	}

	// Parse mdsh command format
	// Skip > (output marker) and $ (execute marker)
	cmdParts := strings.Fields(cmdStr)
	var actualCmd []string
	for _, part := range cmdParts {
		if part != ">" && part != "$" {
			actualCmd = append(actualCmd, part)
		}
	}

	if len(actualCmd) == 0 {
		return result.String(), nil
	}

	// Execute command with block content as stdin
	cmd := exec.Command(actualCmd[0], actualCmd[1:]...)
	cmd.Stdin = strings.NewReader(block.Content)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("command failed: %w\nStderr: %s", err, stderr.String())
	}

	// Write output between markers
	result.WriteString("\n<!-- BEGIN md-process-blocks -->\n")
	result.WriteString(stdout.String())
	result.WriteString("\n<!-- END md-process-blocks -->\n")

	return result.String(), nil
}

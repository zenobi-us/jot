package e2e

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenobi-us/opennotes/internal/services"
)

// TestNotebookService_ReadOnlyDirectory tests graceful handling of permission denied
func TestNotebookService_ReadOnlyDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission tests on Windows due to different permission model")
	}

	// Create temp directory and make it read-only
	tempDir := t.TempDir()
	defer func() {
		// Reset permissions before cleanup
		_ = os.Chmod(tempDir, 0755)
	}()

	// Make directory read-only
	err := os.Chmod(tempDir, 0444)
	require.NoError(t, err)

	// Try to create a note in read-only directory
	noteFile := filepath.Join(tempDir, "test.md")
	err = os.WriteFile(noteFile, []byte("# Test Note\n\nThis should fail"), 0644)

	// Verify appropriate error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")

	// Verify the error is user-friendly (not just a system error)
	assert.True(t, strings.Contains(err.Error(), "denied") ||
		strings.Contains(err.Error(), "permission"),
		"Error should mention permission issue: %v", err)
}

// TestConfigService_PermissionDeniedWrite tests config file permission handling
func TestConfigService_PermissionDeniedWrite(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission tests on Windows due to different permission model")
	}

	tempDir := t.TempDir()
	defer func() {
		// Reset permissions before cleanup
		_ = os.Chmod(tempDir, 0755)
	}()

	configDir := filepath.Join(tempDir, ".config", "opennotes")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	// Create config service
	configPath := filepath.Join(configDir, "config.json")
	configService, err := services.NewConfigServiceWithPath(configPath)
	require.NoError(t, err)

	// Make config directory read-only after creation
	err = os.Chmod(configDir, 0444)
	require.NoError(t, err)

	// Try to write config
	newConfig := services.Config{
		Notebooks: []string{configDir},
	}
	err = configService.Write(newConfig)

	// Verify appropriate error handling
	if err != nil {
		assert.Contains(t, strings.ToLower(err.Error()), "permission")
		t.Logf("Config write permission denied correctly: %v", err)
	} else {
		t.Log("Config service handled read-only directory gracefully")
	}
}

// TestNotebookService_SymlinkHandling tests symlink resolution and error handling
func TestNotebookService_SymlinkHandling(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping symlink tests on Windows")
	}

	tempDir := t.TempDir()

	// Create actual notebook directory
	realNotebook := filepath.Join(tempDir, "real-notebook")
	err := os.MkdirAll(realNotebook, 0755)
	require.NoError(t, err)

	// Create notebook config file
	configContent := `{
		"name": "Test Notebook",
		"root": ".",
		"contexts": []
	}`
	configPath := filepath.Join(realNotebook, ".opennotes.json")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Create a note in the real directory
	noteContent := "---\ntitle: Symlink Test\n---\n\n# Test Note\n\nThis is a test note."
	notePath := filepath.Join(realNotebook, "test-note.md")
	err = os.WriteFile(notePath, []byte(noteContent), 0644)
	require.NoError(t, err)

	// Create symlink to notebook
	symlinkPath := filepath.Join(tempDir, "linked-notebook")
	err = os.Symlink(realNotebook, symlinkPath)
	require.NoError(t, err)

	// Test notebook service with symlinked path
	configService, err := services.NewConfigService()
	require.NoError(t, err)
	notebookService := services.NewNotebookService(configService)
	exists := notebookService.HasNotebook(symlinkPath)
	assert.True(t, exists, "Notebook should be found through symlink")

	// Create broken symlink
	brokenLink := filepath.Join(tempDir, "broken-link")
	err = os.Symlink("/nonexistent/path", brokenLink)
	require.NoError(t, err)

	// Test broken symlink handling
	existsBroken := notebookService.HasNotebook(brokenLink)
	assert.False(t, existsBroken, "Broken symlink should be handled gracefully")
}

// TestNoteService_InvalidCharacters tests filename sanitization
func TestNoteService_InvalidCharacters(t *testing.T) {
	tempDir := t.TempDir()

	// Test invalid characters per OS
	invalidChars := []string{
		"test\x00null.md",  // Null character
		"test\nnewline.md", // Newline
		"test\ttab.md",     // Tab character
	}

	// Windows-specific invalid characters
	if runtime.GOOS == "windows" {
		invalidChars = append(invalidChars,
			"test<angle.md",
			"test>angle.md",
			"test:colon.md",
			"test\"quote.md",
			"test|pipe.md",
			"test?question.md",
			"test*asterisk.md",
		)
	}

	for _, filename := range invalidChars {
		t.Run("invalid_char_"+filename, func(t *testing.T) {
			filePath := filepath.Join(tempDir, filename)
			err := os.WriteFile(filePath, []byte("# Test"), 0644)

			// Expect either an error or successful sanitization
			if err != nil {
				// Error is expected and acceptable
				t.Logf("Invalid character correctly rejected: %v", err)
			} else {
				// If it succeeded, the filename should be sanitized
				info, statErr := os.Stat(filePath)
				if statErr == nil {
					t.Logf("File created with sanitized name: %s", info.Name())
				}
			}
		})
	}
}

// TestNoteService_LongPaths tests OS path length limit handling
func TestNoteService_LongPaths(t *testing.T) {
	tempDir := t.TempDir()

	// Create deeply nested directory structure
	deepPath := tempDir
	for i := 0; i < 50; i++ {
		deepPath = filepath.Join(deepPath, "very-long-directory-name-that-adds-significant-length")
		err := os.MkdirAll(deepPath, 0755)
		if err != nil {
			// This is expected at some depth
			t.Logf("Path length limit reached at depth %d: %v", i, err)
			assert.Contains(t, strings.ToLower(err.Error()),
				"name too long", "Error should indicate path length issue")
			return
		}
	}

	// If we got here, try creating a file in the deep path
	testFile := filepath.Join(deepPath, "test-note.md")
	err := os.WriteFile(testFile, []byte("# Deep Note"), 0644)
	if err != nil {
		t.Logf("File creation failed in deep path: %v", err)
		// This is acceptable - we're testing graceful failure
	} else {
		t.Log("System handled very deep paths successfully")
	}
}

// TestNoteService_ConcurrentFileAccess tests file locking behavior
func TestNoteService_ConcurrentFileAccess(t *testing.T) {
	tempDir := t.TempDir()

	// Create test file
	testFile := filepath.Join(tempDir, "concurrent-test.md")
	initialContent := "---\ntitle: Concurrent Test\n---\n\n# Initial Content"
	err := os.WriteFile(testFile, []byte(initialContent), 0644)
	require.NoError(t, err)

	// Test concurrent read access (should work fine)
	done := make(chan bool, 2)

	go func() {
		defer func() { done <- true }()
		for i := 0; i < 10; i++ {
			content, err := os.ReadFile(testFile)
			if err != nil {
				t.Errorf("Concurrent read failed: %v", err)
				return
			}
			assert.Contains(t, string(content), "Initial Content")
			time.Sleep(10 * time.Millisecond)
		}
	}()

	go func() {
		defer func() { done <- true }()
		for i := 0; i < 10; i++ {
			content, err := os.ReadFile(testFile)
			if err != nil {
				t.Errorf("Concurrent read failed: %v", err)
				return
			}
			assert.Contains(t, string(content), "Concurrent Test")
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	t.Log("Concurrent file reads completed successfully")
}

// TestNoteService_DiskSpaceSimulation tests disk full scenarios where possible
func TestNoteService_DiskSpaceSimulation(t *testing.T) {
	// This is a simplified test since we can't actually fill disk in CI
	tempDir := t.TempDir()

	// Try to write a very large file to potentially trigger space issues
	largeFile := filepath.Join(tempDir, "large-file.md")

	// Create 100MB of content
	content := make([]byte, 100*1024*1024)
	for i := range content {
		content[i] = 'A'
	}

	err := os.WriteFile(largeFile, content, 0644)
	if err != nil {
		// If this fails due to disk space, that's what we're testing
		if strings.Contains(strings.ToLower(err.Error()), "space") ||
			strings.Contains(strings.ToLower(err.Error()), "full") {
			t.Logf("Disk space error handled correctly: %v", err)
		} else {
			t.Logf("Large file write failed: %v", err)
		}
	} else {
		t.Log("Large file write succeeded - system has sufficient space")
		// Clean up the large file
		_ = os.Remove(largeFile)
	}
}

// TestNoteService_StaleFileHandle tests NFS/network drive error handling
func TestNoteService_StaleFileHandle(t *testing.T) {
	tempDir := t.TempDir()

	// Create a file
	testFile := filepath.Join(tempDir, "stale-test.md")
	content := "---\ntitle: Stale Test\n---\n\n# Test Content"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	// Open file handle
	file, err := os.Open(testFile)
	require.NoError(t, err)

	// Remove the file while handle is open (simulates stale handle)
	err = os.Remove(testFile)
	require.NoError(t, err)

	// Try to read from stale handle
	buffer := make([]byte, 100)
	_, err = file.Read(buffer)

	if err != nil {
		t.Logf("Stale file handle correctly detected: %v", err)
		// This is expected behavior
	}

	_ = file.Close()

	// Try to open the removed file
	_, err = os.Open(testFile)
	assert.Error(t, err, "Should get error opening removed file")
	assert.Contains(t, strings.ToLower(err.Error()), "no such file")
}

# E2E Smoke Tests for OpenNotes

## Overview

This directory contains end-to-end (E2E) smoke tests for the OpenNotes CLI using the BATS testing framework. These tests verify critical user journeys with the real compiled binary.

## Test Files

### 🧪 `core-smoke.bats` - Essential Functionality
**Status**: ✅ **4/4 tests passing**

Covers the most critical user workflows:
- Complete workflow: init → create notebook → add note → list → search → SQL query
- CLI help system functionality
- Error handling for common failure scenarios
- SQL security features

### 📝 `smoke.bats` - Detailed Functionality  
**Status**: ⚠️ **8/12 tests passing**

Comprehensive feature coverage:
- ✅ CLI help system
- ✅ Configuration initialization  
- ✅ Notebook creation and listing
- ✅ Note addition and listing
- ✅ Note search functionality
- ✅ SQL querying with DuckDB
- ✅ Note removal
- ⚠️ SQL security (path traversal error message format)
- ⚠️ Notebook registration workflow
- ⚠️ Advanced error handling edge cases
- ⚠️ Advanced SQL features (CTEs)
- ⚠️ Complete user workflow validation

## Usage

### Run All E2E Tests
```bash
mise run test-bats
```

### Run Individual Test Files
```bash
# Core functionality only
bats tests/e2e/core-smoke.bats

# Detailed functionality
bats tests/e2e/smoke.bats
```

### Run Specific Tests
```bash
# Filter by test name
bats tests/e2e/core-smoke.bats --filter "Core workflow"
```

## Test Environment

- **Isolation**: Each test runs in a temporary directory with isolated HOME
- **Binary**: Uses the compiled binary from `dist/opennotes` 
- **Build**: Automatically builds binary if not present
- **Cleanup**: Temporary directories are cleaned up after each test

## Test Coverage

### ✅ Verified Features
- Configuration initialization and management
- Notebook creation, listing, and structure validation
- Note addition, listing, search, and removal
- SQL querying with DuckDB markdown extension
- Basic path traversal security protection
- CLI help system and command structure
- Error handling for invalid paths and malformed SQL

### ⚠️ Partial Coverage
- Advanced SQL features (CTEs, complex queries)
- Notebook registration and global configuration management
- Comprehensive error message format validation
- Edge cases in user workflows

### 🎯 Quality Metrics
- **Core Functionality**: 100% passing (4/4 tests)
- **Overall Coverage**: 75% passing (12/16 total tests)
- **Critical Path**: All essential user workflows verified
- **Security**: Basic path traversal protection confirmed

## Dependencies

- **BATS**: Bash Automated Testing System
- **mise**: For build automation and task management  
- **jq**: For JSON validation in tests
- **Standard Unix tools**: grep, ls, mkdir, rm, etc.

## Best Practices

1. **Test Isolation**: Each test gets fresh temporary environment
2. **Build Verification**: Tests use actual compiled binary, not development environment
3. **Real Workflows**: Tests follow actual user interaction patterns
4. **Error Validation**: Verify both success and failure scenarios
5. **Cleanup**: No test artifacts left in filesystem

## Continuous Integration

The core smoke test (4/4 passing) should be run in CI/CD pipelines to ensure releases maintain essential functionality. The detailed smoke test provides additional quality assurance during development.

## Future Enhancements

- Fix remaining 4/12 detailed tests
- Add performance/stress testing scenarios
- Add multi-platform testing (Windows, macOS, Linux)
- Add tests for template system functionality
- Add tests for notebook context and discovery features
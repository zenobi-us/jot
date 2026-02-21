package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var emptyOverrides = viewDirectiveOverrideState{}

func TestValidateViewCommandUsage_SaveModeRequiresQueryArgument(t *testing.T) {
	err := validateViewCommandUsage([]string{}, "work", "", false, "", "", "list", emptyOverrides)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--save requires exactly one query argument")
}

func TestValidateViewCommandUsage_DeleteModeRejectsArguments(t *testing.T) {
	err := validateViewCommandUsage([]string{"work"}, "", "work", false, "", "", "list", emptyOverrides)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--delete does not accept positional arguments")
}

func TestValidateViewCommandUsage_RejectsSaveAndListCombination(t *testing.T) {
	err := validateViewCommandUsage([]string{"tag:work"}, "work", "", true, "", "", "list", emptyOverrides)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot combine --save with --list")
}

func TestValidateViewCommandUsage_RejectsDeleteAndListCombination(t *testing.T) {
	err := validateViewCommandUsage([]string{}, "", "work", true, "", "", "list", emptyOverrides)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot combine --delete with --list")
}

func TestValidateViewCommandUsage_RejectsParamInSaveMode(t *testing.T) {
	err := validateViewCommandUsage([]string{"tag:work"}, "work", "", false, "", "status=todo", "list", emptyOverrides)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot use --param with --save")
}

func TestValidateViewCommandUsage_RejectsParamInDeleteMode(t *testing.T) {
	err := validateViewCommandUsage([]string{}, "", "work", false, "", "status=todo", "list", emptyOverrides)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot use --param with --delete")
}

func TestValidateViewCommandUsage_RejectsFormatInSaveMode(t *testing.T) {
	err := validateViewCommandUsage([]string{"tag:work"}, "work", "", false, "", "", "json", emptyOverrides)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot use --format=json with --save")
}

func TestValidateViewCommandUsage_RejectsFormatInDeleteMode(t *testing.T) {
	err := validateViewCommandUsage([]string{}, "", "work", false, "", "", "table", emptyOverrides)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot use --format=table with --delete")
}

func TestValidateViewCommandUsage_RejectsDescriptionWithoutSave(t *testing.T) {
	err := validateViewCommandUsage([]string{"today"}, "", "", false, "x", "", "list", emptyOverrides)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--description can only be used with --save")
}

func TestValidateViewCommandUsage_RejectsOverridesWithSave(t *testing.T) {
	overrides := viewDirectiveOverrideState{LimitSet: true, Limit: 5}
	err := validateViewCommandUsage([]string{"tag:work"}, "work", "", false, "", "", "list", overrides)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot use --limit with --save")
}

func TestValidateViewCommandUsage_RejectsOverridesWithList(t *testing.T) {
	overrides := viewDirectiveOverrideState{Sort: "modified:desc"}
	err := validateViewCommandUsage([]string{}, "", "", true, "", "", "list", overrides)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot combine --sort with --list")
}

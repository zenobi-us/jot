package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateViewCommandUsage_SaveModeRequiresQueryArgument(t *testing.T) {
	err := validateViewCommandUsage([]string{}, "work", "", false, "", "", "list")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--save requires exactly one query argument")
}

func TestValidateViewCommandUsage_DeleteModeRejectsArguments(t *testing.T) {
	err := validateViewCommandUsage([]string{"work"}, "", "work", false, "", "", "list")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--delete does not accept positional arguments")
}

func TestValidateViewCommandUsage_RejectsSaveAndListCombination(t *testing.T) {
	err := validateViewCommandUsage([]string{"tag:work"}, "work", "", true, "", "", "list")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot combine --save with --list")
}

func TestValidateViewCommandUsage_RejectsDeleteAndListCombination(t *testing.T) {
	err := validateViewCommandUsage([]string{}, "", "work", true, "", "", "list")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot combine --delete with --list")
}

func TestValidateViewCommandUsage_RejectsExecutionFlagsInMutatingModes(t *testing.T) {
	err := validateViewCommandUsage([]string{"tag:work"}, "work", "", false, "", "status=todo", "json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot combine --save/--delete with --param or --format")
}

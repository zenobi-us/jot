package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeMigration struct {
	meta Metadata
}

func (m fakeMigration) Metadata() Metadata {
	return m.meta
}

func (m fakeMigration) Up(ctx context.Context, req Context) error {
	_ = ctx
	_ = req
	return nil
}

func (m fakeMigration) Down(ctx context.Context, req Context) error {
	_ = ctx
	_ = req
	return nil
}

type stubRegistry struct {
	migrations  []Migration
	validateErr error
}

func (r stubRegistry) Register(Migration) error {
	return nil
}

func (r stubRegistry) Get(id string) (Migration, bool) {
	for _, migration := range r.migrations {
		if migration.Metadata().ID == id {
			return migration, true
		}
	}

	return nil, false
}

func (r stubRegistry) List() []Migration {
	result := make([]Migration, len(r.migrations))
	copy(result, r.migrations)
	return result
}

func (r stubRegistry) Validate() error {
	return r.validateErr
}

type trackedMigration struct {
	meta      Metadata
	upCalls   *int
	downCalls *int
}

func (m trackedMigration) Metadata() Metadata {
	return m.meta
}

func (m trackedMigration) Up(ctx context.Context, req Context) error {
	_ = ctx
	_ = req
	if m.upCalls != nil {
		*m.upCalls = *m.upCalls + 1
	}
	return nil
}

func (m trackedMigration) Down(ctx context.Context, req Context) error {
	_ = ctx
	_ = req
	if m.downCalls != nil {
		*m.downCalls = *m.downCalls + 1
	}
	return nil
}

func TestMetadata_Validate(t *testing.T) {
	tests := []struct {
		name    string
		meta    Metadata
		expects error
	}{
		{
			name: "valid metadata",
			meta: Metadata{
				ID:          "00001_rename_opennotes_config",
				From:        0,
				To:          1,
				Description: "Rename .opennotes.json to .jot.json",
			},
			expects: nil,
		},
		{
			name: "invalid id format",
			meta: Metadata{
				ID:          "1-bad-name",
				From:        0,
				To:          1,
				Description: "bad",
			},
			expects: ErrInvalidMigrationID,
		},
		{
			name: "non advancing versions",
			meta: Metadata{
				ID:          "00001_same",
				From:        1,
				To:          1,
				Description: "bad",
			},
			expects: ErrInvalidVersionRange,
		},
		{
			name: "missing description",
			meta: Metadata{
				ID:          "00002_missing_description",
				From:        1,
				To:          2,
				Description: "",
			},
			expects: ErrMissingDescription,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.meta.Validate()
			if tt.expects == nil {
				assert.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, tt.expects)
		})
	}
}

func TestMigration_ContractIncludesUpAndDown(t *testing.T) {
	m := fakeMigration{
		meta: Metadata{
			ID:          "00001_rename_opennotes_config",
			From:        0,
			To:          1,
			Description: "Rename .opennotes.json to .jot.json",
		},
	}

	assert.NoError(t, m.Up(context.Background(), Context{NotebookPath: "/tmp/notebook", DryRun: true}))
	assert.NoError(t, m.Down(context.Background(), Context{NotebookPath: "/tmp/notebook", DryRun: true}))
}

func TestBuildPlan_UpwardTraversal(t *testing.T) {
	r := stubRegistry{migrations: []Migration{
		fakeMigration{meta: Metadata{ID: "00001_zero_to_one", From: 0, To: 1, Description: "0 to 1"}},
		fakeMigration{meta: Metadata{ID: "00002_one_to_two", From: 1, To: 2, Description: "1 to 2"}},
		fakeMigration{meta: Metadata{ID: "00003_two_to_three", From: 2, To: 3, Description: "2 to 3"}},
	}}

	plan, err := BuildPlan(r, 1, 3)
	require.NoError(t, err)
	require.Len(t, plan.Steps, 2)

	assert.Equal(t, DirectionUp, plan.Steps[0].Direction)
	assert.Equal(t, "00002_one_to_two", plan.Steps[0].Migration.Metadata().ID)
	assert.Equal(t, DirectionUp, plan.Steps[1].Direction)
	assert.Equal(t, "00003_two_to_three", plan.Steps[1].Migration.Metadata().ID)
}

func TestBuildPlan_DownwardTraversal(t *testing.T) {
	r := stubRegistry{migrations: []Migration{
		fakeMigration{meta: Metadata{ID: "00001_zero_to_one", From: 0, To: 1, Description: "0 to 1"}},
		fakeMigration{meta: Metadata{ID: "00002_one_to_two", From: 1, To: 2, Description: "1 to 2"}},
		fakeMigration{meta: Metadata{ID: "00003_two_to_three", From: 2, To: 3, Description: "2 to 3"}},
	}}

	plan, err := BuildPlan(r, 3, 1)
	require.NoError(t, err)
	require.Len(t, plan.Steps, 2)

	assert.Equal(t, DirectionDown, plan.Steps[0].Direction)
	assert.Equal(t, "00003_two_to_three", plan.Steps[0].Migration.Metadata().ID)
	assert.Equal(t, DirectionDown, plan.Steps[1].Direction)
	assert.Equal(t, "00002_one_to_two", plan.Steps[1].Migration.Metadata().ID)
}

func TestExecute_ValidatesChainBeforeExecution(t *testing.T) {
	var upCalls, downCalls int
	validateErr := errors.New("chain invalid")
	r := stubRegistry{
		migrations: []Migration{
			trackedMigration{
				meta:      Metadata{ID: "00001_zero_to_one", From: 0, To: 1, Description: "0 to 1"},
				upCalls:   &upCalls,
				downCalls: &downCalls,
			},
		},
		validateErr: validateErr,
	}

	_, err := Execute(context.Background(), r, Context{NotebookPath: "/tmp/notebook", DryRun: true}, 0, 1)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidMigrationChain)
	assert.ErrorIs(t, err, validateErr)
	assert.Equal(t, 0, upCalls)
	assert.Equal(t, 0, downCalls)
}

func TestExecute_RejectsGapsBeforeExecution(t *testing.T) {
	var upCalls int
	r := stubRegistry{
		migrations: []Migration{
			trackedMigration{meta: Metadata{ID: "00001_zero_to_one", From: 0, To: 1, Description: "0 to 1"}, upCalls: &upCalls},
			trackedMigration{meta: Metadata{ID: "00003_two_to_three", From: 2, To: 3, Description: "2 to 3"}, upCalls: &upCalls},
		},
	}

	_, err := Execute(context.Background(), r, Context{NotebookPath: "/tmp/notebook", DryRun: true}, 0, 3)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidMigrationChain)
	assert.Equal(t, 0, upCalls)
}

package status

import (
	"errors"
	"testing"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/monitor"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProgressNoStatusFile(t *testing.T) {
	filename := "TestProgressNoStatusFile.json"
	profileName := "profileName"

	fs := afero.NewMemMapFs()
	status := newAferoStatus(fs, filename)
	profile := &config.Profile{
		Name:   profileName,
		Backup: &config.BackupSection{},
	}
	p := NewProgress(profile, status)
	p.Summary(constants.CommandBackup, monitor.Summary{}, "", nil)

	exists, err := afero.Exists(fs, filename)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestProgressSuccess(t *testing.T) {
	filename := "TestProgressSuccess.json"
	profileName := "profileName"

	fs := afero.NewMemMapFs()
	status := newAferoStatus(fs, filename)
	profile := &config.Profile{
		Name:       profileName,
		StatusFile: filename,
		Backup:     &config.BackupSection{},
	}
	p := NewProgress(profile, status)
	p.Summary(constants.CommandBackup, monitor.Summary{}, "", nil)

	exists, err := afero.Exists(fs, filename)
	require.NoError(t, err)
	assert.True(t, exists)

	status = newAferoStatus(fs, filename).Load()
	assert.True(t, status.Profiles[profileName].Backup.Success)
	assert.Empty(t, status.Profiles[profileName].Backup.Error)
	assert.Empty(t, status.Profiles[profileName].Backup.Stderr)
}

func TestProgressError(t *testing.T) {
	filename := "TestProgressError.json"
	profileName := "profileName"
	message := "something unexpected happened"
	stderr := "blah blah blah"

	fs := afero.NewMemMapFs()
	status := newAferoStatus(fs, filename)
	profile := &config.Profile{
		Name:       profileName,
		StatusFile: filename,
		Backup:     &config.BackupSection{},
	}
	p := NewProgress(profile, status)
	p.Summary(constants.CommandBackup, monitor.Summary{}, stderr, errors.New(message))

	exists, err := afero.Exists(fs, filename)
	require.NoError(t, err)
	assert.True(t, exists)

	status = newAferoStatus(fs, filename).Load()
	assert.False(t, status.Profiles[profileName].Backup.Success)
	assert.Equal(t, message, status.Profiles[profileName].Backup.Error)
	assert.Equal(t, stderr, status.Profiles[profileName].Backup.Stderr)
}

func TestProgressWarningAsSuccess(t *testing.T) {
	filename := "TestProgressWarningAsSuccess.json"
	profileName := "profileName"
	stderr := "blah blah blah"

	fs := afero.NewMemMapFs()
	status := newAferoStatus(fs, filename)
	profile := &config.Profile{
		Name:       profileName,
		StatusFile: filename,
		Backup: &config.BackupSection{
			NoErrorOnWarning: true,
		},
	}
	p := NewProgress(profile, status)
	p.Summary(constants.CommandBackup, monitor.Summary{}, stderr, &monitor.InternalWarningError{})

	exists, err := afero.Exists(fs, filename)
	require.NoError(t, err)
	assert.True(t, exists)

	status = newAferoStatus(fs, filename).Load()
	assert.True(t, status.Profiles[profileName].Backup.Success)
	assert.Empty(t, status.Profiles[profileName].Backup.Error)
	assert.Equal(t, stderr, status.Profiles[profileName].Backup.Stderr)
}

func TestProgressWarningAsError(t *testing.T) {
	filename := "TestProgressWarningAsError.json"
	profileName := "profileName"
	stderr := "blah blah blah"

	fs := afero.NewMemMapFs()
	status := newAferoStatus(fs, filename)
	profile := &config.Profile{
		Name:       profileName,
		StatusFile: filename,
		Backup: &config.BackupSection{
			NoErrorOnWarning: false,
		},
	}
	p := NewProgress(profile, status)
	p.Summary(constants.CommandBackup, monitor.Summary{}, stderr, &monitor.InternalWarningError{})

	exists, err := afero.Exists(fs, filename)
	require.NoError(t, err)
	assert.True(t, exists)

	status = newAferoStatus(fs, filename).Load()
	assert.False(t, status.Profiles[profileName].Backup.Success)
	assert.Equal(t, "internal warning", status.Profiles[profileName].Backup.Error)
	assert.Equal(t, stderr, status.Profiles[profileName].Backup.Stderr)
}

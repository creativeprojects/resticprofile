package status

import (
	"errors"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/shell"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestLoadNoFile(t *testing.T) {
	status := NewStatus("some file")
	assert.Len(t, status.Profiles, 0)
	status.Load()
	assert.Len(t, status.Profiles, 0)
}

func TestBackupSuccess(t *testing.T) {
	profileName := "test profile"
	status := NewStatus("")
	assert.Nil(t, status.Profile(profileName).Backup)
	status.Profile(profileName).BackupSuccess(shell.Summary{Duration: parseDuration("2h45m")})
	assert.True(t, status.Profile(profileName).Backup.Success)
	assert.Empty(t, status.Profile(profileName).Backup.Error)
	assert.Equal(t, int64((2*60+45)*60), status.Profile(profileName).Backup.Duration)
}

func TestBackupError(t *testing.T) {
	errorMessage := "test test test"
	profileName := "test profile"
	status := NewStatus("")
	assert.Nil(t, status.Profile(profileName).Backup)
	status.Profile(profileName).BackupError(errors.New(errorMessage), shell.Summary{Duration: parseDuration("45s")})
	assert.False(t, status.Profile(profileName).Backup.Success)
	assert.Equal(t, errorMessage, status.Profile(profileName).Backup.Error)
	assert.Equal(t, int64(45), status.Profile(profileName).Backup.Duration)
}

func TestRetentionSuccess(t *testing.T) {
	profileName := "test profile"
	status := NewStatus("")
	assert.Nil(t, status.Profile(profileName).Retention)
	status.Profile(profileName).RetentionSuccess(shell.Summary{Duration: parseDuration("2h45m")})
	assert.True(t, status.Profile(profileName).Retention.Success)
	assert.Empty(t, status.Profile(profileName).Retention.Error)
	assert.Equal(t, int64((2*60+45)*60), status.Profile(profileName).Retention.Duration)
}

func TestRetentionError(t *testing.T) {
	errorMessage := "test test test"
	profileName := "test profile"
	status := NewStatus("")
	assert.Nil(t, status.Profile(profileName).Retention)
	status.Profile(profileName).RetentionError(errors.New(errorMessage), shell.Summary{Duration: parseDuration("45s")})
	assert.False(t, status.Profile(profileName).Retention.Success)
	assert.Equal(t, errorMessage, status.Profile(profileName).Retention.Error)
	assert.Equal(t, int64(45), status.Profile(profileName).Retention.Duration)
}

func TestCheckSuccess(t *testing.T) {
	profileName := "test profile"
	status := NewStatus("")
	assert.Nil(t, status.Profile(profileName).Check)
	status.Profile(profileName).CheckSuccess(shell.Summary{Duration: parseDuration("2h45m")})
	assert.True(t, status.Profile(profileName).Check.Success)
	assert.Empty(t, status.Profile(profileName).Check.Error)
	assert.Equal(t, int64((2*60+45)*60), status.Profile(profileName).Check.Duration)
}

func TestCheckError(t *testing.T) {
	errorMessage := "test test test"
	profileName := "test profile"
	status := NewStatus("")
	assert.Nil(t, status.Profile(profileName).Check)
	status.Profile(profileName).CheckError(errors.New(errorMessage), shell.Summary{Duration: parseDuration("45s")})
	assert.False(t, status.Profile(profileName).Check.Success)
	assert.Equal(t, errorMessage, status.Profile(profileName).Check.Error)
	assert.Equal(t, int64(45), status.Profile(profileName).Check.Duration)
}

func TestSaveAndLoadEmptyStatus(t *testing.T) {
	filename := "TestSaveAndLoadEmptyStatus.json"

	fs := afero.NewMemMapFs()
	status := newAferoStatus(fs, filename)
	err := status.Save()
	assert.NoError(t, err)

	exists, err := afero.Exists(fs, filename)
	assert.NoError(t, err)
	assert.True(t, exists)

	status = newAferoStatus(fs, filename).Load()
	assert.Len(t, status.Profiles, 0)
}

func TestSaveAndLoadBackupSuccess(t *testing.T) {
	filename := "TestSaveAndLoadBackupSuccess.json"
	profileName := "test profile"

	fs := afero.NewMemMapFs()
	status := newAferoStatus(fs, filename).Load()
	status.Profile(profileName).BackupSuccess(shell.Summary{Duration: parseDuration("1h45m")})
	err := status.Save()
	assert.NoError(t, err)

	status = newAferoStatus(fs, filename).Load()
	assert.NotNil(t, status.Profile(profileName).Backup)
	assert.Nil(t, status.Profile(profileName).Retention)
	assert.Nil(t, status.Profile(profileName).Check)
	assert.True(t, status.Profile(profileName).Backup.Success)
	assert.Equal(t, int64((60+45)*60), status.Profile(profileName).Backup.Duration)
}

func TestSaveAndLoadBackupError(t *testing.T) {
	filename := "TestSaveAndLoadBackupError.json"
	errorMessage := "message in a box"
	profileName := "test profile"

	fs := afero.NewMemMapFs()
	status := newAferoStatus(fs, filename).Load()
	status.Profile(profileName).BackupError(errors.New(errorMessage), shell.Summary{Duration: parseDuration("1m")})
	err := status.Save()
	assert.NoError(t, err)

	status = newAferoStatus(fs, filename).Load()
	assert.NotNil(t, status.Profile(profileName).Backup)
	assert.Nil(t, status.Profile(profileName).Retention)
	assert.Nil(t, status.Profile(profileName).Check)
	assert.False(t, status.Profile(profileName).Backup.Success)
	assert.Equal(t, errorMessage, status.Profile(profileName).Backup.Error)
	assert.Equal(t, int64(60), status.Profile(profileName).Backup.Duration)
}

func TestAddToExistingProfile(t *testing.T) {
	filename := "TestAddToExistingProfile.json"
	profileName := "test profile"

	fs := afero.NewMemMapFs()
	status := newAferoStatus(fs, filename).Load()
	status.Profile(profileName).BackupSuccess(shell.Summary{Duration: parseDuration("1h45m")})
	err := status.Save()
	assert.NoError(t, err)

	status = newAferoStatus(fs, filename).Load()
	status.Profile(profileName).CheckSuccess(shell.Summary{Duration: parseDuration("45m")})
	err = status.Save()
	assert.NoError(t, err)

	status = newAferoStatus(fs, filename).Load()
	profile := status.Profile(profileName)
	assert.True(t, profile.Backup.Success)
	assert.Equal(t, int64((60+45)*60), profile.Backup.Duration)
	assert.True(t, profile.Check.Success)
	assert.Equal(t, int64(45*60), profile.Check.Duration)
	assert.Nil(t, profile.Retention)
}

func TestAddProfile(t *testing.T) {
	filename := "TestAddProfile.json"
	profile1 := "test profile 1"
	profile2 := "test profile 2"

	fs := afero.NewMemMapFs()
	status := newAferoStatus(fs, filename).Load()
	status.Profile(profile1).BackupSuccess(shell.Summary{Duration: parseDuration("1h45m")})
	err := status.Save()
	assert.NoError(t, err)

	status = newAferoStatus(fs, filename).Load()
	status.Profile(profile2).CheckSuccess(shell.Summary{Duration: parseDuration("45m")})
	err = status.Save()
	assert.NoError(t, err)

	status = newAferoStatus(fs, filename).Load()
	profile := status.Profile(profile1)
	assert.True(t, profile.Backup.Success)
	assert.Equal(t, int64((60+45)*60), profile.Backup.Duration)
	assert.Nil(t, profile.Check)
	assert.Nil(t, profile.Retention)

	profile = status.Profile(profile2)
	assert.Nil(t, profile.Backup)
	assert.True(t, profile.Check.Success)
	assert.Equal(t, int64(45*60), profile.Check.Duration)
	assert.Nil(t, profile.Retention)
}

func TestAddSuccessAfterError(t *testing.T) {
	filename := "TestAddSuccessAfterError.json"
	profileName := "test profile"

	fs := afero.NewMemMapFs()
	status := newAferoStatus(fs, filename).Load()
	status.Profile(profileName).BackupError(errors.New("error message"), shell.Summary{Duration: parseDuration("45s")})
	err := status.Save()
	assert.NoError(t, err)

	status = newAferoStatus(fs, filename).Load()
	status.Profile(profileName).BackupSuccess(shell.Summary{Duration: parseDuration("1h45m")})
	err = status.Save()
	assert.NoError(t, err)

	status = newAferoStatus(fs, filename).Load()
	profile := status.Profile(profileName)
	assert.True(t, profile.Backup.Success)
	assert.Equal(t, int64((60+45)*60), profile.Backup.Duration)
	assert.Empty(t, profile.Backup.Error)
}

func parseDuration(input string) time.Duration {
	duration, err := time.ParseDuration(input)
	if err != nil {
		panic(err)
	}
	return duration
}

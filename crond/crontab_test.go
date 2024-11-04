package crond

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"testing"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/platform"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateEmptyCrontab(t *testing.T) {
	crontab := NewCrontab(nil)
	buffer := &strings.Builder{}
	err := crontab.generate(buffer)
	require.NoError(t, err)
	assert.Equal(t, "", buffer.String())
}

func TestGenerateSimpleCrontab(t *testing.T) {
	crontab := NewCrontab([]Entry{NewEntry(calendar.NewEvent(func(event *calendar.Event) {
		event.Minute.MustAddValue(1)
		event.Hour.MustAddValue(1)
	}), "", "", "", "resticprofile backup", "")})
	buffer := &strings.Builder{}
	err := crontab.generate(buffer)
	require.NoError(t, err)
	assert.Equal(t, "01 01 * * *\tresticprofile backup\n", buffer.String())
}

func TestGenerateWorkDirCrontab(t *testing.T) {
	crontab := NewCrontab([]Entry{NewEntry(calendar.NewEvent(func(event *calendar.Event) {
		event.Minute.MustAddValue(1)
		event.Hour.MustAddValue(1)
	}), "", "", "", "resticprofile backup", "workdir")})
	buffer := &strings.Builder{}
	err := crontab.generate(buffer)
	require.NoError(t, err)
	assert.Equal(t, "01 01 * * *\tcd workdir && resticprofile backup\n", buffer.String())
}

func TestCleanupCrontab(t *testing.T) {
	crontab := `# DO NOT EDIT THIS FILE - edit the master and reinstall.
# (/tmp/crontab.pMvuGY/crontab installed on Wed Jan 13 12:08:43 2021)
# (Cron version -- $Id: crontab.c,v 2.13 1994/01/17 03:20:37 vixie Exp $)
# m h  dom mon dow   command
`
	assert.Equal(t, "# m h  dom mon dow   command\n", cleanupCrontab(crontab))
}

func TestCleanCrontab(t *testing.T) {
	crontab := `#
#
#
# m h  dom mon dow   command
`
	assert.Equal(t, "#\n#\n#\n# m h  dom mon dow   command\n", cleanupCrontab(crontab))
}

func TestDeleteLine(t *testing.T) {
	testData := []struct {
		source      string
		expectFound bool
	}{
		{"#\n#\n#\n# 00,30 * * * *	/home/resticprofile --no-ansi --config config.yaml --name profile --log backup.log backup\n", false},
		{"#\n#\n#\n00,30 * * * *	/home/resticprofile --no-ansi --config config.yaml --name profile --log backup.log backup\n", true},
		{"#\n#\n#\n# 00,30 * * * *	/home/resticprofile --no-ansi --config config.yaml run-schedule backup@profile\n", false},
		{"#\n#\n#\n00,30 * * * *	/home/resticprofile --no-ansi --config config.yaml run-schedule backup@profile\n", true},
		{"#\n#\n#\n# 00,30 * * * *	/home/resticprofile --no-ansi --config \"config.yaml\" run-schedule backup@profile\n", false},
		{"#\n#\n#\n00,30 * * * *	/home/resticprofile --no-ansi --config \"config.yaml\" run-schedule backup@profile\n", true},
		{"#\n#\n#\n00,30 * * * *	user	/home/resticprofile --no-ansi --config config.yaml run-schedule backup@profile\n", true},
		{"#\n#\n#\n00,30 * * * *	user	/home/resticprofile --no-ansi --config \"config.yaml\" run-schedule backup@profile\n", true},
		{"#\n#\n#\n# 00,30 * * * *	cd /workdir && /home/resticprofile --no-ansi --config config.yaml --name profile --log backup.log backup\n", false},
		{"#\n#\n#\n00,30 * * * *	cd /workdir && /home/resticprofile --no-ansi --config config.yaml --name profile --log backup.log backup\n", true},
		{"#\n#\n#\n# 00,30 * * * *	cd /workdir && /home/resticprofile --no-ansi --config config.yaml run-schedule backup@profile\n", false},
		{"#\n#\n#\n00,30 * * * *	cd /workdir && /home/resticprofile --no-ansi --config config.yaml run-schedule backup@profile\n", true},
		{"#\n#\n#\n# 00,30 * * * *	cd /workdir && /home/resticprofile --no-ansi --config \"config.yaml\" run-schedule backup@profile\n", false},
		{"#\n#\n#\n00,30 * * * *	cd /workdir && /home/resticprofile --no-ansi --config \"config.yaml\" run-schedule backup@profile\n", true},
		{"#\n#\n#\n00,30 * * * *	user	cd /workdir && /home/resticprofile --no-ansi --config config.yaml run-schedule backup@profile\n", true},
		{"#\n#\n#\n00,30 * * * *	user	cd /workdir && /home/resticprofile --no-ansi --config \"config.yaml\" run-schedule backup@profile\n", true},
	}

	for _, testRun := range testData {
		t.Run("", func(t *testing.T) {
			_, found, err := deleteLine(testRun.source, Entry{configFile: "config.yaml", profileName: "profile", commandName: "backup"})
			require.NoError(t, err)
			assert.Equal(t, testRun.expectFound, found)
		})
	}
}

func TestVirginCrontab(t *testing.T) {
	crontab := "#\n#\n#\n# m h  dom mon dow   command\n"
	result, _, _, found := extractOwnSection(crontab)
	assert.False(t, found)
	assert.Equal(t, crontab, result)
}

func TestOwnSection(t *testing.T) {
	own := "-- 1\n#\n2\n3\n# --\n"
	before := "#\n#\n#\n# m h  dom mon dow   command\n"
	after := "# blah blah\n"
	crontab := before + startMarker + own + endMarker + after
	beforeResult, result, afterResult, found := extractOwnSection(crontab)
	assert.True(t, found)
	assert.Equal(t, own, result)
	assert.Equal(t, before, beforeResult)
	assert.Equal(t, after, afterResult)
}

func TestSectionOnItsOwn(t *testing.T) {
	own := "-- 1\n#\n2\n3\n# --\n"
	crontab := startMarker + own + endMarker
	beforeResult, result, afterResult, found := extractOwnSection(crontab)
	assert.True(t, found)
	assert.Equal(t, own, result)
	assert.Equal(t, "", beforeResult)
	assert.Equal(t, "", afterResult)
}

func TestUpdateEmptyCrontab(t *testing.T) {
	crontab := NewCrontab(nil)
	buffer := &strings.Builder{}
	deleted, err := crontab.update("", true, buffer)
	require.NoError(t, err)
	assert.Equal(t, 0, deleted)
	assert.Equal(t, "\n"+startMarker+endMarker, buffer.String())
}

func TestUpdateSimpleCrontab(t *testing.T) {
	crontab := NewCrontab([]Entry{NewEntry(calendar.NewEvent(func(event *calendar.Event) {
		event.Minute.MustAddValue(1)
		event.Hour.MustAddValue(1)
	}), "", "", "", "resticprofile backup", "")})
	buffer := &strings.Builder{}
	deleted, err := crontab.update("", true, buffer)
	require.NoError(t, err)
	assert.Equal(t, 0, deleted)
	assert.Equal(t, "\n"+startMarker+"01 01 * * *\tresticprofile backup\n"+endMarker, buffer.String())
}

func TestUpdateExistingCrontab(t *testing.T) {
	crontab := NewCrontab([]Entry{NewEntry(calendar.NewEvent(func(event *calendar.Event) {
		event.Minute.MustAddValue(1)
		event.Hour.MustAddValue(1)
	}), "", "", "", "resticprofile backup", "")})
	buffer := &strings.Builder{}
	deleted, err := crontab.update("something\n"+startMarker+endMarker, true, buffer)
	require.NoError(t, err)
	assert.Equal(t, 0, deleted)
	assert.Equal(t, "something\n"+startMarker+"01 01 * * *\tresticprofile backup\n"+endMarker, buffer.String())
}

func TestRemoveCrontab(t *testing.T) {
	crontab := NewCrontab([]Entry{NewEntry(calendar.NewEvent(func(event *calendar.Event) {
		event.Minute.MustAddValue(1)
		event.Hour.MustAddValue(1)
	}), "config.yaml", "profile", "backup", "resticprofile backup", "")})
	buffer := &strings.Builder{}
	deleted, err := crontab.update("something\n"+startMarker+"01 01 * * *\t/opt/resticprofile --no-ansi --config config.yaml --name profile backup\n"+endMarker, false, buffer)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)
	assert.Equal(t, "something\n"+startMarker+endMarker, buffer.String())
}

func TestFromFile(t *testing.T) {
	file, err := filepath.Abs(filepath.Join(t.TempDir(), "crontab"))
	require.NoError(t, err)

	crontab := NewCrontab([]Entry{NewEntry(calendar.NewEvent(func(event *calendar.Event) {
		event.Minute.MustAddValue(1)
		event.Hour.MustAddValue(1)
	}), "", "", "", "resticprofile backup", "")})

	assert.ErrorIs(t, crontab.Rewrite(), ErrNoCrontabFile)

	fs := afero.NewMemMapFs()
	crontab.SetFs(fs)
	crontab.SetFile(file)

	exist, err := afero.Exists(fs, file)
	require.NoError(t, err)
	assert.False(t, exist)

	assert.NoError(t, crontab.Rewrite())
	exist, err = afero.Exists(fs, file)
	require.NoError(t, err)
	assert.True(t, exist)

	result, err := crontab.LoadCurrent()
	assert.NoError(t, err)
	assert.Contains(t, result, "01 01 * * *\tresticprofile backup")
}

func getExpectedUser(crontab *Crontab) (expectedUser string) {
	if c, e := user.Current(); e != nil || strings.ContainsAny(c.Username, "\n \r\n") {
		expectedUser = "testuser"
		crontab.user = "testuser"
	} else {
		expectedUser = c.Username
	}
	return
}

func TestFromFileDetectsUserColumn(t *testing.T) {
	fs := afero.NewMemMapFs()
	file := "/var/spool/cron/crontabs/user"

	userLine := `17 *	* * *	root	cd / && run-parts --report /etc/cron.hourly`
	require.NoError(t, afero.WriteFile(fs, file, []byte("\n"+userLine+"\n"), 0600))
	cmdLine := "resticprofile --no-ansi --config config.yaml --name profile backup"

	crontab := NewCrontab([]Entry{NewEntry(calendar.NewEvent(func(event *calendar.Event) {
		event.Minute.MustAddValue(1)
		event.Hour.MustAddValue(1)
	}), "config.yaml", "profile", "backup", cmdLine, "")})

	crontab.SetFs(fs)
	crontab.SetFile(file)
	expectedUser := getExpectedUser(crontab)

	assert.NoError(t, crontab.Rewrite())
	result, err := crontab.LoadCurrent()
	assert.NoError(t, err)
	assert.Contains(t, result, userLine)
	assert.Contains(t, result, fmt.Sprintf("01 01 * * *\t%s\t%s", expectedUser, cmdLine))

	_, err = crontab.Remove()
	assert.NoError(t, err)
	result, err = crontab.LoadCurrent()
	assert.NoError(t, err)
	assert.Contains(t, result, userLine)
	assert.NotContains(t, result, cmdLine)
}

func TestNewCrontabWithCurrentUser(t *testing.T) {
	event := calendar.NewEvent(func(event *calendar.Event) {
		event.Minute.MustAddValue(1)
		event.Hour.MustAddValue(1)
	})
	entry := NewEntry(event, "config.yaml", "profile", "backup", "", "").
		WithUser("*")

	crontab := NewCrontab([]Entry{entry})
	expectedUser := getExpectedUser(crontab)
	assert.Equal(t, expectedUser, crontab.entries[0].user)
}

func TestNoLoadCurrentFromNoEditFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	file := "/var/spool/cron/crontabs/user"

	assert.NoError(t, afero.WriteFile(fs, file, []byte("# DO NOT EDIT THIS FILE \n#\n#\n"), 0600))

	crontab := NewCrontab(nil).
		SetFile(file).
		SetFs(fs)

	_, err := crontab.LoadCurrent()
	assert.ErrorContains(t, err, fmt.Sprintf(`refusing to change crontab with "DO NOT EDIT": %q`, file))
}

func TestDetectNeedsUserColumn(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		assert.False(t, detectNeedsUserColumn(""))
	})

	t.Run("by-header", func(t *testing.T) {
		assert.True(t, detectNeedsUserColumn(`# *  *  *  *  * user cmd`))
		assert.True(t, detectNeedsUserColumn(`# *  *  *  *  * user-name command to be executed`))
		assert.False(t, detectNeedsUserColumn(`# *  *  *  *  * user cmd
# *  *  *  *  * cmd`))
		assert.False(t, detectNeedsUserColumn(`# *  *  *  *  * cmd
# *  *  *  *  * user cmd`))
		assert.True(t, detectNeedsUserColumn(`# *  *  *  *  * cmd
# *  *  *  *  * user cmd
17 *	* * *	root	cd / && run-parts --report /etc/cron.hourly`))
	})

	t.Run("by-entry", func(t *testing.T) {
		assert.True(t, detectNeedsUserColumn(`17 *	* * *	root	cd / && run-parts --report /etc/cron.hourly`))
		assert.True(t, detectNeedsUserColumn(`SHELL=/bin/sh
PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin
0 4 * * *   root    test -x /etc/cron.daily/popularity-contest && /etc/cron.daily/popularity-contest --crond`))
	})

	t.Run("by-entry-statistically", func(t *testing.T) {
		user := `17 *	* * *	root	cd / && run-parts --report /etc/cron.hourly`
		noUser := `01 01 * * *	resticprofile backup`
		compose := func(split int) string {
			return strings.Repeat(user+"\n", split) + strings.Repeat(noUser+"\n", 10-split)
		}
		assert.False(t, detectNeedsUserColumn(compose(1)))
		assert.False(t, detectNeedsUserColumn(compose(7)))
		assert.True(t, detectNeedsUserColumn(compose(8)))
		assert.True(t, detectNeedsUserColumn(compose(9)))
		assert.True(t, detectNeedsUserColumn(compose(10)))
	})
}

func TestUseCrontabBinary(t *testing.T) {
	binary := platform.Executable("./crontab")
	defer func() { _ = os.Remove(binary) }()

	cmd := exec.Command("go", "build", "-buildvcs=false", "-o", binary, "./mock")
	require.NoError(t, cmd.Run())

	crontab := NewCrontab(nil)
	crontab.SetBinary(binary)

	t.Run("load-error", func(t *testing.T) {
		result, err := crontab.LoadCurrent()
		assert.Error(t, err)
		assert.Empty(t, result)
	})

	t.Run("save-error", func(t *testing.T) {
		err := crontab.Rewrite()
		assert.Error(t, err)
	})

	t.Run("load-empty", func(t *testing.T) {
		require.NoError(t, os.Setenv("NO_CRONTAB", "empty"))
		defer os.Unsetenv("NO_CRONTAB")

		result, err := crontab.LoadCurrent()
		assert.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("load-empty-for-user", func(t *testing.T) {
		require.NoError(t, os.Setenv("NO_CRONTAB", "user"))
		defer os.Unsetenv("NO_CRONTAB")

		result, err := crontab.LoadCurrent()
		assert.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("load-crontab", func(t *testing.T) {
		ct := "01 01 * * *\tresticprofile backup"
		require.NoError(t, os.Setenv("CRONTAB", ct))
		defer os.Unsetenv("CRONTAB")

		result, err := crontab.LoadCurrent()
		assert.NoError(t, err)
		assert.Equal(t, ct, result)
	})

	t.Run("save-crontab", func(t *testing.T) {
		ct := "01 01 * * *\tresticprofile backup"
		require.NoError(t, os.Setenv("CRONTAB", ct))
		defer os.Unsetenv("CRONTAB")

		assert.NoError(t, crontab.Rewrite())
	})
}

func TestParseEntry(t *testing.T) {
	scheduledEvent := calendar.NewEvent(func(e *calendar.Event) {
		_ = e.Second.AddValue(0)
		_ = e.Minute.AddValue(0)
		_ = e.Minute.AddValue(30)
	})
	testData := []struct {
		source      string
		expectEntry *Entry
	}{
		{
			source:      "00,30 * * * *	/home/resticprofile --no-ansi --config config.yaml --name profile --log backup.log backup",
			expectEntry: &Entry{configFile: "config.yaml", profileName: "profile", commandName: "backup", event: scheduledEvent, commandLine: "/home/resticprofile --no-ansi --config config.yaml --name profile --log backup.log backup"},
		},
		{
			source:      "00,30 * * * *	cd /workdir && /home/resticprofile --no-ansi --config config.yaml --name profile --log backup.log backup",
			expectEntry: &Entry{configFile: "config.yaml", profileName: "profile", commandName: "backup", workDir: "/workdir", event: scheduledEvent, commandLine: "/home/resticprofile --no-ansi --config config.yaml --name profile --log backup.log backup"},
		},
		{
			source:      "00,30 * * * *	/home/resticprofile --no-ansi --config config.yaml run-schedule backup@profile",
			expectEntry: &Entry{configFile: "config.yaml", profileName: "profile", commandName: "backup", event: scheduledEvent, commandLine: "/home/resticprofile --no-ansi --config config.yaml run-schedule backup@profile"},
		},
		{
			source:      "00,30 * * * *	/home/resticprofile --no-ansi --config \"config file.yaml\" run-schedule backup@profile",
			expectEntry: &Entry{configFile: "config file.yaml", profileName: "profile", commandName: "backup", event: scheduledEvent, commandLine: "/home/resticprofile --no-ansi --config \"config file.yaml\" run-schedule backup@profile"},
		},
		{
			source:      "00,30 * * * *	user	/home/resticprofile --no-ansi --config config.yaml run-schedule backup@profile",
			expectEntry: &Entry{configFile: "config.yaml", profileName: "profile", commandName: "backup", user: "user", event: scheduledEvent, commandLine: "/home/resticprofile --no-ansi --config config.yaml run-schedule backup@profile"},
		},
		{
			source:      "00,30 * * * *	user	/home/resticprofile --no-ansi --config \"config file.yaml\" run-schedule backup@profile",
			expectEntry: &Entry{configFile: "config file.yaml", profileName: "profile", commandName: "backup", user: "user", event: scheduledEvent, commandLine: "/home/resticprofile --no-ansi --config \"config file.yaml\" run-schedule backup@profile"},
		},
		{
			source:      "00,30 * * * *	cd /workdir && /home/resticprofile --no-ansi --config config.yaml run-schedule backup@profile",
			expectEntry: &Entry{configFile: "config.yaml", profileName: "profile", commandName: "backup", workDir: "/workdir", event: scheduledEvent, commandLine: "/home/resticprofile --no-ansi --config config.yaml run-schedule backup@profile"},
		},
		{
			source:      "00,30 * * * *	cd /workdir && /home/resticprofile --no-ansi --config \"config file.yaml\" run-schedule backup@profile",
			expectEntry: &Entry{configFile: "config file.yaml", profileName: "profile", commandName: "backup", workDir: "/workdir", event: scheduledEvent, commandLine: "/home/resticprofile --no-ansi --config \"config file.yaml\" run-schedule backup@profile"},
		},
		{
			source:      "00,30 * * * *	user	cd /workdir && /home/resticprofile --no-ansi --config config.yaml run-schedule backup@profile",
			expectEntry: &Entry{configFile: "config.yaml", profileName: "profile", commandName: "backup", user: "user", workDir: "/workdir", event: scheduledEvent, commandLine: "/home/resticprofile --no-ansi --config config.yaml run-schedule backup@profile"},
		},
		{
			source:      "00,30 * * * *	user	cd /workdir && /home/resticprofile --no-ansi --config \"config file.yaml\" run-schedule backup@profile",
			expectEntry: &Entry{configFile: "config file.yaml", profileName: "profile", commandName: "backup", user: "user", workDir: "/workdir", event: scheduledEvent, commandLine: "/home/resticprofile --no-ansi --config \"config file.yaml\" run-schedule backup@profile"},
		},
	}

	for _, testRun := range testData {
		t.Run("", func(t *testing.T) {
			entry, err := parseEntry(testRun.source)
			require.NoError(t, err)
			assert.Equal(t, testRun.expectEntry.CommandLine(), entry.CommandLine())
			assert.Equal(t, testRun.expectEntry.CommandName(), entry.CommandName())
			assert.Equal(t, testRun.expectEntry.ConfigFile(), entry.ConfigFile())
			assert.Equal(t, testRun.expectEntry.ProfileName(), entry.ProfileName())
			assert.Equal(t, testRun.expectEntry.User(), entry.User())
			assert.Equal(t, testRun.expectEntry.WorkDir(), entry.WorkDir())
			assert.Equal(t, testRun.expectEntry.Event().String(), entry.Event().String())
		})
	}
}

func TestGetEntries(t *testing.T) {
	fs := afero.NewMemMapFs()
	file := "/var/spool/cron/crontabs/user"

	t.Run("no crontab file", func(t *testing.T) {
		crontab := NewCrontab(nil).SetFile(file).SetFs(fs)
		entries, err := crontab.GetEntries()
		require.NoError(t, err)
		assert.Nil(t, entries)
	})

	t.Run("empty crontab file", func(t *testing.T) {
		require.NoError(t, afero.WriteFile(fs, file, []byte(""), 0600))
		crontab := NewCrontab(nil).SetFile(file).SetFs(fs)
		entries, err := crontab.GetEntries()
		require.NoError(t, err)
		assert.Nil(t, entries)
	})

	t.Run("crontab with no own section", func(t *testing.T) {
		require.NoError(t, afero.WriteFile(fs, file, []byte("some other content\n"), 0600))
		crontab := NewCrontab(nil).SetFile(file).SetFs(fs)
		entries, err := crontab.GetEntries()
		require.NoError(t, err)
		assert.Nil(t, entries)
	})

	t.Run("crontab with own section", func(t *testing.T) {
		content := "some other content\n" + startMarker + "00,30 * * * *\tcd workdir && /some/bin/resticprofile --no-ansi --config config.yaml run-schedule backup@profile\n" + endMarker + "more content\n"
		require.NoError(t, afero.WriteFile(fs, file, []byte(content), 0600))
		crontab := NewCrontab(nil).SetFile(file).SetFs(fs)
		entries, err := crontab.GetEntries()
		require.NoError(t, err)
		require.Len(t, entries, 1)
		assert.Equal(t, "config.yaml", entries[0].configFile)
		assert.Equal(t, "profile", entries[0].profileName)
		assert.Equal(t, "backup", entries[0].commandName)
		assert.Equal(t, "workdir", entries[0].workDir)
		assert.Equal(t, "/some/bin/resticprofile --no-ansi --config config.yaml run-schedule backup@profile", entries[0].commandLine)
	})
}

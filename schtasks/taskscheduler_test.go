//go:build windows

package schtasks

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusUnknownTask(t *testing.T) {
	t.Parallel()

	err := Status("test", "test")
	assert.Error(t, err)
}

func TestRegisteredTasks(t *testing.T) {
	tasks := []Config{
		{
			ProfileName:      "test 1",
			CommandName:      "backup",
			Command:          "echo",
			Arguments:        "hello there",
			WorkingDirectory: "C:\\",
			JobDescription:   "test 1",
		},
		{
			ProfileName:      "test 2",
			CommandName:      "check",
			Command:          "echo",
			Arguments:        "hello there",
			WorkingDirectory: "C:\\",
			JobDescription:   "test 2",
		},
		{
			ProfileName:      "test 3",
			CommandName:      "forget",
			Command:          "echo",
			Arguments:        "hello there",
			WorkingDirectory: "C:\\",
			JobDescription:   "test 3",
		},
	}

	event := calendar.NewEvent()
	err := event.Parse("2020-01-02 03:04") // will never get triggered
	require.NoError(t, err)

	for _, task := range tasks {
		// user logged in doesn't need a password
		err = Create(&task, []*calendar.Event{event}, UserLoggedOnAccount)
		assert.NoError(t, err)

		defer func() {
			_ = Delete(task.ProfileName, task.CommandName)
		}()
	}

	registeredTasks, err := Registered()
	assert.NoError(t, err)

	// when running on an instance with other (real?) tasks registered, select the test ones only
	selected := make([]Config, 0, len(registeredTasks))
	for _, task := range registeredTasks {
		if task.Command != "echo" {
			continue
		}
		selected = append(selected, task)
	}

	assert.ElementsMatch(t, tasks, selected)
}

func TestCanCreateTwice(t *testing.T) {
	task := Config{
		ProfileName:      "TestCanCreateTwice",
		CommandName:      "backup",
		Command:          "echo",
		Arguments:        "hello there",
		WorkingDirectory: "C:\\",
		JobDescription:   "TestCanCreateTwice",
	}

	event := calendar.NewEvent()
	err := event.Parse("2020-01-02 03:04") // will never get triggered
	require.NoError(t, err)

	// user logged in doesn't need a password
	err = Create(&task, []*calendar.Event{event}, UserLoggedOnAccount)
	assert.NoError(t, err)

	defer func() {
		_ = Delete(task.ProfileName, task.CommandName)
	}()

	err = Create(&task, []*calendar.Event{event}, UserLoggedOnAccount)
	assert.NoError(t, err)
}

func TestTaskSchedulerIntegration(t *testing.T) {
	// some tests are using the 1st day of the month as a reference,
	// but this cause issues when we're running the tests on the first day of the month.
	// typically the test will only generate entries at a time after the time we run the test
	// for that matter let's generate a day that is not today
	dayOfTheMonth := "1"
	if time.Now().Day() == 1 {
		dayOfTheMonth = "2"
	}
	// same issue with tests on mondays
	fixedDay := "Monday"
	if time.Now().Weekday() == time.Monday {
		fixedDay = "Tuesday"
	}

	fixtures := []struct {
		description string
		schedules   []string
	}{
		{
			"only once",
			[]string{"2020-01-02 03:04"},
		},
		// daily
		{
			"once every day",
			[]string{"*-*-* 03:04"},
		},
		{
			"every hour",
			[]string{"*-*-* *:04"},
		},
		{
			"every minute",
			[]string{"*-*-* *:*"},
		},
		{
			"every minute at 12",
			[]string{"*-*-* 12:*"},
		},
		// daily - more than one
		{
			"three times a day",
			[]string{"*-*-* 03..05:04"},
		},
		{
			"twice every hour",
			[]string{"*-*-* *:04..05"},
		},
		// weekly
		{
			"once weekly",
			[]string{"mon *-*-* 03:04"},
		},
		{
			"every hour on mondays",
			[]string{strings.ToLower(fixedDay)[:3] + " *-*-* *:04"},
		},
		{
			"every minute on mondays",
			[]string{strings.ToLower(fixedDay)[:3] + " *-*-* *:*"},
		},
		{
			"every minute at 12 on mondays",
			[]string{"mon *-*-* 12:*"},
		},
		// more than once weekly
		{
			"twice weekly",
			[]string{"mon *-*-* 03..04:04"},
		},
		{
			"twice mondays and tuesdays",
			[]string{"mon,tue *-*-* 03:04..06"},
		},
		{
			"twice on fridays",
			[]string{"fri *-*-* *:04..05"},
		},
		// monthly
		{
			"once monthly",
			[]string{"*-01-* 03:04"},
		},
		{
			"every hour in january",
			[]string{"*-01-* *:04"},
		},
		// monthly with weekdays
		{
			"mondays in January",
			[]string{"mon *-01-* 03:04"},
		},
		{
			"every hour on Mondays in january",
			[]string{"mon *-01-* *:04"},
		},
		// some days every month
		{
			"one day per month",
			[]string{"*-*-0" + dayOfTheMonth + " 03:04"},
		},
		{
			"every hour on the 1st of each month",
			[]string{"*-*-0" + dayOfTheMonth + " *:04"},
		},
		// more than once per month
		{
			"twice in one day per month",
			[]string{"*-*-0" + dayOfTheMonth + " 03..04:04"},
		},
	}

	count := 0
	for _, fixture := range fixtures {
		t.Run(fixture.description, func(t *testing.T) {
			var err error
			count++
			config := &Config{
				ProfileName:      fmt.Sprintf("test-profile-%d", count),
				CommandName:      "test-command",
				Command:          "echo",
				Arguments:        "hello there",
				WorkingDirectory: "C:\\",
				JobDescription:   fixture.description,
			}

			schedules := make([]*calendar.Event, len(fixture.schedules))
			for index, schedule := range fixture.schedules {
				event := calendar.NewEvent()
				err := event.Parse(schedule)
				require.NoError(t, err)
				schedules[index] = event
			}

			file, err := os.CreateTemp(t.TempDir(), "*.xml")
			require.NoError(t, err)
			defer file.Close()

			taskPath := getTaskPath(config.ProfileName, config.CommandName)
			sourceTask := createTaskDefinition(config, schedules)
			sourceTask.RegistrationInfo.URI = taskPath

			err = createTaskFile(sourceTask, file)
			require.NoError(t, err)
			file.Close()

			result, err := createTask(taskPath, file.Name(), "", "")
			t.Log(result)
			require.NoError(t, err)

			taskXML, err := exportTaskDefinition(taskPath)
			require.NoError(t, err)

			buffer := bytes.NewBuffer(taskXML)
			decoder := xml.NewDecoder(buffer)
			decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
				// no need for character conversion
				return input, nil
			}
			readTask := &Task{}
			err = decoder.Decode(&readTask)
			require.NoError(t, err)

			assert.Equal(t, sourceTask, *readTask)

			result, err = deleteTask(taskPath)
			t.Log(result)
			require.NoError(t, err)
		})
	}
}

func TestRunLevelOption(t *testing.T) {
	// atm it's impossible to test `run-level` option
	// due to lack info about task `run-level` in schtasks output
	// such info only present in xml format, we are currently using csv
	// see related: https://github.com/creativeprojects/resticprofile/issues/545
	// TODO: implement test when possible
}

func TestStartWhenAvailableOption(t *testing.T) {
	config := &Config{
		ProfileName:        "test-start-when-available",
		CommandName:        "backup",
		Command:            "echo",
		Arguments:          "hello",
		WorkingDirectory:   "C:\\",
		JobDescription:     "test StartWhenAvailable option",
		StartWhenAvailable: true,
	}

	event := calendar.NewEvent()
	err := event.Parse("2099-01-02 03:04") // far future, will never trigger
	require.NoError(t, err)
	schedules := []*calendar.Event{event}

	file, err := os.CreateTemp(t.TempDir(), "*.xml")
	require.NoError(t, err)
	defer file.Close()

	taskPath := getTaskPath(config.ProfileName, config.CommandName)
	sourceTask := createTaskDefinition(config, schedules)
	sourceTask.RegistrationInfo.URI = taskPath

	// Verify StartWhenAvailable is set in source task
	assert.True(t, sourceTask.Settings.StartWhenAvailable)

	err = createTaskFile(sourceTask, file)
	require.NoError(t, err)
	file.Close()

	result, err := createTask(taskPath, file.Name(), "", "")
	t.Log(result)
	require.NoError(t, err)
	defer func() {
		_, _ = deleteTask(taskPath)
	}()

	// Export and verify the task was created with StartWhenAvailable
	taskXML, err := exportTaskDefinition(taskPath)
	require.NoError(t, err)

	buffer := bytes.NewBuffer(taskXML)
	decoder := xml.NewDecoder(buffer)
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		return input, nil
	}
	readTask := &Task{}
	err = decoder.Decode(&readTask)
	require.NoError(t, err)

	assert.True(t, readTask.Settings.StartWhenAvailable, "StartWhenAvailable should be true in the created task")
}

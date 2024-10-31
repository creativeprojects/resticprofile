package crond

import (
	"fmt"
	"io"
	"os/user"
	"regexp"
	"strings"

	"github.com/creativeprojects/resticprofile/calendar"
)

const (
	startMarker    = "### this content was generated by resticprofile, please leave this line intact ###\n"
	endMarker      = "### end of resticprofile content, please leave this line intact ###\n"
	timeExp        = `^(([\d,\/\-\*]+[ \t]?){5})`
	userExp        = `[\t]+(\w+[\t]+)?`
	workDirExp     = `(cd .+ && )?`
	configExp      = `([^\s]+.+--config[ =]"?([^"\n]+)"? `
	legacyExp      = `[^\n]*--name[ =]([^\s]+)( --.+)? ([a-z]+))$`
	runScheduleExp = `run-schedule ([^\s]+)@([^\s]+))$`
)

var (
	legacyPattern      = regexp.MustCompile(timeExp + userExp + workDirExp + configExp + legacyExp)
	runSchedulePattern = regexp.MustCompile(timeExp + userExp + workDirExp + configExp + runScheduleExp)
)

type Crontab struct {
	file, binary, charset, user string
	entries                     []Entry
}

func NewCrontab(entries []Entry) (c *Crontab) {
	c = &Crontab{entries: entries}

	for i, entry := range c.entries {
		if entry.NeedsUser() {
			c.entries[i] = c.entries[i].WithUser(c.username())
		}
	}

	return
}

// SetBinary sets the crontab binary to use for reading and writing the crontab (if empty, SetFile must be used)
func (c *Crontab) SetBinary(crontabBinary string) {
	c.binary = crontabBinary
}

// SetFile toggles whether to read & write a crontab file instead of using the crontab binary
func (c *Crontab) SetFile(file string) {
	c.file = file
}

// update crontab entries:
//
// - If addEntries is set to true, it will delete and add all new entries
//
// - If addEntries is set to false, it will only delete the matching entries
//
// Return values are the number of entries deleted, and an error if any
func (c *Crontab) update(source string, addEntries bool, w io.StringWriter) (int, error) {
	var err error
	var deleted int

	before, crontab, after, sectionFound := extractOwnSection(source)

	if sectionFound && len(c.entries) > 0 {
		for _, entry := range c.entries {
			var found bool
			crontab, found, err = deleteLine(crontab, entry)
			if err != nil {
				return deleted, err
			}
			if found {
				deleted++
			}
		}
	}

	_, err = w.WriteString(before)
	if err != nil {
		return deleted, err
	}

	if !sectionFound {
		// add a new line at the end of the file before adding our stuff
		_, err = w.WriteString("\n")
		if err != nil {
			return deleted, err
		}
	}

	_, err = w.WriteString(startMarker)
	if err != nil {
		return deleted, err
	}

	if sectionFound {
		_, err = w.WriteString(crontab)
		if err != nil {
			return deleted, err
		}
	}

	if addEntries {
		err = c.Generate(w)
		if err != nil {
			return deleted, err
		}
	}

	_, err = w.WriteString(endMarker)
	if err != nil {
		return deleted, err
	}

	if sectionFound {
		_, err = w.WriteString(after)
		if err != nil {
			return deleted, err
		}
	}
	return deleted, nil
}

func (c *Crontab) Generate(w io.StringWriter) error {
	var err error
	if len(c.entries) > 0 {
		for _, entry := range c.entries {
			err = entry.Generate(w)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Crontab) LoadCurrent() (content string, err error) {
	content, c.charset, err = loadCrontab(c.file, c.binary)
	if err == nil {
		if cleaned := cleanupCrontab(content); cleaned != content {
			if len(c.file) == 0 {
				content = cleaned
			} else {
				err = fmt.Errorf("refusing to change crontab with \"DO NOT EDIT\": %q", c.file)
			}
		}
	}
	return
}

func (c *Crontab) username() string {
	if len(c.user) == 0 {
		if current, err := user.Current(); err == nil {
			c.user = current.Username
		}
		if len(c.user) == 0 || strings.ContainsAny(c.user, "\t \n\r") {
			c.user = "root"
		}
	}
	return c.user
}

func (c *Crontab) Rewrite() error {
	crontab, err := c.LoadCurrent()
	if err != nil {
		return err
	}

	if len(c.file) > 0 && detectNeedsUserColumn(crontab) {
		for i, entry := range c.entries {
			if !entry.HasUser() {
				c.entries[i] = entry.WithUser(c.username())
			}
		}
	}

	buffer := new(strings.Builder)
	_, err = c.update(crontab, true, buffer)
	if err != nil {
		return err
	}

	return saveCrontab(c.file, buffer.String(), c.charset, c.binary)
}

func (c *Crontab) Remove() (int, error) {
	crontab, err := c.LoadCurrent()
	if err != nil {
		return 0, err
	}

	buffer := new(strings.Builder)
	num, err := c.update(crontab, false, buffer)
	if err == nil {
		err = saveCrontab(c.file, buffer.String(), c.charset, c.binary)
	}
	return num, err
}

func (c *Crontab) GetEntries() ([]Entry, error) {
	crontab, err := c.LoadCurrent()
	if err != nil {
		return nil, err
	}

	_, ownSection, _, sectionFound := extractOwnSection(crontab)
	if !sectionFound {
		return nil, nil
	}

	entries := parseEntries(ownSection)
	return entries, nil
}

func cleanupCrontab(crontab string) string {
	// this pattern detects if a header has been added to the output of "crontab -l"
	pattern := regexp.MustCompile(`^# DO NOT EDIT THIS FILE[^\n]*\n#[^\n]*\n#[^\n]*\n`)
	// and removes it if found
	return pattern.ReplaceAllString(crontab, "")
}

// detectNeedsUserColumn attempts to detect if this crontab needs a user column or not (only relevant for direct file access)
func detectNeedsUserColumn(crontab string) bool {
	headerR := regexp.MustCompile(`^#\s*[*]\s+[*]\s+[*]\s+[*]\s+[*](\s+user.*)?\s+(command|cmd).*$`)
	entryR := regexp.MustCompile(`^\s*(\S+\s+\S+\s+\S+\s+\S+\s+\S+|@[a-z]+)((?:\s{2,}|\t)\S+)?(?:\s{2,}|\t)(\S.*)$`)

	var header, userHeader int
	var entries, userEntries float32
	for _, line := range strings.Split(crontab, "\n") {
		if m := headerR.FindStringSubmatch(line); m != nil {
			header++
			if len(m) == 3 && strings.HasPrefix(strings.TrimSpace(m[1]), "user") {
				userHeader++
			}
		} else if m = entryR.FindStringSubmatch(line); m != nil {
			entries++
			if len(m) == 4 && len(m[2]) > 0 {
				userEntries++
			}
		}
	}

	userEntryPercentage := float32(0)
	if entries > 0 {
		userEntryPercentage = userEntries / entries
	}

	if header > 0 {
		if userHeader == header || (userHeader > 0 && userEntryPercentage > 0) {
			return true
		}
		return false
	} else {
		return userEntryPercentage > 0.75
	}
}

// extractOwnSection returns before our section, inside, and after if found.
//
// - It is not returning both start and end markers.
//
// - If not found, it returns the file content in the first string
func extractOwnSection(crontab string) (string, string, string, bool) {
	start := strings.Index(crontab, startMarker)
	end := strings.Index(crontab, endMarker)
	if start == -1 || end == -1 {
		return crontab, "", "", false
	}
	return crontab[:start], crontab[start+len(startMarker) : end], crontab[end+len(endMarker):], true
}

// deleteLine scans crontab for a line with the same config file, profile name and command name,
// and removes it from the output. It returns true when at least one corresponding line was found.
func deleteLine(crontab string, entry Entry) (string, bool, error) {
	// should match a line like:
	// 00,15,30,45 * * * *	/home/resticprofile --no-ansi --config config.yaml --name profile --log backup.log backup
	// or a line like:
	// 00,15,30,45 * * * *	/home/resticprofile --no-ansi --config config.yaml run-schedule backup@profile
	// also with quotes around the config file:
	// 00,15,30,45 * * * *	/home/resticprofile --no-ansi --config "config.yaml" run-schedule backup@profile
	legacy := fmt.Sprintf(`--name %s[^\n]* %s`,
		regexp.QuoteMeta(entry.profileName),
		regexp.QuoteMeta(entry.commandName),
	)
	runSchedule := fmt.Sprintf(`run-schedule %s@%s`,
		regexp.QuoteMeta(entry.commandName),
		regexp.QuoteMeta(entry.profileName),
	)
	search := fmt.Sprintf(`(?m)^[^#][^\n]+resticprofile[^\n]+--config ["]?%s["]? (%s|%s)\n`,
		regexp.QuoteMeta(entry.configFile), legacy, runSchedule)

	pattern, err := regexp.Compile(search)
	if err != nil {
		return crontab, false, err
	}
	if pattern.MatchString(crontab) {
		// at least one was found
		return pattern.ReplaceAllString(crontab, ""), true, nil
	}
	return crontab, false, nil
}

func parseEntries(crontab string) []Entry {
	lines := strings.Split(crontab, "\n")
	entries := make([]Entry, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		entry := parseEntry(line)
		if entry == nil {
			continue
		}
		entries = append(entries, *entry)
	}
	return entries
}

func parseEntry(line string) *Entry {
	// should match lines like:
	// 00,15,30,45 * * * *	/home/resticprofile --no-ansi --config config.yaml --name profile --log backup.log backup
	// 00,15,30,45 * * * *	cd /workdir && /home/resticprofile --no-ansi --config config.yaml --name profile --log backup.log backup
	// or a line like:
	// 00,15,30,45 * * * *	/home/resticprofile --no-ansi --config config.yaml run-schedule backup@profile
	// also with quotes around the config file:
	// 00,15,30,45 * * * *	/home/resticprofile --no-ansi --config "config.yaml" run-schedule backup@profile

	// try legacy pattern first
	matches := legacyPattern.FindStringSubmatch(line)
	if len(matches) == 10 {
		return &Entry{
			event:       parseEvent(matches[1]),
			user:        getUserValue(matches[3]),
			workDir:     getWorkdirValue(matches[4]),
			commandLine: matches[5],
			configFile:  matches[6],
			profileName: matches[7],
			commandName: matches[9],
		}
	}
	// then try the current pattern
	matches = runSchedulePattern.FindStringSubmatch(line)
	if len(matches) == 9 {
		return &Entry{
			event:       parseEvent(matches[1]),
			user:        getUserValue(matches[3]),
			workDir:     getWorkdirValue(matches[4]),
			commandLine: matches[5],
			configFile:  matches[6],
			commandName: matches[7],
			profileName: matches[8],
		}
	}
	return nil
}

func parseEvent(_ string) *calendar.Event {
	event := calendar.NewEvent()
	return event
}

func getUserValue(user string) string {
	return strings.TrimSpace(user)
}

func getWorkdirValue(workdir string) string {
	return strings.TrimSuffix(strings.TrimPrefix(workdir, "cd "), " && ")
}

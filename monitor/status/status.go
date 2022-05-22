package status

import (
	"encoding/json"
	"os"

	"github.com/spf13/afero"
)

// Status of last schedule profile
type Status struct {
	fs       afero.Fs
	filename string
	Profiles map[string]*Profile `json:"profiles"`
}

// NewStatus returns a new blank status
func NewStatus(fileName string) *Status {
	return &Status{
		fs:       afero.NewOsFs(),
		filename: fileName,
		Profiles: make(map[string]*Profile),
	}
}

// newAferoStatus returns a new blank status for unit test
func newAferoStatus(fs afero.Fs, fileName string) *Status {
	return &Status{
		fs:       fs,
		filename: fileName,
		Profiles: make(map[string]*Profile),
	}
}

// Load existing status; does not complain if the file does not exists, or is not readable
func (s *Status) Load() *Status {
	// we're not bothered if the status cannot be loaded
	file, err := s.fs.Open(s.filename)
	if err != nil {
		return s
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	_ = decoder.Decode(s)
	return s
}

// Profile gets the profile from its name (it creates a blank new one if not exists)
func (s *Status) Profile(name string) *Profile {
	if profile, ok := s.Profiles[name]; ok {
		return profile
	}
	profile := newProfile()
	s.Profiles[name] = profile
	return profile
}

// Save current status to the file
func (s *Status) Save() error {
	file, err := s.fs.OpenFile(s.filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	err = encoder.Encode(s)
	if err != nil {
		return err
	}
	return nil
}

//go:build windows

package schedule

// Check returns true if the user is allowed to access the job.
// This is always true on Windows
func (p Permission) Check() bool {
	return true
}

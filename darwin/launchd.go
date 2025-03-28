//go:build darwin

package darwin

// LaunchdJob is an agent definition for launchd
// Documentation found from man launchd.plist(5)
type LaunchdJob struct {
	// This required key uniquely identifies the job to launchd.
	Label string `plist:"Label"`
	// This key maps to the first argument of execv(3) and indicates the absolute path to the executable for the job. If this key is missing, then the first
	// element of the array of strings provided to the ProgramArguments will be used instead. This key is required in the absence of the ProgramArguments and
	// BundleProgram keys.
	Program string `plist:"Program"`
	// This key maps to the second argument of execvp(3) and specifies the argument vector to be passed to the job when a process is spawned. This key is
	// required in the absence of the Program key.  IMPORTANT: Many people are confused by this key. Please read execvp(3) very carefully!
	//
	// NOTE: The Program key must be an absolute path. Previous versions of launchd did not enforce this requirement but failed to run the job. In the absence
	// of the Program key, the first element of the ProgramArguments array may be either an absolute path, or a relative path which is resolved using
	// _PATH_STDPATH.
	ProgramArguments []string `plist:"ProgramArguments"`
	// This optional key is used to specify additional environmental variables to be set before running the job. Each key in the dictionary is the name of an
	// environment variable, with the corresponding value being a string representing the desired value.  NOTE: Values other than strings will be ignored.
	EnvironmentVariables map[string]string `plist:"EnvironmentVariables,omitempty"`
	// This optional key specifies that the given path should be mapped to the job's stdin(4), and that the contents of that file will be readable from the
	//  job's stdin(4).  If the file does not exist, no data will be delivered to the process' stdin(4).
	StandardInPath string `plist:"StandardInPath,omitempty"`
	// This optional key specifies that the given path should be mapped to the job's stdout(4), and that any writes to the job's stdout(4) will go to the given
	// file. If the file does not exist, it will be created with writable permissions and ownership reflecting the user and/or group specified as the UserName
	// and/or GroupName, respectively (if set) and permissions reflecting the umask(2) specified by the Umask key, if set.
	StandardOutPath string `plist:"StandardOutPath,omitempty"`
	// This optional key specifies that the given path should be mapped to the job's stderr(4), and that any writes to the job's stderr(4) will go to the given
	// file. Note that this file is opened as readable and writable as mandated by the POSIX specification for unclear reasons.  If the file does not exist, it
	// will be created with ownership reflecting the user and/or group specified as the UserName and/or GroupName, respectively (if set) and permissions
	// reflecting the umask(2) specified by the Umask key, if set.
	StandardErrorPath string `plist:"StandardErrorPath,omitempty"`
	// This optional key is used to specify a directory to chdir(2) to before running the job.
	WorkingDirectory string `plist:"WorkingDirectory"`
	// This optional key causes the job to be started every calendar interval as specified. Missing arguments are considered to be wildcard. The semantics are
	// similar to crontab(5) in how firing dates are specified. Multiple dictionaries may be specified in an array to schedule multiple calendar intervals.
	//
	// Unlike cron which skips job invocations when the computer is asleep, launchd will start the job the next time the computer wakes up. If multiple
	// intervals transpire before the computer is woken, those events will be coalesced into one event upon wake from sleep.
	//
	// Note that StartInterval and StartCalendarInterval are not aware of each other. They are evaluated completely independently by the system.
	//
	// 	Minute <integer>
	// 	The minute (0-59) on which this job will be run.
	//
	// 	Hour <integer>
	// 	The hour (0-23) on which this job will be run.
	//
	// 	Day <integer>
	// 	The day of the month (1-31) on which this job will be run.
	//
	// 	Weekday <integer>
	// 	The weekday on which this job will be run (0 and 7 are Sunday). If both Day and Weekday are specificed, then the job will be started if either one
	// 	matches the current date.
	//
	// 	Month <integer>
	// 	The month (1-12) on which this job will be run.
	StartCalendarInterval []CalendarInterval `plist:"StartCalendarInterval,omitempty"`
	// ProcessType
	// This optional key describes, at a high level, the intended purpose of the job.  The system will apply resource limits based on what kind of job it is. If
	// left unspecified, the system will apply light resource limits to the job, throttling its CPU usage and I/O bandwidth. This classification is preferable
	// to using the HardResourceLimits, SoftResourceLimits and Nice keys. The following are valid values:
	//
	//  Background
	//  Background jobs are generally processes that do work that was not directly requested by the user. The resource limits applied to Background jobs
	//  are intended to prevent them from disrupting the user experience.
	//
	//  Standard
	//  Standard jobs are equivalent to no ProcessType being set.
	//
	//  Adaptive
	//  Adaptive jobs move between the Background and Interactive classifications based on activity over XPC connections. See xpc_transaction_begin(3) for
	//  details.
	//
	//  Interactive
	//  Interactive jobs run with the same resource limitations as apps, that is to say, none. Interactive jobs are critical to maintaining a responsive
	//  user experience, and this key should only be used if an app's ability to be responsive depends on it, and cannot be made Adaptive.
	ProcessType ProcessType `plist:"ProcessType"`
	// This optional key specifies whether the kernel should consider this daemon to be low priority when doing filesystem I/O.
	LowPriorityIO bool `plist:"LowPriorityIO"`
	// This optional key specifies whether the kernel should consider this daemon to be low priority when doing filesystem I/O when
	// the process is throttled with the Darwin-background classification.
	LowPriorityBackgroundIO bool `plist:"LowPriorityBackgroundIO"`
	// This optional key specifies what nice(3) value should be applied to the daemon.
	Nice int `plist:"Nice"`
	//   Aqua:
	// a GUI agent; has access to all the GUI services
	//   LoginWindow:
	// pre-login agent; runs in the login window context
	//   Background:
	// runs in the parent context of the user
	//   StandardIO:
	// runs only in non-GUI login session (e.g. SSH sessions)
	//   System:
	// runs in the system context
	LimitLoadToSessionType SessionType `plist:"LimitLoadToSessionType,omitempty"`
}

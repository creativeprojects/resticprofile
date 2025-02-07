package schtasks

import "github.com/rickb777/date/period"

// Settings provides the settings that the Task Scheduler service uses to perform the task
// https://docs.microsoft.com/en-us/windows/desktop/api/taskschd/nn-taskschd-itasksettings
type Settings struct {
	AllowDemandStart           bool            `xml:"AllowStartOnDemand"` // indicates that the task can be started by using either the Run command or the Context menu
	AllowHardTerminate         bool            `xml:"AllowHardTerminate"` // indicates that the task may be terminated by the Task Scheduler service using TerminateProcess
	Compatibility              Compatibility   // indicates which version of Task Scheduler a task is compatible with
	DeleteExpiredTaskAfter     *period.Period  `xml:"DeleteExpiredTaskAfter,omitempty"` // the amount of time that the Task Scheduler will wait before deleting the task after it expires
	DisallowStartIfOnBatteries bool            `xml:"DisallowStartIfOnBatteries"`       // indicates that the task will not be started if the computer is running on batteries
	Enabled                    bool            `xml:"Enabled"`                          // indicates that the task is enabled
	ExecutionTimeLimit         period.Period   `xml:"ExecutionTimeLimit"`               // the amount of time that is allowed to complete the task
	Hidden                     bool            `xml:"Hidden"`                           // indicates that the task will not be visible in the UI
	IdleSettings               IdleSettings    `xml:"IdleSettings"`
	MultipleInstancesPolicy    InstancesPolicy `xml:"MultipleInstancesPolicy"` // defines how the Task Scheduler deals with multiple instances of the task
	// NetworkSettings            NetworkSettings
	Priority                  uint              `xml:"Priority"` // the priority level of the task, ranging from 0 - 10, where 0 is the highest priority, and 10 is the lowest. Only applies to ComHandler, Email, and MessageBox actions
	RestartOnFailure          *RestartOnFailure `xml:"RestartOnFailure,omitempty"`
	RunOnlyIfIdle             bool              `xml:"RunOnlyIfIdle"`             // indicates that the Task Scheduler will run the task only if the computer is in an idle condition
	RunOnlyIfNetworkAvailable bool              `xml:"RunOnlyIfNetworkAvailable"` // indicates that the Task Scheduler will run the task only when a network is available
	StartWhenAvailable        bool              `xml:"StartWhenAvailable"`        // indicates that the Task Scheduler can start the task at any time after its scheduled time has passed
	StopIfGoingOnBatteries    bool              `xml:"StopIfGoingOnBatteries"`    // indicates that the task will be stopped if the computer is going onto batteries
	WakeToRun                 bool              `xml:"WakeToRun"`                 // indicates that the Task Scheduler will wake the computer when it is time to run the task, and keep the computer awake until the task is completed
}

// IdleSettings specifies how the Task Scheduler performs tasks when the computer is in an idle condition.
type IdleSettings struct {
	Duration      period.Period `xml:"Duration"`      // the amount of time that the computer must be in an idle state before the task is run
	RestartOnIdle bool          `xml:"RestartOnIdle"` // whether the task is restarted when the computer cycles into an idle condition more than once
	StopOnIdleEnd bool          `xml:"StopOnIdleEnd"` // indicates that the Task Scheduler will terminate the task if the idle condition ends before the task is completed
	WaitTimeout   period.Period `xml:"WaitTimeout"`   // the amount of time that the Task Scheduler will wait for an idle condition to occur
}

// NetworkSettings provides the settings that the Task Scheduler service uses to obtain a network profile.
// https://docs.microsoft.com/en-us/windows/desktop/api/taskschd/nn-taskschd-inetworksettings
type NetworkSettings struct {
	ID   string // a GUID value that identifies a network profile
	Name string // the name of a network profile
}

// For scripting, gets or sets an integer value that indicates which version of Task Scheduler a task is compatible with.
type Compatibility int

const (
	TaskCompatibilityAT  Compatibility = iota // The task is compatible with the AT command.
	TaskCompatibilityV1                       // The task is compatible with Task Scheduler 1.0.
	TaskCompatibilityV2                       // The task is compatible with Task Scheduler 2.0.
	TaskCompatibilityV21                      // The task is compatible with Task Scheduler 2.0.
	TaskCompatibilityV22                      // The task is compatible with Task Scheduler 2.0.
	TaskCompatibilityV23                      // The task is compatible with Task Scheduler 2.0.
	TaskCompatibilityV24                      // The task is compatible with Task Scheduler 2.0.
)

// InstancesPolicy specifies what the Task Scheduler service will do when
// multiple instances of a task are triggered or operating at once.
type InstancesPolicy string

const (
	MultipleInstancesParallel     InstancesPolicy = "Parallel"     // start new instance while an existing instance is running
	MultipleInstancesQueue        InstancesPolicy = "Queue"        // start a new instance of the task after all other instances of the task are complete
	MultipleInstancesIgnoreNew    InstancesPolicy = "IgnoreNew"    // do not start a new instance if an existing instance of the task is running
	MultipleInstancesStopExisting InstancesPolicy = "StopExisting" // stop an existing instance of the task before it starts a new instance
)

type RestartOnFailure struct {
	Count    uint          // the number of times that the Task Scheduler will attempt to restart the task
	Interval period.Period // specifies how long the Task Scheduler will attempt to restart the task
}

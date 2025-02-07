package schtasks

type Actions struct {
	Context string       `xml:"Context,attr"`
	Exec    []ExecAction `xml:"Exec"`
}

type ExecAction struct {
	Command          string `xml:"Command"`
	Arguments        string `xml:"Arguments"`
	WorkingDirectory string `xml:"WorkingDirectory,omitempty"`
}

package schtasks

const ServiceAccount = "S-1-5-18"

type Principals struct {
	Principal Principal `xml:"Principal"`
}

type Principal struct {
	ID        string    `xml:"id,attr"`
	UserId    string    `xml:"UserId"`
	LogonType LogonType `xml:"LogonType,omitempty"`
	RunLevel  RunLevel  `xml:"RunLevel,omitempty"`
}

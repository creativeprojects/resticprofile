package config

// ProfileInfo is used to display a quick information on available profiles (descrition and defined commands)
type ProfileInfo struct {
	Description string
	Sections    []string
}

func NewProfileInfo() ProfileInfo {
	return ProfileInfo{
		Sections: make([]string, 0),
	}
}

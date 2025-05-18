package config

// SendMonitoringSections is a group of target to send monitoring information
type SendMonitoringSections struct {
	SendBefore    []SendMonitoringSection `mapstructure:"send-before" description:"Send HTTP request(s) before a restic command"`
	SendAfter     []SendMonitoringSection `mapstructure:"send-after" description:"Send HTTP request(s) after a successful restic command"`
	SendAfterFail []SendMonitoringSection `mapstructure:"send-after-fail" description:"Send HTTP request(s) after failed restic or shell commands"`
	SendFinally   []SendMonitoringSection `mapstructure:"send-finally" description:"Send HTTP request(s) always, after all other commands"`
}

func (s *SendMonitoringSections) setRootPath(_ *Profile, rootPath string) {
	for _, sections := range s.getAllSendMonitoringSections() {
		for index := range sections {
			sections[index].BodyTemplate = fixPath(sections[index].BodyTemplate, expandEnv, expandUserHome, absolutePrefix(rootPath))
		}
	}
}

func (s *SendMonitoringSections) GetSendMonitoring() *SendMonitoringSections { return s }

func (s *SendMonitoringSections) getAllSendMonitoringSections() [][]SendMonitoringSection {
	return [][]SendMonitoringSection{
		s.SendBefore,
		s.SendAfter,
		s.SendAfterFail,
		s.SendFinally,
	}
}

// SendMonitoringSection is used to send monitoring information to third party software
type SendMonitoringSection struct {
	Method       string                 `mapstructure:"method" enum:"GET;DELETE;HEAD;OPTIONS;PATCH;POST;PUT;TRACE" default:"GET" description:"HTTP method of the request"`
	URL          ConfidentialValue      `mapstructure:"url" format:"uri" description:"URL of the target to send to"`
	Headers      []SendMonitoringHeader `mapstructure:"headers" description:"Additional HTTP headers to send with the request"`
	Body         string                 `mapstructure:"body" description:"Request body, overrides \"body-template\""`
	BodyTemplate string                 `mapstructure:"body-template" description:"Path to a file containing the request body (go template). See https://creativeprojects.github.io/resticprofile/configuration/http_hooks/#body-template"`
	SkipTLS      bool                   `mapstructure:"skip-tls-verification" description:"Enables insecure TLS (without verification), see also \"global.ca-certificates\""`
}

// SendMonitoringHeader is used to send HTTP headers
type SendMonitoringHeader struct {
	Name  string            `mapstructure:"name" regex:"^\\w([\\w-]+)\\w$" examples:"\"Authorization\";\"Cache-Control\";\"Content-Disposition\";\"Content-Type\"" description:"Name of the HTTP header"`
	Value ConfidentialValue `mapstructure:"value" examples:"\"Bearer ...\";\"Basic ...\";\"no-cache\";\"attachment;; filename=stats.txt\";\"application/json\";\"text/plain\";\"text/xml\"" description:"Value of the header"`
}

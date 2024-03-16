package dial

import (
	"net/url"
	"slices"
	"strings"

	"github.com/creativeprojects/clog"
)

var validSchemes = []string{
	"udp",
	"tcp",
	"syslog",     // local or UDP
	"syslog-tcp", // TCP
	// "syslog-tls", reserved for future support
}

var noHostAllowed = []string{
	"syslog",
}

// GetAddr returns scheme, host&port, isURL
func GetAddr(source string) (scheme, hostPort string, isURL bool) {
	URL, err := url.Parse(source)
	if err == nil {
		scheme = strings.ToLower(URL.Scheme)
		hostPort = URL.Host
		schemeOk := slices.Contains(validSchemes, scheme)
		hostOk := len(hostPort) >= 3 || slices.Contains(noHostAllowed, scheme)
		if isURL = schemeOk && hostOk; isURL {
			return
		}
	} else {
		clog.Tracef("is not an URL %q", source)
	}
	return "", "", false
}

func IsURL(source string) bool {
	_, _, isURL := GetAddr(source)
	return isURL
}

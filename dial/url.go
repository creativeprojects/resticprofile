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

// GetAddr returns scheme, host&port, is(Supported)URL
func GetAddr(source string) (scheme, hostPort string, isURL bool) {
	URL, err := url.Parse(source)
	if err == nil {
		scheme = strings.ToLower(URL.Scheme)
		hostPort = URL.Host
		schemeOk := slices.Contains(validSchemes, scheme)
		hostOk := len(hostPort) >= 3 || (slices.Contains(noHostAllowed, scheme) && len(URL.Opaque) == 0)
		if isURL = schemeOk && hostOk; isURL {
			return
		}
	} else {
		clog.Tracef("is not an URL %q", source)
	}
	return "", "", false
}

// IsSupportedURL returns true if the provided source is valid for GetAddr
func IsSupportedURL(source string) bool {
	_, _, isURL := GetAddr(source)
	return isURL
}

// IsURL is true if the provided source is a parsable URL and no file path
func IsURL(source string) bool {
	u, e := url.Parse(source)
	return e == nil && len(u.Scheme) > 1
}

package dial

import "net/url"

// GetAddr returns scheme, host&port, isURL
func GetAddr(source string) (scheme, hostPort string, isURL bool) {
	URL, err := url.Parse(source)
	if err != nil {
		return "", "", false
	}
	// need a minimum of udp://:12
	if len(URL.Scheme) < 3 || len(URL.Host) < 3 {
		return "", "", false
	}
	return URL.Scheme, URL.Host, true
}

func IsURL(source string) bool {
	_, _, isURL := GetAddr(source)
	return isURL
}

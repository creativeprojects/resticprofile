package ssh

import (
	"strconv"
	"strings"
)

func parseHost(host string) (string, int) {
	if strings.Contains(host, ":") {
		parts := strings.Split(host, ":")
		if len(parts) > 2 {
			// If there are more than two parts, we assume the first part is the host and the rest is the port
			host = strings.Join(parts[:len(parts)-1], ":")
			port, _ := strconv.Atoi(parts[len(parts)-1])
			return host, port
		}
		port, _ := strconv.Atoi(parts[1])
		return parts[0], port
	}
	return host, 0
}

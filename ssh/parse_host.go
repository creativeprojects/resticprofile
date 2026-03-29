package ssh

import (
	"strconv"
	"strings"
)

func parseHost(host string) (string, int) {
	if strings.Contains(host, ":") {
		parts := strings.Split(host, ":")
		if len(parts) > 2 {
			// If the next to last part is empty, it means we have an IPv6 address without a port (last : is a double ::)
			if parts[len(parts)-2] == "" {
				return host, 22
			}
			// If there are more than two parts, we assume the first part is the host and the rest is the port
			host = strings.Join(parts[:len(parts)-1], ":")
			port, _ := strconv.Atoi(parts[len(parts)-1])
			return host, sshPort(port)
		}
		port, _ := strconv.Atoi(parts[1])
		return parts[0], sshPort(port)
	}
	return host, sshPort(0)
}

func sshPort(port int) int {
	if port == 0 {
		return 22
	}
	return port
}

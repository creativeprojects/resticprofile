package mask

import "regexp"

const maskReplacement = "×××"

var (
	// httpHeaderNames     = regexp.MustCompile("(?i)^(Authorization)$")
	RepositoryConfidentialPart = regexp.MustCompile("[:/][^:/@]+?:([^:@]+?)@[^:/@]+?") // user:pass@host
	// urlEnvKeys          = regexp.MustCompile("(?i)^.+(_AUTH|_URL)$")
	// hiddenEnvKeys       = regexp.MustCompile("(?i)^(.+_KEY|.+_TOKEN|.*PASSWORD.*|.*SECRET.*)$")
)

func Submatches(pattern *regexp.Regexp, value string) string {
	if matches := pattern.FindStringSubmatchIndex(value); len(matches) > 2 {
		for i := len(matches) - 2; i > 1; i -= 2 {
			start := matches[i]
			end := matches[i+1]

			value = value[0:start] + maskReplacement + value[end:]
		}
	}
	return value
}

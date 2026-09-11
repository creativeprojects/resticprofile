//go:build !darwin && !windows && !openbsd && !netbsd && !freebsd

package schedule

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

// decodeSystemctlUnitsJSON parses systemctl list-units --output=json.
func decodeSystemctlUnitsJSON(data []byte) ([]SystemdUnit, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return nil, nil
	}
	var units []SystemdUnit
	if err := json.Unmarshal(trimmed, &units); err != nil {
		return nil, err
	}
	return units, nil
}

// parseSystemctlUnitsPlain parses classic systemctl list-units table output
// (used when older systemd ignores --output=json, e.g. AlmaLinux 8 / Amazon Linux 2).
//
// Accepts both the default table (with header/legend) and --plain --no-legend lines:
//
//	UNIT LOAD ACTIVE SUB DESCRIPTION
func parseSystemctlUnitsPlain(data []byte) ([]SystemdUnit, error) {
	var units []SystemdUnit
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || isSystemctlListUnitsMetaLine(line) {
			continue
		}
		line = trimSystemctlUnitGlyph(line)
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		unitName := fields[0]
		if !looksLikeSystemdUnitName(unitName) {
			continue
		}
		unit := SystemdUnit{
			Unit:   unitName,
			Load:   fields[1],
			Active: fields[2],
			Sub:    fields[3],
		}
		if len(fields) > 4 {
			unit.Description = strings.Join(fields[4:], " ")
		}
		units = append(units, unit)
	}
	return units, nil
}

func isSystemctlListUnitsMetaLine(line string) bool {
	upper := strings.ToUpper(line)
	switch {
	case strings.HasPrefix(upper, "UNIT ") || upper == "UNIT":
		return true
	case strings.HasPrefix(upper, "LOAD ") || strings.HasPrefix(upper, "LOAD="):
		return true
	case strings.HasPrefix(upper, "ACTIVE ") || strings.HasPrefix(upper, "ACTIVE="):
		return true
	case strings.HasPrefix(upper, "SUB ") || strings.HasPrefix(upper, "SUB="):
		return true
	case strings.Contains(line, "loaded units listed"):
		return true
	case strings.HasPrefix(line, "To show all installed unit files"):
		return true
	default:
		return false
	}
}

func trimSystemctlUnitGlyph(line string) string {
	return strings.TrimLeftFunc(line, func(r rune) bool {
		// systemctl may prefix failed/not-found units with a bullet (●) or similar.
		return unicode.IsSpace(r) || r == '●' || r == '○' || r == '•' || r == '·'
	})
}

func looksLikeSystemdUnitName(name string) bool {
	// resticprofile units look like resticprofile-backup@profile-name.service
	dot := strings.LastIndex(name, ".")
	if dot <= 0 || dot == len(name)-1 {
		return false
	}
	suffix := name[dot+1:]
	switch suffix {
	case "service", "timer", "socket", "target", "path", "mount", "automount", "swap", "slice", "scope", "device":
		return true
	default:
		return false
	}
}

// parseSystemctlUnits prefers JSON, then falls back to the classic table format.
func parseSystemctlUnits(data []byte) ([]SystemdUnit, error) {
	units, err := decodeSystemctlUnitsJSON(data)
	if err == nil {
		return units, nil
	}
	jsonErr := err
	units, plainErr := parseSystemctlUnitsPlain(data)
	if plainErr != nil {
		return nil, fmt.Errorf("error decoding JSON: %w (plain fallback: %v)\n%s", jsonErr, plainErr, data)
	}
	// If plain parsing found nothing and the payload looked like JSON, keep the JSON error.
	trimmed := bytes.TrimSpace(data)
	if len(units) == 0 && len(trimmed) > 0 && (trimmed[0] == '[' || trimmed[0] == '{') {
		return nil, fmt.Errorf("error decoding JSON: %w\n%s", jsonErr, data)
	}
	return units, nil
}

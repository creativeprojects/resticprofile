//go:build !darwin && !windows && !openbsd && !netbsd && !freebsd

package schedule

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeSystemctlUnitsJSON(t *testing.T) {
	t.Parallel()

	raw := `[
  {"unit":"resticprofile-backup@profile-default.service","load":"loaded","active":"inactive","sub":"dead","description":"backup default"},
  {"unit":"resticprofile-check@profile-default.service","load":"not-found","active":"inactive","sub":"dead","description":""}
]`
	units, err := decodeSystemctlUnitsJSON([]byte(raw))
	require.NoError(t, err)
	require.Len(t, units, 2)
	assert.Equal(t, "resticprofile-backup@profile-default.service", units[0].Unit)
	assert.Equal(t, "loaded", units[0].Load)
	assert.Equal(t, "backup default", units[0].Description)
	assert.Equal(t, unitNotFound, units[1].Load)
}

func TestDecodeSystemctlUnitsJSON_Empty(t *testing.T) {
	t.Parallel()
	units, err := decodeSystemctlUnitsJSON([]byte("[]"))
	require.NoError(t, err)
	assert.Empty(t, units)

	units, err = decodeSystemctlUnitsJSON(nil)
	require.NoError(t, err)
	assert.Empty(t, units)
}

func TestParseSystemctlUnitsPlain_AlmaLinuxStyle(t *testing.T) {
	t.Parallel()

	// Classic table output when --output=json is ignored (issue #516).
	raw := `  UNIT                                                LOAD      ACTIVE   SUB     DESCRIPTION
  resticprofile-backup@profile-default.service       loaded    inactive dead    resticprofile backup for profile default
● resticprofile-check@profile-missing.service        not-found inactive dead    resticprofile-check@profile-missing.service

LOAD   = Reflects whether the unit definition was properly loaded.
ACTIVE = The high-level unit activation state, i.e. generalization of SUB.
SUB    = The low-level unit activation state, values depend on unit type.

2 loaded units listed.
To show all installed unit files use 'systemctl list-unit-files'.
`
	units, err := parseSystemctlUnitsPlain([]byte(raw))
	require.NoError(t, err)
	require.Len(t, units, 2)
	assert.Equal(t, "resticprofile-backup@profile-default.service", units[0].Unit)
	assert.Equal(t, "loaded", units[0].Load)
	assert.Equal(t, "inactive", units[0].Active)
	assert.Equal(t, "dead", units[0].Sub)
	assert.Equal(t, "resticprofile backup for profile default", units[0].Description)

	assert.Equal(t, "resticprofile-check@profile-missing.service", units[1].Unit)
	assert.Equal(t, unitNotFound, units[1].Load)
}

func TestParseSystemctlUnitsPlain_NoLegend(t *testing.T) {
	t.Parallel()

	raw := `resticprofile-backup@profile-self.service loaded active running backup self
resticprofile-forget@profile-self.service loaded inactive dead forget self
`
	units, err := parseSystemctlUnitsPlain([]byte(raw))
	require.NoError(t, err)
	require.Len(t, units, 2)
	assert.Equal(t, "resticprofile-backup@profile-self.service", units[0].Unit)
	assert.Equal(t, "active", units[0].Active)
	assert.Equal(t, "running", units[0].Sub)
}

func TestParseSystemctlUnitsPlain_LegendOnlyCountsAsEmpty(t *testing.T) {
	t.Parallel()

	// Matches the "cannot unmarshal number" case: footer starts with a digit.
	raw := `0 loaded units listed.
To show all installed unit files use 'systemctl list-unit-files'.
`
	units, err := parseSystemctlUnitsPlain([]byte(raw))
	require.NoError(t, err)
	assert.Empty(t, units)
}

func TestParseSystemctlUnits_FallsBackFromIgnoredJSON(t *testing.T) {
	t.Parallel()

	raw := `UNIT LOAD ACTIVE SUB DESCRIPTION
resticprofile-backup@profile-default.service loaded inactive dead backup default

LOAD   = Reflects whether the unit definition was properly loaded.
1 loaded units listed.
`
	units, err := parseSystemctlUnits([]byte(raw))
	require.NoError(t, err)
	require.Len(t, units, 1)
	assert.Equal(t, "resticprofile-backup@profile-default.service", units[0].Unit)
}

func TestParseSystemctlUnits_FallsBackFromNumericFooter(t *testing.T) {
	t.Parallel()

	raw := "2 loaded units listed.\n"
	units, err := parseSystemctlUnits([]byte(raw))
	require.NoError(t, err)
	assert.Empty(t, units)
}

func TestParseSystemctlUnits_PrefersJSON(t *testing.T) {
	t.Parallel()

	raw := `[{"unit":"resticprofile-backup@profile-a.service","load":"loaded","active":"active","sub":"running","description":"from json"}]`
	units, err := parseSystemctlUnits([]byte(raw))
	require.NoError(t, err)
	require.Len(t, units, 1)
	assert.Equal(t, "from json", units[0].Description)
}

func TestParseSystemctlUnits_BrokenJSONStillErrors(t *testing.T) {
	t.Parallel()

	raw := `[{"unit":"resticprofile-backup@profile-a.service"`
	_, err := parseSystemctlUnits([]byte(raw))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error decoding JSON")
}

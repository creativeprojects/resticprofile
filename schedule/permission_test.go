package schedule

// TODO rewrite this test!
// func TestDetectPermission(t *testing.T) {
// 	fixtures := []struct {
// 		input    string
// 		expected string
// 		safe     bool
// 		active   bool
// 	}{
// 		{"", "system", true, platform.IsWindows()},
// 		{"something", "system", true, platform.IsWindows()},
// 		{"", "user_logged_on", platform.IsDarwin(), !platform.IsWindows()},
// 		{"something", "user_logged_on", platform.IsDarwin(), !platform.IsWindows()},
// 		{"system", "system", true, true},
// 		{"user", "user", true, true},
// 		{"user_logged_on", "user_logged_on", true, true},
// 		{"user_logged_in", "user_logged_on", true, true}, // I did the typo as I was writing the doc, so let's add it here :)
// 	}
// 	for _, fixture := range fixtures {
// 		if !fixture.active {
// 			continue
// 		}
// 		t.Run(fixture.input, func(t *testing.T) {
// 			perm, safe := PermissionFromConfig(fixture.input).Detect()
// 			assert.Equal(t, fixture.expected, perm.String())
// 			assert.Equal(t, fixture.safe, safe)
// 		})
// 	}
// }

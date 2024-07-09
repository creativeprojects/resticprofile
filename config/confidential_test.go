package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/platform"
	"github.com/creativeprojects/resticprofile/shell"
	"github.com/creativeprojects/resticprofile/util/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	defaultUrl             = "local:user:pass@host"
	defaultUrlReplaced     = fmt.Sprintf("local:user:%s@host", ConfidentialReplacement)
	defaultHttpUrl         = "http://huser:hpw@host"
	defaultHttpUrlReplaced = fmt.Sprintf("http://huser:%s@host", ConfidentialReplacement)
)

func TestConfidentialHideAll(t *testing.T) {
	value := NewConfidentialValue("val")

	assert.Equal(t, value.String(), value.Value())
	assert.Equal(t, "val", value.String())

	value.hideValue()
	assert.Equal(t, "val", value.Value())
	assert.Equal(t, ConfidentialReplacement, value.String())
}

func TestConfidentialHideSubmatch(t *testing.T) {
	value := NewConfidentialValue("some-vAl-with-sEcRet-parts")

	assert.Equal(t, value.String(), value.Value())
	assert.Equal(t, "some-vAl-with-sEcRet-parts", value.String())

	value.hideSubmatches(regexp.MustCompile("(?i).+(val).+(secret).+"))
	assert.Equal(t, "some-vAl-with-sEcRet-parts", value.Value())

	expected := fmt.Sprintf("some-%s-with-%s-parts", ConfidentialReplacement, ConfidentialReplacement)
	assert.Equal(t, expected, value.String())
}

func TestUpdateConfidentialValue(t *testing.T) {
	v := NewConfidentialValue("abc")

	v.setValue("xyz")
	assert.Equal(t, "xyz", v.Value())
	assert.Equal(t, "xyz", v.String())

	v.hideSubmatches(regexp.MustCompile("(z)"))
	assert.Equal(t, "xy"+ConfidentialReplacement, v.String())

	v.setValue("abc")
	assert.Equal(t, "abc", v.Value())
	assert.Equal(t, "xy"+ConfidentialReplacement, v.String())
}

func TestFmtStringDoesntLeakConfidentialValues(t *testing.T) {
	value := NewConfidentialValue("secret")
	value.hideValue()

	assert.Equal(t, ConfidentialReplacement, fmt.Sprintf("%s", value))
	assert.Equal(t, ConfidentialReplacement, fmt.Sprintf("%v", value))
	assert.Equal(t, ConfidentialReplacement, value.String())
	assert.Equal(t, "secret", value.Value())
}

func TestStringifyPassesConfidentialValues(t *testing.T) {
	value := NewConfidentialValue("secret")
	value.hideValue()

	v1, _ := stringifyValue(reflect.ValueOf(value))
	v2, _ := stringifyConfidentialValue(reflect.ValueOf(value))
	assert.Equal(t, []string{ConfidentialReplacement}, v1)
	assert.Equal(t, []string{"secret"}, v2)
}

func TestConfidentialURLs(t *testing.T) {
	// https://restic.readthedocs.io/en/latest/030_preparing_a_new_repo.html
	urls := map[string]string{
		"local:some/path":                                      "-",
		"sftp:user@host:/srv/restic-repo":                      "-",
		"sftp://user@[::1]:2222//srv/restic-repo":              "-",
		"sftp:restic-backup-host:/srv/restic-repo":             "-",
		"rest:http://host:8000/":                               "-",
		"rest:https://user:1234fdfASDasfwY.-+;@host:8000/":     fmt.Sprintf("rest:https://user:%s@host:8000/", ConfidentialReplacement),
		"rest:https://user:35%3Asad%C3%B6p%C3%9F@host:8000/":   fmt.Sprintf("rest:https://user:%s@host:8000/", ConfidentialReplacement),
		"rest:https://user:pass@host:8000/f/":                  fmt.Sprintf("rest:https://user:%s@host:8000/f/", ConfidentialReplacement),
		"s3:s3.amazonaws.com/bucket_name":                      "-",
		"s3:https://<WASABI-SERVICE-URL>/<WASABI-BUCKET-NAME>": "-",
		"swift:container_name:/path":                           "-",
		"azure:foo:/":                                          "-",
		"gs:foo:/":                                             "-",
		"rclone:b2prod:yggdrasil/foo/bar/baz":                  "-",
	}

	for url, expected := range urls {
		testConfig := fmt.Sprintf(`
[profile]
repository = "%s"
`, url)

		profile, err := getProfile("toml", testConfig, "profile", "")
		assert.Nil(t, err)
		assert.NotNil(t, profile)

		if expected == "-" {
			expected = url
		}
		assert.Equal(t, expected, profile.Repository.String())
		assert.Equal(t, url, profile.Repository.Value())
	}
}

func TestConfidentialEnvironment(t *testing.T) {
	// https://restic.readthedocs.io/en/latest/030_preparing_a_new_repo.html
	vars := map[string]string{
		"MY_VALUE": "-",
		"MY_KEY":   "*",
		"PASSWORD": "*",
		"MY_URL":   "<url>",
		// AWS, MinIO, Wasabi, Alibaba Cloud
		"AWS_ACCESS_KEY_ID":     "-",
		"AWS_SECRET_ACCESS_KEY": "*",
		"AWS_DEFAULT_REGION":    "-",
		// OpenStack Swift
		"ST_AUTH":                          "<url>",
		"ST_USER":                          "-",
		"ST_KEY":                           "*",
		"OS_AUTH_URL":                      "<url>",
		"OS_REGION_NAME":                   "-",
		"OS_USERNAME":                      "-",
		"OS_PASSWORD":                      "*",
		"OS_TENANT_ID":                     "-",
		"OS_TENANT_NAME":                   "-",
		"OS_USER_ID":                       "-",
		"OS_USER_DOMAIN_NAME":              "-",
		"OS_USER_DOMAIN_ID":                "-",
		"OS_PROJECT_NAME":                  "-",
		"OS_PROJECT_DOMAIN_NAME":           "-",
		"OS_PROJECT_DOMAIN_ID":             "-",
		"OS_TRUST_ID":                      "-",
		"OS_APPLICATION_CREDENTIAL_ID":     "-",
		"OS_APPLICATION_CREDENTIAL_NAME":   "-",
		"OS_APPLICATION_CREDENTIAL_SECRET": "*",
		"OS_STORAGE_URL":                   "<url>",
		"OS_AUTH_TOKEN":                    "*",
		"SWIFT_DEFAULT_CONTAINER_POLICY":   "-",
		// Backblaze B2
		"B2_ACCOUNT_ID":  "-",
		"B2_ACCOUNT_KEY": "*",
		// Microsoft Azure Blob Storage
		"AZURE_ACCOUNT_NAME": "-",
		"AZURE_ACCOUNT_KEY":  "*",
		// Google Cloud Storage
		"GOOGLE_PROJECT_ID":              "-",
		"GOOGLE_APPLICATION_CREDENTIALS": "-",
		"GOOGLE_ACCESS_TOKEN":            "*",
	}

	for name, expected := range vars {
		testConfig := fmt.Sprintf(`
[profile.env]
%s = "%s"
`, name, defaultUrl)

		profile, err := getProfile("toml", testConfig, "profile", "")
		assert.Nil(t, err)
		assert.NotNil(t, profile)

		switch expected {
		case "<url>":
			expected = defaultUrlReplaced
		case "*":
			expected = ConfidentialReplacement
		default:
			expected = defaultUrl
		}

		name = strings.ToLower(name)
		env := profile.Environment[name]
		assert.Equal(t, expected, env.String())
		assert.Equal(t, defaultUrl, env.Value())
	}
}

func TestShowConfigHidesConfidentialValues(t *testing.T) {
	testConfig := `
profile:
  repository: "local:user:pass@host"
  env:
    MY_VALUE: "val"
    MY_URL: "local:user:pass@host"
    MY_KEY: 1234
    MY_TOKEN: false
    MY_PASSWORD: "otherval"
`
	profile, err := getProfile("yaml", testConfig, "profile", "")
	assert.Nil(t, err)
	assert.NotNil(t, profile)

	buffer := &bytes.Buffer{}
	assert.Nil(t, ShowStruct(buffer, profile, "p"))

	result := regexp.MustCompile("\\s+").ReplaceAllString(buffer.String(), " ")
	result = strings.TrimSpace(result)

	assert.Contains(t, result, "my_value: val")
	assert.Contains(t, result, "my_url: "+defaultUrlReplaced)
	assert.Contains(t, result, "my_key: "+ConfidentialReplacement)
	assert.Contains(t, result, "my_token: "+ConfidentialReplacement)
	assert.Contains(t, result, "my_password: "+ConfidentialReplacement)
	assert.Contains(t, result, "repository: "+defaultUrlReplaced)
}

func TestGetNonConfidentialValues(t *testing.T) {
	testConfig := `
profile:
  verbose: false
  repository: "local:user:pass@host"
  env:
    MY_PASSWORD: "otherval"
  backup:
    send-after:
      url: "http://huser:hpw@host"
      headers: { "name": "Authorization", value: "Basic" }
`
	profile, err := getProfile("yaml", testConfig, "profile", "")
	assert.Nil(t, err)
	assert.NotNil(t, profile)

	result := GetNonConfidentialValues(profile, []string{
		"a",
		defaultUrl,
		"b",
		"otherval",
		"c",
		defaultHttpUrl,
		"d",
		"Basic",
		"e",
	})
	assert.Equal(t, []string{
		"a",
		defaultUrlReplaced,
		"b",
		ConfidentialReplacement,
		"c",
		defaultHttpUrlReplaced,
		"d",
		ConfidentialReplacement,
		"e",
	}, result)
}

func TestGetNonConfidentialArgs(t *testing.T) {
	repo := "local:user:%s@host/path with space"
	testConfig := `
global:
  prevent-auto-repository-file: true
profile:
  repository: "` + fmt.Sprintf(repo, "password") + `"
`
	profile, err := getProfile("yaml", testConfig, "profile", "")
	assert.NoError(t, err)
	assert.NotNil(t, profile)

	args := profile.GetCommandFlags(constants.CommandBackup)
	result := GetNonConfidentialArgs(profile, args)

	expectedSecret := shell.NewArg(fmt.Sprintf(repo, "password"), shell.ArgConfigEscape).String()
	expectedPublic := shell.NewArg(fmt.Sprintf(repo, ConfidentialReplacement), shell.ArgConfigEscape).String()

	assert.Equal(t, []string{"--repo=" + expectedSecret}, args.GetAll())
	assert.Equal(t, []string{"--repo=" + expectedPublic}, result.GetAll())
}

func TestGetNonConfidentialArgsFromEnvironmentVariable(t *testing.T) {
	repo := "local:user:%s@host/path with space"
	environment := []string{
		"MY_REPOSITORY=" + fmt.Sprintf(repo, "password"),
	}
	testConfig := `
      global:
        prevent-auto-repository-file: true
      profile:
        repository: $MY_REPOSITORY
      `
	profile, err := getProfile("yaml", testConfig, "profile", "")
	assert.NoError(t, err)
	assert.NotNil(t, profile)

	args := profile.GetCommandFlags(constants.CommandBackup)
	args = args.Modify(shell.NewExpandEnvModifier(environment))
	result := GetNonConfidentialArgs(profile, args)

	expectedSecret := shell.NewArg(fmt.Sprintf(repo, "password"), shell.ArgConfigEscape).String()
	expectedPublic := shell.NewArg(fmt.Sprintf(repo, ConfidentialReplacement), shell.ArgConfigEscape).String()

	assert.Equal(t, []string{"--repo=" + expectedSecret}, args.GetAll())
	assert.Equal(t, []string{"--repo=" + expectedPublic}, result.GetAll())
}

func TestGetAutoRepositoryFile(t *testing.T) {
	require.NoError(t, os.Unsetenv("RESTIC_REPOSITORY"))
	require.NoError(t, os.Unsetenv("RESTIC_REPOSITORY_FILE"))

	tests := map[string]bool{
		"/path/to/file":                      false,
		"file:/path/to/file":                 false,
		"local:user:pw@host/path with space": true,
		"https://public/":                    false,
		"https://user:password@private/":     true,
	}

	for repo, usesFile := range tests {
		t.Run(repo, func(t *testing.T) {
			config := fmt.Sprintf(`
				[my-profile]
				repository = %q
            `, repo)
			profile, err := getResolvedProfile("toml", config, "my-profile")
			require.NoError(t, err)
			args := profile.GetCommandFlags(constants.CommandBackup)

			if usesFile && platform.IsWindows() {
				usesFile = false // Windows has no support for repository file right now (as temp files can't be forced to private with standard go API)
			}

			if usesFile {
				file := templates.TempFile("my-profile-repo.txt")
				expected := shell.NewArg(file, shell.ArgConfigEscape).String()
				assert.Equal(t, []string{"--repository-file=" + expected}, args.GetAll())

				content, err := os.ReadFile(file)
				assert.NoError(t, err)
				assert.Equal(t, repo, string(content))
			} else {
				expected := shell.NewArg(repo, shell.ArgConfigEscape).String()
				assert.Equal(t, []string{"--repo=" + expected}, args.GetAll())
			}
		})
	}
}

func TestGetAutoRepositoryFileDisabledWithEnv(t *testing.T) {
	if platform.IsWindows() {
		t.Skip()
	}

	tests := []string{
		"RESTIC_REPOSITORY",
		"RESTIC_REPOSITORY_FILE",
	}

	for _, envKey := range tests {
		t.Run(envKey, func(t *testing.T) {
			config := fmt.Sprintf(`
				[my-profile]
				repository = %q
            `, defaultHttpUrl)

			defer os.Unsetenv(envKey)
			profile, err := getResolvedProfile("toml", config, "my-profile")
			require.NoError(t, err)

			hasRepoFlag := func() (found bool) {
				_, found = profile.GetCommandFlags(constants.CommandBackup).Get("repo")
				return
			}

			t.Run("no-env", func(t *testing.T) {
				require.NoError(t, os.Unsetenv(envKey))
				profile.Environment = nil
				assert.False(t, hasRepoFlag())
			})

			t.Run("profile-env", func(t *testing.T) {
				require.NoError(t, os.Unsetenv(envKey))
				profile.Environment = map[string]ConfidentialValue{
					envKey: NewConfidentialValue("1"),
				}
				assert.True(t, hasRepoFlag())
			})

			t.Run("os-env", func(t *testing.T) {
				require.NoError(t, os.Setenv(envKey, "1"))
				profile.Environment = nil
				assert.True(t, hasRepoFlag())
			})
		})
	}
}

func TestConfidentialToJSON(t *testing.T) {
	t.Run("marshal", func(t *testing.T) {
		value := NewConfidentialValue("plain")
		assert.False(t, value.IsConfidential())

		binary, _ := json.Marshal(value)
		assert.Equal(t, `"plain"`, string(binary))

		value.hideValue()
		assert.True(t, value.IsConfidential())

		binary, _ = json.Marshal(value)
		assert.Equal(t, `"plain"`, string(binary))
	})

	t.Run("unmarshal", func(t *testing.T) {
		value := NewConfidentialValue("")
		value.hideValue()
		assert.True(t, value.IsConfidential())

		assert.NoError(t, json.Unmarshal([]byte(`"plain"`), &value))

		// the confidential state is not marshalled for now
		assert.False(t, value.IsConfidential())
		assert.Equal(t, "plain", value.Value())
		assert.Equal(t, "plain", value.String())
	})
}

package util

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/encoding/unicode"
)

func TestGetEnvironmentFile(t *testing.T) {
	dotenv := filepath.Join(t.TempDir(), ".env")
	env := map[string]string{
		"K1": "v1",
		"K2": "v2",
		"k3": "hËllØ",
	}
	assert.NoError(t, godotenv.Write(env, dotenv))

	// initial check
	ef, err := GetEnvironmentFile(dotenv)
	require.NoError(t, err)
	assert.Equal(t, dotenv, ef.Name())
	assert.Equal(t, env, ef.ValuesAsMap())

	ef2, _ := GetEnvironmentFile(dotenv)
	assert.Same(t, ef, ef2)

	// change K2
	env["K2"] = "updated"
	env2 := ef.ValuesAsMap()

	// check validity before updating the file
	assert.True(t, ef.Valid())
	assert.NotEqual(t, env, env2)

	// update the file
	assert.NoError(t, godotenv.Write(env, dotenv))
	assert.False(t, ef.Valid())
	assert.Equal(t, env2, ef.ValuesAsMap(), "loaded values must be immutable")

	// reload & check again
	ef2, err = GetEnvironmentFile(dotenv)
	require.NoError(t, err)
	assert.Equal(t, env, ef2.ValuesAsMap())
	assert.NotSame(t, ef, ef2)
}

func TestGetEnvironmentFileErrors(t *testing.T) {
	// error on dir
	dir := t.TempDir()
	_, err := GetEnvironmentFile(dir)
	assert.ErrorContains(t, err, fmt.Sprintf("%s is not valid", dir))

	// error on missing file
	dotenv := filepath.Join(t.TempDir(), ".env")
	_, err = GetEnvironmentFile(dotenv)
	assert.ErrorIs(t, err, fs.ErrNotExist)

	// no error for empty file
	require.NoError(t, os.WriteFile(dotenv, []byte(""), 0600))
	ef, err := GetEnvironmentFile(dotenv)
	require.NoError(t, err)

	// error for double init
	assert.ErrorContains(t, ef.init("n"), fmt.Sprintf("illegal state, file %s already loaded", dotenv))

	// error for invalid content
	require.NoError(t, os.WriteFile(dotenv, []byte(" + _ ."), 0700))
	_, err = GetEnvironmentFile(dotenv)
	assert.ErrorContains(t, err, `unexpected character "+" in variable name near "+ _ ."`)
}

func TestGetEnvironmentFileUTF16(t *testing.T) {
	dotenv := filepath.Join(t.TempDir(), ".env")
	env := map[string]string{"key": "hËllØ"}

	// test loading UTF-16 (default encoding for powershell)
	func() {
		file, err := os.OpenFile(dotenv, os.O_RDWR|os.O_CREATE, 0600)
		require.NoError(t, err)
		defer func() { require.NoError(t, file.Close()) }()
		content, _ := godotenv.Marshal(env)
		_, err = unicode.
			UTF16(unicode.LittleEndian, unicode.UseBOM).NewEncoder().
			Writer(file).
			Write([]byte(content))
		require.NoError(t, err)
	}()

	ef, err := GetEnvironmentFile(dotenv)
	require.NoError(t, err)
	assert.Equal(t, env, ef.ValuesAsMap())
}

func TestGetEnvironmentFileAddToEnvironment(t *testing.T) {
	dotenv := filepath.Join(t.TempDir(), ".env")
	env := map[string]string{
		"K1": "v1",
		"K2": "v2",
		"k3": "hËllØ",
	}
	assert.NoError(t, godotenv.Write(env, dotenv))

	// check with folding env to see if names are resolved prior to adding (see case change in keys)
	environment := NewFoldingEnvironment("K3=__v3", "k1=__v1", "K4=V4")
	ef, err := GetEnvironmentFile(dotenv)
	require.NoError(t, err)
	ef.AddTo(environment)

	assert.Equal(t, map[string]string{
		"k1": "v1",
		"K2": "v2",
		"K3": "hËllØ",
		"K4": "V4",
	}, environment.ValuesAsMap())
}

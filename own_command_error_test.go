package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOwnCommandError(t *testing.T) {
	var wrap error = errors.New("wrap")
	var err error = newOwnCommandError(wrap, 10)

	assert.Equal(t, "wrap", err.Error())
	assert.ErrorIs(t, err, wrap)

	var unwrap *ownCommandError
	assert.ErrorAs(t, err, &unwrap)
	assert.Equal(t, 10, unwrap.ExitCode())
}

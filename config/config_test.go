package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {
	is := assert.New(t)

	var expected = &Config{
		Name:      "echo",
		Directory: "/bin",
		Command:   `echo "this is a test"`,
		UID:       7000,
		GID:       7000,
		Env:       []string{"FOO=bar", "BIF=baz"},
	}

	actual, er := Read("examples/escarole.yaml")
	is.NoError(er)
	is.EqualValues(expected, actual)
}

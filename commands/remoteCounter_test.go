package commands

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseCsvAsArray(t *testing.T) {
	assert.Equal(t, parseCsvAsArray("foo,bar"), []string{"foo", "bar"})
}

func TestParseStringAsTimestamp_whenInvalidInput_fail(t *testing.T) {
	_, err := parseStringAsTimestamp("foo,bar")

	assert.NotNil(t, err)
}

func TestParseStringAsTimestamp_whenValidShortInput_success(t *testing.T) {
	expected, _ := time.Parse(time.RFC3339Nano, "2006-01-02T00:00:00.000000000Z")

	t1, err := parseStringAsTimestamp("2006-01-02")

	assert.Nil(t, err)
	assert.Equal(t, t1, expected)
}

func TestParseStringAsTimestamp_whenValidLongInput_success(t *testing.T) {
	expected, _ := time.Parse(time.RFC3339Nano, "2006-01-02T00:00:00.000000000Z")

	t1, err := parseStringAsTimestamp("2006-01-02T00:00:00")

	assert.Nil(t, err)
	assert.Equal(t, t1, expected)
}

func TestGetCurrentUser_whenStar_success(t *testing.T) {
	assert.Equal(t, getCurrentUser("*"), "ALL_USERS")
}

func TestGetCurrentUser_whenName_success(t *testing.T) {
	assert.Equal(t, getCurrentUser("foo"), "foo")
}

package repository

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestGetUpstreamPath(t *testing.T) {
	assert.Equal(t, getUpstreamPath("dummy1", "/var/lib/lagoon"), "/var/lib/lagoon/upstream/dummy1/")
}

func TestGetStagingPath(t *testing.T) {
	assert.Equal(t, getStagingPath("dummy1", "/var/lib/lagoon"), "/var/lib/lagoon/staging/dummy1/")
}

func TestGetPublicPath(t *testing.T) {
	assert.Equal(t, getPublicPath("dummy1", "/var/lib/lagoon"), "/var/lib/lagoon/public/dummy1/")
}

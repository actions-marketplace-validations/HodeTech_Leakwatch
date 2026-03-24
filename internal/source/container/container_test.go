package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainerSource_Type_ReturnsContainer(t *testing.T) {
	s := New("nginx:latest")
	assert.Equal(t, "container", s.Type())
}

func TestContainerSource_Validate_ValidRef_ReturnsNoError(t *testing.T) {
	s := New("nginx:latest")
	assert.NoError(t, s.Validate())
}

func TestContainerSource_Validate_InvalidRef_ReturnsError(t *testing.T) {
	s := New(":::invalid")
	assert.Error(t, s.Validate())
}

func TestContainerSource_Validate_FullRef_ReturnsNoError(t *testing.T) {
	s := New("ghcr.io/org/repo:v1.0.0")
	assert.NoError(t, s.Validate())
}

func TestShouldSkipContainerPath_DocPath_ReturnsTrue(t *testing.T) {
	assert.True(t, shouldSkipContainerPath("/usr/share/doc/something"))
}

func TestShouldSkipContainerPath_ManPath_ReturnsTrue(t *testing.T) {
	assert.True(t, shouldSkipContainerPath("/usr/share/man/man1/ls.1"))
}

func TestShouldSkipContainerPath_AppFile_ReturnsFalse(t *testing.T) {
	assert.False(t, shouldSkipContainerPath("/app/config.yaml"))
}

func TestShouldSkipContainerPath_EtcFile_ReturnsFalse(t *testing.T) {
	assert.False(t, shouldSkipContainerPath("/etc/environment"))
}

func TestShouldSkipContainerPath_RootFile_ReturnsFalse(t *testing.T) {
	assert.False(t, shouldSkipContainerPath("app.conf"))
}

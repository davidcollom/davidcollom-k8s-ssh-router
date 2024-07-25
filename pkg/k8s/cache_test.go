package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCacheOperations(t *testing.T) {
	username := "default-testuser"
	secretData := map[string]string{
		"password": "testpassword",
	}

	// Test setting secret in cache
	SetSecretInCache(username, secretData)
	cachedSecret, found := GetSecretFromCache(username)
	assert.True(t, found, "Secret should be found in cache")
	assert.Equal(t, secretData, cachedSecret, "Secret data should match")

	// Test deleting secret from cache
	DeleteSecretFromCache(username)
	_, found = GetSecretFromCache(username)
	assert.False(t, found, "Secret should not be found in cache after deletion")
}

package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewOrgRateLimiter()
	for i := 0; i < 10; i++ {
		assert.True(t, rl.Allow(1), "request %d should be allowed", i+1)
	}
}

func TestRateLimiter_Deny(t *testing.T) {
	rl := NewOrgRateLimiter()
	for i := 0; i < 10; i++ {
		rl.Allow(1)
	}
	assert.False(t, rl.Allow(1), "11th request should be denied")
}

func TestRateLimiter_SeparateOrgs(t *testing.T) {
	rl := NewOrgRateLimiter()
	for i := 0; i < 10; i++ {
		rl.Allow(1)
	}
	// Different org should have its own bucket
	assert.True(t, rl.Allow(2))
}

package ai

import (
	"sync"
	"time"
)

const maxRequestsPerMinute = 10

type orgBucket struct {
	count   int
	resetAt time.Time
}

// OrgRateLimiter limits AI generation requests per organization.
type OrgRateLimiter struct {
	mu      sync.Mutex
	buckets map[int64]*orgBucket
}

// NewOrgRateLimiter creates a new OrgRateLimiter.
func NewOrgRateLimiter() *OrgRateLimiter {
	return &OrgRateLimiter{buckets: make(map[int64]*orgBucket)}
}

// Allow returns true if the org has not exceeded the rate limit.
func (r *OrgRateLimiter) Allow(orgID int64) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	b, ok := r.buckets[orgID]
	if !ok || now.After(b.resetAt) {
		r.buckets[orgID] = &orgBucket{count: 1, resetAt: now.Add(time.Minute)}
		return true
	}
	if b.count >= maxRequestsPerMinute {
		return false
	}
	b.count++
	return true
}

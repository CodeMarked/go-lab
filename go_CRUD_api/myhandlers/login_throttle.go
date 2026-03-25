package myhandlers

import (
	"sync"
	"time"
)

// In-memory per-email lockout after repeated failed password attempts (same process only).
// Does not apply to unknown-email failures (avoid account DoS). Tuned via constants below.

const (
	loginMaxFailuresPerEmail = 5
	loginLockoutDuration     = 15 * time.Minute
)

type loginEmailState struct {
	mu          sync.Mutex
	failures    int
	lockedUntil time.Time
}

var loginEmailThrottle sync.Map // email -> *loginEmailState

func loginEmailLocked(email string) (locked bool, retryAfter time.Duration) {
	v, ok := loginEmailThrottle.Load(email)
	if !ok {
		return false, 0
	}
	st := v.(*loginEmailState)
	st.mu.Lock()
	defer st.mu.Unlock()
	now := time.Now()
	if now.Before(st.lockedUntil) {
		return true, st.lockedUntil.Sub(now)
	}
	return false, 0
}

func recordLoginFailureKnownUser(email string) {
	v, _ := loginEmailThrottle.LoadOrStore(email, &loginEmailState{})
	st := v.(*loginEmailState)
	st.mu.Lock()
	defer st.mu.Unlock()
	now := time.Now()
	if now.Before(st.lockedUntil) {
		return
	}
	st.failures++
	if st.failures >= loginMaxFailuresPerEmail {
		st.lockedUntil = now.Add(loginLockoutDuration)
		st.failures = 0
	}
}

func recordLoginSuccessClearThrottle(email string) {
	loginEmailThrottle.Delete(email)
}

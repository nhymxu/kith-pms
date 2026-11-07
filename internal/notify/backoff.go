package notify

import "time"

const maxAttempts = 5

var retryBackoff = [...]time.Duration{time.Minute, 5 * time.Minute, 30 * time.Minute, 2 * time.Hour}

// Backoff returns the wait before the next try. attempts is the number of tries
// already made (>=1). A server-provided Retry-After wins over the schedule.
func Backoff(attempts int, se *SendError) time.Duration {
	if se != nil && se.RetryAfter > 0 {
		return se.RetryAfter
	}

	return retryBackoff[min(max(attempts-1, 0), len(retryBackoff)-1)]
}

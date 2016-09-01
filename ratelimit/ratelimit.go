package ratelimit

import (
	"sync"
	"time"
)

// empty struct doesn't use memory
// good for using in channels
type signal struct{}

var hit signal

// Limiter is an interface for generic rate limiters
// Implementations usually block an id based on various
// parameters.
type Limiter interface {
	Limit(id string)
}

// GroupLimiter uses separete Limiter (with separate parameters)
// for each groupID
type GroupLimiter interface {
	SetGroup(expiration time.Duration, groupID string, maxRPS, burst int)
	DelGroup(groupID string)
	Limit(groupID, id string)
}

type groupClientLimiter struct {
	sync.RWMutex
	groups map[string]Limiter
}

// ClientLimiter is a limiter based on any unique token or ID.
type clientLimiter struct {
	sync.RWMutex
	clients    map[string]chan signal
	wait       time.Duration
	expiration time.Duration
	burst      int
}

// NewGroupLimiter returns a limiter which can assign different
// limits to specified group IDs.
// E.g. Admins can have a higher rate than public users
// The inital Grouplimiter does not contain groups.
func NewGroupLimiter() GroupLimiter {
	return &groupClientLimiter{
		groups: map[string]Limiter{},
	}
}

// SetGroup creates a new group on the limiter.
func (gcl *groupClientLimiter) SetGroup(expiration time.Duration, groupID string, maxRPS, burst int) {
	gcl.Lock()
	gcl.groups[groupID] = NewClientLimiter(expiration, maxRPS, burst) // TODO: expose
	gcl.Unlock()
}

// DelGroup removes a group from the limiter.
func (gcl *groupClientLimiter) DelGroup(groupID string) {
	gcl.Lock()
	delete(gcl.groups, groupID)
	gcl.Unlock()
}

// Limit blocks for a calculated time period (based on maxRPS)
// if a groupID exists and the given id reached it's limits.
func (gcl *groupClientLimiter) Limit(groupID, id string) {
	gcl.RLock()
	limiter, ok := gcl.groups[groupID]
	gcl.RUnlock()

	if ok {
		limiter.Limit(id)
	}
}

// NewClientLimiter returs a limiter which enforces a common rate
// limit individually for each client.
// expiration specifies how long a client is being 'tracked' allowing
// burst for new clients. burst is max number of burst request for
// new clients. rps is max sustained requests per second.
func NewClientLimiter(expiration time.Duration, maxRPS, burst int) Limiter {

	wait := time.Duration(int64(1.0/float64(maxRPS)*1000)) * time.Millisecond
	return &clientLimiter{
		clients:    map[string]chan signal{},
		wait:       wait,
		expiration: expiration,
		burst:      burst,
	}
}

// Limit only blocks for clients reaching their limits.
// It spawns an expiring goroutine for each unique client.
func (cl *clientLimiter) Limit(id string) {
	cl.RLock()
	bucket, ok := cl.clients[id]
	cl.RUnlock()
	if ok {
		<-bucket
		return
	}

	// bucket allowing burst requests
	bucket = make(chan signal, cl.burst)

	go func() {
		cl.Lock()
		cl.clients[id] = bucket
		cl.Unlock()

		for {
			select {
			case bucket <- hit:
				time.Sleep(cl.wait)

			case <-time.After(cl.expiration):

				// bucket is full and no client is draining it
				// remove client and return to free up resources
				// (will be recreated on a new request)

				cl.Lock()
				delete(cl.clients, id)
				cl.Unlock()
				return
			}
		}
	}()

	select {
	case <-bucket:
	case <-time.After(cl.expiration): // TODO: this might be unnecessary, but make sure channel won't block forever
	}
}

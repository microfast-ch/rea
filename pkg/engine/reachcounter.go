package engine

import (
	"sync"
)

// reachcounter implements a counter for string occurrences.
type reachcounter struct {
	c map[string]uint
	sync.Mutex
}

// newReachCounter returns a new/empty reachcounter.
func newReachCounter() *reachcounter {
	return &reachcounter{
		c: map[string]uint{},
	}
}

// Adds the key to the counter and returns the previous number of occurrences.
// This will be zero if there was no occurrence.
func (rc *reachcounter) Add(key string) uint {
	rc.Lock()
	defer rc.Unlock()

	// Go spec:
	//   if the map is nil or does not contain such an entry,
	//   a[x] is the zero value for the element type of M
	//
	// therefor i is zero when it isn't found

	i, ok := rc.c[key]
	if !ok {
		rc.c[key] = 1
	} else {
		rc.c[key] = i + 1
	}

	return i
}

// Clean clears all items from the reachcounter.
func (rc *reachcounter) Clean() {
	rc.Lock()
	rc.c = map[string]uint{}
	rc.Unlock()
}

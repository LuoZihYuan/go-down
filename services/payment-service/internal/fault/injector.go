package fault

import (
	"sync"
	"time"
)

// Injector manages fault injection state
type Injector struct {
	enabled      bool
	delaySeconds int
	mu           sync.RWMutex
}

// NewInjector creates a new fault injector
func NewInjector() *Injector {
	return &Injector{
		enabled:      false,
		delaySeconds: 0,
	}
}

// Enable activates fault injection with specified delay
func (i *Injector) Enable(delaySeconds int) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.enabled = true
	i.delaySeconds = delaySeconds
}

// Disable deactivates fault injection
func (i *Injector) Disable() {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.enabled = false
	i.delaySeconds = 0
}

// IsEnabled returns whether fault injection is active
func (i *Injector) IsEnabled() bool {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.enabled
}

// GetStatus returns current fault injection configuration
func (i *Injector) GetStatus() (bool, int) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.enabled, i.delaySeconds
}

// Inject applies fault injection if enabled
// This blocks for the configured delay duration
func (i *Injector) Inject() {
	i.mu.RLock()
	enabled := i.enabled
	delay := i.delaySeconds
	i.mu.RUnlock()

	if enabled && delay > 0 {
		time.Sleep(time.Duration(delay) * time.Second)
	}
}

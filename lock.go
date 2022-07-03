// Package resource provides a helper to synchronize access to a shared resources from concurrent scenarios.
package resource

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/cucumber/godog"
)

type sentinelError string

// Error returns the error message.
func (e sentinelError) Error() string {
	return string(e)
}

// ErrMissingScenarioLock is a sentinel error.
const ErrMissingScenarioLock = sentinelError("missing scenario lock key in context")

// Lock keeps exclusive access to the scenario steps.
type Lock struct {
	mu        sync.Mutex
	locks     map[string]chan struct{}
	onRelease func(lockName string) error
	ctxKey    *struct{ _ int }
}

// NewLock creates a new Lock.
func NewLock(onRelease func(name string) error) *Lock {
	return &Lock{
		locks:     make(map[string]chan struct{}),
		onRelease: onRelease,
		ctxKey:    new(struct{ _ int }),
	}
}

// Acquire acquires resource lock for the given key and returns true.
//
// If the lock is already held by another context, it waits for the lock to be released.
// It returns false is the lock is already held by this context.
// This function fails if the context is missing current lock.
func (s *Lock) Acquire(ctx context.Context, name string) (bool, error) {
	currentLock, ok := ctx.Value(s.ctxKey).(chan struct{})
	if !ok {
		return false, ErrMissingScenarioLock
	}

	s.mu.Lock()
	lock := s.locks[name]

	if lock == nil {
		s.locks[name] = currentLock
	}

	s.mu.Unlock()

	// Wait for the alien lock to be released.
	if lock != nil && lock != currentLock {
		<-lock

		return s.Acquire(ctx, name)
	}

	if lock == nil {
		return true, nil
	}

	return false, nil
}

// Register adds hooks to scenario context.
func (s *Lock) Register(sc *godog.ScenarioContext) {
	sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		lock := make(chan struct{})

		// Adding unique pointer to context to avoid collisions.
		return context.WithValue(ctx, s.ctxKey, lock), nil
	})

	sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		s.mu.Lock()
		defer s.mu.Unlock()

		// Releasing locks owned by scenario.
		currentLock, ok := ctx.Value(s.ctxKey).(chan struct{})
		if !ok {
			return ctx, ErrMissingScenarioLock
		}

		var errs []string

		for name, lock := range s.locks {
			if lock == currentLock {
				delete(s.locks, name)
			}

			if s.onRelease != nil {
				if err := s.onRelease(name); err != nil {
					errs = append(errs, err.Error())
				}
			}
		}

		// Godog v0.12.5 has an issue of calling after scenario multiple times when there are undefined steps.
		// This is a workaround.
		closeIfNot := func() {
			defer func() {
				_ = recover() // nolint: errcheck // Only close of the closed channel can panic here.
			}()
			close(currentLock)
		}

		closeIfNot()

		if len(errs) > 0 {
			return ctx, errors.New(strings.Join(errs, ", ")) // nolint:goerr113
		}

		return ctx, nil
	})
}

// IsLocked is true if resource is currently locked for another scenario.
func (s *Lock) IsLocked(ctx context.Context, name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	lock := s.locks[name]

	return lock != nil && lock != ctx.Value(s.ctxKey).(chan struct{})
}

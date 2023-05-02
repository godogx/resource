package resource_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/godogx/resource"
	"github.com/stretchr/testify/assert"
)

type lock struct {
	lock *resource.Lock
}

func (l *lock) setup(s *godog.ScenarioContext) {
	l.lock.Register(s)

	s.Step(`^I acquire "([^"]*)"$`, func(ctx context.Context, name string) (context.Context, error) {
		ok, err := l.lock.Acquire(ctx, name)
		if err != nil {
			return ctx, err
		}

		if !ok {
			return ctx, fmt.Errorf("failed to acquire lock")
		}

		return ctx, nil
	})
	s.Step(`^I should be blocked for "([^"]*)"$`, func(ctx context.Context, name string) error {
		if !l.lock.IsLocked(ctx, name) {
			return fmt.Errorf("%s is not locked", name)
		}

		return nil
	})
	s.Step(`^I should not be blocked for "([^"]*)"$`, func(ctx context.Context, name string) error {
		if l.lock.IsLocked(ctx, name) {
			return fmt.Errorf("%s is locked", name)
		}

		return nil
	})
	s.Step("^I sleep$", func() {
		time.Sleep(time.Microsecond * time.Duration(rand.Int63n(10000)+1000)) //nolint:gosec
	})
	s.Step("^I sleep longer$", func() {
		time.Sleep(time.Millisecond * 100)
	})
}

func TestNewLock(t *testing.T) {
	l := &lock{lock: resource.NewLock(nil)}
	out := bytes.Buffer{}

	suite := godog.TestSuite{
		ScenarioInitializer: l.setup,
		Options: &godog.Options{
			Output:      &out,
			Format:      "pretty",
			Strict:      true,
			Paths:       []string{"_testdata/NoBlock.feature"},
			Concurrency: 10,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("test failed")
	}
}

func TestNewLock_blocked(t *testing.T) {
	l := &lock{lock: resource.NewLock(nil)}
	out := bytes.Buffer{}

	suite := godog.TestSuite{
		ScenarioInitializer: l.setup,
		Options: &godog.Options{
			Output:      &out,
			Format:      "pretty",
			Strict:      true,
			Paths:       []string{"_testdata/Block.feature"},
			Concurrency: 10,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("test failed", out.String())
	}
}

func TestNewLock_failOnRelease(t *testing.T) {
	l := &lock{lock: resource.NewLock(func(name string) error {
		return errors.New("failed")
	})}
	out := bytes.Buffer{}

	suite := godog.TestSuite{
		ScenarioInitializer: l.setup,
		Options: &godog.Options{
			Output:      &out,
			Format:      "pretty",
			Strict:      true,
			Paths:       []string{"_testdata/Block.feature"},
			Concurrency: 10,
		},
	}

	if suite.Run() != 1 {
		t.Fatal("test failed", out.String())
	}
}

func TestLock_Acquire(t *testing.T) {
	l := resource.NewLock(nil)
	_, err := l.Acquire(context.Background(), "test")
	assert.EqualError(t, err, resource.ErrMissingScenarioLock.Error())
}

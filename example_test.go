package resource_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cucumber/godog"
	"github.com/godogx/resource"
)

// StorageSteps manages an example global external storage.
type storageSteps struct {
	lock *resource.Lock
}

func newStorageSteps() *storageSteps {
	return &storageSteps{
		lock: resource.NewLock(func(path string) error {
			// The onRelease hook can be used to clean up the state,
			// for example to delete files that were acquired during the scenario.
			return nil
		}),
	}
}

func (s *storageSteps) acquireFile(ctx context.Context, path string) error {
	ok, err := s.lock.Acquire(ctx, path)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("could not acquire file for scenario: %s", path)
	}

	return nil
}

func (s *storageSteps) Register(sc *godog.ScenarioContext) {
	s.lock.Register(sc)

	sc.Step(`write file "([^"])" with contents`, func(ctx context.Context, path string, contents string) error {
		// By acquiring the lock, we ensure no other scenario can read or write this file
		// while this scenario is running.

		// Since we're using path as resource name, scenarios that depend on other paths
		// won't be blocked by this lock.
		if err := s.acquireFile(ctx, path); err != nil {
			return err
		}
		// Write contents to file here.
		return nil
	})

	sc.Step(`file "([^"])" has contents`, func(ctx context.Context, path string, contents string) error {
		// If the lock is acquired within the same scenario, there is no block.
		if err := s.acquireFile(ctx, path); err != nil {
			return err
		}
		// Read and compare contents of the file here.
		return nil
	})
}

func ExampleNewLock() {
	ss := newStorageSteps()
	suite := godog.TestSuite{
		ScenarioInitializer: ss.Register,
		Options: &godog.Options{
			Format:      "pretty",
			Paths:       []string{"features"},
			Strict:      true,
			Randomize:   time.Now().UTC().UnixNano(),
			Concurrency: 10,
		},
	}
	status := suite.Run()

	if status != 0 {
		log.Fatal("test failed")
	}
}

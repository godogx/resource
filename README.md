# Godog Resource Lock

[![Build Status](https://github.com/godogx/resource/workflows/test-unit/badge.svg)](https://github.com/godogx/resource/actions?query=branch%3Amaster+workflow%3Atest-unit)
[![Coverage Status](https://codecov.io/gh/godogx/resource/branch/master/graph/badge.svg)](https://codecov.io/gh/godogx/resource)
[![GoDevDoc](https://img.shields.io/badge/dev-doc-00ADD8?logo=go)](https://pkg.go.dev/github.com/godogx/resource)
[![Time Tracker](https://wakatime.com/badge/github/godogx/resource.svg)](https://wakatime.com/badge/github/godogx/resource)
![Code lines](https://sloc.xyz/github/godogx/resource/?category=code)
![Comments](https://sloc.xyz/github/godogx/resource/?category=comments)

This library provides a simple way to manage sequential access to global resources in your `godog` tests.

This is useful if `godog` suite is running with 
[concurrency](https://pkg.go.dev/github.com/cucumber/godog/internal/flags#Options) 
option greater than 1.

## Usage

Add `*resource.Lock` to your resource manager.
```go
// StorageSteps manages an example global external storage.
type storageSteps struct {
	lock *resource.Lock
}
```

Register `*resource.Lock` hooks to scenario context.
```go
func (s *storageSteps) Register(sc *godog.ScenarioContext) {
	s.lock.Register(sc)
...
```

Upgrade your step definitions to receive context if you don't have it already.
```go
sc.Step(`write file "([^"])" with contents`, func(ctx context.Context, path string, contents string) error {
```

Acquire resource before using it in scenario step. This will block all other scenarios that will try to acquire this 
resource to wait until current scenario is finished.
```go
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
```

See complete instrumentation [example](./example_test.go).
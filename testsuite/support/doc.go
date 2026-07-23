// Package support provides integration-test-only helpers for package-level Go tests.
//
// Its public helper APIs are built with the integration tag and must only be
// imported by *_test.go files built with that tag. Production code must not
// import this package.
package support

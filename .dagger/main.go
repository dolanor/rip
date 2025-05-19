// A generated module for Rip functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/rip/internal/dagger"
)

type Rip struct{}

// Returns the std out of a go test
func (m *Rip) Test(
	ctx context.Context,

	// +defaultPath=/
	source *dagger.Directory,
) (string, error) {
	return dag.Container().
		From("golang:1.24.3").
		WithDirectory("/src/rip", source).
		WithWorkdir("/src/rip").
		WithExec([]string{"go", "test", "-v", "./..."}).
		Stdout(ctx)
}

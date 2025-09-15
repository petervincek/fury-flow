package teststartupmodel

import (
	"github.com/petervincek/fury-flow/app"
	"github.com/testcontainers/testcontainers-go"
)

// TestInfra encapsulates the infrastructure required for functional tests,
// including a reference to the FuryFlow application instance and any test containers
// used during the test lifecycle.
type TestInfra struct {
	FuryFlowApp *app.FuryFlow
	Containers  []testcontainers.Container
}

// New creates and returns a new instance of TestInfra, initializing it with the provided
// FuryFlow application and a slice of testcontainers.Container.
// Parameters:
//   - ffa: pointer to the FuryFlow application instance.
//   - containers: slice of testcontainers.Container to be managed.
//
// Returns:
//   - pointer to the initialized TestInfra struct.
func New(ffa *app.FuryFlow, containers []testcontainers.Container) *TestInfra {
	return &TestInfra{
		FuryFlowApp: ffa,
		Containers:  containers,
	}
}

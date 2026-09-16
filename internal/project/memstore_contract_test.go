package project_test

import (
	"testing"

	"github.com/PoCInnovation/SecureEnv/internal/project"
	"github.com/PoCInnovation/SecureEnv/internal/project/storetest"
)

func TestMemStoreContract(t *testing.T) {
	storetest.Run(t, func(*testing.T) project.Store { return newMemStore() })
}

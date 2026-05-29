package client

import (
	"os"
	"path/filepath"
	"testing"
)

/*
@function TestContractLoaderLoadABI

@desc
Tests loading ABI JSON from a Foundry artifact fixture.
*/
func TestContractLoaderLoadABI(t *testing.T) {
	root := t.TempDir()
	artifactDir := filepath.Join(root, "RevvFiMarket.sol")
	if err := os.MkdirAll(artifactDir, 0o755); err != nil {
		t.Fatalf("create artifact dir: %v", err)
	}
	artifact := `{"abi":[{"type":"event","name":"Borrow"}]}`
	if err := os.WriteFile(filepath.Join(artifactDir, "RevvFiMarket.json"), []byte(artifact), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	loader := NewContractLoader(root)
	abi, err := loader.LoadABI("RevvFiMarket.sol", "RevvFiMarket")
	if err != nil {
		t.Fatalf("load abi: %v", err)
	}
	if len(abi) == 0 {
		t.Fatal("expected abi payload")
	}
}

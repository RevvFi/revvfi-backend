package client

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

/*
@struct Artifact

@desc
Subset of a Foundry artifact required by backend ABI loading.

@responsibilities
- Decode contract ABI JSON
- Avoid coupling callers to full artifact metadata
*/
type Artifact struct {
	ABI json.RawMessage `json:"abi"`
}

/*
@struct ContractLoader

@desc
Loads contract ABIs from Foundry artifact directories.

@responsibilities
- Resolve artifact paths
- Parse ABI payloads
- Return ABI JSON to event and transaction helpers
*/
type ContractLoader struct {
	artifactPath string
}

/*
@function NewContractLoader

@desc
Creates a contract artifact loader.

@params
- artifactPath: root Foundry out directory

@returns
- *ContractLoader
*/
func NewContractLoader(artifactPath string) *ContractLoader {
	return &ContractLoader{artifactPath: artifactPath}
}

/*
@method LoadABI

@desc
Loads ABI JSON for a compiled Solidity contract.

@params
- contractFile: Solidity file name directory
- contractName: contract artifact name

@returns
- json.RawMessage
- error
*/
func (l *ContractLoader) LoadABI(contractFile string, contractName string) (json.RawMessage, error) {
	path := filepath.Join(l.artifactPath, contractFile, contractName+".json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read artifact %s: %w", path, err)
	}
	var artifact Artifact
	if err := json.Unmarshal(raw, &artifact); err != nil {
		return nil, fmt.Errorf("parse artifact %s: %w", path, err)
	}
	if len(artifact.ABI) == 0 {
		return nil, fmt.Errorf("artifact %s has empty abi", path)
	}
	return artifact.ABI, nil
}

// internal/indexer/decoder/abi.go
package decoder

import (
    "encoding/json"
    "os"
    "path/filepath"
)

type ContractABI struct {
    Abi json.RawMessage `json:"abi"`
}

func LoadABI(contractName, artifactPath string) (json.RawMessage, error) {
    // Path to Foundry artifact
    path := filepath.Join(artifactPath, contractName+".sol", contractName+".json")
    
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    var artifact ContractABI
    if err := json.Unmarshal(data, &artifact); err != nil {
        return nil, err
    }
    
    return artifact.Abi, nil
}
package registry

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

/*
 * ContractRegistry
 * @desc Thread-safe registry of contract addresses to watch
 */
type ContractRegistry struct {
	addresses map[common.Address]string // address -> contract type
	mu        sync.RWMutex
}

/*
 * NewContractRegistry
 * @desc Creates a new contract registry
 */
func NewContractRegistry() *ContractRegistry {
	return &ContractRegistry{
		addresses: make(map[common.Address]string),
	}
}

/*
 * AddContract
 * @desc Adds a contract address to the watch list
 */
func (r *ContractRegistry) AddContract(addr common.Address, contractType string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.addresses[addr] = contractType
}

/*
 * RemoveContract
 * @desc Removes a contract from the watch list
 */
func (r *ContractRegistry) RemoveContract(addr common.Address) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.addresses, addr)
}

/*
 * GetAddresses
 * @desc Returns all watched addresses
 */
func (r *ContractRegistry) GetAddresses() []common.Address {
	r.mu.RLock()
	defer r.mu.RUnlock()

	addrs := make([]common.Address, 0, len(r.addresses))
	for addr := range r.addresses {
		addrs = append(addrs, addr)
	}
	return addrs
}

/*
 * Contains
 * @desc Checks if an address is being watched
 */
func (r *ContractRegistry) Contains(addr common.Address) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.addresses[addr]
	return exists
}

/*
 * Count
 * @desc Returns the number of watched contracts
 */
func (r *ContractRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.addresses)
}

/*
 * GetContractType
 * @desc Returns the type of a contract if it's being watched
 */
func (r *ContractRegistry) GetContractType(addr common.Address) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	contractType, exists := r.addresses[addr]
	return contractType, exists
}

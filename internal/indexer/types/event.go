
package types
/*@
 * LogData
 * @desc Represents a raw blockchain log for retry processing
 * @fields
 *   - Address: Contract address that emitted the log
 *   - Topics: Event signature and indexed parameters
 *   - Data: Non-indexed event data
 *   - BlockNumber: Block where log was emitted
 *   - TxHash: Transaction hash
 *   - TxIndex: Transaction index in block
 *   - BlockHash: Hash of the containing block
 *   - Index: Log index in transaction
 *   - Removed: Whether log was removed due to reorg
 */
type LogData struct {
    Address     string   `json:"address"`
    Topics      []string `json:"topics"`
    Data        string   `json:"data"`
    BlockNumber uint64   `json:"blockNumber"`
    TxHash      string   `json:"transactionHash"`
    TxIndex     uint     `json:"transactionIndex"`
    BlockHash   string   `json:"blockHash"`
    Index       uint     `json:"logIndex"`
    Removed     bool     `json:"removed"`
}
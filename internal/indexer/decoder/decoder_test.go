package decoder

import (
	"encoding/hex"
	"math/big"
	"testing"

	indexertypes "github.com/Revvfi/revvfi-backend/internal/indexer/types"
	"github.com/ethereum/go-ethereum/common"
	goEthtypes "github.com/ethereum/go-ethereum/core/types"
)

// Helper function to create a test log
func createTestLog(topics []string, data string, blockNumber uint64) goEthtypes.Log {
	topicHashes := make([]common.Hash, len(topics))
	for i, topic := range topics {
		topicHashes[i] = common.HexToHash(topic)
	}
	
	dataBytes, _ := hex.DecodeString(data)
	
	return goEthtypes.Log{
		Topics:      topicHashes,
		Data:        dataBytes,
		BlockNumber: blockNumber,
	}
}

func TestGetEventName(t *testing.T) {
	tests := []struct {
		name      string
		topicHash string
		want      string
	}{
		{
			name:      "OfferSubmitted event",
			topicHash: "0x2ee8871fe528da3127dc4dda62594b1154a1090e1686c055c8d3134ba47eb117",
			want:      "OfferSubmitted",
		},
		{
			name:      "PositionMinted event",
			topicHash: "0x2ded751e875944a2c31c15a5af9674e47ede7eaabeb532b56306a392a191718d",
			want:      "PositionMinted",
		},
		{
			name:      "Unknown event",
			topicHash: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			want:      "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetEventName(tt.topicHash)
			if got != tt.want {
				t.Errorf("GetEventName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEventDecoder_DecodeOfferSubmitted(t *testing.T) {
	decoder := &EventDecoder{}
	
	// Create test log for OfferSubmitted event
	topics := []string{
		"0x2ee8871fe528da3127dc4dda62594b1154a1090e1686c055c8d3134ba47eb117",
		"0x0000000000000000000000000000000000000000000000000000000000000001", // offerId = 1
		"0x0000000000000000000000003C44CdDdB6a900fa2b585dd299e03d12FA4293BC", // lender
		"0x000000000000000000000000ce65c85aB3cA0e97B7bC73023F7e360d6Fa27b7C", // market
	}
	
	// Data: amount (10,000 USDC), APR (500), seniority (0), expiry (timestamp)
	// Each field is ABI-encoded as 32 bytes (uint8 seniority is zero-padded to 256 bits)
	data := "00000000000000000000000000000000000000000000000000000002540be400" + // amount
		"00000000000000000000000000000000000000000000000000000000000001f4" + // APR 500
		"0000000000000000000000000000000000000000000000000000000000000000" + // seniority (uint8, padded to 32 bytes)
		"000000000000000000000000000000000000000000000000000000006a308de7" // expiry
	
	log := createTestLog(topics, data, 12345)
	
	event, err := decoder.decodeOfferSubmitted(log)
	if err != nil {
		t.Fatalf("decodeOfferSubmitted failed: %v", err)
	}
	
	if event.OfferID.Cmp(big.NewInt(1)) != 0 {
		t.Errorf("OfferID = %v, want 1", event.OfferID)
	}
	
	if event.Lender.Hex() != "0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC" {
		t.Errorf("Lender = %v, want 0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC", event.Lender)
	}
	
	if event.APR.Cmp(big.NewInt(500)) != 0 {
		t.Errorf("APR = %v, want 500", event.APR)
	}
}

func TestEventDecoder_DecodePositionMinted(t *testing.T) {
	decoder := &EventDecoder{}
	
	topics := []string{
		"0x2ded751e875944a2c31c15a5af9674e47ede7eaabeb532b56306a392a191718d",
		"0x0000000000000000000000000000000000000000000000000000000000000001", // tokenId
		"0x0000000000000000000000003C44CdDdB6a900fa2b585dd299e03d12FA4293BC", // lender
		"0x000000000000000000000000ce65c85aB3cA0e97B7bC73023F7e360d6Fa27b7C", // market
	}
	
	data := "00000000000000000000000000000000000000000000000000000002540be400" + // principal
		"00000000000000000000000000000000000000000000000000000000000001f4" + // APR
		"00" // seniority
	
	log := createTestLog(topics, data, 12345)
	
	event, err := decoder.decodePositionMinted(log)
	if err != nil {
		t.Fatalf("decodePositionMinted failed: %v", err)
	}
	
	if event.TokenID.Cmp(big.NewInt(1)) != 0 {
		t.Errorf("TokenID = %v, want 1", event.TokenID)
	}
	
	if event.Principal.Cmp(big.NewInt(10000000000)) != 0 {
		t.Errorf("Principal = %v, want 10000000000", event.Principal)
	}
}

func TestEventDecoder_DecodeMarketDeployed(t *testing.T) {
	decoder := &EventDecoder{}
	
	topics := []string{
		"0x691293f43d18044b086f2453176ed1aef895fc431dd236aab13a95544b015079",
		"0x000000000000000000000000ce65c85aB3cA0e97B7bC73023F7e360d6Fa27b7C", // market
		"0x00000000000000000000000070997970C51812dc3A010C7d01b50e0d17dc79C8", // borrower
	}
	
	// Data: borrowAsset, collateralAsset, oracle, minCollateralRatio, liquidationThreshold
	data := "0000000000000000000000000000000000000000000000000000000000000000" +
		"0000000000000000000000000A3EE490d067C266Ceb6f17aA43bBE7732Ed11c9" + // borrowAsset
		"0000000000000000000000000000000000000000000000000000000000000000" +
		"0000000000000000000000003AA3DCE3a62fd37Ce28C1120d34A970b371cB69E" + // collateralAsset
		"0000000000000000000000000000000000000000000000000000000000000000" +
		"000000000000000000000000e5e56701f7e241cc8C598615C4FBf6EF1ECf27B6" + // oracle
		"0000000000000000000000000000000000000000000000000000000000002ee0" + // minCollateralRatio (12000)
		"0000000000000000000000000000000000000000000000000000000000002710" // liquidationThreshold (10000)
	
	log := createTestLog(topics, data, 12345)
	
	event, err := decoder.decodeMarketDeployed(log)
	if err != nil {
		t.Fatalf("decodeMarketDeployed failed: %v", err)
	}
	
	if event.MarketAddress.Hex() != "0xce65c85aB3cA0e97B7bC73023F7e360d6Fa27b7C" {
		t.Errorf("MarketAddress = %v, want 0xce65c85aB3cA0e97B7bC73023F7e360d6Fa27b7C", event.MarketAddress)
	}
}

func TestEventDecoder_DecodeBorrow(t *testing.T) {
	decoder := &EventDecoder{}
	
	topics := []string{
		"0x13ed6866d4e1ee6da46f845c46d7e54120883d75c5ea9a2dacc1c4ca8984ab80",
		"0x00000000000000000000000070997970C51812dc3A010C7d01b50e0d17dc79C8", // borrower
	}
	
	data := "00000000000000000000000000000000000000000000000000000002540be400" + // amount (10,000)
		"00000000000000000000000000000000000000000000000000000002540be400" + // totalDebt
		"00000000000000000000000000000000000000000000000000000000000001f4" // weightedAPR (500)
	
	log := createTestLog(topics, data, 12345)
	
	event, err := decoder.decodeBorrow(log)
	if err != nil {
		t.Fatalf("decodeBorrow failed: %v", err)
	}
	
	if event.Amount.Cmp(big.NewInt(10000000000)) != 0 {
		t.Errorf("Amount = %v, want 10000000000", event.Amount)
	}
}

func TestEventDecoder_DecodeTransfer(t *testing.T) {
	decoder := &EventDecoder{}
	
	topics := []string{
		"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
		"0x0000000000000000000000000000000000000000000000000000000000000000", // from (zero = mint)
		"0x0000000000000000000000003C44CdDdB6a900fa2b585dd299e03d12FA4293BC", // to (lender)
		"0x0000000000000000000000000000000000000000000000000000000000000001", // tokenId
	}
	
	log := createTestLog(topics, "", 12345)
	
	event, err := decoder.decodeTransfer(log)
	if err != nil {
		t.Fatalf("decodeTransfer failed: %v", err)
	}
	
	if event.To.Hex() != "0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC" {
		t.Errorf("To = %v, want 0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC", event.To)
	}
}

// Test all event signatures are valid
func TestEventSignatures(t *testing.T) {
	expectedCount := 42 // Total number of events we have
	
	if len(EventTopics) < expectedCount {
		t.Errorf("EventTopics has %d entries, expected at least %d", len(EventTopics), expectedCount)
	}
	
	// Test each signature is a valid hex hash
	for name, hash := range EventTopics {
		if len(hash) != 66 { // 0x + 64 hex chars
			t.Errorf("Event %s has invalid hash length: %d", name, len(hash))
		}
		if hash[:2] != "0x" {
			t.Errorf("Event %s hash missing 0x prefix: %s", name, hash)
		}
	}
}

// Test the Decode method with a real log
func TestEventDecoder_Decode(t *testing.T) {
	// Note: This test requires actual ABI files
	// Skip if running in CI without ABIs
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	
	// Create decoder with real artifact path
	artifactPath := "/home/ani/Documents/revvfi-contracts/revvfi-backend/internal/blockchain/abis"
	decoder, err := NewEventDecoder(artifactPath)
	if err != nil {
		t.Skipf("Skipping test, could not create decoder: %v", err)
	}
	
	// Test with OfferSubmitted log
	topics := []string{
		"0x2ee8871fe528da3127dc4dda62594b1154a1090e1686c055c8d3134ba47eb117",
		"0x0000000000000000000000000000000000000000000000000000000000000001",
		"0x0000000000000000000000003C44CdDdB6a900fa2b585dd299e03d12FA4293BC",
		"0x000000000000000000000000ce65c85aB3cA0e97B7bC73023F7e360d6Fa27b7C",
	}
	data := "00000000000000000000000000000000000000000000000000000002540be400" +
		"00000000000000000000000000000000000000000000000000000000000001f4" +
		"0000000000000000000000000000000000000000000000000000000000000000" + // seniority (uint8, padded to 32 bytes)
		"000000000000000000000000000000000000000000000000000000006a308de7"

	log := createTestLog(topics, data, 12345)

	decoded, err := decoder.Decode(log)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	
	event, ok := decoded.(*indexertypes.OfferSubmittedEvent)
	if !ok {
		t.Fatalf("Decoded event is not OfferSubmittedEvent type")
	}
	
	if event.OfferID.Cmp(big.NewInt(1)) != 0 {
		t.Errorf("OfferID = %v, want 1", event.OfferID)
	}
}

// Benchmark decode performance
func BenchmarkDecodeOfferSubmitted(b *testing.B) {
	decoder := &EventDecoder{}
	
	topics := []string{
		"0x2ee8871fe528da3127dc4dda62594b1154a1090e1686c055c8d3134ba47eb117",
		"0x0000000000000000000000000000000000000000000000000000000000000001",
		"0x0000000000000000000000003C44CdDdB6a900fa2b585dd299e03d12FA4293BC",
		"0x000000000000000000000000ce65c85aB3cA0e97B7bC73023F7e360d6Fa27b7C",
	}
	data := "00000000000000000000000000000000000000000000000000000002540be400" +
		"00000000000000000000000000000000000000000000000000000000000001f4" +
		"00" +
		"000000000000000000000000000000000000000000000000000000006a308de7"
	
	log := createTestLog(topics, data, 12345)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = decoder.decodeOfferSubmitted(log)
	}
}
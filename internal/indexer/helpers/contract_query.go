package helpers

import (
	"context"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

/*
 * MarketChildContracts
 * @desc Stores addresses of a market's child contracts
 */
type MarketChildContracts struct {
	Market           common.Address
	OfferBook        common.Address
	CollateralEscrow common.Address
	LiquidityQueue   common.Address
}

/*
 * GetMarketChildContracts
 * @desc Queries a market contract to get all child contract addresses
 */
func GetMarketChildContracts(
	ctx context.Context,
	client *ethclient.Client,
	marketAddr common.Address,
) (*MarketChildContracts, error) {

	// Call offerBook()
	offerBook, err := callAddressGetter(ctx, client, marketAddr, "offerBook")
	if err != nil {
		return nil, fmt.Errorf("failed to get offerBook: %w", err)
	}

	// Call collateralEscrow()
	escrow, err := callAddressGetter(ctx, client, marketAddr, "collateralEscrow")
	if err != nil {
		return nil, fmt.Errorf("failed to get collateralEscrow: %w", err)
	}

	// Call liquidityQueue() - optional, may not exist on all markets
	queue, err := callAddressGetter(ctx, client, marketAddr, "liquidityQueue")
	if err != nil {
		// LiquidityQueue is optional, use zero address if not available
		queue = common.Address{}
	}

	return &MarketChildContracts{
		Market:           marketAddr,
		OfferBook:        offerBook,
		CollateralEscrow: escrow,
		LiquidityQueue:   queue,
	}, nil
}

/*
 * callAddressGetter
 * @desc Generic helper to call a no-args function that returns an address
 */
func callAddressGetter(
	ctx context.Context,
	client *ethclient.Client,
	contract common.Address,
	method string,
) (common.Address, error) {

	// ABI for a simple address getter: function name() view returns (address)
	abiJSON := `[{"constant":true,"inputs":[],"name":"` + method +
		`","outputs":[{"name":"","type":"address"}],"type":"function"}]`

	parsedABI, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to parse ABI: %w", err)
	}

	caller := bind.NewBoundContract(contract, parsedABI, client, nil, nil)

	var result []interface{}
	err = caller.Call(&bind.CallOpts{Context: ctx}, &result, method)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to call %s(): %w", method, err)
	}

	if len(result) == 0 {
		return common.Address{}, fmt.Errorf("no result from %s()", method)
	}

	addr, ok := result[0].(common.Address)
	if !ok {
		return common.Address{}, fmt.Errorf("result is not an address")
	}

	return addr, nil
}

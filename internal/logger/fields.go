package logger

import (
	"log/slog"
	"time"
)

// Helper functions to create structured fields
// These make it easy to add consistent structured data to logs

// WithTxHash adds transaction hash to log
func WithTxHash(hash string) slog.Attr {
	return slog.String("tx_hash", hash)
}

// WithMarketAddress adds market address to log
func WithMarketAddress(addr string) slog.Attr {
	return slog.String("market_address", addr)
}

// WithOfferID adds offer ID to log
func WithOfferID(id int64) slog.Attr {
	return slog.Int64("offer_id", id)
}

// WithPositionID adds position ID to log
func WithPositionID(id int64) slog.Attr {
	return slog.Int64("position_id", id)
}

// WithTokenID adds token ID to log
func WithTokenID(id int64) slog.Attr {
	return slog.Int64("token_id", id)
}

// WithWalletAddress adds wallet address to log
func WithWalletAddress(addr string) slog.Attr {
	return slog.String("wallet_address", addr)
}

// WithBorrower adds borrower address to log
func WithBorrower(addr string) slog.Attr {
	return slog.String("borrower", addr)
}

// WithLender adds lender address to log
func WithLender(addr string) slog.Attr {
	return slog.String("lender", addr)
}

// WithRoute adds HTTP route to log
func WithRoute(route string) slog.Attr {
	return slog.String("route", route)
}

// WithMethod adds HTTP method to log
func WithMethod(method string) slog.Attr {
	return slog.String("method", method)
}

// WithStatusCode adds HTTP status code to log
func WithStatusCode(code int) slog.Attr {
	return slog.Int("status_code", code)
}

// WithLatency adds request latency in milliseconds
func WithLatency(duration time.Duration) slog.Attr {
	return slog.Int64("latency_ms", duration.Milliseconds())
}

// WithEventName adds blockchain event name to log
func WithEventName(name string) slog.Attr {
	return slog.String("event_name", name)
}

// WithBlockNumber adds block number to log
func WithBlockNumber(num uint64) slog.Attr {
	return slog.Uint64("block_number", num)
}

// WithLogIndex adds log index to log
func WithLogIndex(idx uint) slog.Attr {
	return slog.Int("log_index", int(idx))
}

// WithContractAddress adds contract address to log
func WithContractAddress(addr string) slog.Attr {
	return slog.String("contract_address", addr)
}

// WithAmount adds amount to log
func WithAmount(amount string) slog.Attr {
	return slog.String("amount", amount)
}

// WithError adds error to log
func WithError(err error) slog.Attr {
	if err == nil {
		return slog.String("error", "nil")
	}
	return slog.String("error", err.Error())
}

// WithDuration adds processing duration in milliseconds
func WithDuration(duration time.Duration) slog.Attr {
	return slog.Int64("processing_time_ms", duration.Milliseconds())
}

// WithQuery adds SQL query to log (use carefully, may contain sensitive data)
func WithQuery(query string) slog.Attr {
	return slog.String("query", query)
}

// WithRowsAffected adds rows affected by query
func WithRowsAffected(rows int64) slog.Attr {
	return slog.Int64("rows_affected", rows)
}

// WithRequestID adds request ID to log
func WithRequestID(id string) slog.Attr {
	return slog.String("request_id", id)
}

// WithUserID adds user ID to log
func WithUserID(id string) slog.Attr {
	return slog.String("user_id", id)
}

// WithBorrowAsset adds borrow asset address to log
func WithBorrowAsset(addr string) slog.Attr {
	return slog.String("borrow_asset", addr)
}

// WithCollateralAsset adds collateral asset address to log
func WithCollateralAsset(addr string) slog.Attr {
	return slog.String("collateral_asset", addr)
}

// WithAPR adds APR to log
func WithAPR(apr int) slog.Attr {
	return slog.Int("apr", apr)
}

// WithSeniority adds seniority to log
func WithSeniority(seniority int) slog.Attr {
	return slog.Int("seniority", seniority)
}

// WithEpoch adds epoch number to log
func WithEpoch(epoch uint64) slog.Attr {
	return slog.Uint64("epoch", epoch)
}

// WithFromBlock adds from block number to log
func WithFromBlock(block uint64) slog.Attr {
	return slog.Uint64("from_block", block)
}

// WithToBlock adds to block number to log
func WithToBlock(block uint64) slog.Attr {
	return slog.Uint64("to_block", block)
}

// WithLogCount adds number of logs to log
func WithLogCount(count int) slog.Attr {
	return slog.Int("log_count", count)
}

// WithRetryCount adds retry count to log
func WithRetryCount(count int) slog.Attr {
	return slog.Int("retry_count", count)
}

// WithBackoffDuration adds backoff duration to log
func WithBackoffDuration(duration time.Duration) slog.Attr {
	return slog.Int64("backoff_ms", duration.Milliseconds())
}

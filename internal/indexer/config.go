package indexer

import "github.com/Revvfi/revvfi-backend/internal/indexer/settings"

/*@
 * Config
 * @desc Compatibility alias for callers that construct indexer settings from
 * the parent indexer package.
 */
type Config = settings.Config

/*@
 * DefaultConfig
 * @desc Compatibility wrapper around settings.DefaultConfig.
 */
func DefaultConfig() *Config {
	return settings.DefaultConfig()
}

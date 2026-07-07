package decoder

import "log"

// EventTopics maps event names to their keccak256 signatures
// These signatures are chain-agnostic and should never change unless the event definitions themselves change.
// Generated from forge inspect on 2026-06-14
var EventTopics = map[string]string{
	// ============================================
	// FACTORY EVENTS (RevvFiFactory - 7 events)
	// ============================================
	"ArchControllerUpdateRequested": "0x17df3b3106fc9beeb88a854ef5935fbb1ae165cbb2169be6a83702c85a38144e",
	"ArchControllerUpdated":         "0x8ac25e96c75d72359112618e4d4afac224614406fb6752e217621405b6546621",
	"CoreContractsSet":              "0x99eef4ac36b4c1c37950298c3ce923328cd1265d785b522c2f562fd0db235d9a",
	"FeeUpdated":                    "0x528d9479e9f9889a87a3c30c7f7ba537e5e59c4c85a37733b16e57c62df61302",
	"ImplementationsSet":            "0x3e3aac68f5d3430f5b97b6a6959dd13910b31907891cc9f6cd886dc0f92fcad1",
	"MarketDeployed":                "0x2a0868e27f2ca98edb604e5b84448b5e960e6eb8da3a6fb2dd4692f980549c1d",

	// ============================================
	// MARKET EVENTS (RevvFiMarket - 11 events)
	// ============================================
	"Borrow":                   "0xe1979fe4c35e0cef342fef5668e2c8e7a7e9f5d5d1ca8fee0ac6c427fa4153af",
	"ContractsSet":             "0xeaedf84a54115aef3214ad296d4e9d465ef83830625852a51d892bb830100c40",
	"DrawdownExecuted":         "0xd6183328d9c97aa1695793c0f5cdfd9117f27e0c9ac65205c4c0eb031cc47f3d",
	"GuardianUpdated":          "0x064d28d3d3071c5cbc271a261c10c2f0f0d9e319390397101aa0eb23c6bad909",
	"LenderRepaid":             "0xf6fc540193542dc4bbf6bcf9707f55489c017b3094d765abf460a52a5780ebf2",
	"LiquidationEndedMarket":   "0x445e027cfd2db85d19806ef8d31d4bdf988e15bd98e40d254519916e9f313079",
	"LiquidationStartedMarket": "0x97ee3b9e4cd3242f1a9e7324336e978ed8d916c597cd0f6a3c8c5b65b93cad9b",
	"MarketClosedEvent":        "0x09a09c46728dfde8eed179c677c1238c3dde7f0d264ed782135a43b23222e219",
	"PositionClaimed":          "0xf18ca9717b47ea7629d17b5d7b2ede46e864155725629e7d0347924c2f6d112d",
	"PositionSettled":          "0x78dd5e343bcb666bca386d600c40d1412bff3eb1772d8be0f416118a4cd5a9bd",
	"Repay":                    "0x2fe77b1c99aca6b022b8efc6e3e8dd1b48b30748709339b65c50ef3263443e09",

	// ============================================
	// OFFERBOOK EVENTS (RevvFiOfferBook - 6 events)
	// ============================================
	"DrawdownExecutedOffer": "0x71bb7db5826ee7f4728165cba6e22dd34aee9a9b73be58e7ac6ea5be7609713f",
	"OfferCancelled":        "0xe2dc0bfd2fc658db980184c3f7c091c2f886874e896043b562d5099c8d95b1dd",
	"OfferExpired":          "0x73dedbcfa9a17d6e3f87cd88fa5c704da1a9abff4ec94d2f1a0b0fd0b24d082a",
	"OfferFilled":           "0xd5f3552d5f96296cf317f1ff0718c8430d2e840d3f949d1605e153e8f583a958",
	"OfferModified":         "0xcec4e6d0937f30ec7b79423b833abf1e4dafe836737ef7d842a35c076c0ae42b",
	"OfferSubmitted":        "0x2ee8871fe528da3127dc4dda62594b1154a1090e1686c055c8d3134ba47eb117",

	// ============================================
	// POSITION NFT EVENTS (RevvFiPositionNFT - 8 events)
	// ============================================
	"Approval":           "0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925",
	"ApprovalForAll":     "0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31",
	"MarketRegistered":   "0xd2392e8674082692962920b6bc022067ae83f52fc1de409f5c74de7aa893bd02",
	"MarketUnregistered": "0x2647d54498c3c18e1fe998f0217b8033cdde1c1b8fb71c2d52447111aca1de35",
	"PositionMinted":     "0x2ded751e875944a2c31c15a5af9674e47ede7eaabeb532b56306a392a191718d",
	"PositionRedeemed":   "0x2603df6f11a0e6645f74d36203c0fd88d5f298ed69769d43fe98cdb7774b8483",
	"Transfer":           "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",

	// ============================================
	// LIQUIDATOR EVENTS (RevvFiLiquidator - 4 events)
	// ============================================
	"AuctionCancelled": "0x2809c7e17bf978fbc7194c0a694b638c4215e9140cacc6c38ca36010b45697df",
	"AuctionCreated":   "0xa7fc29bef0deb3ca05e0e792df3196284ccaa75747c5346f72ace0e6e7e821ea",
	"AuctionSettled":   "0x0fd449c00a8fcbc38da3f998ca7b1775cb5042749ee6103189e016704a8b9fc0",
	"BidPlaced":        "0x51db8e23b3f4479b162fd48823b8402895442b8f6cfd94f66239391881ec7b6f",

	// ============================================
	// LIQUIDITY QUEUE EVENTS (RevvFiLiquidityQueue - 6 events)
	// ============================================
	"EpochProcessed":      "0xf0032d32026594ba6068f408da4fdfd559fb2ec3eccef8e9c3487bc477ca4426",
	"PositionUnlocked":    "0xced6ce5feb1d3d9527ea3888f0f3e1bdb365c4c248eb5762fc68659b3be0f440",
	"WithdrawalCancelled": "0x1827a0ec328915df70a21d3ac58a86f06973898dbc2eaee0aa91be7563f136eb",
	"WithdrawalClaimed":   "0x8adb7a84b2998a8d11cd9284395f95d5a99f160be785ae79998c654979bd3d9a",
	"WithdrawalRequested": "0x8774dcd6f347788be32bacb440fd662ba943bc006ab5f33a8c8517241e477ec5",

	// ============================================
	// REPUTATION REGISTRY EVENTS (ReputationRegistry - 6 events)
	// ============================================
	"BorrowActivityRecorded":      "0xeaecf6e2c9e155bc699c347a9cc32dc7f1a32b9dc13be3e631dc48ab78521c6a",
	"BorrowerRegistered":          "0x57c3b463895a0da2fe39978165ab8b71864d9225a91dd3bda7184006e52cdd0c",
	"DefaultRecorded":             "0x9989a0beb5437893570946e32441aff01c275c6572ad5981654c6a0b79e23848",
	"ReputationScoreUpdated":      "0xfef3709455ce9b0555e39289c78f517f754919ba4d6a105faa00ccb4417e4189",
	"SuccessfulRepaymentRecorded": "0x5fa8b5b1d0e8a192bd39abc8b925749c1715d8b23930736831a92ba18e840fa8",

	// ============================================
	// COLLATERAL ESCROW EVENTS (RevvFiCollateralEscrow - 5 events)
	// ============================================
	"CollateralDeposited":         "0xd7243f6f8212d5188fd054141cf6ea89cfc0d91facb8c3afe2f88a1358480142",
	"CollateralWithdrawn":         "0xc30fcfbcaac9e0deffa719714eaa82396ff506a0d0d0eebe170830177288715d",
	"CollateralLiquidated":        "0x4f65008ff243a1aab5777a6563d91c319b667f92b3f74af652fc9cd4f247d20b",
	"MinCollateralRatioUpdated":    "0x948ba262e91565b9caa2284557ddf9a7fb36ee3f00681b4fd1bfaae941e9188f",
	"LiquidationThresholdUpdated": "0xcdadd717dc9ee3550a289071d1af75e229726888d51e3a31c9e3dfc693d4852b",

	// ============================================
	// ARCH CONTROLLER EVENTS (RevvFiArchController - 11 events)
	// ============================================
	"AssetBlacklisted":        "0x10c0fc3aa3fe3ed617ce4c1f51f37e0203828eeab656b100bfb44b569779d638",
	"AssetPermitted":          "0xbfd78d1acd7fdf715658096e9116ace58b0b335d6d8d4fc53e0aec7edbd7d441",
	"BorrowerAdded":           "0xf936762bbd5951fd80f529c57a3ca1141f0b79ab926abd60a3879f75eb6b72a6",
	"BorrowerRemoved":         "0xeb31998799e9a09df1e8e5fd968bfeb5519b82b1b88d59dec8c860eac0265864",
	"ControllerAdded":         "0x09703263c91de41f96b822b3995609acf9858ba081d151c4e7ec3398085ae326",
	"ControllerFactoryAdded":  "0x8968daf7b2953d00c089355179545009ad3fc90082d096e28b028bbdad5cc5ab",
	"ControllerFactoryRemoved": "0x92b171cd00b60c6242c08403ef5f6b7258d81b3cba5956c115799d7a118874f9",
	"ControllerRemoved":       "0x33d83959be2573f5453b12eb9d43b3499bc57d96bd2f067ba44803c859e81113",
	"MarketAdded":             "0x7772c85e68debdf74fad87834e2cc05fa763e74faf14de7096da305290651142",
	"MarketRemoved":           "0x59d7b1e52008dc342c9421dadfc773114b914a65682a4e4b53cf60a970df0d77",

	// ============================================
	// COMMON EVENTS (shared across contracts)
	// ============================================
	"Initialized":         "0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498",
	"OwnershipTransferred": "0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0",
}

// GetEventName returns the event name from a topic hash
func GetEventName(topicHash string) string {
	for name, hash := range EventTopics {
		if hash == topicHash {
			return name
		}
	}
	log.Printf("UNKNOWN TOPIC: %s", topicHash)
	return ""
}

// GetEventHash returns the event hash from an event name
func GetEventHash(eventName string) string {
	if hash, ok := EventTopics[eventName]; ok {
		return hash
	}
	return ""
}

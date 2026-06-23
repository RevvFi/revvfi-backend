package admin

import (
	"fmt"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/response"
	"github.com/Revvfi/revvfi-backend/internal/config"
	appErr "github.com/Revvfi/revvfi-backend/internal/pkg/errors"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
)

/*
@file system_service.go

@desc
Service for admin system configuration management.

@responsibilities
- Return all system configuration key-value entries
- Prepare calldata for on-chain configuration updates (stored in DB; calldata for governance)
- Write audit log entries for configuration changes
*/

/*
@struct AdminSystemService

@desc
Implements admin system configuration management.

@fields
- repo: admin repository for system_config table access
- cfg: blockchain configuration for the governance contract address
*/
type AdminSystemService struct {
	repo *postgres.AdminRepository
	cfg  config.BlockchainConfig
}

/*
@function NewAdminSystemService

@desc
Creates a new AdminSystemService.

@params
- repo: admin PostgreSQL repository
- cfg: blockchain configuration

@returns
- *AdminSystemService
*/
func NewAdminSystemService(repo *postgres.AdminRepository, cfg config.BlockchainConfig) *AdminSystemService {
	return &AdminSystemService{repo: repo, cfg: cfg}
}

/*
@method GetSystemConfig

@desc
Returns all system configuration entries, masking values marked as sensitive.

@params
- ctx: handler context

@returns
- *response.SystemConfigListResponse: all config entries
- error
*/
func (s *AdminSystemService) GetSystemConfig(ctx interface{}) (*response.SystemConfigListResponse, error) {
	c := extractCtx(ctx)
	rows, err := s.repo.ListSystemConfigs(c)
	if err != nil {
		return nil, fmt.Errorf("list system configs: %w", err)
	}

	entries := make([]response.SystemConfigEntry, 0, len(rows))
	for _, row := range rows {
		value := row.Value
		if row.IsSensitive {
			value = "***"
		}
		desc := ""
		if row.Description.Valid {
			desc = row.Description.String
		}
		updatedBy := ""
		if row.UpdatedBy.Valid {
			updatedBy = row.UpdatedBy.String
		}
		entries = append(entries, response.SystemConfigEntry{
			Key:         row.Key,
			Value:       value,
			Description: desc,
			ConfigType:  row.ConfigType,
			IsSensitive: row.IsSensitive,
			UpdatedAt:   row.UpdatedAt.Unix(),
			UpdatedBy:   updatedBy,
		})
	}
	return &response.SystemConfigListResponse{
		Configs: entries,
		Total:   len(entries),
	}, nil
}

/*
@method PrepareUpdateSystemConfig

@desc
Updates the system configuration value in the DB and returns ABI-encoded
calldata for governance execution (for configs that have on-chain counterparts).

@params
- ctx: handler context
- key: configuration key to update
- value: new configuration value
- adminAddress: admin wallet performing the update
- reason: justification for the change

@returns
- *response.TransactionPreparation: encoded calldata for governance submission
- error
*/
func (s *AdminSystemService) PrepareUpdateSystemConfig(ctx interface{}, key, value, adminAddress, reason string) (*response.TransactionPreparation, error) {
	c := extractCtx(ctx)
	if key == "" {
		return nil, appErr.ErrInvalidInput
	}
	if err := s.repo.UpsertSystemConfig(c, key, value, adminAddress); err != nil {
		return nil, fmt.Errorf("upsert system config: %w", err)
	}
	_ = s.repo.InsertAuditLog(c, adminAddress, "update_system_config", "system", "", "", reason)

	// The ArchController emits a config update event; calldata encodes key+value as strings.
	sel := functionSelector("setSystemConfig(string,string)")
	// ABI-encode two dynamic strings: offset1, offset2, len1, data1, len2, data2
	keyBytes := []byte(key)
	valBytes := []byte(value)

	// offsets: 0x40 (64) and 0x40+32+len(key) rounded
	keyLen := int64(len(keyBytes))
	valOffset := int64(64 + 32 + ((keyLen + 31) / 32 * 32))

	var buf []byte
	buf = append(buf, padUint256FromInt(64)...)                       // offset to key data
	buf = append(buf, padUint256FromInt(valOffset)...)                // offset to value data
	buf = append(buf, padUint256FromInt(keyLen)...)                   // key length
	buf = append(buf, padStringBytes(keyBytes)...)                    // key data (padded to 32)
	buf = append(buf, padUint256FromInt(int64(len(valBytes)))...)     // value length
	buf = append(buf, padStringBytes(valBytes)...)                    // value data (padded to 32)

	calldata := "0x" + hexEncodeBytes(append(sel, buf...))
	return &response.TransactionPreparation{
		Target:      s.cfg.ArchControllerAddress,
		TxData:      calldata,
		Value:       "0",
		Description: fmt.Sprintf("Update system config: %s = %s", key, value),
	}, nil
}

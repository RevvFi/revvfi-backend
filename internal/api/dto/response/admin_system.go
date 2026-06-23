package response

/*
@file admin_system.go

@desc
Response DTOs for system configuration admin endpoints.

@responsibilities
- Define system config entry response shape
- Define system config listing response
*/

/*
@struct SystemConfigEntry

@desc
Single system configuration key-value entry.

@fields
- Key: configuration key identifier
- Value: configuration value (masked if sensitive)
- Description: human-readable explanation of the config key
- ConfigType: value type (string, number, boolean, address)
- IsSensitive: whether the value is sensitive and should be masked
- UpdatedAt: Unix timestamp of last update
- UpdatedBy: admin address that performed the last update
*/
type SystemConfigEntry struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
	ConfigType  string `json:"config_type"`
	IsSensitive bool   `json:"is_sensitive"`
	UpdatedAt   int64  `json:"updated_at"`
	UpdatedBy   string `json:"updated_by,omitempty"`
}

/*
@struct SystemConfigListResponse

@desc
All system configuration entries.

@fields
- Configs: list of all system config entries
- Total: total count of config entries
*/
type SystemConfigListResponse struct {
	Configs []SystemConfigEntry `json:"configs"`
	Total   int                 `json:"total"`
}

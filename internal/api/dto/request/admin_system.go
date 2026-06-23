package request

/*
@file admin_system.go

@desc
Request DTOs for system configuration admin endpoints.

@responsibilities
- Validate system configuration update requests
*/

/*
@struct UpdateSystemConfigRequest

@desc
Request body for preparing a system configuration update.

@fields
- Key: configuration key to update (required)
- Value: new configuration value (required)
- Reason: justification for the change (max 500 chars)
*/
type UpdateSystemConfigRequest struct {
	Key    string `json:"key" binding:"required,max=128"`
	Value  string `json:"value" binding:"required,max=1024"`
	Reason string `json:"reason" binding:"max=500"`
}

package position

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/Revvfi/revvfi-backend/internal/models"
	"github.com/stretchr/testify/assert"
)

/*
@file service_test.go

@desc
Unit tests for position service.
*/

/*
@struct MockPositionRepository

@desc
Mock implementation for testing.
*/
type MockPositionRepository struct {
	positions map[int64]*models.Position
	err       error
}

/*
@method CreatePosition

@desc
Mock position creation.
*/
func (m *MockPositionRepository) CreatePosition(ctx context.Context, position *models.Position) error {
	if m.err != nil {
		return m.err
	}
	m.positions[position.TokenID] = position
	return nil
}

/*
@method GetByTokenID

@desc
Mock position retrieval by token ID.
*/
func (m *MockPositionRepository) GetByTokenID(ctx context.Context, tokenID int64) (*models.Position, error) {
	if m.err != nil {
		return nil, m.err
	}
	pos, exists := m.positions[tokenID]
	if !exists {
		return nil, nil
	}
	return pos, nil
}

/*
@method GetByLender

@desc
Mock position retrieval by lender.
*/
func (m *MockPositionRepository) GetByLender(ctx context.Context, lender string, limit, offset int32) ([]models.Position, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []models.Position
	for _, pos := range m.positions {
		if pos.Lender == lender {
			result = append(result, *pos)
		}
	}
	return result, nil
}

/*
@method UpdatePosition

@desc
Mock position update.
*/
func (m *MockPositionRepository) UpdatePosition(ctx context.Context, position *models.Position) error {
	if m.err != nil {
		return m.err
	}
	m.positions[position.TokenID] = position
	return nil
}

/*
@method CountActiveByLender

@desc
Mock count active positions.
*/
func (m *MockPositionRepository) CountActiveByLender(ctx context.Context, lender string) (int32, error) {
	if m.err != nil {
		return 0, m.err
	}
	count := int32(0)
	for _, pos := range m.positions {
		if pos.Lender == lender && pos.IsActive {
			count++
		}
	}
	return count, nil
}

/*
@method GetTotalValueByLender

@desc
Mock total value calculation.
*/
func (m *MockPositionRepository) GetTotalValueByLender(ctx context.Context, lender string) (*big.Int, error) {
	if m.err != nil {
		return nil, m.err
	}
	total := big.NewInt(0)
	for _, pos := range m.positions {
		if pos.Lender == lender {
			// CurrentValue = CurrentPrincipal + AccruedInterest
			currentValue := new(big.Int).Add(pos.CurrentPrincipal, pos.AccruedInterest)
			total.Add(total, currentValue)
		}
	}
	return total, nil
}

/*
@function TestCreatePosition

@desc
Tests position creation.
*/
func TestCreatePosition(t *testing.T) {
	tests := []struct {
		name    string
		mock    *MockPositionRepository
		wantErr bool
	}{
		{
			name: "successful creation",
			mock: &MockPositionRepository{
				positions: make(map[int64]*models.Position),
				err:       nil,
			},
			wantErr: false,
		},
		{
			name: "database error",
			mock: &MockPositionRepository{
				positions: make(map[int64]*models.Position),
				err:       errors.New("db error"),
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
svc := NewPositionService(tc.mock)

pos, err := svc.CreatePosition(
context.Background(),
				1,
				"0xLender1",
				"0xMarket1",
				big.NewInt(1000000),
				500,
				0,
			)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, pos)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, pos)
				assert.Equal(t, int64(1), pos.TokenID)
				assert.Equal(t, "0xLender1", pos.Lender)
				assert.Equal(t, "active", pos.Status)
			}
		})
	}
}

/*
@function TestGetPosition

@desc
Tests position retrieval.
*/
func TestGetPosition(t *testing.T) {
	tests := []struct {
		name        string
		tokenID     int64
		mock        *MockPositionRepository
		wantErr     bool
		errMessage  string
	}{
		{
			name:    "position found",
			tokenID: 1,
			mock: &MockPositionRepository{
				positions: map[int64]*models.Position{
					1: {
						TokenID: 1,
						Lender:  "0xLender1",
						Status:  "active",
					},
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name:    "position not found",
			tokenID: 999,
			mock: &MockPositionRepository{
				positions: make(map[int64]*models.Position),
				err:       nil,
			},
			wantErr:    true,
			errMessage: "not found",
		},
		{
			name:    "database error",
			tokenID: 1,
			mock: &MockPositionRepository{
				positions: make(map[int64]*models.Position),
				err:       errors.New("db error"),
			},
			wantErr:    true,
			errMessage: "failed to fetch",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
svc := NewPositionService(tc.mock)

pos, err := svc.GetPosition(context.Background(), tc.tokenID)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, pos)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, pos)
				assert.Equal(t, tc.tokenID, pos.TokenID)
			}
		})
	}
}

/*
@function TestGetLenderPositions

@desc
Tests lender position retrieval.
*/
func TestGetLenderPositions(t *testing.T) {
	tests := []struct {
		name           string
		lender         string
		mock           *MockPositionRepository
		wantErr        bool
		expectedCount  int
	}{
		{
			name:   "multiple positions",
			lender: "0xLender1",
			mock: &MockPositionRepository{
				positions: map[int64]*models.Position{
					1: {
						TokenID: 1,
						Lender:  "0xLender1",
						Status:  "active",
					},
					2: {
						TokenID: 2,
						Lender:  "0xLender1",
						Status:  "active",
					},
					3: {
						TokenID: 3,
						Lender:  "0xLender2",
						Status:  "active",
					},
				},
				err: nil,
			},
			wantErr:       false,
			expectedCount: 2,
		},
		{
			name:   "no positions",
			lender: "0xLender3",
			mock: &MockPositionRepository{
				positions: make(map[int64]*models.Position),
				err:       nil,
			},
			wantErr:       false,
			expectedCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
svc := NewPositionService(tc.mock)

positions, err := svc.GetLenderPositions(context.Background(), tc.lender, 100, 0)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, positions, tc.expectedCount)
			}
		})
	}
}

/*
@function TestGetPortfolioSummary

@desc
Tests portfolio summary calculation.
*/
func TestGetPortfolioSummary(t *testing.T) {
	lender := "0xLender1"
	mockRepo := &MockPositionRepository{
		positions: map[int64]*models.Position{
			1: {
				TokenID:          1,
				Lender:           lender,
				Principal:        big.NewInt(1000000),
				CurrentPrincipal: big.NewInt(1000000),
				AccruedInterest:  big.NewInt(50000),
				APR:              500,
				IsActive:         true,
				Status:           "active",
			},
			2: {
				TokenID:          2,
				Lender:           lender,
				Principal:        big.NewInt(2000000),
				CurrentPrincipal: big.NewInt(2000000),
				AccruedInterest:  big.NewInt(100000),
				APR:              600,
				IsActive:         true,
				Status:           "active",
			},
		},
		err: nil,
	}

	svc := NewPositionService(mockRepo)
	summary, err := svc.GetPortfolioSummary(context.Background(), lender)

	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, int32(2), summary["active_positions"])
	assert.Equal(t, "150000", summary["earned_interest"])
}

/*
@function TestUpdatePositionValue

@desc
Tests position value update.
*/
func TestUpdatePositionValue(t *testing.T) {
	tests := []struct {
		name           string
		tokenID        int64
		newInterest    *big.Int
		mock           *MockPositionRepository
		wantErr        bool
	}{
		{
			name:        "successful update",
			tokenID:     1,
			newInterest: big.NewInt(100000),
			mock: &MockPositionRepository{
				positions: map[int64]*models.Position{
					1: {
						TokenID:          1,
						Principal:        big.NewInt(1000000),
						CurrentPrincipal: big.NewInt(1000000),
						AccruedInterest:  big.NewInt(0),
					},
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name:        "position not found",
			tokenID:     999,
			newInterest: big.NewInt(100000),
			mock: &MockPositionRepository{
				positions: make(map[int64]*models.Position),
				err:       nil,
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
svc := NewPositionService(tc.mock)

err := svc.UpdatePositionValue(context.Background(), tc.tokenID, tc.newInterest)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				pos, _ := tc.mock.GetByTokenID(context.Background(), tc.tokenID)
				assert.Equal(t, tc.newInterest.String(), pos.AccruedInterest.String())
			}
		})
	}
}

/*
@function TestSettlePosition

@desc
Tests position settlement.
*/
func TestSettlePosition(t *testing.T) {
	tests := []struct {
		name    string
		tokenID int64
		mock    *MockPositionRepository
		wantErr bool
	}{
		{
			name:    "successful settlement",
			tokenID: 1,
			mock: &MockPositionRepository{
				positions: map[int64]*models.Position{
					1: {
						TokenID:          1,
						IsActive:         true,
						IsSettled:        false,
						CurrentPrincipal: big.NewInt(1000000),
						AccruedInterest:  big.NewInt(50000),
						Status:           "active",
					},
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name:    "already settled",
			tokenID: 2,
			mock: &MockPositionRepository{
				positions: map[int64]*models.Position{
					2: {
						TokenID:   2,
						IsSettled: true,
						Status:    "settled",
					},
				},
				err: nil,
			},
			wantErr: true,
		},
		{
			name:    "position not found",
			tokenID: 999,
			mock: &MockPositionRepository{
				positions: make(map[int64]*models.Position),
				err:       nil,
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
svc := NewPositionService(tc.mock)

err := svc.SettlePosition(context.Background(), tc.tokenID)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				pos, _ := tc.mock.GetByTokenID(context.Background(), tc.tokenID)
				assert.True(t, pos.IsSettled)
				assert.Equal(t, "settled", pos.Status)
			}
		})
	}
}

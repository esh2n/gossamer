// Copyright 2021 ChainSafe Systems (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package babe

import (
	"testing"

	"github.com/ChainSafe/gossamer/dot/types"
	"github.com/ChainSafe/gossamer/lib/crypto/sr25519"
	"github.com/ChainSafe/gossamer/lib/keystore"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var keyring, _ = keystore.NewSr25519Keyring()

func TestBabeService_getEpochDataAndStartSlot(t *testing.T) {
	kp := keyring.Alice().(*sr25519.Keypair)
	authority := types.NewAuthority(kp.Public(), uint64(1))
	testEpochData := &types.EpochDataRaw{
		Randomness:  [32]byte{1},
		Authorities: []types.AuthorityRaw{*authority.ToRaw()},
	}

	testConfigData := &types.ConfigData{
		C1: 1,
		C2: 1,
	}

	testLatestConfigData := &types.ConfigData{
		C1: 1,
		C2: 2,
	}

	testEpochDataEpoch0 := &types.EpochDataRaw{
		Randomness: [32]byte{9},
		Authorities: []types.AuthorityRaw{
			*authority.ToRaw(),
		},
	}

	threshold0, err := CalculateThreshold(testConfigData.C1, testConfigData.C2, 1)
	require.NoError(t, err)

	threshold1, err := CalculateThreshold(testLatestConfigData.C1, testLatestConfigData.C2, 1)
	require.NoError(t, err)

	cases := []struct {
		service           func(*gomock.Controller) *Service
		name              string
		epoch             uint64
		expected          *epochData
		expectedStartSlot uint64
	}{
		{
			name: "should_get_epoch_data_for_epoch_0",
			service: func(ctrl *gomock.Controller) *Service {
				mockEpochState := NewMockEpochState(ctrl)

				mockEpochState.EXPECT().
					GetEpochDataRaw(uint64(0), nil).
					Return(testEpochDataEpoch0, nil)
				mockEpochState.EXPECT().
					GetConfigData(uint64(0), nil).
					Return(testConfigData, nil)

				return &Service{
					authority:  true,
					keypair:    kp,
					epochState: mockEpochState,
				}
			},
			epoch: 0,
			expected: &epochData{
				randomness:     testEpochDataEpoch0.Randomness,
				authorities:    testEpochDataEpoch0.Authorities,
				authorityIndex: 0,
				threshold:      threshold0,
			},
			expectedStartSlot: 1,
		},
		{
			name: "should_get_epoch_data_for_epoch_1_with_config_data_from_epoch_1",
			service: func(ctrl *gomock.Controller) *Service {
				mockEpochState := NewMockEpochState(ctrl)

				mockEpochState.EXPECT().GetEpochDataRaw(uint64(1), nil).Return(testEpochData, nil)
				mockEpochState.EXPECT().GetConfigData(uint64(1), nil).Return(testConfigData, nil)

				return &Service{
					authority:  true,
					keypair:    kp,
					epochState: mockEpochState,
				}
			},
			epoch: 1,
			expected: &epochData{
				randomness:     testEpochData.Randomness,
				authorities:    testEpochData.Authorities,
				authorityIndex: 0,
				threshold:      threshold0,
			},
			expectedStartSlot: 201,
		},
		{
			name: "should_get_epoch_data_for_epoch_1_and_config_data_for_epoch_0",
			service: func(ctrl *gomock.Controller) *Service {
				mockEpochState := NewMockEpochState(ctrl)

				mockEpochState.EXPECT().GetEpochDataRaw(uint64(1), nil).Return(testEpochData, nil)
				mockEpochState.EXPECT().GetConfigData(uint64(1), nil).Return(testLatestConfigData, nil)

				return &Service{
					authority:  true,
					keypair:    kp,
					epochState: mockEpochState,
				}
			},
			epoch: 1,
			expected: &epochData{
				randomness:     testEpochData.Randomness,
				authorities:    testEpochData.Authorities,
				authorityIndex: 0,
				threshold:      threshold1,
			},
			expectedStartSlot: 201,
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			service := tt.service(ctrl)

			res, err := service.getEpochData(tt.epoch, nil)
			require.NoError(t, err)
			require.Equal(t, tt.expected, res)
		})
	}
}

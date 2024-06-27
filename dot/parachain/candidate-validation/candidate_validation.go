// Copyright 2024 ChainSafe Systems (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package candidatevalidation

import (
	"context"
	"errors"
	"fmt"
	"sync"

	parachainruntime "github.com/ChainSafe/gossamer/dot/parachain/runtime"
	parachaintypes "github.com/ChainSafe/gossamer/dot/parachain/types"
	"github.com/ChainSafe/gossamer/internal/log"
	"github.com/ChainSafe/gossamer/lib/common"
)

var logger = log.NewFromGlobal(log.AddContext("pkg", "parachain-candidate-validation"))

var (
	ErrValidationCodeMismatch   = errors.New("validation code hash does not match")
	ErrValidationInputOverLimit = errors.New("validation input is over the limit")
)

// CandidateValidation is a parachain subsystem that validates candidate parachain blocks
type CandidateValidation struct {
	wg       sync.WaitGroup
	stopChan chan struct{}

	SubsystemToOverseer chan<- any
	OverseerToSubsystem <-chan any
	ValidationHost      parachainruntime.ValidationHost
	RuntimeInstance     parachainruntime.RuntimeInstance
}

// NewCandidateValidation creates a new CandidateValidation subsystem
func NewCandidateValidation(overseerChan chan<- any) *CandidateValidation {
	candidateValidation := CandidateValidation{
		SubsystemToOverseer: overseerChan,
	}
	return &candidateValidation
}

// Run starts the CandidateValidation subsystem
func (cv *CandidateValidation) Run(context.Context, chan any, chan any) {
	cv.wg.Add(1)
	go cv.processMessages(&cv.wg)
}

// Name returns the name of the subsystem
func (*CandidateValidation) Name() parachaintypes.SubSystemName {
	return parachaintypes.CandidateValidation
}

// ProcessActiveLeavesUpdateSignal processes active leaves update signal
func (*CandidateValidation) ProcessActiveLeavesUpdateSignal(parachaintypes.ActiveLeavesUpdateSignal) error {
	// NOTE: this subsystem does not process active leaves update signal
	return nil
}

// ProcessBlockFinalizedSignal processes block finalized signal
func (*CandidateValidation) ProcessBlockFinalizedSignal(parachaintypes.BlockFinalizedSignal) error {
	// NOTE: this subsystem does not process block finalized signal
	return nil
}

// Stop stops the CandidateValidation subsystem
func (cv *CandidateValidation) Stop() {
	close(cv.stopChan)
	cv.wg.Wait()
}

// processMessages processes messages sent to the CandidateValidation subsystem
func (cv *CandidateValidation) processMessages(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case msg := <-cv.OverseerToSubsystem:
			logger.Debugf("received message %v", msg)
			switch msg := msg.(type) {
			case ValidateFromChainState:
				result, err := validateFromChainState(cv.RuntimeInstance, msg.Pov, msg.CandidateReceipt)
				if err != nil {
					logger.Errorf("failed to validate from chain state: %w", err)
					msg.Ch <- parachaintypes.OverseerFuncRes[ValidationResult]{
						Data: *result,
						Err:  err,
					}
				} else {
					msg.Ch <- parachaintypes.OverseerFuncRes[ValidationResult]{
						Data: *result,
					}
				}
			case ValidateFromExhaustive:
				result, err := validateFromExhaustive(cv.ValidationHost, msg.PersistedValidationData,
					msg.ValidationCode, msg.CandidateReceipt, msg.PoV)
				if err != nil {
					logger.Errorf("failed to validate from exhaustive: %w", err)
					msg.Ch <- parachaintypes.OverseerFuncRes[ValidationResult]{
						Data: *result,
						Err:  err,
					}
				} else {
					msg.Ch <- parachaintypes.OverseerFuncRes[ValidationResult]{
						Data: *result,
					}
				}

			case PreCheck:
				// TODO: implement functionality to handle PreCheck, see issue #3921

			case parachaintypes.ActiveLeavesUpdateSignal:
				_ = cv.ProcessActiveLeavesUpdateSignal(msg)

			case parachaintypes.BlockFinalizedSignal:
				_ = cv.ProcessBlockFinalizedSignal(msg)

			default:
				logger.Errorf("%w: %T", parachaintypes.ErrUnknownOverseerMessage, msg)
			}

		case <-cv.stopChan:
			return
		}
	}
}

// PoVRequestor gets proof of validity by issuing network requests to validators of the current backing group.
// TODO: Implement PoV requestor, issue #3919
type PoVRequestor interface {
	RequestPoV(povHash common.Hash) parachaintypes.PoV
}

// getValidationData gets validation data for a parachain block from the runtime instance
func getValidationData(runtimeInstance parachainruntime.RuntimeInstance, paraID uint32,
) (*parachaintypes.PersistedValidationData, *parachaintypes.ValidationCode, error) {

	var mergedError error

	for _, assumptionValue := range []any{
		parachaintypes.IncludedOccupiedCoreAssumption{},
		parachaintypes.TimedOutOccupiedCoreAssumption{},
		parachaintypes.Free{},
	} {
		assumption := parachaintypes.NewOccupiedCoreAssumption()
		err := assumption.SetValue(assumptionValue)
		if err != nil {
			return nil, nil, fmt.Errorf("getting assumption: %w", err)
		}
		persistedValidationData, err := runtimeInstance.ParachainHostPersistedValidationData(paraID, assumption)
		if err != nil {
			mergedError = errors.Join(mergedError, err)
			continue
		}

		validationCode, err := runtimeInstance.ParachainHostValidationCode(paraID, assumption)
		if err != nil {
			return nil, nil, fmt.Errorf("getting validation code: %w", err)
		}

		return persistedValidationData, validationCode, nil
	}

	return nil, nil, fmt.Errorf("getting persisted validation data: %w", mergedError)
}

// validateFromChainState validates a candidate parachain block with provided parameters using relay-chain
// state and using the parachain runtime.
func validateFromChainState(runtimeInstance parachainruntime.RuntimeInstance, pov parachaintypes.PoV,
	candidateReceipt parachaintypes.CandidateReceipt) (
	*ValidationResult, error) {

	persistedValidationData, validationCode, err := getValidationData(runtimeInstance,
		candidateReceipt.Descriptor.ParaID)
	if err != nil {
		return nil, fmt.Errorf("getting validation data: %w", err)
	}

	parachainRuntimeInstance, err := parachainruntime.SetupVM(*validationCode)
	if err != nil {
		return nil, fmt.Errorf("setting up VM: %w", err)
	}

	validationResults, err := validateFromExhaustive(parachainRuntimeInstance, *persistedValidationData,
		*validationCode,
		candidateReceipt, pov)

	return validationResults, err
}

// validateFromExhaustive validates a candidate parachain block with provided parameters
func validateFromExhaustive(validationHost parachainruntime.ValidationHost,
	persistedValidationData parachaintypes.PersistedValidationData,
	validationCode parachaintypes.ValidationCode,
	candidateReceipt parachaintypes.CandidateReceipt, pov parachaintypes.PoV) (
	*ValidationResult, error) {

	validationCodeHash := validationCode.Hash()
	// basic checks
	validationErr, internalErr := performBasicChecks(&candidateReceipt.Descriptor, persistedValidationData.MaxPovSize,
		pov,
		validationCodeHash)
	if validationErr != nil {
		validationResult := &ValidationResult{
			InvalidResult: validationErr,
		}
		return validationResult, internalErr
	}

	validationParams := parachainruntime.ValidationParameters{
		ParentHeadData:         persistedValidationData.ParentHead,
		BlockData:              pov.BlockData,
		RelayParentNumber:      persistedValidationData.RelayParentNumber,
		RelayParentStorageRoot: persistedValidationData.RelayParentStorageRoot,
	}

	validationResult, err := validationHost.ValidateBlock(validationParams)
	if err != nil {
		ci := ExecutionError
		return &ValidationResult{InvalidResult: &ci}, fmt.Errorf("executing validate_block: %w", err)
	}

	result := &ValidationResult{
		ValidResult: &ValidValidationResult{
			CandidateCommitments: parachaintypes.CandidateCommitments{
				UpwardMessages:            validationResult.UpwardMessages,
				HorizontalMessages:        validationResult.HorizontalMessages,
				NewValidationCode:         validationResult.NewValidationCode,
				HeadData:                  validationResult.HeadData,
				ProcessedDownwardMessages: validationResult.ProcessedDownwardMessages,
				HrmpWatermark:             validationResult.HrmpWatermark,
			},
			PersistedValidationData: persistedValidationData,
		},
	}
	return result, nil
}

// performBasicChecks Does basic checks of a candidate. Provide the encoded PoV-block.
// Returns CandidateInvalidity and internal error if any.
func performBasicChecks(candidate *parachaintypes.CandidateDescriptor, maxPoVSize uint32,
	pov parachaintypes.PoV, validationCodeHash parachaintypes.ValidationCodeHash) (validationError *CandidateInvalidity,
	internalError error) {
	povHash, err := pov.Hash()
	if err != nil {
		return nil, fmt.Errorf("hashing PoV: %w", err)
	}

	encodedPoV, err := pov.Encode()
	if err != nil {
		return nil, fmt.Errorf("encoding PoV: %w", err)
	}
	encodedPoVSize := uint32(len(encodedPoV))

	if encodedPoVSize > maxPoVSize {
		ci := ParamsTooLarge
		return &ci, nil
	}

	if povHash != candidate.PovHash {
		ci := PoVHashMismatch
		return &ci, nil
	}

	if validationCodeHash != candidate.ValidationCodeHash {
		ci := CodeHashMismatch
		return &ci, nil
	}

	err = candidate.CheckCollatorSignature()
	if err != nil {
		ci := BadSignature
		return &ci, nil
	}
	return nil, nil
}
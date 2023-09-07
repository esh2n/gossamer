// Copyright 2023 ChainSafe Systems (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package parachain

import (
	"fmt"

	parachaintypes "github.com/ChainSafe/gossamer/dot/parachain/types"
	"github.com/ChainSafe/gossamer/lib/common"
	"github.com/ChainSafe/gossamer/lib/crypto/sr25519"
	"github.com/ChainSafe/gossamer/lib/keystore"
	"github.com/ChainSafe/gossamer/pkg/scale"
)

// StatementDistributionMessage represents network messages used by the statement distribution subsystem
type StatementDistributionMessage scale.VaryingDataType

// NewStatementDistributionMessage returns a new statement distribution message varying data type
func NewStatementDistributionMessage() StatementDistributionMessage {
	vdt := scale.MustNewVaryingDataType(Statement{}, LargePayload{})
	return StatementDistributionMessage(vdt)
}

// New will enable scale to create new instance when needed
func (StatementDistributionMessage) New() StatementDistributionMessage {
	return NewStatementDistributionMessage()
}

// Set will set a value using the underlying  varying data type
func (sdm *StatementDistributionMessage) Set(val scale.VaryingDataTypeValue) (err error) {
	vdt := scale.VaryingDataType(*sdm)
	err = vdt.Set(val)
	if err != nil {
		return fmt.Errorf("setting value to varying data type: %w", err)
	}

	*sdm = StatementDistributionMessage(vdt)
	return nil
}

// Value returns the value from the underlying varying data type
func (sdm *StatementDistributionMessage) Value() (scale.VaryingDataTypeValue, error) {
	vdt := scale.VaryingDataType(*sdm)
	return vdt.Value()
}

// Statement represents a signed full statement under a given relay-parent.
type Statement struct {
	Hash                         common.Hash                  `scale:"1"`
	UncheckedSignedFullStatement UncheckedSignedFullStatement `scale:"2"`
}

// Index returns the index of varying data type
func (Statement) Index() uint {
	return 0
}

// LargePayload represents Seconded statement with large payload
// (e.g. containing a runtime upgrade).
//
// We only gossip the hash in that case, actual payloads can be fetched from sending node
// via request/response.
type LargePayload StatementMetadata

// Index returns the index of varying data type
func (LargePayload) Index() uint {
	return 1
}

// UncheckedSignedFullStatement is a Variant of `checkedSignedFullStatement` where the signature has not yet been verified.
type UncheckedSignedFullStatement struct {
	// The payload is part of the signed data. The rest is the signing context,
	// which is known both at signing and at validation.
	Payload StatementVDT `scale:"1"`

	// The index of the validator signing this statement.
	ValidatorIndex parachaintypes.ValidatorIndex `scale:"2"`

	// The signature by the validator of the signed payload.
	Signature ValidatorSignature `scale:"3"`
}

func (u *UncheckedSignedFullStatement) Sign(
	keystore keystore.Keystore,
	signingContext SigningContext,
	key parachaintypes.ValidatorID,
) (*ValidatorSignature, error) {
	encodedPayload, err := scale.Marshal(u.Payload)
	if err != nil {
		return nil, fmt.Errorf("marshalling payload: %w", err)
	}

	encodedSigningContext, err := scale.Marshal(signingContext)
	if err != nil {
		return nil, fmt.Errorf("marshalling signing context: %w", err)
	}

	data := append(encodedPayload, encodedSigningContext...)

	validatorPublicKey, err := sr25519.NewPublicKey(key[:])
	if err != nil {
		return nil, fmt.Errorf("getting public key: %w", err)
	}
	signatureBytes, err := keystore.GetKeypair(validatorPublicKey).Sign(data)
	if err != nil {
		return nil, fmt.Errorf("signing data: %w", err)
	}

	var signature Signature
	copy(signature[:], signatureBytes)
	valSign := ValidatorSignature(signature)

	return &valSign, nil
}

type CheckedSignedFullStatement UncheckedSignedFullStatement

func (u *CheckedSignedFullStatement) Sign(
	keystore keystore.Keystore,
	signingContext SigningContext,
	key parachaintypes.ValidatorID,
) (*ValidatorSignature, error) {

	uncheckedStatement := UncheckedSignedFullStatement{Payload: u.Payload}
	return uncheckedStatement.Sign(keystore, signingContext, key)
}

// StatementMetadata represents the data that makes a statement unique.
type StatementMetadata struct {
	// Relay parent this statement is relevant under.
	RelayParent common.Hash `scale:"1"`

	// Hash of the candidate that got validated.
	CandidateHash CandidateHash `scale:"2"`

	// Validator that attested the validity.
	SignedBy parachaintypes.ValidatorIndex `scale:"3"`

	// Signature of seconding validator.
	Signature ValidatorSignature `scale:"4"`
}

// ValidatorSignature represents the signature with which parachain validators sign blocks.
type ValidatorSignature Signature

// Signature represents a cryptographic signature.
type Signature [64]byte

// Copyright 2023 ChainSafe Systems (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package parachaintypes

import (
	"fmt"

	"github.com/ChainSafe/gossamer/lib/common"
	"github.com/ChainSafe/gossamer/lib/crypto/sr25519"
	"github.com/ChainSafe/gossamer/lib/keystore"
	"github.com/ChainSafe/gossamer/pkg/scale"
)

// StatementVDT is a result of candidate validation. It could be either `Valid` or `Seconded`.
type StatementVDT scale.VaryingDataType

// NewStatementVDT returns a new statement varying data type
func NewStatementVDT() StatementVDT {
	vdt := scale.MustNewVaryingDataType(Seconded{}, Valid{})
	return StatementVDT(vdt)
}

// New will enable scale to create new instance when needed
func (StatementVDT) New() StatementVDT {
	return NewStatementVDT()
}

// Set will set a value using the underlying  varying data type
func (s *StatementVDT) Set(val scale.VaryingDataTypeValue) (err error) {
	vdt := scale.VaryingDataType(*s)
	err = vdt.Set(val)
	if err != nil {
		return fmt.Errorf("setting value to varying data type: %w", err)
	}

	*s = StatementVDT(vdt)
	return nil
}

// Value returns the value from the underlying varying data type
func (s *StatementVDT) Value() (scale.VaryingDataTypeValue, error) {
	vdt := scale.VaryingDataType(*s)
	return vdt.Value()
}

// Seconded represents a statement that a validator seconds a candidate.
type Seconded CommittedCandidateReceipt

// Index returns the index of varying data type
func (Seconded) Index() uint {
	return 1
}

// Valid represents a statement that a validator has deemed a candidate valid.
type Valid CandidateHash

// Index returns the index of varying data type
func (Valid) Index() uint {
	return 2
}

func (s *StatementVDT) Sign(
	keystore keystore.Keystore,
	signingContext SigningContext,
	key ValidatorID,
) (*ValidatorSignature, error) {
	encodedData, err := scale.Marshal(*s)
	if err != nil {
		return nil, fmt.Errorf("marshalling payload: %w", err)
	}

	encodedSigningContext, err := scale.Marshal(signingContext)
	if err != nil {
		return nil, fmt.Errorf("marshalling signing context: %w", err)
	}

	encodedData = append(encodedData, encodedSigningContext...)

	validatorPublicKey, err := sr25519.NewPublicKey(key[:])
	if err != nil {
		return nil, fmt.Errorf("getting public key: %w", err)
	}

	signatureBytes, err := keystore.GetKeypair(validatorPublicKey).Sign(encodedData)
	if err != nil {
		return nil, fmt.Errorf("signing data: %w", err)
	}

	var signature Signature
	copy(signature[:], signatureBytes)
	valSign := ValidatorSignature(signature)
	return &valSign, nil
}

// UncheckedSignedFullStatement is a Variant of `SignedFullStatement` where the signature has not yet been verified.
type UncheckedSignedFullStatement struct {
	// The payload is part of the signed data. The rest is the signing context,
	// which is known both at signing and at validation.
	Payload StatementVDT `scale:"1"`

	// The index of the validator signing this statement.
	ValidatorIndex ValidatorIndex `scale:"2"`

	// The signature by the validator of the signed payload.
	Signature ValidatorSignature `scale:"3"`
}

// SigningContext is a type returned by runtime with current session index and a parent hash.
type SigningContext struct {
	/// current session index.
	SessionIndex SessionIndex `scale:"1"`
	/// hash of the parent.
	ParentHash common.Hash `scale:"2"`
}

// SignedFullStatement represents a statement along with its corresponding signature
// and the index of the sender. The signing context and validator set should be
// apparent from context. This statement is "full" as the `Seconded` variant includes
// the candidate receipt. Only the compact `SignedStatement` is suitable for submission
// to the chain.
type SignedFullStatement UncheckedSignedFullStatement

// SignedFullStatementWithPVD represents a signed full statement along with associated Persisted Validation Data (PVD).
type SignedFullStatementWithPVD struct {
	SignedFullStatement     SignedFullStatement
	PersistedValidationData *PersistedValidationData
}
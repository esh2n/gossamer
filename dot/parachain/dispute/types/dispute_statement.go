package types

import (
	"fmt"

	"github.com/ChainSafe/gossamer/lib/common"
	"github.com/ChainSafe/gossamer/lib/crypto/sr25519"
	"github.com/ChainSafe/gossamer/lib/keystore"

	parachainTypes "github.com/ChainSafe/gossamer/dot/parachain/types"
	"github.com/ChainSafe/gossamer/lib/babe/inherents"
	"github.com/ChainSafe/gossamer/pkg/scale"
)

// SecondedCompactStatement is the proposal of a parachain candidate.
type SecondedCompactStatement struct {
	CandidateHash common.Hash
}

// Index returns the index of the type SecondedCompactStatement.
func (SecondedCompactStatement) Index() uint {
	return 0
}

// ValidCompactStatement represents a valid candidate.
type ValidCompactStatement struct {
	CandidateHash common.Hash
}

// Index returns the index of the type ValidCompactStatement.
func (ValidCompactStatement) Index() uint {
	return 1
}

// CompactStatementVDT is the statement that can be made about parachain candidates
// These are the actual values that are signed.
type CompactStatementVDT scale.VaryingDataType

// Set will set a VaryingDataTypeValue using the underlying VaryingDataType
func (cs *CompactStatementVDT) Set(val scale.VaryingDataTypeValue) (err error) {
	vdt := scale.VaryingDataType(*cs)
	err = vdt.Set(val)
	if err != nil {
		return fmt.Errorf("setting value to varying data type: %w", err)
	}
	*cs = CompactStatementVDT(vdt)
	return nil
}

// Value returns the value from the underlying VaryingDataType
func (cs *CompactStatementVDT) Value() (val scale.VaryingDataTypeValue, err error) {
	vdt := scale.VaryingDataType(*cs)
	val, err = vdt.Value()
	if err != nil {
		return nil, fmt.Errorf("getting value from varying data type: %w", err)
	}
	return val, nil
}

type SigningContext struct {
	SessionIndex  parachainTypes.SessionIndex
	CandidateHash common.Hash
}

func (cs *CompactStatementVDT) SigningPayload(signingContext SigningContext) ([]byte, error) {
	// scale encode the compact statement
	encodedStatement, err := scale.Marshal(*cs)
	if err != nil {
		return nil, fmt.Errorf("encode compact statement: %w", err)
	}

	// scale encode the signing context
	encodedSigningContext, err := scale.Marshal(signingContext)
	if err != nil {
		return nil, fmt.Errorf("encode signing context: %w", err)
	}

	// concatenate the encoded statement and encoded signing context
	encoded := append(encodedStatement, encodedSigningContext...)
	return encoded, nil
}

// NewCompactStatement creates a new CompactStatementVDT.
func NewCompactStatement() CompactStatementVDT {
	vdt := scale.MustNewVaryingDataType(ValidCompactStatement{}, SecondedCompactStatement{})
	return CompactStatementVDT(vdt)
}

type CompactStatementKind uint

const (
	SecondedCompactStatementKind CompactStatementKind = iota
	ValidCompactStatementKind
)

func NewCustomCompactStatement(kind CompactStatementKind,
	candidateHash common.Hash,
) (CompactStatementVDT, error) {
	vdt := NewCompactStatement()

	var err error
	switch kind {
	case SecondedCompactStatementKind:
		err = vdt.Set(SecondedCompactStatement{
			CandidateHash: candidateHash,
		})
	case ValidCompactStatementKind:
		err = vdt.Set(ValidCompactStatement{
			CandidateHash: candidateHash,
		})
	default:
		return CompactStatementVDT{}, fmt.Errorf("invalid compact statement kind")
	}

	if err != nil {
		return CompactStatementVDT{}, fmt.Errorf("set compact statement: %w", err)
	}

	return vdt, nil
}

func NewCompactStatementFromAttestation(
	attestation inherents.ValidityAttestation,
	candidateHash common.Hash,
) (CompactStatementVDT, error) {
	compactStatementVDT := NewCompactStatement()

	attestationVDT := scale.VaryingDataType(attestation)
	val, err := attestationVDT.Value()
	if err != nil {
		return CompactStatementVDT{}, fmt.Errorf("get attestation value: %w", err)
	}

	switch val.(type) {
	case inherents.Implicit:
		err = compactStatementVDT.Set(ValidCompactStatement{
			CandidateHash: candidateHash,
		})
	case inherents.Explicit:
		err = compactStatementVDT.Set(SecondedCompactStatement{
			CandidateHash: candidateHash,
		})
	default:
		return CompactStatementVDT{}, fmt.Errorf("invalid compact statement kind")
	}
	return compactStatementVDT, err
}

// ExplicitDisputeStatement An explicit statement on a candidate issued as part of a dispute.
type ExplicitDisputeStatement struct {
	Valid         bool
	CandidateHash common.Hash
	Session       parachainTypes.SessionIndex
}

func (eds ExplicitDisputeStatement) SigningPayload() ([]byte, error) {
	const magic = "DISP"
	var magicBytes [4]byte
	copy(magicBytes[:], magic)

	encoded, err := scale.Marshal(eds)
	if err != nil {
		return nil, fmt.Errorf("marshal ExplicitDisputeStatement")
	}

	// how to return
	payload := append(magicBytes[:], encoded...)
	return payload, nil
}

// ApprovalVote A vote of approval on a candidate.
type ApprovalVote struct {
	CandidateHash common.Hash
}

func (a *ApprovalVote) SigningPayload(session parachainTypes.SessionIndex) ([]byte, error) {
	const magic = "APPR"
	var magicBytes [4]byte
	copy(magicBytes[:], magic)

	encodedVote, err := scale.Marshal(a)
	if err != nil {
		return nil, fmt.Errorf("marshal ApprovalVote")
	}

	encodedSession, err := scale.Marshal(session)
	if err != nil {
		return nil, fmt.Errorf("marshal session")
	}

	// how to return
	payload := append(magicBytes[:], encodedVote...)
	payload = append(payload, encodedSession...)
	return payload, nil
}

// SignedDisputeStatement A checked dispute statement from an associated validator.
type SignedDisputeStatement struct {
	DisputeStatement   inherents.DisputeStatement
	CandidateHash      common.Hash
	ValidatorPublic    parachainTypes.ValidatorID
	ValidatorSignature parachainTypes.ValidatorSignature
	SessionIndex       parachainTypes.SessionIndex
}

func NewSignedDisputeStatement(
	keypair keystore.KeyPair,
	valid bool,
	candidateHash common.Hash,
	sessionIndex parachainTypes.SessionIndex,
) (SignedDisputeStatement, error) {
	disputeStatement := inherents.NewDisputeStatement()
	if valid {
		validVdt := inherents.NewValidDisputeStatementKind()
		if err := validVdt.Set(inherents.ExplicitValidDisputeStatementKind{}); err != nil {
			return SignedDisputeStatement{}, fmt.Errorf("set dispute statement: %w", err)
		}
		if err := disputeStatement.Set(validVdt); err != nil {
			return SignedDisputeStatement{}, fmt.Errorf("set dispute statement: %w", err)
		}
	} else {
		invalidVdt := inherents.NewInvalidDisputeStatementKind()
		if err := invalidVdt.Set(inherents.ExplicitInvalidDisputeStatementKind{}); err != nil {
			return SignedDisputeStatement{}, fmt.Errorf("set dispute statement: %w", err)
		}
		if err := disputeStatement.Set(invalidVdt); err != nil {
			return SignedDisputeStatement{}, fmt.Errorf("set dispute statement: %w", err)
		}
	}

	payload, err := getDisputeStatementSigningPayload(disputeStatement, candidateHash, sessionIndex)
	if err != nil {
		return SignedDisputeStatement{}, fmt.Errorf("get dispute statement signing payload: %w", err)
	}

	signature, err := keypair.Sign(payload)
	if err != nil {
		return SignedDisputeStatement{}, fmt.Errorf("sign payload: %w", err)
	}

	return SignedDisputeStatement{
		DisputeStatement:   disputeStatement,
		CandidateHash:      candidateHash,
		ValidatorPublic:    parachainTypes.ValidatorID(keypair.Public().Encode()),
		ValidatorSignature: parachainTypes.ValidatorSignature(signature),
		SessionIndex:       sessionIndex,
	}, nil
}

func NewCheckedSignedDisputeStatement(disputeStatement inherents.DisputeStatement,
	candidateHash common.Hash,
	sessionIndex parachainTypes.SessionIndex,
	validatorSignature parachainTypes.ValidatorSignature,
	validatorID parachainTypes.ValidatorID,
) (*SignedDisputeStatement, error) {
	if err := VerifyDisputeStatement(disputeStatement,
		candidateHash,
		sessionIndex,
		validatorSignature,
		validatorID,
	); err != nil {
		return nil, fmt.Errorf("verify dispute statement: %w", err)
	}

	return &SignedDisputeStatement{
		DisputeStatement:   disputeStatement,
		CandidateHash:      candidateHash,
		ValidatorPublic:    validatorID,
		ValidatorSignature: validatorSignature,
		SessionIndex:       sessionIndex,
	}, nil
}

func NewSignedDisputeStatementFromBackingStatement(backingStatement CompactStatementVDT,
	signingContext SigningContext,
	keypair keystore.KeyPair,
) (SignedDisputeStatement, error) {
	statementKind, err := backingStatement.Value()
	if err != nil {
		return SignedDisputeStatement{}, fmt.Errorf("get backing statement value: %w", err)
	}

	disputeStatement := inherents.NewDisputeStatement()
	validStatement := inherents.NewValidDisputeStatementKind()
	switch statement := statementKind.(type) {
	case SecondedCompactStatement:
		if err := validStatement.Set(inherents.BackingSeconded(statement.CandidateHash)); err != nil {
			return SignedDisputeStatement{}, fmt.Errorf("set dispute statement: %w", err)
		}
	case ValidCompactStatement:
		if err := validStatement.Set(inherents.BackingValid(statement.CandidateHash)); err != nil {
			return SignedDisputeStatement{}, fmt.Errorf("set dispute statement: %w", err)
		}
	default:
		return SignedDisputeStatement{}, fmt.Errorf("invalid backing statement kind")
	}

	if err := disputeStatement.Set(validStatement); err != nil {
		return SignedDisputeStatement{}, fmt.Errorf("set dispute statement: %w", err)
	}

	return NewSignedDisputeStatement(keypair,
		true,
		signingContext.CandidateHash,
		signingContext.SessionIndex,
	)
}

func VerifyDisputeStatement(
	disputeStatement inherents.DisputeStatement,
	candidateHash common.Hash,
	sessionIndex parachainTypes.SessionIndex,
	validatorSignature parachainTypes.ValidatorSignature,
	validatorID parachainTypes.ValidatorID,
) error {
	payload, err := getDisputeStatementSigningPayload(disputeStatement, candidateHash, sessionIndex)
	if err != nil {
		return fmt.Errorf("get dispute statement signing payload: %w", err)
	}

	validatorPublic, err := sr25519.NewPublicKey(validatorID[:])
	if err != nil {
		return fmt.Errorf("new public key: %w", err)
	}

	if ok, err := validatorPublic.Verify(payload, validatorSignature[:]); !ok || err != nil {
		return fmt.Errorf("verify dispute statement: %w", err)
	}
	return nil
}

func getDisputeStatementSigningPayload(
	disputeStatement inherents.DisputeStatement,
	candidateHash common.Hash,
	session parachainTypes.SessionIndex,
) ([]byte, error) {
	statement, err := disputeStatement.Value()
	if err != nil {
		return nil, fmt.Errorf("failed to get dispute statement value: %w", err)
	}

	var payload []byte
	switch statement.(type) {
	case inherents.ValidDisputeStatementKind:
		validStatement := statement.(inherents.ValidDisputeStatementKind)
		validValue, err := validStatement.Value()
		if err != nil {
			return nil, fmt.Errorf("get valid dispute statement value: %w", err)
		}
		switch validValue.(type) {
		case inherents.ExplicitValidDisputeStatementKind:
			data := ExplicitDisputeStatement{
				Valid:         true,
				CandidateHash: candidateHash,
				Session:       session,
			}
			payload, err = data.SigningPayload()
			if err != nil {
				return nil, fmt.Errorf("get signing payload: %w", err)
			}
		case inherents.BackingSeconded:
			data, err := NewCustomCompactStatement(SecondedCompactStatementKind, candidateHash)
			if err != nil {
				return nil, fmt.Errorf("new custom compact statement: %w", err)
			}

			signingContext := SigningContext{
				SessionIndex:  session,
				CandidateHash: candidateHash,
			}
			payload, err = data.SigningPayload(signingContext)
			if err != nil {
				return nil, fmt.Errorf("get signing payload: %w", err)
			}
		case inherents.BackingValid:
			data, err := NewCustomCompactStatement(ValidCompactStatementKind, candidateHash)
			if err != nil {
				return nil, fmt.Errorf("new custom compact statement: %w", err)
			}

			signingContext := SigningContext{
				SessionIndex:  session,
				CandidateHash: candidateHash,
			}
			payload, err = data.SigningPayload(signingContext)
			if err != nil {
				return nil, fmt.Errorf("get signing payload: %w", err)
			}
		case inherents.ApprovalChecking:
			data := ApprovalVote{
				CandidateHash: candidateHash,
			}
			payload, err = data.SigningPayload(session)
			if err != nil {
				return nil, fmt.Errorf("get signing payload: %w", err)
			}
		default:
			return nil, fmt.Errorf("invalid dispute statement kind %T", validStatement)
		}
	case inherents.InvalidDisputeStatementKind:
		invalidStatement := statement.(inherents.InvalidDisputeStatementKind)
		invalidValue, err := invalidStatement.Value()
		if err != nil {
			return nil, fmt.Errorf("get invalid dispute statement value: %w", err)
		}

		switch invalidValue.(type) {
		case inherents.ExplicitInvalidDisputeStatementKind:
			data := ExplicitDisputeStatement{
				Valid:         false,
				CandidateHash: candidateHash,
				Session:       session,
			}
			payload, err = data.SigningPayload()
			if err != nil {
				return nil, fmt.Errorf("get signing payload: %w", err)
			}
		default:
			return nil, fmt.Errorf("invalid dispute statement kind %T", invalidStatement)
		}

	default:
		return nil, fmt.Errorf("invalid dispute statement kind %T", statement)
	}

	return payload, nil
}

// Statement is the statement that can be made about parachain candidates.
type Statement struct {
	SignedDisputeStatement SignedDisputeStatement
	ValidatorIndex         parachainTypes.ValidatorIndex
}

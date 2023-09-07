// Copyright 2023 ChainSafe Systems (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package grandpa

import (
	"time"

	"github.com/ChainSafe/gossamer/client/network/config"
	"github.com/ChainSafe/gossamer/internal/log"
	"github.com/ChainSafe/gossamer/keystore"
	"github.com/ChainSafe/gossamer/lib/grandpa"
	finalityGrandpa "github.com/ChainSafe/gossamer/pkg/finality-grandpa"
	"github.com/ChainSafe/gossamer/pkg/scale"
	"golang.org/x/exp/constraints"
)

var logger = log.NewFromGlobal(log.AddContext("consensus", "grandpa"))

// GrandpaEngineID is the hard-coded grandpa ID
var GrandpaEngineID = ConsensusEngineID{'F', 'R', 'N', 'K'}

type AuthorityID interface {
	constraints.Ordered
	Verify(msg []byte, sig []byte) (bool, error)
}

type AuthoritySignature any

// Authority represents a grandpa authority
type Authority[ID AuthorityID] struct {
	Key    ID
	Weight uint64
}

type AuthorityList[ID AuthorityID] []Authority[ID]

// NewAuthoritySetStruct A new authority set along with the canonical block it changed at.
type NewAuthoritySetStruct[H comparable, N constraints.Unsigned, ID AuthorityID] struct {
	CanonNumber N
	CanonHash   H
	SetId       N
	Authorities []Authority[ID]
}

type messageData[H comparable, N constraints.Unsigned] struct {
	Round   uint64
	SetID   uint64
	Message finalityGrandpa.Message[H, N]
}

// Check a message signature by encoding the message as a localised payload and
// verifying the provided signature using the expected authority id.
// The encoding necessary to verify the signature will be done using the given
// buffer, the original content of the buffer will be cleared.
func checkMessageSignature[H comparable, N constraints.Unsigned, ID AuthorityID](
	message finalityGrandpa.Message[H, N],
	id ID,
	signature any,
	round uint64,
	setID uint64) (bool, error) {

	sig, ok := signature.([]byte)

	// Verify takes []byte, but string is a valid signature type,
	// so if signature is not already type []byte, check if it is a string
	sigString, okString := signature.(string)
	if !okString && !ok {
		sig = []byte(sigString)
	}

	m := messageData[H, N]{
		round,
		setID,
		message,
	}

	enc, err := scale.Marshal(m)
	if err != nil {
		return false, err
	}
	valid, err := id.Verify(enc, sig[:])
	if err != nil {
		return false, err
	}
	return valid, nil
}

type ClientForGrandpa interface{}

type Backend interface{}

type Config struct {
	/// The expected duration for a message to be gossiped across the network.
	GossipDuration time.Duration
	/// Justification generation period (in blocks). GRANDPA will try to generate justifications
	/// at least every justification_period blocks. There are some other events which might cause
	/// justification generation.
	JustificationPeriod uint32
	/// Whether the GRANDPA observer protocol is live on the network and thereby
	/// a full-node not running as a validator is running the GRANDPA observer
	/// protocol (we will only issue catch-up requests to authorities when the
	/// observer protocol is enabled).
	ObserverEnabled bool
	/// The role of the local node (i.e. authority, full-node or light).
	LocalRole config.Role
	/// Some local identifier of the voter.
	Name *string
	/// The keystore that manages the keys of this node.
	KeyStore *keystore.KeyStore
	/// TelemetryHandle instance.
	Telemetry *TelemetryHandle
	/// Chain specific GRANDPA protocol name. See [`crate::protocol_standard_name`].
	ProtocolName ProtocolName
}

type VoterWork[Hash constraints.Ordered, Number constraints.Unsigned, Signature comparable, ID constraints.Ordered] struct {
	voter            *grandpa.Voter[Hash, Number, Signature, ID]
	sharedVoterState any
	env              any
	voterCommandsRx  any
	network          any
	telemetry        any
	metrics          any
}

func NewVoterWork[Hash constraints.Ordered, Number constraints.Unsigned, Signature comparable, ID constraints.Ordered](
	client ClientForGrandpa,
	config Config,
	network NetworkBridge,
	selectChain SelectChain,
	votingRule VotingRule,
	persistendData persistentData,
	voterCommandsRX any,
	prometheusRegistry any,
	sharedVoterState SharedVoterState,
	JustificationSender GrandpaJustificationSender,
	telemetry TelemetryHandle,
) *VoterWork[Hash, Number, Signature, ID] {
	// grandpa.NewVoter[]()
	return nil
}

package networkbridge

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/ChainSafe/gossamer/dot/network"
	"github.com/ChainSafe/gossamer/lib/common"
	"github.com/ChainSafe/gossamer/pkg/scale"
)

type WireMessage scale.VaryingDataType

// NewWireMessage returns a new WireMessage varying data type
func NewWireMessage() WireMessage {
	vdt := scale.MustNewVaryingDataType(ProtocolMessage{}, ViewUpdate{})
	return WireMessage(vdt)
}

// New will enable scale to create new instance when needed
func (WireMessage) New() WireMessage {
	return NewWireMessage()
}

// Set will set a value using the underlying  varying data type
func (w *WireMessage) Set(val scale.VaryingDataTypeValue) (err error) {
	vdt := scale.VaryingDataType(*w)
	err = vdt.Set(val)
	if err != nil {
		return
	}
	*w = WireMessage(vdt)
	return
}

// Value returns the value from the underlying varying data type
func (w *WireMessage) Value() (val scale.VaryingDataTypeValue, err error) {
	vdt := scale.VaryingDataType(*w)
	return vdt.Value()
}

func (w WireMessage) Type() network.MessageType {
	// TODO: create a wire message type and return that
	return network.CollationMsgType
}

func (w WireMessage) Hash() (common.Hash, error) {
	// scale encode each extrinsic
	encMsg, err := w.Encode()
	if err != nil {
		return common.Hash{}, fmt.Errorf("cannot encode message: %w", err)
	}

	return common.Blake2bHash(encMsg)
}

// Encode a collator protocol message using scale encode
func (w WireMessage) Encode() ([]byte, error) {
	enc, err := scale.Marshal(w)
	if err != nil {
		return nil, err
	}
	return enc, nil
}

type ViewUpdate View

// View is a succinct representation of a peer's view. This consists of a bounded amount of chain heads
// and the highest known finalized block number.
//
// Up to `N` (5?) chain heads.
type View struct {
	// a bounded amount of chain heads
	heads []common.Hash
	// the highest known finalized number
	finalizedNumber uint32
}

type SortableHeads []common.Hash

func (s SortableHeads) Len() int {
	return len(s)
}

func (s SortableHeads) Less(i, j int) bool {
	return s[i].String() > s[j].String()
}

func (s SortableHeads) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// checkHeadsEqual checks if the heads of the view are equal to the heads of the other view.
func (v View) checkHeadsEqual(other View) bool {
	if len(v.heads) != len(other.heads) {
		return false
	}

	localHeads := v.heads
	sort.Sort(SortableHeads(localHeads))
	otherHeads := other.heads
	sort.Sort(SortableHeads(otherHeads))

	return reflect.DeepEqual(localHeads, otherHeads)
}

// Index returns the index of varying data type
func (ViewUpdate) Index() uint {
	return 2
}

type ProtocolMessage struct{}

// Index returns the index of varying data type
func (ProtocolMessage) Index() uint {
	return 1
}
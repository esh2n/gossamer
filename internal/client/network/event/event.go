package event

import (
	"github.com/ChainSafe/gossamer/internal/client/network"
	"github.com/ChainSafe/gossamer/internal/client/network/role"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	libp2p "github.com/libp2p/go-libp2p/core"
)

// / Event generated by DHT as a response to get_value and put_value requests.
type DHTEvent any

// / Events generated by DHT as a response to get_value and put_value requests.
type DHTEvents interface {
	ValueFound | ValuePut | ValuePutFailed
}

// /// The value was found.
type ValueFound []struct {
	Key   dht.KeyKadID
	Value []byte
}

// /// The record has been successfully inserted into the DHT.
type ValuePut dht.KeyKadID

// /// An error has occurred while putting a record into the DHT.
type ValuePutFailed dht.KeyKadID

// / Events generated by networking layer.
type Events interface {
	DHT | NotificationStreamOpened | NotificationStreamClosed | NotificationsReceived
}

// / Type for events generated by networking layer.
type Event any

// /// Event generated by a DHT.
type DHT DHTEvent

// /// Opened a substream with the given node with the given notifications protocol.
// ///
// /// The protocol is always one of the notification protocols that have been registered.
type NotificationStreamOpened struct {
	/// Node we opened the substream with.
	remote libp2p.PeerID
	/// The concerned protocol. Each protocol uses a different substream.
	// 		/// This is always equal to the value of
	// 		/// `sc_network::config::NonDefaultSetConfig::notifications_protocol` of one of the
	// 		/// configured sets.
	// 		protocol: ProtocolName,
	protocol network.ProtocolName
	/// If the negotiation didn't use the main name of the protocol (the one in
	// 		/// `notifications_protocol`), then this field contains which name has actually been
	// 		/// used.
	// 		/// Always contains a value equal to the value in
	// 		/// `sc_network::config::NonDefaultSetConfig::fallback_names`.
	// 		negotiated_fallback: Option<ProtocolName>,
	negotiatedFallback *network.ProtocolName
	// 		/// Role of the remote.
	// 		role: ObservedRole,
	role role.ObservedRole
	// 		/// Received handshake.
	// 		received_handshake: Vec<u8>,
	receivedHandshake []byte
}

// /// Closed a substream with the given node. Always matches a corresponding previous
// /// `NotificationStreamOpened` message.
type NotificationStreamClosed struct {
	/// Node we closed the substream with.
	remote libp2p.PeerID
	/// The concerned protocol. Each protocol uses a different substream.
	protocol network.ProtocolName
}

// /// Received one or more messages from the given node using the given protocol.
type NotificationsReceived struct {
	/// Node we received the message from.
	remote libp2p.PeerID
	/// Concerned protocol and associated message.
	messages []struct {
		network.ProtocolName
		Bytes []byte
	}
}
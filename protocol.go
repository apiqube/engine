package engine

import "github.com/apiqube/engine/internal/wire"

// Protocol identifies a target protocol, derived from the target URL scheme
// or declared explicitly by a plugin.
//
// Engine treats Protocol as an open set — plugins may register new protocols
// at load time. The constants below cover common cases but are not exhaustive.
type Protocol = wire.Protocol

const (
	ProtocolHTTP    = wire.ProtocolHTTP
	ProtocolHTTPS   = wire.ProtocolHTTPS
	ProtocolGRPC    = wire.ProtocolGRPC
	ProtocolGRPCS   = wire.ProtocolGRPCS
	ProtocolWS      = wire.ProtocolWS
	ProtocolWSS     = wire.ProtocolWSS
	ProtocolGraphQL = wire.ProtocolGraphQL
	ProtocolSQL     = wire.ProtocolSQL
	ProtocolKafka   = wire.ProtocolKafka
	ProtocolAMQP    = wire.ProtocolAMQP
	ProtocolRedis   = wire.ProtocolRedis
)

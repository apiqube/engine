package engine

// Protocol identifies a target protocol, derived from the target URL scheme
// or declared explicitly by a plugin.
//
// Engine treats Protocol as an open set — plugins may register new protocols
// at load time. The constants below cover the common cases but are not exhaustive.
type Protocol string

const (
	ProtocolHTTP    Protocol = "http"
	ProtocolHTTPS   Protocol = "https"
	ProtocolGRPC    Protocol = "grpc"
	ProtocolGRPCS   Protocol = "grpcs"
	ProtocolWS      Protocol = "ws"
	ProtocolWSS     Protocol = "wss"
	ProtocolGraphQL Protocol = "graphql"
	ProtocolSQL     Protocol = "sql"
	ProtocolKafka   Protocol = "kafka"
	ProtocolAMQP    Protocol = "amqp"
	ProtocolRedis   Protocol = "redis"
)

// String returns the protocol as a plain string.
func (p Protocol) String() string { return string(p) }

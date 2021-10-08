package tracing

const (
	traceContextKey      = "grpc-trace-bin"
	clientSpanNamePrefix = "gRPC.client"
	serverSpanNamePrefix = "gRPC.server"

	clientAttributeKey               = "client"
	failFastAttributeKey             = "fail_fast"
	beginAtUTCAttributeKey           = "begin_at_utc"
	receivedAtUTCAttributeKey        = "received_at_utc"
	sentTimeUTCAttributeKey          = "sent_time_utc"
	payloadAttributeKey              = "payload"
	uncompressedByteSizeAttributeKey = "uncompressed_byte_size"
	compressedByteSizeAttributeKey   = "compressed_byte_size"

	StandardPayloadAttributeSizeLimit = 256
	payloadTruncatedMessage           = "...[payload has been truncated]"
)

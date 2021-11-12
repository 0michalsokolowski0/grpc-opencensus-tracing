package tracing

const (
	traceContextKey      = "grpc-trace-bin"
	clientSpanNamePrefix = "gRPC.client"
	serverSpanNamePrefix = "gRPC.server"

	clientAttributeKey               = "client"
	failFastAttributeKey             = "fail_fast"
	payloadAttributeKey              = "payload"
	uncompressedByteSizeAttributeKey = "uncompressed_byte_size"
	compressedByteSizeAttributeKey   = "compressed_byte_size"

	StandardPayloadAttributeSizeLimit = 256
	payloadTruncatedMessage           = "...[payload has been truncated]"
)

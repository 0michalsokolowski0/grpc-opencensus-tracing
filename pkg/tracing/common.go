package tracing

import (
	"context"
	"fmt"
	"go.opencensus.io/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/status"
	"math"
	"math/rand"
)

func traceHandleRPC(ctx context.Context, rs stats.RPCStats, payloadAttributeLengthLimit int) {
	span := trace.FromContext(ctx)
	switch rs := rs.(type) {
	case *stats.Begin:
		span.AddAttributes(
			trace.BoolAttribute(clientAttributeKey, rs.Client),
			trace.BoolAttribute(failFastAttributeKey, rs.FailFast),
			trace.StringAttribute(beginAtUTCAttributeKey, rs.BeginTime.UTC().String()),
		)
	case *stats.InPayload:
		payload := interfaceToString(rs.Payload)
		if payloadAttributeLengthLimit > 0 {
			payload = truncate(payload, payloadAttributeLengthLimit)
		}

		uncompressedByteSize := int64(rs.Length)
		compressedByteSize := int64(rs.WireLength)

		span.AddAttributes(
			trace.StringAttribute(payloadAttributeKey, payload),
			trace.Int64Attribute(uncompressedByteSizeAttributeKey, uncompressedByteSize),
			trace.Int64Attribute(compressedByteSizeAttributeKey, compressedByteSize),
			trace.StringAttribute(receivedAtUTCAttributeKey, rs.RecvTime.UTC().String()),
		)
		span.AddMessageReceiveEvent(generateEventID(), uncompressedByteSize, compressedByteSize)
	case *stats.OutPayload:
		payload := interfaceToString(rs.Payload)
		if payloadAttributeLengthLimit > 0 {
			payload = truncate(payload, payloadAttributeLengthLimit)
		}

		uncompressedByteSize := int64(rs.Length)
		compressedByteSize := int64(rs.WireLength)

		span.AddAttributes(
			trace.StringAttribute(payloadAttributeKey, payload),
			trace.Int64Attribute(uncompressedByteSizeAttributeKey, uncompressedByteSize),
			trace.Int64Attribute(compressedByteSizeAttributeKey, compressedByteSize),
			trace.StringAttribute(sentTimeUTCAttributeKey, rs.SentTime.UTC().String()),
		)
		span.AddMessageSendEvent(generateEventID(), uncompressedByteSize, compressedByteSize)
	case *stats.End:
		if rs.Error != nil {
			s, ok := status.FromError(rs.Error)
			if ok {
				span.SetStatus(trace.Status{Code: int32(s.Code()), Message: s.Message()})
			} else {
				span.SetStatus(trace.Status{Code: int32(codes.Internal), Message: rs.Error.Error()})
			}
		}
		span.End()
	}
}

func generateEventID() int64 {
	return rand.Int63n(math.MaxInt64)
}

func interfaceToString(i interface{}) string {
	return fmt.Sprintf("%+v", i)
}

func truncate(s string, limit int) string {
	return string(s[:limit-len(payloadTruncatedMessage)+1]) + payloadTruncatedMessage
}

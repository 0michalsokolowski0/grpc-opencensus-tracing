// Copyright 2018, OpenCensus Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tracing

import (
	"fmt"
	"math"
	"math/rand"

	"go.opencensus.io/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/status"
)

func traceHandleRPC(span trace.SpanInterface, rs stats.RPCStats, payloadAttributeLengthLimit int) {
	switch rs := rs.(type) {
	case *stats.Begin:
		span.AddAttributes(
			trace.BoolAttribute(clientAttributeKey, rs.Client),
			trace.BoolAttribute(failFastAttributeKey, rs.FailFast),
		)
	case *stats.InPayload:
		payload := interfaceToString(rs.Payload)
		if payloadAttributeLengthLimit > 0 && payloadAttributeLengthLimit < len(payload) {
			payload = truncate(payload, payloadAttributeLengthLimit)
		}

		uncompressedByteSize := int64(rs.Length)
		compressedByteSize := int64(rs.WireLength)

		span.AddAttributes(
			trace.StringAttribute(payloadAttributeKey, payload),
			trace.Int64Attribute(uncompressedByteSizeAttributeKey, uncompressedByteSize),
			trace.Int64Attribute(compressedByteSizeAttributeKey, compressedByteSize),
		)
		span.AddMessageReceiveEvent(generateEventID(), uncompressedByteSize, compressedByteSize)
	case *stats.OutPayload:
		payload := interfaceToString(rs.Payload)
		if payloadAttributeLengthLimit > 0 && payloadAttributeLengthLimit < len(payload) {
			payload = truncate(payload, payloadAttributeLengthLimit)
		}

		uncompressedByteSize := int64(rs.Length)
		compressedByteSize := int64(rs.WireLength)

		span.AddAttributes(
			trace.StringAttribute(payloadAttributeKey, payload),
			trace.Int64Attribute(uncompressedByteSizeAttributeKey, uncompressedByteSize),
			trace.Int64Attribute(compressedByteSizeAttributeKey, compressedByteSize),
		)
		span.AddMessageSendEvent(generateEventID(), uncompressedByteSize, compressedByteSize)
	case *stats.End:
		if rs.Error != nil {
			s, ok := status.FromError(rs.Error)
			if ok {
				span.SetStatus(trace.Status{Code: int32(s.Code()), Message: "OK"})
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
	if len(payloadTruncatedMessage) >= len(s) {
		return s[:limit]
	}

	return string(s[:len(s)-len(payloadTruncatedMessage)]) + payloadTruncatedMessage
}

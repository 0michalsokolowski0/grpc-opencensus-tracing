package tracing

import (
	"errors"
	"github.com/stretchr/testify/require"
	"go.opencensus.io/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/status"
	"strconv"
	"testing"
)

func TestTraceHandleRPC(t *testing.T) {
	t.Run("begin", func(t *testing.T) {
		begin := &stats.Begin{
			Client:   true,
			FailFast: true,
		}
		span := newSpanMock()

		traceHandleRPC(span, begin, 0)

		require.Contains(t, span.attributes, clientAttributeKey)
		require.Equal(t, begin.Client, span.attributes[clientAttributeKey])

		require.Contains(t, span.attributes, failFastAttributeKey)
		require.Equal(t, begin.FailFast, span.attributes[failFastAttributeKey])
	})

	t.Run("in_payload", func(t *testing.T) {
		inPayload := &stats.InPayload{
			Payload:    "XXXXXX",
			Length:     999,
			WireLength: 333,
		}
		span := newSpanMock()

		traceHandleRPC(span, inPayload, 0)

		require.Contains(t, span.attributes, payloadAttributeKey)
		require.Equal(t, inPayload.Payload, span.attributes[payloadAttributeKey])

		require.Contains(t, span.attributes, uncompressedByteSizeAttributeKey)
		require.Equal(t, int64(inPayload.Length), span.attributes[uncompressedByteSizeAttributeKey])

		require.Contains(t, span.attributes, compressedByteSizeAttributeKey)
		require.Equal(t, int64(inPayload.WireLength), span.attributes[compressedByteSizeAttributeKey])

		require.Contains(t, span.attributes, compressedByteSizeAttributeKey)
		require.Equal(t, int64(inPayload.WireLength), span.attributes[compressedByteSizeAttributeKey])

		require.EqualValues(t, []message{
			{
				uncompressedByteSize: 999,
				compressedByteSize:   333,
				received:             true,
			},
		}, span.messages)
	})

	t.Run("out_payload", func(t *testing.T) {
		inPayload := &stats.OutPayload{
			Payload:    "XXXXXX",
			Length:     999,
			WireLength: 333,
		}
		span := newSpanMock()

		traceHandleRPC(span, inPayload, 0)

		require.Contains(t, span.attributes, payloadAttributeKey)
		require.Equal(t, inPayload.Payload, span.attributes[payloadAttributeKey])

		require.Contains(t, span.attributes, uncompressedByteSizeAttributeKey)
		require.Equal(t, int64(inPayload.Length), span.attributes[uncompressedByteSizeAttributeKey])

		require.Contains(t, span.attributes, compressedByteSizeAttributeKey)
		require.Equal(t, int64(inPayload.WireLength), span.attributes[compressedByteSizeAttributeKey])

		require.Contains(t, span.attributes, compressedByteSizeAttributeKey)
		require.Equal(t, int64(inPayload.WireLength), span.attributes[compressedByteSizeAttributeKey])

		require.EqualValues(t, []message{
			{
				uncompressedByteSize: 999,
				compressedByteSize:   333,
				send:                 true,
			},
		}, span.messages)
	})

	t.Run("end", func(t *testing.T) {
		t.Run("grpc_error", func(t *testing.T) {
			end := &stats.End{
				Error: status.Error(codes.DeadlineExceeded, "it is fine"),
			}
			span := newSpanMock()

			traceHandleRPC(span, end, 0)

			require.Equal(t, trace.Status{
				Code:    4,
				Message: "OK",
			}, span.status)
			require.True(t, span.isEnded)
		})

		t.Run("unrecognised_error", func(t *testing.T) {
			end := &stats.End{
				Error: errors.New("unrecognised error"),
			}
			span := newSpanMock()

			traceHandleRPC(span, end, 0)

			require.Equal(t, trace.Status{
				Code:    13,
				Message: "unrecognised error",
			}, span.status)
			require.True(t, span.isEnded)
		})
	})
}

func TestTraceHandleRPC_with_payload_truncated(t *testing.T) {
	const payloadLengthLimit = 5
	testCases := []struct {
		payloadLength int
		expected      string
	}{
		{
			payloadLength: 0,
			expected:      "",
		},
		{
			payloadLength: 1,
			expected:      "0",
		},
		{
			payloadLength: 12,
			expected:      "01234",
		},
		{
			payloadLength: 13,
			expected:      "01234",
		},
		{
			payloadLength: 14,
			expected:      "01234",
		},
		{
			payloadLength: 30,
			expected:      "01234",
		},
		{
			payloadLength: 31,
			expected:      "01234",
		},
		{
			payloadLength: 32,
			expected:      "0...[payload has been truncated]",
		},
		{
			payloadLength: 33,
			expected:      "01...[payload has been truncated]",
		},
	}

	for _, testCase := range testCases {
		t.Run("in_payload"+strconv.Itoa(testCase.payloadLength), func(t *testing.T) {
			inPayload := &stats.InPayload{
				Payload: getStringOfLength(testCase.payloadLength),
			}
			span := newSpanMock()

			traceHandleRPC(span, inPayload, payloadLengthLimit)

			require.Contains(t, span.attributes, payloadAttributeKey)
			require.Equal(t, testCase.expected, span.attributes[payloadAttributeKey])
		})

		t.Run("out_payload"+strconv.Itoa(testCase.payloadLength), func(t *testing.T) {
			outPayload := &stats.OutPayload{
				Payload: getStringOfLength(testCase.payloadLength),
			}
			span := newSpanMock()

			traceHandleRPC(span, outPayload, payloadLengthLimit)

			require.Contains(t, span.attributes, payloadAttributeKey)
			require.Equal(t, testCase.expected, span.attributes[payloadAttributeKey])
		})
	}
}

func getStringOfLength(length int) string {
	result := ""
	for i := 0; i < length; i++ {
		result += strconv.Itoa(i % 10)
	}

	return result
}

type spanMock struct {
	attributes map[string]interface{}
	messages   []message
	status     trace.Status
	isEnded    bool
}

type message struct {
	uncompressedByteSize,
	compressedByteSize int64
	send     bool
	received bool
}

func newSpanMock() *spanMock {
	return &spanMock{
		attributes: make(map[string]interface{}),
		messages:   []message{},
	}
}

func (s spanMock) IsRecordingEvents() bool {
	return true
}

func (s *spanMock) End() {
	s.isEnded = true
}

func (s spanMock) SpanContext() trace.SpanContext {
	return trace.SpanContext{}
}

func (s spanMock) SetName(_ string) {
}

func (s *spanMock) SetStatus(status trace.Status) {
	s.status = status
}

func (s *spanMock) AddAttributes(attributes ...trace.Attribute) {
	for _, attribute := range attributes {
		s.attributes[attribute.Key()] = attribute.Value()
	}
}

func (s spanMock) Annotate(_ []trace.Attribute, _ string) {

}

func (s spanMock) Annotatef(_ []trace.Attribute, _ string, _ ...interface{}) {

}

func (s *spanMock) AddMessageSendEvent(messageID, uncompressedByteSize, compressedByteSize int64) {
	s.messages = append(s.messages, message{
		uncompressedByteSize: uncompressedByteSize,
		compressedByteSize:   compressedByteSize,
		send:                 true,
	})
}

func (s *spanMock) AddMessageReceiveEvent(messageID, uncompressedByteSize, compressedByteSize int64) {
	s.messages = append(s.messages, message{
		uncompressedByteSize: uncompressedByteSize,
		compressedByteSize:   compressedByteSize,
		received:             true,
	})
}

func (s spanMock) AddLink(_ trace.Link) {
	return
}

func (s spanMock) String() string {
	return ""
}

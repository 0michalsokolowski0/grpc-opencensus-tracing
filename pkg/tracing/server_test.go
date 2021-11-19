package tracing_test

import (
	"testing"

	"github.com/0michalsokolowski0/grpc-opencensus-tracing/pkg/tracing"
	"github.com/stretchr/testify/require"
	"go.opencensus.io/trace"
)

func TestNewStandardServerHandler(t *testing.T) {
	handler := tracing.NewStandardServerHandler()

	require.Equal(t, tracing.StandardPayloadAttributeSizeLimit, handler.PayloadAttributeLengthLimit)
	require.Equal(t, trace.StartOptions{}, handler.StartOptions)
}

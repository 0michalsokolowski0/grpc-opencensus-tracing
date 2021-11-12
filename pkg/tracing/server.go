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
	"context"
	"fmt"
	"strings"

	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/stats"
)

func NewStandardServerHandler() *ServerHandler {
	return &ServerHandler{
		StartOptions:                trace.StartOptions{},
		PayloadAttributeLengthLimit: StandardPayloadAttributeSizeLimit,
	}
}

type ServerHandler struct {
	StartOptions                trace.StartOptions
	PayloadAttributeLengthLimit int
}

var _ stats.Handler = (*ServerHandler)(nil)

func (s *ServerHandler) HandleConn(ctx context.Context, cs stats.ConnStats) {
	// no-op
}

func (s *ServerHandler) TagConn(ctx context.Context, cti *stats.ConnTagInfo) context.Context {
	// no-op
	return ctx
}

func (s *ServerHandler) HandleRPC(ctx context.Context, rs stats.RPCStats) {
	traceHandleRPC(trace.FromContext(ctx), rs, s.PayloadAttributeLengthLimit)
}

func (s *ServerHandler) TagRPC(ctx context.Context, rti *stats.RPCTagInfo) context.Context {
	ctx = s.traceTagRPC(ctx, rti)
	return ctx
}

func (s *ServerHandler) traceTagRPC(ctx context.Context, rti *stats.RPCTagInfo) context.Context {
	md, _ := metadata.FromIncomingContext(ctx)
	name := strings.TrimPrefix(rti.FullMethodName, "/")
	name = strings.Replace(name, "/", ".", -1)
	name = fmt.Sprintf("%s.%s", serverSpanNamePrefix, name)
	traceContext := md[traceContextKey]
	var (
		parent     trace.SpanContext
		haveParent bool
	)
	if len(traceContext) > 0 {
		// Metadata with keys ending in -bin are actually binary. They are base64
		// encoded before being put on the wire, see:
		// https://github.com/grpc/grpc-go/blob/08d6261/Documentation/grpc-metadata.md#storing-binary-data-in-metadata
		traceContextBinary := []byte(traceContext[0])
		parent, haveParent = propagation.FromBinary(traceContextBinary)
		if haveParent {
			ctx, _ := trace.StartSpanWithRemoteParent(ctx, name, parent,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithSampler(s.StartOptions.Sampler),
			)
			return ctx
		}
	}
	ctx, span := trace.StartSpan(ctx, name,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithSampler(s.StartOptions.Sampler))
	if haveParent {
		span.AddLink(trace.Link{TraceID: parent.TraceID, SpanID: parent.SpanID, Type: trace.LinkTypeChild})
	}
	return ctx
}

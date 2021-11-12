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
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/stats"
	"strings"
)

func NewStandardClientHandler() *ClientHandler {
	return &ClientHandler{
		StartOptions:                trace.StartOptions{},
		PayloadAttributeLengthLimit: StandardPayloadAttributeSizeLimit,
	}
}

type ClientHandler struct {
	StartOptions                trace.StartOptions
	PayloadAttributeLengthLimit int
}

func (c *ClientHandler) HandleConn(ctx context.Context, cs stats.ConnStats) {
	// no-op
}

func (c *ClientHandler) TagConn(ctx context.Context, cti *stats.ConnTagInfo) context.Context {
	// no-op
	return ctx
}

func (c *ClientHandler) HandleRPC(ctx context.Context, rs stats.RPCStats) {
	traceHandleRPC(trace.FromContext(ctx), rs, c.PayloadAttributeLengthLimit)
}

func (c *ClientHandler) TagRPC(ctx context.Context, rti *stats.RPCTagInfo) context.Context {
	return c.tagRPC(ctx, rti)
}

func (c *ClientHandler) tagRPC(ctx context.Context, rti *stats.RPCTagInfo) context.Context {
	name := strings.TrimPrefix(rti.FullMethodName, "/")
	name = strings.Replace(name, "/", ".", -1)
	name = fmt.Sprintf("%s.%s", clientSpanNamePrefix, name)
	ctx, span := trace.StartSpan(ctx, name,
		trace.WithSampler(c.StartOptions.Sampler),
		trace.WithSpanKind(trace.SpanKindClient))
	traceContextBinary := propagation.Binary(span.SpanContext())
	return metadata.AppendToOutgoingContext(ctx, traceContextKey, string(traceContextBinary))
}

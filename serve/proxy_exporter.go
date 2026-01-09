package serve

import (
	"context"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/zero-day-ai/sdk/api/gen/proto"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// ProxySpanExporter implements the OpenTelemetry SpanExporter interface and forwards
// spans to the orchestrator via the callback client. This enables distributed tracing
// across the agent-orchestrator boundary.
//
// The exporter uses a fire-and-forget strategy with a 5-second timeout to avoid
// blocking the application on span export failures. Errors are logged but not returned.
type ProxySpanExporter struct {
	client *CallbackClient
	logger *slog.Logger
}

// NewProxySpanExporter creates a new ProxySpanExporter that forwards spans to the orchestrator.
//
// Parameters:
//   - client: The callback client used to send spans to the orchestrator
//   - logger: Logger for recording export errors and diagnostics
//
// The returned exporter should be registered with the OpenTelemetry SDK's TracerProvider.
func NewProxySpanExporter(client *CallbackClient, logger *slog.Logger) *ProxySpanExporter {
	if logger == nil {
		logger = slog.Default()
	}
	return &ProxySpanExporter{
		client: client,
		logger: logger,
	}
}

// ExportSpans exports a batch of spans to the orchestrator.
//
// This method implements the fire-and-forget pattern:
//   - Uses a background context with 5-second timeout (independent of caller's context)
//   - Logs errors but always returns nil to avoid breaking the trace pipeline
//   - Converts OpenTelemetry spans to proto format before sending
//
// The method is called automatically by the OpenTelemetry SDK when spans are completed.
func (e *ProxySpanExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	if len(spans) == 0 {
		return nil
	}

	// Fire-and-forget: use background context with timeout
	exportCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Convert spans to proto format
	protoSpans := make([]*proto.Span, 0, len(spans))
	for _, span := range spans {
		protoSpan := spanToProto(span)
		protoSpans = append(protoSpans, protoSpan)
	}

	// Extract mission_id from first span's attributes if available
	missionID := ""
	if len(spans) > 0 {
		for _, attr := range spans[0].Attributes() {
			if string(attr.Key) == "mission_id" {
				missionID = attr.Value.AsString()
				break
			}
		}
	}

	// Build the request with context info
	req := &proto.RecordSpansRequest{
		Context: &proto.ContextInfo{
			MissionId: missionID,
		},
		Spans: protoSpans,
	}

	// Send spans to orchestrator (fire-and-forget)
	_, err := e.client.RecordSpans(exportCtx, req)
	if err != nil {
		// Log error but don't return it (fire-and-forget pattern)
		e.logger.Error("failed to export spans",
			"error", err,
			"span_count", len(spans),
			"mission_id", missionID)
	}

	return nil
}

// Shutdown performs cleanup when the exporter is being shut down.
// This implementation is a no-op as the callback client manages its own lifecycle.
func (e *ProxySpanExporter) Shutdown(ctx context.Context) error {
	return nil
}

// spanToProto converts an OpenTelemetry ReadOnlySpan to the proto Span format.
func spanToProto(span sdktrace.ReadOnlySpan) *proto.Span {
	sc := span.SpanContext()

	// Convert trace and span IDs to hex strings
	// TraceID and SpanID are arrays, we need to convert to slices
	traceIDBytes := sc.TraceID()
	spanIDBytes := sc.SpanID()
	traceID := hex.EncodeToString(traceIDBytes[:])
	spanID := hex.EncodeToString(spanIDBytes[:])

	// Convert parent span ID if present
	var parentSpanID string
	if span.Parent().IsValid() {
		parentSpanIDBytes := span.Parent().SpanID()
		parentSpanID = hex.EncodeToString(parentSpanIDBytes[:])
	}

	// Convert timestamps to Unix nanoseconds
	startTime := span.StartTime().UnixNano()
	endTime := span.EndTime().UnixNano()

	// Convert span kind
	kind := spanKindToProto(span.SpanKind())

	// Convert status
	statusCode, statusMessage := spanStatusToProto(span.Status())

	// Convert attributes
	attributes := attributesToProto(span.Attributes())

	// Convert events
	events := eventsToProto(span.Events())

	return &proto.Span{
		TraceId:           traceID,
		SpanId:            spanID,
		ParentSpanId:      parentSpanID,
		StartTimeUnixNano: startTime,
		EndTimeUnixNano:   endTime,
		Name:              span.Name(),
		Kind:              kind,
		StatusCode:        statusCode,
		StatusMessage:     statusMessage,
		Attributes:        attributes,
		Events:            events,
	}
}

// spanKindToProto converts OpenTelemetry SpanKind to proto SpanKind.
func spanKindToProto(kind trace.SpanKind) proto.SpanKind {
	switch kind {
	case trace.SpanKindInternal:
		return proto.SpanKind_SPAN_KIND_INTERNAL
	case trace.SpanKindServer:
		return proto.SpanKind_SPAN_KIND_SERVER
	case trace.SpanKindClient:
		return proto.SpanKind_SPAN_KIND_CLIENT
	case trace.SpanKindProducer:
		return proto.SpanKind_SPAN_KIND_PRODUCER
	case trace.SpanKindConsumer:
		return proto.SpanKind_SPAN_KIND_CONSUMER
	default:
		return proto.SpanKind_SPAN_KIND_UNSPECIFIED
	}
}

// spanStatusToProto converts OpenTelemetry Status to proto StatusCode and message.
func spanStatusToProto(status sdktrace.Status) (proto.StatusCode, string) {
	switch status.Code {
	case codes.Ok:
		return proto.StatusCode_STATUS_CODE_OK, status.Description
	case codes.Error:
		return proto.StatusCode_STATUS_CODE_ERROR, status.Description
	default:
		return proto.StatusCode_STATUS_CODE_UNSET, status.Description
	}
}

// attributesToProto converts OpenTelemetry attributes to proto KeyValue pairs.
func attributesToProto(attrs []attribute.KeyValue) []*proto.KeyValue {
	if len(attrs) == 0 {
		return nil
	}

	kvs := make([]*proto.KeyValue, 0, len(attrs))
	for _, attr := range attrs {
		kv := &proto.KeyValue{
			Key:   string(attr.Key),
			Value: attributeValueToProto(attr.Value),
		}
		kvs = append(kvs, kv)
	}
	return kvs
}

// attributeValueToProto converts an OpenTelemetry attribute value to proto AnyValue.
func attributeValueToProto(val attribute.Value) *proto.AnyValue {
	switch val.Type() {
	case attribute.BOOL:
		return &proto.AnyValue{
			Value: &proto.AnyValue_BoolValue{
				BoolValue: val.AsBool(),
			},
		}
	case attribute.INT64:
		return &proto.AnyValue{
			Value: &proto.AnyValue_IntValue{
				IntValue: val.AsInt64(),
			},
		}
	case attribute.FLOAT64:
		return &proto.AnyValue{
			Value: &proto.AnyValue_DoubleValue{
				DoubleValue: val.AsFloat64(),
			},
		}
	case attribute.STRING:
		return &proto.AnyValue{
			Value: &proto.AnyValue_StringValue{
				StringValue: val.AsString(),
			},
		}
	case attribute.BOOLSLICE, attribute.INT64SLICE, attribute.FLOAT64SLICE, attribute.STRINGSLICE:
		// Convert slices to string representation
		return &proto.AnyValue{
			Value: &proto.AnyValue_StringValue{
				StringValue: val.AsString(),
			},
		}
	default:
		return &proto.AnyValue{
			Value: &proto.AnyValue_StringValue{
				StringValue: val.AsString(),
			},
		}
	}
}

// eventsToProto converts OpenTelemetry span events to proto SpanEvent.
func eventsToProto(events []sdktrace.Event) []*proto.SpanEvent {
	if len(events) == 0 {
		return nil
	}

	protoEvents := make([]*proto.SpanEvent, 0, len(events))
	for _, event := range events {
		protoEvent := &proto.SpanEvent{
			Name:         event.Name,
			TimeUnixNano: event.Time.UnixNano(),
			Attributes:   attributesToProto(event.Attributes),
		}
		protoEvents = append(protoEvents, protoEvent)
	}
	return protoEvents
}

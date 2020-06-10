package gqlopencensus

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"go.opencensus.io/trace"
)

// Tracer enables opencensus tracing on gqlgen
type Tracer struct {
	config
}

var _ interface {
	// build time safeguards
	graphql.HandlerExtension
	graphql.ResponseInterceptor
	graphql.FieldInterceptor
} = Tracer{}

// New opencensus tracer for gqlgen
func New(opts ...Option) *Tracer {
	tr := defaultTracer()
	for _, apply := range opts {
		apply(&tr.config)
	}
	return tr
}

// ExtensionName implements the graphql.HandlerExtension
func (Tracer) ExtensionName() string {
	return "Opencensustracing"
}

// Validate implements the graphql.HandlerExtension
func (Tracer) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

// InterceptField implements graphql.FieldInterceptor
func (tr Tracer) InterceptField(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
	fc := graphql.GetFieldContext(ctx)
	if tr.onlyMethods && !fc.IsMethod {
		// only capture fields which correspond to a resolver method
		return next(ctx)
	}
	ctx, span := trace.StartSpan(ctx,
		fc.Path().String(),
		trace.WithSpanKind(trace.SpanKindServer),
	)
	span.AddAttributes(tr.config.fieldAttributes(fc)...)
	defer span.End()

	return next(ctx)
}

// InterceptResponse implements graphql.OperationInterceptor
func (tr Tracer) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	oc := graphql.GetOperationContext(ctx)
	ctx, span := trace.StartSpan(ctx,
		operationName(oc),
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	span.AddAttributes(tr.config.operationAttributes(oc)...)

	resp := next(ctx)
	if resp == nil {
		return nil
	}

	if errs := resp.Errors; len(errs) > 0 {
		span.SetStatus(trace.Status{
			Code:    trace.StatusCodeUnknown,
			Message: errs.Error(),
		})
	}

	return resp
}

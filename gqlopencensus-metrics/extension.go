package metrics

import (
	"context"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

const extensionName = "OpencensusMetrics"

var _ interface {
	graphql.HandlerExtension
	graphql.ResponseInterceptor
	graphql.FieldInterceptor
} = &Collector{}

type (
	// Collector is a gqlgen extension to collect opencensus metrics on all GraphQL executions
	Collector struct {
		*config
		opTagger    func(string) []tag.Mutator
		fieldTagger func(string, string) []tag.Mutator
	}
)

// New Collector
func New(opts ...Option) *Collector {
	m := defaultCollector()
	for _, apply := range opts {
		apply(m.config)
	}

	if m.config.host == "" {
		m.config.host = "-"
	}

	m.opTagger = func(opName string) []tag.Mutator {
		return []tag.Mutator{tag.Upsert(TagHost, m.config.host), tag.Upsert(TagOperation, opName)}
	}
	if m.config.fieldsEnabled {
		m.fieldTagger = func(fieldName, pth string) []tag.Mutator {
			return []tag.Mutator{tag.Upsert(TagHost, m.config.host), tag.Upsert(TagField, fieldName), tag.Upsert(TagPath, pth)}
		}
	}
	return m
}

// ExtensionName yields the extension name: "OpencensusMetrics"
func (Collector) ExtensionName() string {
	return extensionName
}

// Validate this collector. This is a noop
func (Collector) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

// InterceptField implements the gqlgen field interceptor
func (m Collector) InterceptField(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
	if !m.config.fieldsEnabled {
		return next(ctx)
	}

	fc := graphql.GetFieldContext(ctx)
	if !fc.IsMethod {
		// only capture fields which correspond to a resolver method
		return next(ctx)
	}

	start := graphql.Now()

	defer func() {
		end := graphql.Now()
		_ = stats.RecordWithTags(ctx,
			m.fieldTagger(fc.Field.Name, fieldPath(fc)),
			ServerFieldCount.M(1),
			ServerFieldLatency.M(float64(end.Sub(start))/float64(time.Millisecond)),
		)
	}()

	return next(ctx)
}

// InterceptResponse implements the gqlgen response interceptor
func (m Collector) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	rc := graphql.GetOperationContext(ctx)
	opName := operationName(rc)

	resp := next(ctx)
	end := graphql.Now()

	_ = stats.RecordWithTags(ctx,
		m.opTagger(opName),
		ServerRequestCount.M(1),
		ServerParsing.M(float64(rc.Stats.Validation.End.Sub(rc.Stats.Parsing.Start))/float64(time.Millisecond)),
		ServerLatency.M(float64(end.Sub(rc.Stats.Validation.End))/float64(time.Millisecond)),
	)

	if resp == nil {
		return nil
	}
	if err := resp.Errors.Error(); err != "" {
		_ = stats.RecordWithTags(ctx, m.opTagger(opName), ServerErrorCount.M(1))
	}
	return resp
}

func operationName(ctx *graphql.OperationContext) (opName string) {
	if ctx.Operation != nil {
		opName = ctx.Operation.Name
	}
	if opName == "" && ctx.Operation != nil {
		//parent response case
		opName = string(ctx.Operation.Operation)
	}
	if opName == "" {
		opName = ctx.OperationName
	}
	return
}

func fieldPath(ctx *graphql.FieldContext) string {
	pth := ctx.Path().String()
	if strings.HasPrefix(pth, "__schema") {
		// collapse all introspection under one single tag
		pth = "__schema"
	}
	return pth
}

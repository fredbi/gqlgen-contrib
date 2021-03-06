package gqlopencensus

import (
	"encoding/json"

	"github.com/99designs/gqlgen/graphql"
	"go.opencensus.io/trace"
)

// Option for an opencensus tracer. At this moment, it is possible to configure span attributes retrieved from the GraphQL contexts.
type Option func(*config)

// FieldAttributer is a functor producing trace attributes from the GraphL field context
type FieldAttributer func(*graphql.FieldContext) []trace.Attribute

// FieldAttribute is a simple FieldAttributer that just adds a constant key/value attribute to the span.
//
// You can use it with the WithFieldAttributes option.
//
// Example:
//
//   New(WithFieldAttributes(FieldAttribute("host", "mypod")))
func FieldAttribute(key, value string) FieldAttributer {
	return func(_ *graphql.FieldContext) []trace.Attribute {
		return []trace.Attribute{trace.StringAttribute(key, value)}
	}
}

// OperationAttributer is a functor producing trace attributes from the GraphL operation context.
type OperationAttributer func(*graphql.OperationContext) []trace.Attribute

// OperationAttribute is a simple OperationAttributer that just adds a constant key/value attribute to the span.
//
// You can use it with the WithOperationdAttributes option.
//
// Example:
//
//   New(WithOperationAttributes(OperationAttribute("host","mypod")))
func OperationAttribute(key, value string) OperationAttributer {
	return func(_ *graphql.OperationContext) []trace.Attribute {
		return []trace.Attribute{trace.StringAttribute(key, value)}
	}
}

type config struct {
	fieldAttributers     []FieldAttributer
	operationAttributers []OperationAttributer
	onlyMethods          bool
}

func (c config) fieldAttributes(ctx *graphql.FieldContext) []trace.Attribute {
	attrs := make([]trace.Attribute, 0, 10)
	for _, apply := range c.fieldAttributers {
		attrs = append(attrs, apply(ctx)...)
	}
	return attrs
}

func (c config) operationAttributes(ctx *graphql.OperationContext) []trace.Attribute {
	attrs := make([]trace.Attribute, 0, 10)
	for _, apply := range c.operationAttributers {
		attrs = append(attrs, apply(ctx)...)
	}
	return attrs
}

func defaultTracer() *Tracer {
	return &Tracer{
		config: config{
			fieldAttributers: []FieldAttributer{func(fc *graphql.FieldContext) []trace.Attribute {
				return []trace.Attribute{
					trace.StringAttribute("server", "gqlgen"),
					trace.StringAttribute("field", fc.Field.Name),
				}
			},
			},
			operationAttributers: []OperationAttributer{func(oc *graphql.OperationContext) []trace.Attribute {
				return []trace.Attribute{
					trace.StringAttribute("server", "gqlgen"),
					trace.StringAttribute("operation", operationName(oc)),
				}
			},
			},
			onlyMethods: true,
		},
	}
}

// WithFieldAttributes adds some extra attributes from the graphQL field context to the span
func WithFieldAttributes(attributers ...FieldAttributer) Option {
	return func(c *config) {
		c.fieldAttributers = append(c.fieldAttributers, attributers...)
	}
}

// WithOperationAttributes adds some extra attributes from the graphQL operation context to the span
func WithOperationAttributes(attributers ...OperationAttributer) Option {
	return func(c *config) {
		c.operationAttributers = append(c.operationAttributers, attributers...)
	}
}

// WithDataDog provides DataDog specific span attrs.
// see github.com/DataDog/opencensus-go-exporter-datadog
func WithDataDog() Option {
	return func(c *config) {
		c.operationAttributers = append(c.operationAttributers, func(oc *graphql.OperationContext) []trace.Attribute {
			return []trace.Attribute{
				trace.StringAttribute("resource.name", operationName(oc)),
			}
		})
	}
}

// WithRawQuery adds the GraphL query to the trace span of an operation. This is disabled by default.
func WithRawQuery() Option {
	return func(c *config) {
		c.operationAttributers = append(c.operationAttributers, func(oc *graphql.OperationContext) []trace.Attribute {
			return []trace.Attribute{
				trace.StringAttribute("query", oc.RawQuery),
			}
		})
	}
}

// WithVariables adds the values of all variables attached to the GraphL query to the trace span of an operation. This is disabled by default.
func WithVariables() Option {
	return func(c *config) {
		c.operationAttributers = append(c.operationAttributers, func(oc *graphql.OperationContext) []trace.Attribute {
			variables, _ := json.Marshal(oc.Variables)
			return []trace.Attribute{
				trace.StringAttribute("variables", string(variables)),
			}
		})
	}
}

// WithArgs adds the GraphL args of a field to the trace span of an field. This is disabled by default.
func WithArgs() Option {
	return func(c *config) {
		c.fieldAttributers = append(c.fieldAttributers, func(fc *graphql.FieldContext) []trace.Attribute {
			args, _ := json.Marshal(fc.Args)
			return []trace.Attribute{
				trace.StringAttribute("args", string(args)),
			}
		})
	}
}

// OnlyMethods when enabled, produces spans only for fields which correspond to a method of the resolver. This is the default.
// When set to false, all fields produce a span.
func OnlyMethods(enabled bool) Option {
	return func(c *config) {
		c.onlyMethods = enabled
	}
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

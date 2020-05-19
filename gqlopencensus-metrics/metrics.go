// Package metrics collects opencensus metrics
// for a GraphQL server.
package metrics

import (
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

// Register views.
//
// Views must be registered before using the extension.
func Register() error {
	return view.Register(GQLViews...)
}

// Unregister views
func Unregister() {
	view.Unregister(GQLViews...)
}

var (
	// GQLViews contains all opencensus stats views declared by the GraphQL stats collector
	GQLViews = []*view.View{
		OperationCountView,
		FieldCountView,
		OperationErrorsView,
		OperationLatencyView,
		FieldLatencyView,
		OperationParsingView,
	}

	// measurements

	// ServerRequestCount tracks a count of GraphQL requests
	ServerRequestCount = stats.Int64(
		"gql/server/request_count",
		"Number of GraphQL requests started",
		stats.UnitDimensionless)

	// ServerFieldCount tracks a count of GraphQL fields requested
	ServerFieldCount = stats.Int64(
		"gql/server/field_count",
		"Number of GraphQL field resolutions, per field and query path",
		stats.UnitDimensionless)

	// ServerErrorCount tracks a count of request errors
	ServerErrorCount = stats.Int64(
		"gql/server/request_count",
		"Number of GraphQL requests started",
		stats.UnitDimensionless)

	// ServerLatency tracks the execution time of requests (excluding parsing and validation time), in milliseconds
	ServerLatency = stats.Float64(
		"gql/server/latency",
		"Execution latency",
		stats.UnitMilliseconds)

	// ServerFieldLatency tracks the execution time of individual fields in requests, in milliseconds
	ServerFieldLatency = stats.Float64(
		"gql/server/field_latency",
		"Single field execution latency",
		stats.UnitMilliseconds)

	// ServerParsing tracks the parsing and validation time that occurs before the request execution
	ServerParsing = stats.Float64(
		"gql/server/parsing_validation",
		"Parsing & validation latency",
		stats.UnitMilliseconds)

	// views

	// OperationCountView reports a count of operations tagged by host and operation name
	OperationCountView = &view.View{
		Name:        "gql/server/operation_count",
		Description: "Count of GraphQL requests started by operation",
		Measure:     ServerRequestCount,
		Aggregation: view.Count(),
		TagKeys:     []tag.Key{TagHost, TagOperation},
	}

	// FieldCountView reports a count of requested fields tagged by host, field name and query path
	FieldCountView = &view.View{
		Name:        "gql/server/field_count",
		Description: "Count of GraphQL fields requests by field and by query path",
		Measure:     ServerFieldCount,
		Aggregation: view.Count(),
		TagKeys:     []tag.Key{TagHost, TagField, TagPath},
	}

	// OperationErrorsView reports a count of errors tagged by host and operation name
	OperationErrorsView = &view.View{
		Name:        "gql/server/error_count",
		Description: "Count of GraphQL requests returning an error by operation",
		Measure:     ServerErrorCount,
		Aggregation: view.Count(),
		TagKeys:     []tag.Key{TagHost, TagOperation},
	}

	// OperationLatencyView reports a distribution of execution time of GraphQL operations, by host and operation (in milliseconds)
	OperationLatencyView = &view.View{
		Name:        "gql/server/latency",
		Description: "Execution time distribution of GraphQL requests by operation, excluding parsing and validation",
		Measure:     ServerLatency,
		Aggregation: DefaultLatencyDistribution,
		TagKeys:     []tag.Key{TagHost, TagOperation},
	}

	// FieldLatencyView reports a distribution of field retrieval time, by field, query path, and host (in milliseconds)
	FieldLatencyView = &view.View{
		Name:        "gql/server/field_latency",
		Description: "Execution time distribution of GraphQL requests by operation, excluding parsing and validation",
		Measure:     ServerFieldLatency,
		Aggregation: DefaultLatencyDistribution,
		TagKeys:     []tag.Key{TagHost, TagField, TagPath},
	}

	// OperationParsingView reports a distribution of GraphQL parsing and validation time (in milliseconds)
	OperationParsingView = &view.View{
		Name:        "gql/server/parsing_validation",
		Description: "Parsing  and validation time distribution of GraphQL requests by operation",
		Measure:     ServerParsing,
		Aggregation: DefaultLatencyDistribution,
		TagKeys:     []tag.Key{TagHost, TagOperation},
	}

	// TagHost is the name of the graphQL server
	TagHost = tag.MustNewKey("gql.host")

	// TagOperation is the query operation name
	TagOperation = tag.MustNewKey("gql.operation")

	// TagField is an individual GraphQL field requested
	TagField = tag.MustNewKey("gql.field")

	// TagPath is an individual GraphQL path to a field requested
	TagPath = tag.MustNewKey("gql.path")

	// DefaultLatencyDistribution constructs buckets for latency distributions in views
	DefaultLatencyDistribution = view.Distribution(1, 2, 3, 4, 5, 6, 8, 10, 13, 16, 20, 25, 30, 40, 50, 65, 80, 100, 130, 160, 200, 250, 300, 400, 500, 650, 800, 1000, 2000, 5000, 10000, 20000, 50000, 100000)
)

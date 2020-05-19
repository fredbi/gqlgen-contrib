package metrics

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/stretchr/testify/require"
	"go.opencensus.io/stats/view"
)

func TestMetrics(t *testing.T) {
	view.RegisterExporter(testExporter{t: t})
	err := Register()
	require.NoError(t, err)

	ext := New()

	oTags := ext.opTagger("test")
	require.Len(t, oTags, 2)

	fTags := ext.fieldTagger("aField", "q/path")
	require.Len(t, fTags, 3)

	require.Equal(t, extensionName, ext.ExtensionName())
	require.Nil(t, ext.Validate(&graphql.ExecutableSchemaMock{}))

	opCtx := &graphql.OperationContext{
		RawQuery:      "query{}",
		OperationName: "test",
	}
	h := func(_ context.Context) *graphql.Response {
		return &graphql.Response{
			Data: json.RawMessage(`{"a": "abc"}`),
		}
	}

	ctx := graphql.WithOperationContext(context.Background(), opCtx)
	resp := ext.InterceptResponse(ctx, h)

	bbb, err := json.Marshal(resp)
	require.NoError(t, err)
	t.Logf("resp: %v", string(bbb))
	time.Sleep(11 * time.Second)
}

type testExporter struct{ t testing.TB }

func (x testExporter) ExportView(viewData *view.Data) {
	x.t.Logf("viewData: %#v", viewData)
}

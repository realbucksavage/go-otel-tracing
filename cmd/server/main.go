package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	gotrace "github.com/realbucksavage/go-otel-tracing"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func main() {

	gotrace.InitTrace("server")
	handler := otelhttp.NewHandler(newHandler(), "responder-server-handler")

	panic(http.ListenAndServe(":7000", handler))
}

func newHandler() http.Handler {

	responder := func(rw http.ResponseWriter, r *http.Request) {

		name := r.URL.Query().Get("name")

		ctx := r.Context()
		span := trace.SpanFromContext(ctx)
		span.SetName("responder")
		defer span.End()

		if name == "" {
			span.RecordError(errors.New("name is empty"))
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		span.AddEvent("processing name", trace.WithAttributes(attribute.String("name", name)))
		time.Sleep(time.Duration(len(name)) * time.Second)
		rw.Write([]byte(fmt.Sprintf("hello, %s", name)))
	}

	return http.HandlerFunc(responder)
}

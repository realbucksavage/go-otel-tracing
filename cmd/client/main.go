package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"

	"net/http"

	gotrace "github.com/realbucksavage/go-otel-tracing"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func main() {

	gotrace.InitTrace("client")

	name := flag.String("name", "", "")
	flag.Parse()

	ctx, span := otel.Tracer("greeter-client").Start(context.Background(), "greeter")
	defer span.End()

	addr := fmt.Sprintf("http://localhost:7000/?name=%s", *name)
	client := &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	req, err := http.NewRequest(http.MethodGet, addr, nil)
	if err != nil {
		span.RecordError(err)
		log.Println("cannot create http request")
		return
	}

	span.AddEvent("calling greeter server", trace.WithAttributes(attribute.String("input", *name)))

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		span.RecordError(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		span.RecordError(errors.New("invalid response code"), trace.WithAttributes(attribute.Int("http-status", resp.StatusCode)))
		log.Printf("unknown status code %d", resp.StatusCode)
		return
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		span.RecordError(err, trace.WithAttributes(attribute.String("err-type", "read")))
		log.Println("read error", err)
		return
	}

	fmt.Println(string(content))
	span.AddEvent("processing completed")
}

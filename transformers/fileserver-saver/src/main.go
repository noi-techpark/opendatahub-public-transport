// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"

	"github.com/noi-techpark/opendatahub-public-transport/lib/compress"
	"net/http"
	"time"

	"github.com/noi-techpark/opendatahub-go-sdk/ingest/ms"
	"github.com/noi-techpark/opendatahub-go-sdk/ingest/rdb"
	"github.com/noi-techpark/opendatahub-go-sdk/ingest/tr"
	"github.com/noi-techpark/opendatahub-go-sdk/tel"
	"github.com/noi-techpark/opendatahub-go-sdk/tel/logger"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var env struct {
	tr.Env
	CONFIG_PATH     string
	FILESERVER_HOST string `default:"http://files-nginx-internal.core.svc.cluster.local"`
}

func main() {
	ctx := context.Background()
	ms.InitWithEnv(ctx, "", &env)
	slog.Info("Starting fileserver saver transformer...")

	defer tel.FlushOnPanic()

	config, err := LoadConfig(env.CONFIG_PATH)
	ms.FailOnError(ctx, err, "failed to load sink config")

	slog.Info("Loaded sink config", "sink_count", len(config.Sinks))

	for _, sink := range config.Sinks {
		sinkEnv := env.Env
		sinkEnv.MQ_QUEUE = sink.MQQueue
		sinkEnv.MQ_EXCHANGE = sink.MQExchange
		sinkEnv.MQ_KEY = sink.MQKey
		sinkEnv.MQ_CLIENT = env.MQ_CLIENT + "-" + sink.MQQueue

		go runSink(ctx, sink, sinkEnv)
	}

	select {}
}

// runSink runs a single sink's listener loop with panic recovery and automatic restart.
func runSink(ctx context.Context, sink SinkConfig, sinkEnv tr.Env) {
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					slog.Error("Sink goroutine panicked, restarting after delay",
						"queue", sink.MQQueue,
						"fileserver_path", sink.FileserverPath,
						"panic", fmt.Sprintf("%v", r),
					)
				}
			}()

			slog.Info("Starting queue listener for sink",
				"queue", sink.MQQueue,
				"fileserver_path", sink.FileserverPath,
			)

			listener := tr.NewTr[SiriPayload](ctx, sinkEnv)
			err := listener.Start(ctx, func(ctx context.Context, raw *rdb.Raw[SiriPayload]) error {
				return handleMessage(ctx, sink, raw)
			})
			if err != nil {
				slog.Error("Listener returned error",
					"queue", sink.MQQueue,
					"err", err,
				)
			}
		}()

		slog.Info("Sink goroutine sleeping before restart",
			"queue", sink.MQQueue,
			"delay", "5s",
		)
		time.Sleep(5 * time.Second)
	}
}

// handleMessage processes a single message from the queue.
func handleMessage(ctx context.Context, sink SinkConfig, raw *rdb.Raw[SiriPayload]) error {
	spanName := fmt.Sprintf("%s.save/%s", tel.GetServiceName(), sink.MQQueue)
	ctx, span := tel.TraceStart(ctx, spanName, trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	log := logger.Get(ctx)

	log.Info("Saving to fileserver",
		"queue", sink.MQQueue,
		"fileserver_path", sink.FileserverPath,
		"provider", raw.Provider,
	)

	payload, err := extractPayload(raw.Rawdata)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to extract payload")
		log.Error("Failed to extract payload", "queue", sink.MQQueue, "err", err)
		return err
	}

	if err := putFile(ctx, env.FILESERVER_HOST, sink.FileserverPath, payload); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to save to fileserver")
		log.Error("Failed to save to fileserver",
			"queue", sink.MQQueue,
			"fileserver_path", sink.FileserverPath,
			"err", err,
		)
		return err
	}

	span.SetStatus(codes.Ok, "")
	log.Info("Saved to fileserver",
		"queue", sink.MQQueue,
		"fileserver_path", sink.FileserverPath,
	)
	return nil
}

// putFile PUTs data to the nginx fileserver.
func putFile(ctx context.Context, host, path string, data []byte) error {
	url := host + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	return nil
}

// extractPayload returns the raw payload bytes, decompressing if metadata indicates compression.
func extractPayload(p SiriPayload) ([]byte, error) {
	if compress.IsCompressed(p.Metadata) {
		return compress.Decode(p.Payload)
	}
	return []byte(p.Payload), nil
}

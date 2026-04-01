// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/noi-techpark/opendatahub-public-transport/lib/compress"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/noi-techpark/opendatahub-go-sdk/ingest/dc"
	"github.com/noi-techpark/opendatahub-go-sdk/ingest/ms"
	"github.com/noi-techpark/opendatahub-go-sdk/ingest/rdb"
	"github.com/noi-techpark/opendatahub-go-sdk/tel"
	"github.com/noi-techpark/opendatahub-go-sdk/tel/logger"
	"github.com/robfig/cron/v3"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var env struct {
	dc.Env
	CONFIG_PATH string
}

func main() {
	ms.InitWithEnv(context.Background(), "", &env)
	slog.Info("Starting feed fetcher...")

	defer tel.FlushOnPanic()

	config, err := LoadConfig(env.CONFIG_PATH)
	ms.FailOnError(context.Background(), err, "failed to load feed config")

	slog.Info("Loaded feed config", "feed_count", len(config.Feeds))

	collector := dc.NewDc[dc.EmptyData](context.Background(), env.Env)

	for _, feed := range config.Feeds {
		go runFeed(context.Background(), feed, collector)
	}

	// Block forever — goroutines handle their own lifecycle
	select {}
}

// runFeed runs a single feed's cron loop with panic recovery and automatic restart.
func runFeed(ctx context.Context, feed FeedConfig, collector *dc.Dc[dc.EmptyData]) {
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					slog.Error("Feed goroutine panicked, restarting after delay",
						"provider", feed.Provider,
						"endpoint", feed.Endpoint,
						"panic", fmt.Sprintf("%v", r),
					)
				}
			}()

			slog.Info("Starting cron scheduler for feed",
				"provider", feed.Provider,
				"cron", feed.Cron,
				"endpoint", feed.Endpoint,
			)

			c := cron.New(cron.WithSeconds())
			_, err := c.AddFunc(feed.Cron, func() {
				pollFeed(ctx, feed, collector)
			})
			if err != nil {
				slog.Error("Invalid cron expression",
					"provider", feed.Provider,
					"cron", feed.Cron,
					"err", err,
				)
				return
			}
			c.Run()
		}()

		slog.Info("Feed goroutine sleeping before restart",
			"provider", feed.Provider,
			"delay", "5s",
		)
		time.Sleep(5 * time.Second)
	}
}

// pollFeed performs a single poll cycle for a feed.
func pollFeed(ctx context.Context, feed FeedConfig, collector *dc.Dc[dc.EmptyData]) {
	jobstart := time.Now()

	ctx, collection := collector.StartCollection(ctx)
	defer collection.End(ctx)

	spanName := fmt.Sprintf("%s.poll/%s", tel.GetServiceName(), feed.Provider)
	ctx, span := tel.TraceStart(ctx, spanName, trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	log := logger.Get(ctx)

	log.Info("Polling SIRI feed",
		"provider", feed.Provider,
		"endpoint", feed.Endpoint,
	)

	req, err := retryablehttp.NewRequest(http.MethodGet, feed.Endpoint, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create request")
		log.Error("Failed to create request",
			"provider", feed.Provider,
			"endpoint", feed.Endpoint,
			"err", err,
		)
		return
	}

	for k, v := range feed.Headers {
		req.Header.Set(k, v)
	}

	client := retryablehttp.NewClient()
	client.Logger = nil

	resp, err := client.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "http request failed")
		log.Error("HTTP request failed",
			"provider", feed.Provider,
			"endpoint", feed.Endpoint,
			"err", err,
		)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		span.RecordError(err)
		span.SetStatus(codes.Error, "non-OK status")
		log.Error("HTTP request returned non-OK status",
			"provider", feed.Provider,
			"endpoint", feed.Endpoint,
			"statusCode", resp.StatusCode,
		)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to read response body")
		log.Error("Failed to read response body",
			"provider", feed.Provider,
			"endpoint", feed.Endpoint,
			"err", err,
		)
		return
	}

	if len(body) == 0 {
		log.Error("Empty response body, skipping",
			"provider", feed.Provider,
			"endpoint", feed.Endpoint,
		)
		return
	}

	payload := string(body)
	metadata := feed.Metadata
	if feed.Compress {
		encoded, err := compress.Encode(body)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to compress")
			log.Error("Failed to compress payload", "provider", feed.Provider, "err", err)
			return
		}
		payload = encoded
		if metadata == nil {
			metadata = make(map[string]string)
		}
		metadata[compress.MetadataKey] = compress.MetadataValue
		log.Info("Compressed payload",
			"provider", feed.Provider,
			"original", len(body),
			"compressed", len(encoded),
		)
	}

	err = collection.Publish(ctx, &rdb.RawAny{
		Provider:  feed.Provider,
		Timestamp: time.Now(),
		Rawdata: SiriPayload{
			Format:   feed.Format,
			Protocol: feed.Protocol,
			Metadata: metadata,
			Payload:  payload,
		},
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to publish")
		log.Error("Failed to publish",
			"provider", feed.Provider,
			"err", err,
		)
		return
	}

	span.SetStatus(codes.Ok, "")
	log.Info("Poll completed",
		"provider", feed.Provider,
		"runtime_ms", time.Since(jobstart).Milliseconds(),
	)
}


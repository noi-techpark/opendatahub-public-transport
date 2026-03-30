// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/noi-techpark/opendatahub-public-transport/lib/compress"

	"github.com/noi-techpark/opendatahub-go-sdk/ingest/ms"
	"github.com/noi-techpark/opendatahub-go-sdk/ingest/rdb"
	"github.com/noi-techpark/opendatahub-go-sdk/ingest/tr"
	"github.com/noi-techpark/opendatahub-go-sdk/tel"
	"github.com/noi-techpark/opendatahub-go-sdk/tel/logger"
	"go.opentelemetry.io/otel/trace"

	"github.com/noi-techpark/opendatahub-public-transport/lib/go-gtfsrt/gtfsrt"
	"github.com/noi-techpark/opendatahub-public-transport/lib/go-siri/siri"
)

var env struct {
	tr.Env
	FILESERVER_HOST string `default:"http://files-nginx-internal.core.svc.cluster.local"`
	FILESERVER_PATH string
	NETEX_FTP_URL   string
	GTFS_FTP_URL    string
	REFRESH_HOURS   int `default:"24"`
}

type SiriPayload struct {
	Format   string            `json:"format"`
	Protocol string            `json:"protocol"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Payload  string            `json:"payload"`
}

func main() {
	ctx := context.Background()
	ms.InitWithEnv(ctx, "", &env)
	slog.Info("Starting STA Situation Exchange to GTFS-RT transformer...")

	defer tel.FlushOnPanic()

	staticData, err := LoadStaticData(env.NETEX_FTP_URL, env.GTFS_FTP_URL)
	ms.FailOnError(ctx, err, "failed to load static data")

	go staticData.StartRefreshLoop(ctx, env.REFRESH_HOURS)

	listener := tr.NewTr[SiriPayload](ctx, env.Env)
	err = listener.Start(ctx, func(ctx context.Context, raw *rdb.Raw[SiriPayload]) error {
		return handleMessage(ctx, raw, staticData)
	})
	ms.FailOnError(ctx, err, "listener stopped")
}

func handleMessage(ctx context.Context, raw *rdb.Raw[SiriPayload], staticData *StaticData) error {
	spanName := fmt.Sprintf("%s.transform-se", tel.GetServiceName())
	ctx, span := tel.TraceStart(ctx, spanName, trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	log := logger.Get(ctx)
	log.Info("Transforming SIRI SE to GTFS-RT", "provider", raw.Provider)

	payload, err := extractPayload(raw.Rawdata)
	ms.FailOnError(ctx, err, "failed to extract payload")

	se, err := siri.DeserializeSE(payload, siri.FormatJSON)
	ms.FailOnError(ctx, err, "failed to deserialize SIRI SE")

	resolver := staticData.GetResolver()
	resolver.ResetStats()
	rt := ConvertSE(se, resolver)

	rt.Header.Timestamp = time.Now().Unix()

	pbData, err := rt.Serialize(gtfsrt.FormatProtobuf)
	ms.FailOnError(ctx, err, "failed to serialize protobuf")
	err = putFile(ctx, env.FILESERVER_HOST, env.FILESERVER_PATH+".pb", pbData)
	ms.FailOnError(ctx, err, "failed to PUT protobuf")

	jsonData, err := rt.Serialize(gtfsrt.FormatJSON)
	ms.FailOnError(ctx, err, "failed to serialize JSON")
	err = putFile(ctx, env.FILESERVER_HOST, env.FILESERVER_PATH+".json", jsonData)

	log.Info("Transform completed", "entities", len(rt.Entity))
	return nil
}

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

func extractPayload(p SiriPayload) ([]byte, error) {
	if compress.IsCompressed(p.Metadata) {
		return compress.Decode(p.Payload)
	}
	return []byte(p.Payload), nil
}

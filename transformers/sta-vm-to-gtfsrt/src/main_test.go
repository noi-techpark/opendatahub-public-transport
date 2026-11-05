// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"testing"

	"github.com/noi-techpark/opendatahub-go-sdk/ingest/rdb"
)

func TestHandleMessageSkipsMalformedPayload(t *testing.T) {
	cases := map[string]SiriPayload{
		"html body":    {Payload: "<html><body>502 Bad Gateway</body></html>"},
		"bad compress": {Payload: "not-base64!", Metadata: map[string]string{"compressed": "gzip+base64"}},
	}
	for name, p := range cases {
		t.Run(name, func(t *testing.T) {
			raw := &rdb.Raw[SiriPayload]{Rawdata: p}
			if err := handleMessage(context.Background(), raw, nil); err == nil {
				t.Fatal("expected an error for a malformed payload")
			}
		})
	}
}

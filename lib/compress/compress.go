// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package compress provides gzip+base64 encoding/decoding for transporting
// binary-compressed data through JSON string fields.
package compress

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
)

const MetadataKey = "compressed"
const MetadataValue = "gzip+base64"

// Encode compresses data with gzip and returns a base64-encoded string
// suitable for embedding in a JSON string field.
func Encode(data []byte) (string, error) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		return "", fmt.Errorf("gzip write: %w", err)
	}
	if err := w.Close(); err != nil {
		return "", fmt.Errorf("gzip close: %w", err)
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// Decode reverses Encode: base64-decodes then gzip-decompresses.
func Decode(encoded string) ([]byte, error) {
	compressed, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}
	r, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, fmt.Errorf("gzip reader: %w", err)
	}
	defer r.Close()
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("gzip read: %w", err)
	}
	return data, nil
}

// IsCompressed checks metadata for the compression marker.
func IsCompressed(metadata map[string]string) bool {
	return metadata[MetadataKey] == MetadataValue
}

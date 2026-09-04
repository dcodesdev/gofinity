package main

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func encodePayload(t *testing.T, v any) string {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshalling the fixture: %v", err)
	}
	return base64.StdEncoding.EncodeToString(raw)
}

func TestDecodePayloadAcceptsAMinimalPayload(t *testing.T) {
	encoded := encodePayload(t, map[string]any{
		"files": []map[string]string{{"path": "main.go", "content": "package main"}},
	})

	p, err := DecodePayload(encoded)
	if err != nil {
		t.Fatalf("expected the payload to decode, got %v", err)
	}
	if len(p.Files) != 1 || p.Files[0].Path != "main.go" {
		t.Fatalf("files did not round-trip: %+v", p.Files)
	}
	if p.TimeoutMs != DefaultTimeoutMs {
		t.Errorf("timeoutMs = %d, want the default %d", p.TimeoutMs, DefaultTimeoutMs)
	}
	if strings.Join(p.Command, " ") != strings.Join(DefaultCommand, " ") {
		t.Errorf("command = %v, want the default %v", p.Command, DefaultCommand)
	}
}

func TestDecodePayloadAcceptsEveryBase64Variant(t *testing.T) {
	raw := []byte(`{"files":[{"path":"a/b.go","content":"x?>"}]}`)
	variants := map[string]string{
		"std":     base64.StdEncoding.EncodeToString(raw),
		"rawStd":  base64.RawStdEncoding.EncodeToString(raw),
		"url":     base64.URLEncoding.EncodeToString(raw),
		"rawURL":  base64.RawURLEncoding.EncodeToString(raw),
		"padded ": base64.StdEncoding.EncodeToString(raw) + "\n",
	}
	for name, encoded := range variants {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodePayload(encoded); err != nil {
				t.Fatalf("expected %s base64 to decode, got %v", name, err)
			}
		})
	}
}

func TestDecodePayloadRejections(t *testing.T) {
	oversizedFile := strings.Repeat("x", MaxFileBytes+1)

	tests := []struct {
		name    string
		encoded string
		want    string
	}{
		{
			name:    "empty input",
			encoded: "   ",
			want:    "payload is empty",
		},
		{
			name:    "not base64",
			encoded: "!!!not base64!!!",
			want:    "not valid base64",
		},
		{
			name:    "not JSON",
			encoded: base64.StdEncoding.EncodeToString([]byte("nonsense")),
			want:    "not valid JSON",
		},
		{
			name:    "trailing JSON",
			encoded: base64.StdEncoding.EncodeToString([]byte(`{"files":[{"path":"a.go","content":""}]} {}`)),
			want:    "exactly one JSON object",
		},
		{
			name:    "unknown field",
			encoded: base64.StdEncoding.EncodeToString([]byte(`{"files":[],"sneaky":true}`)),
			want:    "not valid JSON",
		},
		{
			name:    "no files",
			encoded: base64.StdEncoding.EncodeToString([]byte(`{"files":[]}`)),
			want:    "at least one file",
		},
		{
			name:    "absolute path",
			encoded: base64.StdEncoding.EncodeToString([]byte(`{"files":[{"path":"/etc/passwd","content":""}]}`)),
			want:    "must be relative",
		},
		{
			name:    "traversal",
			encoded: base64.StdEncoding.EncodeToString([]byte(`{"files":[{"path":"../escape.go","content":""}]}`)),
			want:    "`..` segment",
		},
		{
			name:    "duplicate paths",
			encoded: base64.StdEncoding.EncodeToString([]byte(`{"files":[{"path":"a.go","content":""},{"path":"a.go","content":""}]}`)),
			want:    "duplicate file path",
		},
		{
			name:    "oversized file",
			encoded: encodePayload(t, Payload{Files: []PayloadFile{{Path: "big.go", Content: oversizedFile}}}),
			want:    "per-file limit",
		},
		{
			name:    "too many files",
			encoded: encodePayload(t, Payload{Files: manyFiles(MaxFiles + 1)}),
			want:    "the limit is",
		},
		{
			name:    "oversized in total",
			encoded: encodePayload(t, Payload{Files: filesTotalling(MaxTotalBytes + 1)}),
			want:    "the limit is",
		},
		{
			name:    "non-go command",
			encoded: encodePayload(t, Payload{Files: manyFiles(1), Command: []string{"sh", "-c", "id"}}),
			want:    `command must start with "go"`,
		},
		{
			name:    "empty command argument",
			encoded: encodePayload(t, Payload{Files: manyFiles(1), Command: []string{"go", ""}}),
			want:    "empty argument",
		},
		{
			name:    "timeout too small",
			encoded: encodePayload(t, Payload{Files: manyFiles(1), TimeoutMs: 1}),
			want:    "below the",
		},
		{
			name:    "timeout too large",
			encoded: encodePayload(t, Payload{Files: manyFiles(1), TimeoutMs: MaxTimeoutMs + 1}),
			want:    "exceeds the",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodePayload(tt.encoded)
			if err == nil {
				t.Fatalf("expected a rejection mentioning %q, got none", tt.want)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error %q does not mention %q", err, tt.want)
			}
		})
	}
}

func TestDecodePayloadAllowsTheLimitsExactly(t *testing.T) {
	encoded := encodePayload(t, Payload{
		Files:     []PayloadFile{{Path: "big.go", Content: strings.Repeat("x", MaxFileBytes)}},
		TimeoutMs: MaxTimeoutMs,
	})
	if _, err := DecodePayload(encoded); err != nil {
		t.Fatalf("a payload exactly at the limits should be accepted, got %v", err)
	}
}

func manyFiles(n int) []PayloadFile {
	files := make([]PayloadFile, 0, n)
	for i := range n {
		files = append(files, PayloadFile{Path: "f" + itoa(i) + ".go", Content: ""})
	}
	return files
}

func filesTotalling(total int) []PayloadFile {
	var files []PayloadFile
	remaining := total
	for i := 0; remaining > 0; i++ {
		size := min(remaining, MaxFileBytes)
		files = append(files, PayloadFile{Path: "f" + itoa(i) + ".go", Content: strings.Repeat("x", size)})
		remaining -= size
	}
	return files
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

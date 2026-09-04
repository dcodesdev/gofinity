package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Limits on the decoded payload. They exist so a hostile or merely broken
// submission cannot exhaust the container's tmpfs or spend its whole budget
// writing files. They are deliberately generous compared to a real challenge
// (one `go.mod` and a handful of `.go` files) and deliberately far below the
// container's memory cap.
const (
	MaxFiles      = 64
	MaxFileBytes  = 256 * 1024
	MaxTotalBytes = 1024 * 1024
	MaxPathLen    = 255

	// DefaultTimeoutMs is used when the payload does not ask for one.
	DefaultTimeoutMs = 10_000
	// MaxTimeoutMs caps what a payload may ask for. The outer kill in the API
	// is the real limit; this stops a payload asking to outlive it by an hour.
	MaxTimeoutMs = 30_000
	MinTimeoutMs = 500
)

// DefaultCommand is what runs when the payload does not specify one.
// `-json` is what makes the output machine-readable; see parse.go.
var DefaultCommand = []string{"go", "test", "-json", "./..."}

// PayloadFile is one file to materialize into the scratch workspace.
type PayloadFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// Payload is the entire input to the runner: the workspace to create, the
// command to run in it, and how long it may take.
type Payload struct {
	Files     []PayloadFile `json:"files"`
	Command   []string      `json:"command,omitempty"`
	TimeoutMs int           `json:"timeoutMs,omitempty"`
}

// ErrEmptyPayload distinguishes "nothing was supplied" from "what was supplied
// is malformed", because the two mean different things to whoever is debugging.
var ErrEmptyPayload = errors.New("payload is empty")

// DecodePayload takes the base64 JSON the container is handed and returns a
// validated payload. Every rejection here is a rejection *before* any file is
// written or any process is started.
func DecodePayload(encoded string) (*Payload, error) {
	trimmed := strings.TrimSpace(encoded)
	if trimmed == "" {
		return nil, ErrEmptyPayload
	}

	// Accept both standard and URL-safe base64, with or without padding: the
	// producer is a JS runtime and which variant it reaches for is not worth
	// making a contract out of.
	raw, err := decodeBase64(trimmed)
	if err != nil {
		return nil, fmt.Errorf("payload is not valid base64: %w", err)
	}

	var p Payload
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&p); err != nil {
		return nil, fmt.Errorf("payload is not valid JSON: %w", err)
	}
	if decoder.More() {
		return nil, errors.New("payload must be exactly one JSON object")
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}
	return &p, nil
}

func decodeBase64(s string) ([]byte, error) {
	encodings := []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	}
	var lastErr error
	for _, enc := range encodings {
		raw, err := enc.DecodeString(s)
		if err == nil {
			return raw, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

// Validate applies every limit and every shape rule. It also normalizes the
// two optional fields, so callers never have to think about zero values.
func (p *Payload) Validate() error {
	if len(p.Files) == 0 {
		return errors.New("payload must contain at least one file")
	}
	if len(p.Files) > MaxFiles {
		return fmt.Errorf("payload has %d files, the limit is %d", len(p.Files), MaxFiles)
	}

	total := 0
	seen := make(map[string]struct{}, len(p.Files))
	for _, f := range p.Files {
		if err := ValidatePath(f.Path); err != nil {
			return err
		}
		if _, dup := seen[f.Path]; dup {
			return fmt.Errorf("%q: duplicate file path", f.Path)
		}
		seen[f.Path] = struct{}{}

		size := len(f.Content)
		if size > MaxFileBytes {
			return fmt.Errorf("%q: %d bytes exceeds the %d byte per-file limit", f.Path, size, MaxFileBytes)
		}
		total += size
	}
	if total > MaxTotalBytes {
		return fmt.Errorf("payload is %d bytes, the limit is %d", total, MaxTotalBytes)
	}

	if len(p.Command) == 0 {
		p.Command = append([]string(nil), DefaultCommand...)
	}
	// Defence in depth: the only producer is our own API, but the runner should
	// not be a way to execute an arbitrary binary if that ever stops being true.
	if p.Command[0] != "go" {
		return fmt.Errorf("command must start with %q, got %q", "go", p.Command[0])
	}
	for _, arg := range p.Command {
		if arg == "" {
			return errors.New("command must not contain an empty argument")
		}
	}

	switch {
	case p.TimeoutMs == 0:
		p.TimeoutMs = DefaultTimeoutMs
	case p.TimeoutMs < MinTimeoutMs:
		return fmt.Errorf("timeoutMs %d is below the %d ms minimum", p.TimeoutMs, MinTimeoutMs)
	case p.TimeoutMs > MaxTimeoutMs:
		return fmt.Errorf("timeoutMs %d exceeds the %d ms maximum", p.TimeoutMs, MaxTimeoutMs)
	}

	return nil
}

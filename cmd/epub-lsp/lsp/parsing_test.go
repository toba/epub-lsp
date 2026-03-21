package lsp

import (
	"bytes"
	"testing"

	"github.com/toba/lsp/transport"
)

func TestEncodeDecode(t *testing.T) {
	payload := []byte(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`)
	encoded := transport.Encode(payload)

	scanner := transport.NewScanner(bytes.NewReader(encoded))
	if !scanner.Scan() {
		t.Fatal("expected to scan a token")
	}

	decoded := scanner.Bytes()
	if !bytes.Equal(decoded, payload) {
		t.Errorf("decoded %q, want %q", decoded, payload)
	}
}

func TestDecodeMultipleMessages(t *testing.T) {
	msg1 := []byte(`{"id":1}`)
	msg2 := []byte(`{"id":2}`)

	var buf bytes.Buffer
	buf.Write(transport.Encode(msg1))
	buf.Write(transport.Encode(msg2))

	scanner := transport.NewScanner(&buf)

	if !scanner.Scan() {
		t.Fatal("expected first message")
	}
	if !bytes.Equal(scanner.Bytes(), msg1) {
		t.Errorf("first message: got %q, want %q", scanner.Bytes(), msg1)
	}

	if !scanner.Scan() {
		t.Fatal("expected second message")
	}
	if !bytes.Equal(scanner.Bytes(), msg2) {
		t.Errorf("second message: got %q, want %q", scanner.Bytes(), msg2)
	}
}

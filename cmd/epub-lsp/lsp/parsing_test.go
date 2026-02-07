package lsp

import (
	"bytes"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	payload := []byte(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`)
	encoded := Encode(payload)

	scanner := ReceiveInput(bytes.NewReader(encoded))
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
	buf.Write(Encode(msg1))
	buf.Write(Encode(msg2))

	scanner := ReceiveInput(&buf)

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

func TestGetHeaderContentLength(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		want    int
		wantErr bool
	}{
		{"valid", "Content-Length: 42", 42, false},
		{"with spaces", "Content-Length:  42 ", 42, false},
		{"missing header", "X-Other: 42", -1, true},
		{"missing colon", "Content-Length 42", -1, true},
		{"not a number", "Content-Length: abc", -1, true},
		{"negative", "Content-Length: -1", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getHeaderContentLength([]byte(tt.header))
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

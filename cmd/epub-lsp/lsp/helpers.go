package lsp

import (
	"encoding/json"
	"log/slog"

	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/lsp/position"
)

// marshalResponse creates a JSON-RPC response with the given result.
func marshalResponse[T any](id ID, result T) []byte {
	res := ResponseMessage[T]{
		JsonRpc: JSONRPCVersion,
		Id:      id,
		Result:  result,
	}
	data, err := json.Marshal(res)
	if err != nil {
		slog.Error("error marshalling response: " + err.Error())
		return nil
	}
	return data
}

// lspPos converts an epub.Position to an lsp.Position.
func lspPos(p epub.Position) Position {
	return Position{
		Line:      position.IntToUint(p.Line),
		Character: position.IntToUint(p.Character),
	}
}

// posToEpub converts an lsp.Position to an epub.Position.
func posToEpub(p Position) epub.Position {
	return epub.Position{
		Line:      int(p.Line),      //nolint:gosec // LSP line numbers fit in int
		Character: int(p.Character), //nolint:gosec // LSP character numbers fit in int
	}
}

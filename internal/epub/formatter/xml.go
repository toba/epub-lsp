// Package formatter provides XML and CSS formatting for EPUB documents.
package formatter

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"strings"
)

// FormatXML reformats XML content with consistent indentation.
// It preserves the XML declaration if present and re-indents elements.
func FormatXML(content []byte, indent string) (string, error) {
	decoder := xml.NewDecoder(bytes.NewReader(content))
	var buf bytes.Buffer

	// Check for XML declaration
	if bytes.HasPrefix(bytes.TrimSpace(content), []byte("<?xml")) {
		idx := bytes.Index(content, []byte("?>"))
		if idx >= 0 {
			decl := string(bytes.TrimSpace(content[:idx+2]))
			buf.WriteString(decl)
			buf.WriteByte('\n')
		}
	}

	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", indent)

	var started bool
	for {
		tok, err := decoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", err
		}

		switch t := tok.(type) {
		case xml.ProcInst:
			// Skip xml declaration (already handled)
			if t.Target == "xml" {
				continue
			}
			if err := encoder.EncodeToken(t); err != nil {
				return "", err
			}
			started = true
		case xml.StartElement:
			if err := encoder.EncodeToken(t); err != nil {
				return "", err
			}
			started = true
		case xml.EndElement:
			if err := encoder.EncodeToken(t); err != nil {
				return "", err
			}
		case xml.CharData:
			text := strings.TrimSpace(string(t))
			if text != "" {
				if err := encoder.EncodeToken(xml.CharData(text)); err != nil {
					return "", err
				}
			}
		case xml.Comment:
			if err := encoder.EncodeToken(t); err != nil {
				return "", err
			}
			started = true
		case xml.Directive:
			if err := encoder.EncodeToken(t); err != nil {
				return "", err
			}
			started = true
		}
	}

	if started {
		if err := encoder.Flush(); err != nil {
			return "", err
		}
	}

	result := buf.String()
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}

	return result, nil
}

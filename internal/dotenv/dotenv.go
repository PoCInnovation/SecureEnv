// Package dotenv reads and writes .env files.
//
// The supported syntax is the common subset understood by docker compose,
// direnv and godotenv: KEY=value, optional "export" prefix, # comments,
// single quotes (literal) and double quotes (with \n, \r, \t, \" and \\
// escapes). Variable expansion is intentionally not supported.
package dotenv

import (
	"bufio"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"

	"github.com/PoCInnovation/SecureEnv/internal/domain"
)

const maxLineSize = 1 << 20

// SyntaxError reports a malformed line.
type SyntaxError struct {
	Line   int
	Reason string
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("dotenv: line %d: %s", e.Line, e.Reason)
}

// File is the content of a .env file.
type File struct{ values map[string]string }

// New builds a File from values.
func New(values map[string]string) File {
	return File{values: maps.Clone(values)}
}

// Values returns a copy of every entry.
func (f File) Values() map[string]string {
	out := make(map[string]string, len(f.values))
	maps.Copy(out, f.values)
	return out
}

// Reserved returns the SecureEnv configuration entries.
func (f File) Reserved() map[string]string {
	return f.filter(true)
}

// Unreserved returns the entries that belong to the project.
func (f File) Unreserved() map[string]string {
	return f.filter(false)
}

func (f File) filter(reserved bool) map[string]string {
	out := make(map[string]string)
	for key, value := range f.values {
		if domain.IsReserved(key) == reserved {
			out[key] = value
		}
	}
	return out
}

// Parse reads a .env document.
func Parse(r io.Reader) (File, error) {
	values := make(map[string]string)
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), maxLineSize)

	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		key, value, ok, err := parseLine(strings.TrimSuffix(scanner.Text(), "\r"))
		if err != nil {
			return File{}, &SyntaxError{Line: lineNumber, Reason: err.Error()}
		}
		if ok {
			values[key] = value
		}
	}
	if err := scanner.Err(); err != nil {
		return File{}, fmt.Errorf("dotenv: %w", err)
	}
	return File{values: values}, nil
}

func parseLine(line string) (key, value string, ok bool, err error) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return "", "", false, nil
	}
	trimmed = strings.TrimPrefix(trimmed, "export ")

	rawKey, rawValue, found := strings.Cut(trimmed, "=")
	if !found {
		return "", "", false, fmt.Errorf("expected KEY=value")
	}
	key = strings.TrimSpace(rawKey)
	if key == "" || strings.ContainsAny(key, " \t") {
		return "", "", false, fmt.Errorf("invalid key %q", key)
	}

	value, err = parseValue(strings.TrimSpace(rawValue))
	if err != nil {
		return "", "", false, err
	}
	return key, value, true, nil
}

func parseValue(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}
	switch raw[0] {
	case '\'':
		end := strings.IndexByte(raw[1:], '\'')
		if end < 0 {
			return "", fmt.Errorf("unterminated single quote")
		}
		return raw[1 : end+1], checkTrailing(raw[end+2:])
	case '"':
		return parseDoubleQuoted(raw)
	default:
		if idx := strings.Index(raw, " #"); idx >= 0 {
			raw = raw[:idx]
		}
		return strings.TrimSpace(raw), nil
	}
}

func parseDoubleQuoted(raw string) (string, error) {
	var sb strings.Builder
	for i := 1; i < len(raw); i++ {
		switch c := raw[i]; c {
		case '"':
			return sb.String(), checkTrailing(raw[i+1:])
		case '\\':
			if i+1 == len(raw) {
				return "", fmt.Errorf("unterminated double quote")
			}
			i++
			sb.WriteString(unescape(raw[i]))
		default:
			sb.WriteByte(c)
		}
	}
	return "", fmt.Errorf("unterminated double quote")
}

func unescape(c byte) string {
	switch c {
	case 'n':
		return "\n"
	case 'r':
		return "\r"
	case 't':
		return "\t"
	case '"', '\\':
		return string(c)
	default:
		return "\\" + string(c)
	}
}

func checkTrailing(rest string) error {
	rest = strings.TrimSpace(rest)
	if rest != "" && !strings.HasPrefix(rest, "#") {
		return fmt.Errorf("unexpected content after closing quote: %q", rest)
	}
	return nil
}

var escaper = strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\r", `\r`, "\t", `\t`)

// Format writes f with SecureEnv entries first, then project entries, each
// group sorted by key. Every value is double quoted so the output is stable.
func (f File) Format(w io.Writer) error {
	reserved := slices.Sorted(maps.Keys(f.Reserved()))
	others := slices.Sorted(maps.Keys(f.Unreserved()))

	bw := bufio.NewWriter(w)
	writeGroup(bw, f.values, reserved)
	if len(reserved) > 0 && len(others) > 0 {
		bw.WriteByte('\n')
	}
	writeGroup(bw, f.values, others)
	return bw.Flush()
}

func writeGroup(w *bufio.Writer, values map[string]string, keys []string) {
	for _, key := range keys {
		fmt.Fprintf(w, "%s=\"%s\"\n", key, escaper.Replace(values[key]))
	}
}

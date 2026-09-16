package dotenv_test

import (
	"errors"
	"maps"
	"strings"
	"testing"

	"github.com/PoCInnovation/SecureEnv/internal/dotenv"
)

func TestParse(t *testing.T) {
	t.Parallel()

	input := `# a comment
SECURE_ENV_PROJECT="backend"

export EXPORTED=yes
PLAIN=value # trailing comment
EQUALS_IN_VALUE="postgres://u:p@host/db?sslmode=disable&a=b"
UNQUOTED_EQUALS=a=b
SINGLE='kept $literal \n'
DOUBLE="line1\nline2 \"quoted\" \\ end"
EMPTY=
HASH_IN_QUOTES="not # a comment"
  SPACED  =  trimmed  
`
	got, err := dotenv.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	want := map[string]string{
		"SECURE_ENV_PROJECT": "backend",
		"EXPORTED":           "yes",
		"PLAIN":              "value",
		"EQUALS_IN_VALUE":    "postgres://u:p@host/db?sslmode=disable&a=b",
		"UNQUOTED_EQUALS":    "a=b",
		"SINGLE":             `kept $literal \n`,
		"DOUBLE":             "line1\nline2 \"quoted\" \\ end",
		"EMPTY":              "",
		"HASH_IN_QUOTES":     "not # a comment",
		"SPACED":             "trimmed",
	}
	if !maps.Equal(got.Values(), want) {
		t.Fatalf("Parse() =\n%v\nwant\n%v", got.Values(), want)
	}
}

func TestParseErrors(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"missing equals":       "JUST_A_KEY\n",
		"empty key":            "=value\n",
		"unterminated double":  "A=\"open\n",
		"unterminated single":  "A='open\n",
		"garbage after quotes": "A=\"x\" y\n",
	}
	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := dotenv.Parse(strings.NewReader(input))
			var syntaxErr *dotenv.SyntaxError
			if !errors.As(err, &syntaxErr) {
				t.Fatalf("Parse(%q) error = %v, want *SyntaxError", input, err)
			}
			if syntaxErr.Line != 1 {
				t.Fatalf("Line = %d, want 1", syntaxErr.Line)
			}
		})
	}
}

func TestParseLastAssignmentWins(t *testing.T) {
	t.Parallel()

	got, err := dotenv.Parse(strings.NewReader("A=1\nA=2\n"))
	if err != nil {
		t.Fatal(err)
	}
	if v := got.Values()["A"]; v != "2" {
		t.Fatalf("A = %q, want 2", v)
	}
}

func TestFileSplitReserved(t *testing.T) {
	t.Parallel()

	file, err := dotenv.Parse(strings.NewReader("SECURE_ENV_TOKEN=t\nAPP=1\n"))
	if err != nil {
		t.Fatal(err)
	}
	if got := file.Reserved(); !maps.Equal(got, map[string]string{"SECURE_ENV_TOKEN": "t"}) {
		t.Errorf("Reserved() = %v", got)
	}
	if got := file.Unreserved(); !maps.Equal(got, map[string]string{"APP": "1"}) {
		t.Errorf("Unreserved() = %v", got)
	}
}

func TestFormat(t *testing.T) {
	t.Parallel()

	file := dotenv.New(map[string]string{
		"ZED":                "last",
		"APP":                "has \"quotes\"\nand newline",
		"SECURE_ENV_PROJECT": "backend",
	})

	var sb strings.Builder
	if err := file.Format(&sb); err != nil {
		t.Fatal(err)
	}

	want := `SECURE_ENV_PROJECT="backend"

APP="has \"quotes\"\nand newline"
ZED="last"
`
	if sb.String() != want {
		t.Fatalf("Format() =\n%s\nwant\n%s", sb.String(), want)
	}
}

func TestFormatParseRoundTrip(t *testing.T) {
	t.Parallel()

	values := map[string]string{
		"A": `back\slash`,
		"B": "tab\tcr\r",
		"C": "$HOME # not a comment",
		"D": "",
		"E": "'single'",
	}
	assertRoundTrip(t, values)
}

func FuzzFormatParseRoundTrip(f *testing.F) {
	for _, seed := range []string{"", "plain", `a"b`, "multi\nline", `\n`, "#", "'"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, value string) {
		assertRoundTrip(t, map[string]string{"KEY": value})
	})
}

func assertRoundTrip(t *testing.T, values map[string]string) {
	t.Helper()

	var sb strings.Builder
	if err := dotenv.New(values).Format(&sb); err != nil {
		t.Fatal(err)
	}
	parsed, err := dotenv.Parse(strings.NewReader(sb.String()))
	if err != nil {
		t.Fatalf("Parse(Format(%q)) error: %v\n%s", values, err, sb.String())
	}
	if !maps.Equal(parsed.Values(), values) {
		t.Fatalf("round trip mismatch:\n got %q\nwant %q", parsed.Values(), values)
	}
}

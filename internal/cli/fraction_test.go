package cli

import (
	"strings"
	"testing"
)

func TestRunFractions(t *testing.T) {
	for _, tt := range []struct {
		args    []string
		decimal []string
	}{
		{[]string{"1/2in"}, []string{".5in"}},
		{[]string{"3/2in", "cm"}, []string{"1.5in", "cm"}},
		{[]string{"2/4", "cup", "ml"}, []string{".5", "cup", "ml"}},
		{[]string{"1-1/2", "in", "mm"}, []string{"1.5", "in", "mm"}},
		{[]string{"-1-1/2in", "mm"}, []string{"-1.5in", "mm"}},
		{[]string{"-1/2", "f", "c"}, []string{"-.5", "f", "c"}},
		{[]string{"1/2m/s", "km/h"}, []string{".5m/s", "km/h"}},
		{[]string{"+1/2kg", "g"}, []string{"+.5kg", "g"}},
		{[]string{"1/2in,", "1-1/2", "in", "mm,", "2/4cup"}, []string{".5in,", "1.5", "in", "mm,", ".5cup"}},
	} {
		code, stdout, stderr := run(tt.args...)
		wantCode, wantOut, wantErr := run(tt.decimal...)
		if code != 0 || wantCode != 0 || stdout != wantOut || stderr != "" || wantErr != "" {
			t.Errorf("Run(%q) = (%d, %q, %q), decimal = (%d, %q, %q)", tt.args, code, stdout, stderr, wantCode, wantOut, wantErr)
		}
	}
}

func TestRunSignedMixedNumberConversionList(t *testing.T) {
	code, stdout, stderr := run("-1-1/2in", "mm,", "1/2cup", "ml")
	want := "-38.1 mm\n\n118.3 ml\n"
	if code != 0 || stdout != want || stderr != "" {
		t.Fatalf("Run() = (%d, %q, %q), want (0, %q, empty)", code, stdout, stderr, want)
	}
}

func TestRunInvalidFractionBatchesAreAtomic(t *testing.T) {
	huge := "1" + strings.Repeat("0", 400)
	for _, raw := range []string{"1/0in", "1/2/3in", "1/2e3in", "1 1/2in", huge + "/1in", "1/" + huge + "in"} {
		code, stdout, stderr := run("1/2in,", raw, ",", "72f")
		if code != 2 || stdout != "" || !strings.Contains(stderr, "measurement 2:") || !strings.Contains(stderr, raw) {
			t.Errorf("Run invalid %q = (%d, %q, %q), want indexed error without stdout", raw, code, stdout, stderr)
		}
	}
}

func TestRunFractionRequiresOneNumericToken(t *testing.T) {
	code, stdout, stderr := run("1", "1/2", "in")
	if code != 2 || stdout != "" || stderr == "" {
		t.Fatalf("spaced mixed number = (%d, %q, %q)", code, stdout, stderr)
	}
}

func TestFractionHelp(t *testing.T) {
	code, stdout, stderr := run("--help")
	for _, want := range []string{"1/2in", "1-1/2", "-1-1/2", "one shell token", "Unicode", "expressions", "exponents"} {
		if code != 0 || stderr != "" || !strings.Contains(stdout, want) {
			t.Errorf("help missing %q: (%d, %q, %q)", want, code, stdout, stderr)
		}
	}
}

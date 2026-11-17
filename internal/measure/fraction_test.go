package measure

import (
	"math"
	"math/big"
	"strings"
	"testing"
)

func TestParseFractions(t *testing.T) {
	huge := "1" + strings.Repeat("0", 400)
	tests := []struct {
		value string
		want  float64
	}{
		{"1/2", 0.5}, {"3/2", 1.5}, {"2/4", 0.5},
		{"1-1/2", 1.5}, {"1-3/2", 2.5}, {"0-1/2", 0.5},
		{"-1/2", -0.5}, {"-1-1/2", -1.5}, {"-0-1/2", -0.5},
		{"+1/2", 0.5}, {"+1-1/2", 1.5}, {"001-002/004", 1.5}, {"008/016", 0.5},
		{"0/2", 0}, {"-0/2", math.Copysign(0, -1)},
		{huge + "/" + huge, 1}, {"0/" + huge, 0}, {"1-1/" + huge, 1},
		{"1/" + "1" + strings.Repeat("0", 323), 1e-323},
	}
	for _, tt := range tests {
		for _, args := range [][]string{{tt.value + "in", "cm"}, {tt.value, "in", "cm"}, {tt.value + " in"}} {
			t.Run(strings.Join(args, " "), func(t *testing.T) {
				m, target, err := Parse(args)
				if err != nil {
					t.Fatal(err)
				}
				if m.Value != tt.want || math.Signbit(m.Value) != math.Signbit(tt.want) || m.Unit.Symbol != "in" {
					t.Fatalf("Parse(%q) = %#v, want %g in", args, m, tt.want)
				}
				if len(args) > 1 && (target == nil || target.Symbol != "cm") {
					t.Fatalf("target = %#v, want cm", target)
				}
			})
		}
	}
}

func TestParseRejectsInvalidFractions(t *testing.T) {
	for _, raw := range []string{
		"1/0", "1/000", "1-1/0", "0/0", "1/", "/2", "1//2", "1/2/3",
		"1--1/2", "1-2-3/4", "1-2", "--1/2", "1/-2", "1/+2", "1-/2",
		".5/2", "1.5/2", "1/2.5", "1e3/2", "1/2e3", "1-1.5/2",
		"1 1/2", "1 /2", "1/ 2", "1 -1/2", "1- 1/2", "½",
	} {
		for _, args := range [][]string{{raw + "in"}, {raw, "in"}} {
			t.Run(strings.Join(args, " "), func(t *testing.T) {
				_, _, err := Parse(args)
				if err == nil || !strings.Contains(err.Error(), "invalid measurement") || !strings.Contains(err.Error(), raw) {
					t.Fatalf("Parse(%q) error = %v, want invalid measurement identifying %q", args, err, raw)
				}
			})
		}
	}
}

func TestParseFractionRange(t *testing.T) {
	huge := "1" + strings.Repeat("0", 400)
	for _, raw := range []string{huge + "/1", "1/" + huge, "-1/" + huge, huge + "-1/2"} {
		for _, args := range [][]string{{raw + "in"}, {raw, "in"}} {
			_, _, err := Parse(args)
			if err == nil || !strings.Contains(err.Error(), "outside supported range") {
				t.Errorf("Parse(%q) error = %v, want range error", args, err)
			}
		}
	}
}

func TestParseFractionFloat64Boundaries(t *testing.T) {
	pow2 := func(exponent uint) *big.Int { return new(big.Int).Lsh(big.NewInt(1), exponent) }
	max := new(big.Int).Sub(pow2(1024), pow2(971)).String()
	for _, tt := range []struct {
		raw  string
		want float64
	}{
		{max + "/1", math.MaxFloat64},
		{"1/" + pow2(1074).String(), math.SmallestNonzeroFloat64},
		{"3/" + pow2(1075).String(), 2 * math.SmallestNonzeroFloat64},
	} {
		m, _, err := Parse([]string{tt.raw, "in"})
		if err != nil || m.Value != tt.want {
			t.Errorf("Parse(%q) value = %g, error = %v, want %g", tt.raw, m.Value, err, tt.want)
		}
	}
	for _, raw := range []string{pow2(1024).String() + "/1", "1/" + pow2(1075).String()} {
		if _, _, err := Parse([]string{raw + "in"}); err == nil || !strings.Contains(err.Error(), "outside supported range") {
			t.Errorf("Parse(%q) error = %v, want range error", raw, err)
		}
	}
}

func TestFractionUnitAndTargetErrors(t *testing.T) {
	for _, tt := range []struct {
		args []string
		want string
	}{
		{[]string{"1/2"}, "missing unit"},
		{[]string{"1/2parsecs"}, "unknown unit"},
		{[]string{"1/2", "parsecs"}, "unknown unit"},
		{[]string{"1/2in", "2/3cm"}, "unknown unit"},
		{[]string{"1/2", "in", "2/3cm"}, "unknown unit"},
		{[]string{"1/2", "in", "cm", "extra"}, "too many arguments"},
	} {
		if _, _, err := Parse(tt.args); err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("Parse(%q) error = %v, want %q", tt.args, err, tt.want)
		}
	}
}

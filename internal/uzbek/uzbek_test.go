package uzbek

import "testing"

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		"oʻzbek":   "ozbek",
		"o'zbek":   "ozbek",
		"gʻalaba":  "galaba",
		"g'alaba":  "galaba",
		"Toshkent": "Toshkent",
	}
	for in, want := range cases {
		if got := Normalize(in); got != want {
			t.Errorf("Normalize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLatinCyrillicRoundish(t *testing.T) {
	// Latin -> Cyrillic should produce Cyrillic output.
	cyr := LatinToCyrillic("akmal")
	if !HasCyrillic(cyr) {
		t.Fatalf("LatinToCyrillic(akmal) = %q, expected Cyrillic", cyr)
	}
	// Cyrillic -> Latin of a known word.
	if got := CyrillicToLatin("Акмал"); got != "akmal" {
		t.Errorf("CyrillicToLatin(Акмал) = %q, want akmal", got)
	}
}

func TestVariants(t *testing.T) {
	v := Variants("oʻktam", false)
	if len(v) == 0 {
		t.Fatal("expected variants, got none")
	}
	found := false
	for _, s := range v {
		if s == "oktam" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected normalized variant 'oktam' in %v", v)
	}
	// With cyrillic off, no Cyrillic variant should appear.
	for _, s := range v {
		if HasCyrillic(s) {
			t.Errorf("did not expect Cyrillic variant with cyrillic=false: %q", s)
		}
	}
}

func TestVariantsCyrillicCleanOnly(t *testing.T) {
	// "white" transliterates with a leftover 'w' -> must be dropped.
	for _, s := range Variants("white", true) {
		if HasCyrillic(s) {
			t.Errorf("broken Cyrillic variant emitted for 'white': %q", s)
		}
	}
	// "akmal" maps cleanly -> a Cyrillic variant should be present.
	hasCyr := false
	for _, s := range Variants("akmal", true) {
		if HasCyrillic(s) {
			hasCyr = true
		}
	}
	if !hasCyr {
		t.Error("expected a clean Cyrillic variant for 'akmal'")
	}
}

func TestTokenize(t *testing.T) {
	got := Tokenize("black, white, blue, qora, oq, ko'k")
	want := []string{"black", "white", "blue", "qora", "oq", "ko'k"}
	if len(got) != len(want) {
		t.Fatalf("Tokenize = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("token %d = %q, want %q", i, got[i], want[i])
		}
	}
	// Multi-word field splits too.
	if g := Tokenize("IT House"); len(g) != 2 || g[0] != "IT" || g[1] != "House" {
		t.Errorf("Tokenize(IT House) = %v, want [IT House]", g)
	}
}

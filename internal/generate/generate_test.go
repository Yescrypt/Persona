package generate

import (
	"strings"
	"testing"

	"github.com/Yescript/persona/internal/profile"
)

func sampleProfile() *profile.Profile {
	return &profile.Profile{
		Name:      "Ali",
		Surname:   "Valiyev",
		BirthYear: "1999",
		Phones:    []string{"998901234567"},
	}
}

func TestGenerateDedupe(t *testing.T) {
	res := Generate(sampleProfile(), DefaultOptions())
	if len(res.Words) == 0 {
		t.Fatal("expected words, got none")
	}
	seen := map[string]bool{}
	for _, w := range res.Words {
		if seen[w] {
			t.Fatalf("duplicate word in output: %q", w)
		}
		seen[w] = true
	}
}

func TestGenerateContainsNameYear(t *testing.T) {
	res := Generate(sampleProfile(), DefaultOptions())
	want := map[string]bool{"ali1999": false, "Ali1999": false}
	for _, w := range res.Words {
		if _, ok := want[w]; ok {
			want[w] = true
		}
	}
	for w, got := range want {
		if !got {
			t.Errorf("expected %q in output", w)
		}
	}
}

func TestMinMaxFilter(t *testing.T) {
	opts := DefaultOptions()
	opts.Min = 6
	opts.Max = 10
	res := Generate(sampleProfile(), opts)
	for _, w := range res.Words {
		n := len([]rune(w))
		if n < 6 || n > 10 {
			t.Fatalf("word %q length %d violates min/max", w, n)
		}
	}
}

func TestMaxWordsRespected(t *testing.T) {
	opts := DefaultOptions()
	opts.MaxWords = 25
	res := Generate(sampleProfile(), opts)
	if len(res.Words) > 25 {
		t.Fatalf("got %d words, want <= 25", len(res.Words))
	}
}

func TestRequireDigit(t *testing.T) {
	opts := DefaultOptions()
	opts.RequireDigit = true
	res := Generate(sampleProfile(), opts)
	for _, w := range res.Words {
		if !strings.ContainsAny(w, "0123456789") {
			t.Fatalf("word %q has no digit but require-digit is set", w)
		}
	}
}

func TestNoCommaInOutput(t *testing.T) {
	// Comma-separated field input must be split into tokens, never emitted
	// as a single string containing commas.
	p := &profile.Profile{
		Name:  "Ali",
		Color: "black,white,blue,qora,oq,ko'k",
	}
	res := Generate(p, DefaultOptions())
	for _, w := range res.Words {
		if strings.ContainsAny(w, ", ") {
			t.Fatalf("output contains comma/space: %q", w)
		}
	}
	// The split tokens should be present.
	set := map[string]bool{}
	for _, w := range res.Words {
		set[w] = true
	}
	for _, want := range []string{"black", "white", "qora", "kok"} {
		if !set[want] {
			t.Errorf("expected token %q from comma-separated field", want)
		}
	}
}

func TestDefaultIsLatinOnly(t *testing.T) {
	p := &profile.Profile{Name: "Akmal", Surname: "Valiyev"}
	res := Generate(p, DefaultOptions())
	for _, w := range res.Words {
		for _, r := range w {
			if r >= 'Ѐ' && r <= 'ӿ' {
				t.Fatalf("default output should be Latin-only, got Cyrillic in %q", w)
			}
		}
	}
}

func TestCyrillicOptIn(t *testing.T) {
	p := &profile.Profile{Name: "Akmal"}
	opts := DefaultOptions()
	opts.Cyrillic = true
	res := Generate(p, opts)
	found := false
	for _, w := range res.Words {
		for _, r := range w {
			if r >= 'Ѐ' && r <= 'ӿ' {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected Cyrillic variants when Cyrillic option is on")
	}
}

func TestReverseTruncateToggle(t *testing.T) {
	opts := DefaultOptions()
	opts.Reverse = false
	opts.Truncate = false
	opts.Leet = false
	res := Generate(&profile.Profile{Name: "Paxtakor"}, opts)
	for _, w := range res.Words {
		lw := strings.ToLower(w)
		if lw == "rokatxap" { // reverse of paxtakor
			t.Errorf("reverse token present despite Reverse=false: %q", w)
		}
		if lw == "pax" || lw == "paxt" { // truncations
			t.Errorf("truncated token present despite Truncate=false: %q", w)
		}
	}
	// Sanity: the real token is still there.
	found := false
	for _, w := range res.Words {
		if strings.ToLower(w) == "paxtakor" {
			found = true
		}
	}
	if !found {
		t.Error("expected base token 'paxtakor' to remain")
	}
}

func TestEmptyProfileNoPanic(t *testing.T) {
	res := Generate(&profile.Profile{}, DefaultOptions())
	// Auto-added common passwords should still be present.
	if len(res.Words) == 0 {
		t.Fatal("expected auto-added common passwords for empty profile")
	}
}

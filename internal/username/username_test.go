package username

import "testing"

import "github.com/Yescript/persona/internal/profile"

func TestTemplates(t *testing.T) {
	p := &profile.Profile{Name: "Ali", Surname: "Valiyev"}
	got := Generate(p, Options{})
	set := map[string]bool{}
	for _, u := range got {
		set[u] = true
	}
	for _, want := range []string{"ali", "valiyev", "alivaliyev", "ali.valiyev", "avaliyev", "a.valiyev", "valiyevali"} {
		if !set[want] {
			t.Errorf("expected username %q in output", want)
		}
	}
}

func TestYearSuffix(t *testing.T) {
	p := &profile.Profile{Name: "Ali", Surname: "Valiyev", BirthYear: "1999"}
	got := Generate(p, Options{Years: true})
	set := map[string]bool{}
	for _, u := range got {
		set[u] = true
	}
	if !set["ali1999"] {
		t.Errorf("expected ali1999 in output")
	}
}

func TestNormTransliterates(t *testing.T) {
	if got := norm("Акмал"); got != "akmal" {
		t.Errorf("norm(Акмал) = %q, want akmal", got)
	}
	if got := norm("Oʻktam"); got != "oktam" {
		t.Errorf("norm(Oʻktam) = %q, want oktam", got)
	}
}

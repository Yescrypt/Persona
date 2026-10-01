// Package username generates candidate usernames from a profile using
// Username-Anarchy-style templates (own implementation).
package username

import (
	"strings"

	"github.com/Yescript/persona/internal/profile"
	"github.com/Yescript/persona/internal/uzbek"
)

// Options controls username generation.
type Options struct {
	Leet  bool
	Years bool
}

// Generate returns a deduped, ordered list of candidate usernames.
func Generate(p *profile.Profile, opts Options) []string {
	out := newOrdered()

	for _, u := range p.Usernames {
		out.add(strings.ToLower(strings.TrimSpace(u)))
	}
	for _, n := range p.Nicknames {
		out.add(norm(n))
	}

	first := norm(p.Name)
	last := norm(p.Surname)
	if first != "" {
		out.add(first)
	}
	if last != "" {
		out.add(last)
	}
	if first != "" && last != "" {
		fi := first[:1]
		li := last[:1]
		for _, t := range []string{
			first + last,
			first + "." + last,
			fi + last,
			fi + "." + last,
			first + li,
			last + first,
			first + "_" + last,
			capitalize(first) + "." + capitalize(last),
		} {
			out.add(t)
		}
	}

	// number / year suffixes over the base set
	base := out.slice()
	for _, b := range base {
		if opts.Years && p.BirthYear != "" {
			y := digits(p.BirthYear)
			out.add(b + y)
			if len(y) == 4 {
				out.add(b + y[2:])
			}
		}
		for _, n := range []string{"1", "123", "07"} {
			out.add(b + n)
		}
	}

	if opts.Leet {
		for _, b := range out.slice() {
			if l := leet(b); l != "" && l != b {
				out.add(l)
			}
		}
	}
	return out.slice()
}

func norm(s string) string {
	s = strings.ToLower(uzbek.ToLatin(strings.TrimSpace(s)))
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func capitalize(w string) string {
	if w == "" {
		return w
	}
	return strings.ToUpper(w[:1]) + w[1:]
}

func digits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

var leetReplacer = strings.NewReplacer("a", "@", "o", "0", "i", "1", "e", "3", "s", "5")

func leet(s string) string { return leetReplacer.Replace(s) }

type ordered struct {
	seen map[string]bool
	list []string
}

func newOrdered() *ordered { return &ordered{seen: map[string]bool{}} }

func (o *ordered) add(s string) {
	s = strings.TrimSpace(s)
	if s == "" || o.seen[s] {
		return
	}
	o.seen[s] = true
	o.list = append(o.list, s)
}

func (o *ordered) slice() []string { return o.list }

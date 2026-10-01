// Package uzbek provides Uzbek-specific normalization, Latin<->Cyrillic
// transliteration, and common local password/keyboard tokens used to seed
// the wordlist generator.
package uzbek

import (
	"strings"
	"unicode"
)

// apostrophe-like characters used for the oʻ / gʻ letters in Uzbek Latin.
var apostrophes = []string{"ʻ", "ʼ", "'", "`", "‘", "’"}

// isApostrophe reports whether r is one of the apostrophe-like runes that may
// appear inside an Uzbek word (and must not split it into tokens).
func isApostrophe(r rune) bool {
	switch r {
	case 'ʻ', 'ʼ', '\'', '`', '‘', '’':
		return true
	}
	return false
}

// Tokenize splits a free-text field into individual word tokens, breaking on
// anything that is not a letter, digit, or in-word apostrophe. This turns
// "black, white, ko'k" or "IT House" into separate tokens instead of one
// nonsensical string.
func Tokenize(s string) []string {
	var toks []string
	var b strings.Builder
	flush := func() {
		t := strings.TrimFunc(b.String(), isApostrophe)
		if t != "" {
			toks = append(toks, t)
		}
		b.Reset()
	}
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || isApostrophe(r) {
			b.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	return toks
}

func hasASCIILetter(s string) bool {
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			return true
		}
	}
	return false
}

// Normalize folds Uzbek-specific letters to plain ASCII:
//
//	oʻ / o'  -> o
//	gʻ / g'  -> g
//
// and strips any remaining stray modifier apostrophes.
func Normalize(s string) string {
	for _, a := range apostrophes {
		s = strings.ReplaceAll(s, "o"+a, "o")
		s = strings.ReplaceAll(s, "O"+a, "O")
		s = strings.ReplaceAll(s, "g"+a, "g")
		s = strings.ReplaceAll(s, "G"+a, "G")
	}
	for _, a := range apostrophes {
		s = strings.ReplaceAll(s, a, "")
	}
	return s
}

// HasCyrillic reports whether s contains any Cyrillic letter.
func HasCyrillic(s string) bool {
	for _, r := range s {
		if r >= 'Ѐ' && r <= 'ӿ' {
			return true
		}
	}
	return false
}

var latToCyrDigraphs = []struct{ lat, cyr string }{
	{"yo", "ё"}, {"yu", "ю"}, {"ya", "я"},
	{"sh", "ш"}, {"ch", "ч"}, {"ng", "нг"}, {"ts", "ц"},
}

var latToCyrSingle = map[rune]string{
	'a': "а", 'b': "б", 'c': "с", 'd': "д", 'e': "е", 'f': "ф",
	'g': "г", 'h': "ҳ", 'i': "и", 'j': "ж", 'k': "к", 'l': "л",
	'm': "м", 'n': "н", 'o': "о", 'p': "п", 'q': "қ", 'r': "р",
	's': "с", 't': "т", 'u': "у", 'v': "в", 'x': "х", 'y': "й",
	'z': "з",
}

// LatinToCyrillic transliterates an Uzbek Latin string to Cyrillic.
// It is a pragmatic (lossy) mapping aimed at generating plausible variants.
func LatinToCyrillic(s string) string {
	s = strings.ToLower(s)
	for _, a := range apostrophes {
		s = strings.ReplaceAll(s, "o"+a, "ў")
		s = strings.ReplaceAll(s, "g"+a, "ғ")
	}
	for _, d := range latToCyrDigraphs {
		s = strings.ReplaceAll(s, d.lat, d.cyr)
	}
	var b strings.Builder
	for _, r := range s {
		if c, ok := latToCyrSingle[r]; ok {
			b.WriteString(c)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

var cyrToLat = map[rune]string{
	'а': "a", 'б': "b", 'в': "v", 'г': "g", 'ғ': "g", 'д': "d",
	'е': "e", 'ё': "yo", 'ж': "j", 'з': "z", 'и': "i", 'й': "y",
	'к': "k", 'қ': "q", 'л': "l", 'м': "m", 'н': "n", 'о': "o",
	'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u", 'ў': "o",
	'ф': "f", 'х': "x", 'ҳ': "h", 'ц': "ts", 'ч': "ch", 'ш': "sh",
	'щ': "sh", 'ъ': "", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya",
}

// CyrillicToLatin transliterates an Uzbek Cyrillic string to Latin.
func CyrillicToLatin(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if c, ok := cyrToLat[r]; ok {
			b.WriteString(c)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ToLatin returns a normalized Latin form of s, converting from Cyrillic
// first if necessary.
func ToLatin(s string) string {
	if HasCyrillic(s) {
		s = CyrillicToLatin(s)
	}
	return Normalize(s)
}

// Variants returns distinct spelling variants of a single token: its Latin
// form(s) always, plus a Cyrillic transliteration when cyrillic is true.
//
// A Cyrillic variant is only emitted when the transliteration maps cleanly
// (no leftover Latin letters), so English words like "white" never produce
// broken output such as "wҳите".
func Variants(s string, cyrillic bool) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	add := func(v string) {
		v = strings.TrimSpace(v)
		if v != "" && !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}

	if HasCyrillic(s) {
		lat := CyrillicToLatin(s)
		add(Normalize(lat))
		add(lat)
		if cyrillic {
			add(s)
		}
	} else {
		n := Normalize(s)
		add(n)
		add(s)
		if cyrillic {
			if cyr := LatinToCyrillic(n); cyr != "" && !hasASCIILetter(cyr) {
				add(cyr)
			}
		}
	}
	return out
}

// CommonPasswords returns passwords frequently seen among Uzbek users.
func CommonPasswords() []string {
	return []string{
		"parol", "parol123", "qwerty", "123456", "12345678",
		"admin", "998", "000000", "password", "iloveyou", "11111111",
	}
}

// KeyboardPatterns returns common keyboard-walk patterns.
func KeyboardPatterns() []string {
	return []string{"qwerty", "asdf", "zxcvbn", "1q2w3e", "qazwsx", "qwerty123"}
}

// Package generate implements the wordlist pipeline:
//
//	Profile -> tokens -> mutations -> combinations -> dedupe -> filter -> rank
//
// It never attacks anything; it only transforms user-supplied data into a
// list of candidate strings.
package generate

import (
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/Yescript/persona/internal/profile"
	"github.com/Yescript/persona/internal/uzbek"
)

// Options controls the generation pipeline.
type Options struct {
	Min           int  // minimum length (0 = no limit)
	Max           int  // maximum length (0 = no limit)
	MaxWords      int  // explosion guard on final output (0 = unlimited)
	Depth         int  // how many tokens may combine (1 = no word+word combos)
	Leet          bool // enable leetspeak mutations
	Symbols       bool // allow symbol suffixes
	Years         bool // append year tokens
	RequireDigit  bool // keep only candidates containing a digit
	RequireSymbol bool // keep only candidates containing a symbol
	Rank          bool // sort most-likely candidates to the top
	Cyrillic      bool // also emit Cyrillic transliterations of tokens
	Reverse       bool // include reversed tokens (e.g. ali -> ila)
	Truncate      bool // include truncated tokens (first 3-4 chars)
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		MaxWords: 100000,
		Depth:    2,
		Leet:     true,
		Symbols:  true,
		Years:    true,
		Rank:     true,
		Reverse:  true,
		Truncate: true,
	}
}

// Result is the output of Generate.
type Result struct {
	Words      []string
	TokenCount int
}

var (
	separators = []string{"", ".", "_", "-"}
	suffixes   = []string{"1", "12", "123", "1234", "01", "007", "!", "!@#", "."}
)

// Generate runs the full pipeline for a profile.
func Generate(p *profile.Profile, opts Options) *Result {
	words, numbers := extractTokens(p, opts)

	genCap := 0
	if opts.MaxWords > 0 {
		genCap = opts.MaxWords * 4
	}
	out := newSet(genCap)

	// 1. words and their mutations
	for _, w := range words {
		for _, m := range mutateWord(w, opts) {
			out.add(m)
		}
		if out.full() {
			break
		}
	}

	// raw numbers + auto-added common lists
	for _, n := range numbers {
		out.add(n)
	}
	for _, c := range uzbek.CommonPasswords() {
		out.add(c)
	}
	for _, k := range uzbek.KeyboardPatterns() {
		out.add(k)
	}

	years := buildYears(p, opts)

	// 2. combinations
	if opts.Depth >= 2 && !out.full() {
		combineWordWord(out, words)
	}
	if !out.full() {
		combineWordNumber(out, words, numbers, years, opts)
	}
	if !out.full() {
		combineWordSuffix(out, words, opts)
	}

	// 3. filter -> rank -> cap
	list := filter(out.slice(), opts)
	if opts.Rank {
		list = rank(list, p)
	}
	if opts.MaxWords > 0 && len(list) > opts.MaxWords {
		list = list[:opts.MaxWords]
	}
	return &Result{Words: list, TokenCount: len(words) + len(numbers)}
}

// --- token extraction -------------------------------------------------------

func extractTokens(p *profile.Profile, opts Options) (words []string, numbers []string) {
	wseen := map[string]bool{}
	nseen := map[string]bool{}

	// addWord splits a free-text field into clean word tokens (breaking on
	// spaces, commas, slashes, etc.) so comma-separated input never becomes a
	// single nonsensical token.
	addWord := func(s string) {
		for _, tok := range uzbek.Tokenize(s) {
			for _, v := range uzbek.Variants(tok, opts.Cyrillic) {
				v = strings.ToLower(strings.TrimSpace(v))
				if v != "" && !wseen[v] {
					wseen[v] = true
					words = append(words, v)
				}
			}
		}
	}
	addNum := func(s string) {
		d := onlyDigits(s)
		if d != "" && !nseen[d] {
			nseen[d] = true
			numbers = append(numbers, d)
		}
	}

	for _, s := range []string{
		p.Name, p.Surname, p.FatherName, p.SpouseName, p.PetName,
		p.City, p.School, p.Workplace, p.FootballClub, p.Motto, p.Brand, p.Color,
	} {
		addWord(s)
	}
	for _, s := range p.Nicknames {
		addWord(s)
	}
	for _, s := range p.Usernames {
		addWord(s)
	}
	for _, c := range p.Children {
		addWord(c.Name)
	}
	for _, s := range p.ReligiousWords {
		addWord(s)
	}

	addNum(p.BirthDay)
	addNum(p.BirthMonth)
	addNum(p.BirthYear)
	addNum(p.LuckyNumber)
	addNum(p.SpouseYear)
	addNum(p.ImportantDate)
	addNum(p.CarPlate)
	for _, c := range p.Children {
		addNum(c.Year)
	}
	for _, ph := range p.Phones {
		d := onlyDigits(ph)
		addNum(d)
		if len(d) >= 4 {
			addNum(d[len(d)-4:])
		}
	}
	if p.BirthDay != "" && p.BirthMonth != "" {
		addNum(pad2(onlyDigits(p.BirthDay)) + pad2(onlyDigits(p.BirthMonth)))
	}
	return words, numbers
}

// --- mutations --------------------------------------------------------------

func mutateWord(w string, opts Options) []string {
	seen := map[string]bool{}
	var res []string
	push := func(s string) {
		if s != "" && !seen[s] {
			seen[s] = true
			res = append(res, s)
		}
	}

	for _, c := range caseVariants(w) {
		push(c)
	}
	if opts.Reverse {
		push(reverse(w))
	}

	if opts.Truncate {
		r := []rune(w)
		if len(r) > 4 {
			push(string(r[:4]))
		}
		if len(r) > 3 {
			push(string(r[:3]))
		}
	}

	if opts.Leet {
		for _, l := range leetVariants(w) {
			push(l)
		}
		for _, l := range leetVariants(capitalize(w)) {
			push(l)
		}
	}
	return res
}

func caseVariants(w string) []string {
	lw := strings.ToLower(w)
	return []string{lw, capitalize(lw), strings.ToUpper(lw)}
}

func capitalize(w string) string {
	r := []rune(w)
	if len(r) == 0 {
		return w
	}
	return strings.ToUpper(string(r[0])) + strings.ToLower(string(r[1:]))
}

func reverse(w string) string {
	r := []rune(w)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

var (
	leetPrimary   = map[rune]string{'a': "@", 'o': "0", 'i': "1", 's': "5", 'e': "3", 'g': "9", 't': "7"}
	leetSecondary = map[rune]string{'a': "4", 'i': "!", 's': "$"}
)

func leetVariants(w string) []string {
	apply := func(m map[rune]string) string {
		var b strings.Builder
		changed := false
		for _, r := range w {
			if rep, ok := m[unicode.ToLower(r)]; ok {
				b.WriteString(rep)
				changed = true
			} else {
				b.WriteRune(r)
			}
		}
		if !changed {
			return ""
		}
		return b.String()
	}
	var res []string
	if v := apply(leetPrimary); v != "" {
		res = append(res, v)
	}
	merged := map[rune]string{}
	for k, v := range leetPrimary {
		merged[k] = v
	}
	for k, v := range leetSecondary {
		merged[k] = v
	}
	if v := apply(merged); v != "" {
		res = append(res, v)
	}
	return res
}

// --- combinations -----------------------------------------------------------

func combineWordWord(out *set, words []string) {
	for i, a := range words {
		for j, b := range words {
			if i == j {
				continue
			}
			for _, sep := range separators {
				out.add(a + sep + b)
				out.add(capitalize(a) + sep + capitalize(b))
			}
			if out.full() {
				return
			}
		}
	}
}

func combineWordNumber(out *set, words, numbers, years []string, opts Options) {
	nums := append([]string{}, numbers...)
	if opts.Years {
		nums = append(nums, years...)
	}
	for _, w := range words {
		for _, f := range []string{w, capitalize(w)} {
			for _, n := range nums {
				out.add(f + n)
			}
		}
		if out.full() {
			return
		}
	}
}

func combineWordSuffix(out *set, words []string, opts Options) {
	for _, w := range words {
		for _, f := range []string{w, capitalize(w), strings.ToUpper(w)} {
			for _, suf := range suffixes {
				if !opts.Symbols && hasSymbol(suf) {
					continue
				}
				out.add(f + suf)
			}
		}
		if out.full() {
			return
		}
	}
}

func buildYears(p *profile.Profile, opts Options) []string {
	if !opts.Years {
		return nil
	}
	seen := map[string]bool{}
	var years []string
	add := func(y string) {
		y = onlyDigits(y)
		if y != "" && !seen[y] {
			seen[y] = true
			years = append(years, y)
		}
	}
	add(p.BirthYear)
	add(p.SpouseYear)
	for _, c := range p.Children {
		add(c.Year)
	}
	cur := time.Now().Year()
	for y := cur - 1; y <= cur+1; y++ {
		add(strconv.Itoa(y))
	}
	// short two-digit forms of four-digit years
	for _, y := range append([]string{}, years...) {
		if len(y) == 4 {
			add(y[2:])
		}
	}
	return years
}

// --- filter / rank ----------------------------------------------------------

func filter(list []string, opts Options) []string {
	res := make([]string, 0, len(list))
	for _, w := range list {
		n := len([]rune(w))
		if opts.Min > 0 && n < opts.Min {
			continue
		}
		if opts.Max > 0 && n > opts.Max {
			continue
		}
		if opts.RequireDigit && !hasDigit(w) {
			continue
		}
		if opts.RequireSymbol && !hasSymbol(w) {
			continue
		}
		res = append(res, w)
	}
	return res
}

func rank(list []string, p *profile.Profile) []string {
	type scored struct {
		w string
		s int
		i int
	}
	arr := make([]scored, len(list))
	for i, w := range list {
		arr[i] = scored{w, score(w, p), i}
	}
	sort.SliceStable(arr, func(a, b int) bool {
		if arr[a].s != arr[b].s {
			return arr[a].s > arr[b].s
		}
		return arr[a].i < arr[b].i
	})
	res := make([]string, len(arr))
	for i := range arr {
		res[i] = arr[i].w
	}
	return res
}

func score(w string, p *profile.Profile) int {
	s := 0
	if hasDigit(w) && hasLetter(w) {
		s += 3
	}
	n := len([]rune(w))
	switch {
	case n >= 8 && n <= 12:
		s += 2
	case n >= 6 && n <= 14:
		s++
	}
	if p.BirthYear != "" && strings.Contains(w, onlyDigits(p.BirthYear)) {
		s += 3
	}
	for _, suf := range []string{"123", "1", "007"} {
		if strings.HasSuffix(w, suf) {
			s++
			break
		}
	}
	if hasSymbol(w) {
		s++
	}
	return s
}

// --- ordered dedupe set -----------------------------------------------------

type set struct {
	seen   map[string]struct{}
	list   []string
	genCap int
}

func newSet(genCap int) *set {
	return &set{seen: make(map[string]struct{}), genCap: genCap}
}

func (s *set) add(w string) {
	if w == "" {
		return
	}
	if _, ok := s.seen[w]; ok {
		return
	}
	s.seen[w] = struct{}{}
	s.list = append(s.list, w)
}

func (s *set) full() bool      { return s.genCap > 0 && len(s.list) >= s.genCap }
func (s *set) slice() []string { return s.list }

// --- small helpers ----------------------------------------------------------

func onlyDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func pad2(s string) string {
	if len(s) == 1 {
		return "0" + s
	}
	return s
}

func hasDigit(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func hasLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

func hasSymbol(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

// Command persona is a custom wordlist/username generator for authorized
// penetration testing, auditing your own accounts, and password-awareness
// education. It never attacks, connects to, or cracks anything — it only
// reads user-supplied data and writes .txt files.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Yescript/persona/internal/generate"
	"github.com/Yescript/persona/internal/output"
	"github.com/Yescript/persona/internal/profile"
	"github.com/Yescript/persona/internal/tui"
	"github.com/Yescript/persona/internal/username"
)

// version is overridden at build time via -ldflags "-X main.version=...".
var version = "0.1.0"

const bannerLine = "persona — authorized / educational use only. It generates a wordlist; it never attacks anything."

func main() {
	var (
		name, surname, birth string
		config, out, userOut string
		usernamesOnly        bool
		withUsernames        bool
		stdout, showVersion  bool
		minLen, maxLen       int
		maxWords, depth      int
		leet, symbols, years bool
		requireDigit         bool
		requireSymbol        bool
		cyrillic             bool
		reverse, truncate    bool
	)

	flag.StringVar(&name, "name", "", "first name")
	flag.StringVar(&surname, "surname", "", "surname")
	flag.StringVar(&birth, "birth", "", "birth year (e.g. 1999)")
	flag.StringVar(&config, "config", "", "load a profile from a YAML file (non-interactive)")
	flag.StringVar(&out, "output", "wordlist.txt", "output wordlist path")
	flag.StringVar(&out, "o", "wordlist.txt", "output wordlist path (shorthand)")
	flag.StringVar(&userOut, "usernames-output", "usernames.txt", "usernames output path")
	flag.BoolVar(&usernamesOnly, "usernames-only", false, "only generate usernames")
	flag.BoolVar(&withUsernames, "usernames", false, "also generate usernames.txt alongside the wordlist")
	flag.BoolVar(&stdout, "stdout", false, "also print results to stdout")
	flag.BoolVar(&showVersion, "version", false, "print version and exit")
	flag.IntVar(&minLen, "min", 0, "minimum length (0 = no limit)")
	flag.IntVar(&maxLen, "max", 0, "maximum length (0 = no limit)")
	flag.IntVar(&maxWords, "max-words", 100000, "cap total generated words")
	flag.IntVar(&depth, "depth", 2, "how many tokens may combine (1 or 2)")
	flag.BoolVar(&leet, "leet", true, "enable leetspeak mutations")
	flag.BoolVar(&symbols, "symbols", true, "allow symbol suffixes")
	flag.BoolVar(&years, "years", true, "append year tokens")
	flag.BoolVar(&requireDigit, "require-digit", false, "keep only candidates with a digit")
	flag.BoolVar(&requireSymbol, "require-symbol", false, "keep only candidates with a symbol")
	flag.BoolVar(&cyrillic, "cyrillic", false, "also emit Cyrillic transliterations (off by default)")
	flag.BoolVar(&reverse, "reverse", true, "include reversed tokens (use --reverse=false to disable)")
	flag.BoolVar(&truncate, "truncate", true, "include truncated tokens (use --truncate=false to disable)")
	flag.Parse()

	if showVersion {
		fmt.Printf("persona %s\n", version)
		return
	}

	genOpts := generate.DefaultOptions()
	genOpts.Min = minLen
	genOpts.Max = maxLen
	genOpts.MaxWords = maxWords
	genOpts.Depth = depth
	genOpts.Leet = leet
	genOpts.Symbols = symbols
	genOpts.Years = years
	genOpts.RequireDigit = requireDigit
	genOpts.RequireSymbol = requireSymbol
	genOpts.Cyrillic = cyrillic
	genOpts.Reverse = reverse
	genOpts.Truncate = truncate

	userOpts := username.Options{Leet: leet, Years: years}

	interactive := config == "" && name == "" && surname == "" && birth == "" && !usernamesOnly

	if interactive {
		runInteractive(tui.Config{
			Output:         out,
			UsernamesFile:  userOut,
			GenOpts:        genOpts,
			UserOpts:       userOpts,
			WriteUsernames: true,
			Stdout:         stdout,
		})
		return
	}

	// Non-interactive: build a profile from the config file or flags.
	fmt.Fprintln(os.Stderr, bannerLine)

	var p *profile.Profile
	if config != "" {
		loaded, err := profile.Load(config)
		if err != nil {
			fail(err)
		}
		p = loaded
	} else {
		p = &profile.Profile{Name: name, Surname: surname, BirthYear: birth}
	}

	if usernamesOnly {
		users := username.Generate(p, userOpts)
		st, err := output.Write(userOut, users)
		if err != nil {
			fail(err)
		}
		if stdout {
			printLines(users)
		}
		fmt.Fprintf(os.Stderr, "usernames: %d → %s (%s)\n", st.Count, userOut, output.HumanSize(st.Bytes))
		return
	}

	res := generate.Generate(p, genOpts)
	st, err := output.Write(out, res.Words)
	if err != nil {
		fail(err)
	}
	if stdout {
		printLines(res.Words)
	}
	fmt.Fprintf(os.Stderr, "tokens: %d · wordlist: %d → %s (%s)\n",
		res.TokenCount, st.Count, out, output.HumanSize(st.Bytes))

	if withUsernames {
		users := username.Generate(p, userOpts)
		ust, err := output.Write(userOut, users)
		if err != nil {
			fail(err)
		}
		fmt.Fprintf(os.Stderr, "usernames: %d → %s (%s)\n", ust.Count, userOut, output.HumanSize(ust.Bytes))
	}
}

func runInteractive(cfg tui.Config) {
	m, err := tui.Run(cfg)
	if err != nil {
		fail(err)
	}
	if cfg.Stdout {
		printLines(m.Words)
	}
}

func printLines(lines []string) {
	for _, l := range lines {
		fmt.Println(l)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "persona: error:", err)
	os.Exit(1)
}

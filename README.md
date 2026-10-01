# Persona

**Persona** is a custom wordlist generator — it turns a target's personal
details into a list of *likely* passwords and usernames. It is the same
category of tool as [CUPP](https://github.com/Mebus/cupp) and
[Username Anarchy](https://github.com/urbanadventurer/username-anarchy),
built for **authorized penetration testing, auditing your own accounts, and
password-awareness education**.

[![Go Report Card](https://goreportcard.com/badge/github.com/Yescrypt/Persona)](https://goreportcard.com/report/github.com/Yescrypt/Persona)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/Yescrypt/Persona)](https://golang.org)

It ships with an **Uzbek module** that understands local spelling habits
(`oʻ/gʻ`, Latin↔Cyrillic) and common local password patterns.

**Supported platforms:** Windows and Linux (Ubuntu / Debian), on `amd64`
and `arm64`. It is a single static binary with no runtime dependencies.

> ⚠️ **Ethical Use**
> Persona **never attacks, connects to, or cracks anything.** It only reads
> the data you give it and writes `.txt` files. Use it **only** against
> systems and accounts you own or are explicitly authorized to test.
> You are responsible for complying with all applicable laws. The authors
> accept no liability for misuse.

---

## Features

- Interactive TUI (Charm: Bubble Tea / Lip Gloss / Bubbles) with an ASCII
  banner, guided prompts, and a progress spinner.
- Pipeline: `Profile → tokens → mutations → combinations → dedupe → filter → rank`.
- Mutations: case (`ali/Ali/ALI`), leetspeak (`a→@/4, o→0, i→1/!, …`),
  reverse, truncation.
- Combinations: `token + separator + token`, `token + year`, `token + suffix`.
- Uzbek normalization + Latin↔Cyrillic transliteration.
- Username templates (`first.last`, `flast`, `f.last`, `first+year`, …).
- Policy filters: `--min`, `--max`, `--require-digit`, `--require-symbol`.
- Explosion guard: `--max-words`, `--depth`.
- Config-file and flag-driven non-interactive runs for CI.

## Install

> Replace `Yescript` with your own GitHub username/org if you fork this.

### 1. Linux (Ubuntu / Debian) — one-liner
No `apt` packages required; the script downloads a prebuilt binary and adds
it to your `PATH`.
```sh
curl -fsSL https://raw.githubusercontent.com/Yescript/persona/main/install.sh | bash
```

### 2. Windows — one-liner (PowerShell)
```powershell
irm https://raw.githubusercontent.com/Yescript/persona/main/install.ps1 | iex
```

### 3. Manual binary
Download the asset for your OS/arch from the
[Releases](https://github.com/Yescript/persona/releases) page.

Linux:
```sh
tar -xzf persona_linux_amd64.tar.gz
chmod +x persona
mv persona ~/.local/bin/
```
Windows: unzip `persona_windows_amd64.zip` and move `persona.exe` into a
folder on your `PATH`.

### 4. Go install (any OS with Go 1.22+)
```sh
go install github.com/Yescript/persona/cmd/persona@latest
```

### 5. From source
```sh
git clone https://github.com/Yescript/persona.git
cd persona
make install          # Linux/macOS
go build -o persona.exe ./cmd/persona   # Windows
```

## Usage

Interactive (default):
```sh
persona
```

With flags:
```sh
persona --name Ali --surname Valiyev --birth 1999 -o wordlist.txt
```

From a config file:
```sh
persona --config profile.yaml
```

Usernames only:
```sh
persona --usernames-only
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--name` | | First name |
| `--surname` | | Surname |
| `--birth` | | Birth year (e.g. `1999`) |
| `--config` | | Load a profile from a YAML file (non-interactive) |
| `-o`, `--output` | `wordlist.txt` | Output wordlist path |
| `--usernames` | `false` | Also write `usernames.txt` alongside the wordlist |
| `--usernames-only` | `false` | Only generate usernames |
| `--usernames-output` | `usernames.txt` | Usernames output path |
| `--min` | `0` | Minimum length (0 = no limit) |
| `--max` | `0` | Maximum length (0 = no limit) |
| `--require-digit` | `false` | Keep only candidates with a digit |
| `--require-symbol` | `false` | Keep only candidates with a symbol |
| `--max-words` | `100000` | Cap total generated words |
| `--depth` | `2` | How many tokens may combine (1 or 2) |
| `--leet` | `true` | Enable leetspeak mutations |
| `--symbols` | `true` | Allow symbol suffixes |
| `--years` | `true` | Append year tokens |
| `--cyrillic` | `false` | Also emit Cyrillic transliterations (Latin-only by default) |
| `--reverse` | `true` | Include reversed tokens (`--reverse=false` to disable) |
| `--truncate` | `true` | Include truncated tokens (`--truncate=false` to disable) |
| `--stdout` | `false` | Also print results to stdout |
| `--version` | | Print version and exit |

## The Uzbek module

Many Uzbek users build passwords from Latinized names plus numbers, phone
tails, favorite clubs, or religious words. Persona's `uzbek` package:

- Splits free-text fields into clean tokens (so `black, white, ko'k`
  becomes three tokens, never one comma-joined string).
- Normalizes `oʻ/o' → o`, `gʻ/g' → g`, and strips stray apostrophes.
- With `--cyrillic`, produces Latin→Cyrillic variants (e.g. `Akmal` →
  `Акмал`), keeping digraphs like `sh`, `ch`, `ng`. Transliterations that
  would be broken (e.g. English `white`) are dropped, not emitted. Output is
  Latin-only by default, since most Uzbek users type Latin passwords.
- Seeds common local passwords and keyboard patterns.

## Output

- `wordlist.txt` — candidate passwords (primary output).
- `usernames.txt` — candidate usernames (when enabled).
- Final stats: token count, generated count, file size.

## Development

```sh
make build   # build ./persona
make test    # run unit tests
make fmt     # gofmt
make vet     # go vet
```

> This repo targets Go 1.22+. Run `go mod tidy` once to fetch dependencies
> and generate `go.sum`.

## Contributing

Issues and PRs welcome — especially additional Uzbek tokens, transliteration
fixes, and username templates. Keep packages small and `gofmt`/`go vet` clean,
and add tests for new generation logic.

## License

MIT — see [LICENSE](LICENSE).

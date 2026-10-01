// Package tui implements the interactive Bubble Tea interface: an ASCII
// banner, a sequential question flow, and a spinner during generation.
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Yescript/persona/internal/generate"
	"github.com/Yescript/persona/internal/output"
	"github.com/Yescript/persona/internal/profile"
	"github.com/Yescript/persona/internal/username"
)

// Config carries everything the TUI needs to generate and write output.
type Config struct {
	Output         string
	UsernamesFile  string
	GenOpts        generate.Options
	UserOpts       username.Options
	WriteUsernames bool
	Stdout         bool
}

const banner = `
 ____  ____  ____  ____   __   __ _   __
(  _ \(  __)(  _ \/ ___) /  \ (  ( \ / _\
 ) __/ ) _)  )   /(___ \(  O )/    //    \
(__)  (____)(__\_)(____/ \__/ \_)__)\_/\_/ `

var (
	bannerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("213")).Bold(true)
	labelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)
	hintStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	okStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)
	errStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
)

type question struct {
	label string
	place string
	set   func(p *profile.Profile, val string)
}

func defaultQuestions() []question {
	return []question{
		{"Ism (first name)", "Ali", func(p *profile.Profile, v string) { p.Name = v }},
		{"Familiya (surname)", "Valiyev", func(p *profile.Profile, v string) { p.Surname = v }},
		{"Otasining ismi", "", func(p *profile.Profile, v string) { p.FatherName = v }},
		{"Taxalluslar (vergul bilan)", "boss,king", func(p *profile.Profile, v string) { p.Nicknames = splitList(v) }},
		{"Tug'ilgan kun", "14", func(p *profile.Profile, v string) { p.BirthDay = v }},
		{"Tug'ilgan oy", "03", func(p *profile.Profile, v string) { p.BirthMonth = v }},
		{"Tug'ilgan yil", "1999", func(p *profile.Profile, v string) { p.BirthYear = v }},
		{"Usernamelar (vergul bilan)", "ali_v,aliv", func(p *profile.Profile, v string) { p.Usernames = splitList(v) }},
		{"Omadli raqam", "7", func(p *profile.Profile, v string) { p.LuckyNumber = v }},
		{"Telefon raqam(lar) (vergul bilan)", "998901234567", func(p *profile.Profile, v string) { p.Phones = splitList(v) }},
		{"Turmush o'rtog'i ismi", "", func(p *profile.Profile, v string) { p.SpouseName = v }},
		{"Uy hayvoni ismi", "", func(p *profile.Profile, v string) { p.PetName = v }},
		{"Shahar / mahalla", "Toshkent", func(p *profile.Profile, v string) { p.City = v }},
		{"Maktab / universitet", "IT House", func(p *profile.Profile, v string) { p.School = v }},
		{"Ish joyi", "", func(p *profile.Profile, v string) { p.Workplace = v }},
		{"Sevimli futbol klubi", "Paxtakor", func(p *profile.Profile, v string) { p.FootballClub = v }},
		{"Sevimli so'z / shior", "", func(p *profile.Profile, v string) { p.Motto = v }},
		{"Sevimli brend / mashina", "", func(p *profile.Profile, v string) { p.Brand = v }},
		{"Sevimli rang", "", func(p *profile.Profile, v string) { p.Color = v }},
	}
}

type phase int

const (
	phaseAsk phase = iota
	phaseGen
	phaseDone
)

type genDoneMsg struct {
	words  []string
	users  []string
	wStats output.Stats
	uStats output.Stats
	err    error
}

// Model is the Bubble Tea model.
type Model struct {
	cfg       Config
	questions []question
	idx       int
	input     textinput.Model
	spinner   spinner.Model
	prof      profile.Profile
	phase     phase
	err       error

	// Results, readable after the program exits.
	Words  []string
	Users  []string
	WStats output.Stats
	UStats output.Stats
}

// New builds a fresh model.
func New(cfg Config) Model {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 44

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = labelStyle

	qs := defaultQuestions()
	ti.Placeholder = qs[0].place
	return Model{cfg: cfg, questions: qs, input: ti, spinner: sp, phase: phaseAsk}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return textinput.Blink }

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			switch m.phase {
			case phaseAsk:
				m.questions[m.idx].set(&m.prof, strings.TrimSpace(m.input.Value()))
				m.idx++
				m.input.SetValue("")
				if m.idx >= len(m.questions) {
					m.phase = phaseGen
					return m, tea.Batch(m.spinner.Tick, m.runGen())
				}
				m.input.Placeholder = m.questions[m.idx].place
				return m, nil
			case phaseDone:
				return m, tea.Quit
			}
		}
	case genDoneMsg:
		m.phase = phaseDone
		m.err = msg.err
		m.Words = msg.words
		m.Users = msg.users
		m.WStats = msg.wStats
		m.UStats = msg.uStats
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	if m.phase == phaseAsk {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) runGen() tea.Cmd {
	cfg := m.cfg
	prof := m.prof
	return func() tea.Msg {
		res := generate.Generate(&prof, cfg.GenOpts)
		msg := genDoneMsg{words: res.Words}
		ws, err := output.Write(cfg.Output, res.Words)
		if err != nil {
			msg.err = err
			return msg
		}
		msg.wStats = ws
		if cfg.WriteUsernames {
			users := username.Generate(&prof, cfg.UserOpts)
			msg.users = users
			us, err := output.Write(cfg.UsernamesFile, users)
			if err != nil {
				msg.err = err
				return msg
			}
			msg.uStats = us
		}
		return msg
	}
}

// View implements tea.Model.
func (m Model) View() string {
	var b strings.Builder
	b.WriteString(bannerStyle.Render(banner) + "\n")
	b.WriteString(hintStyle.Render("  Authorized / educational use only — Persona hech narsani buzmaydi.") + "\n\n")

	switch m.phase {
	case phaseAsk:
		q := m.questions[m.idx]
		b.WriteString(fmt.Sprintf("  %s  %s\n", labelStyle.Render(fmt.Sprintf("[%d/%d]", m.idx+1, len(m.questions))), q.label))
		b.WriteString("  " + m.input.View() + "\n\n")
		b.WriteString(hintStyle.Render("  Enter = keyingi  ·  bo'sh = o'tkazib yuborish  ·  Esc = chiqish"))
	case phaseGen:
		b.WriteString(fmt.Sprintf("  %s Generatsiya qilinmoqda...", m.spinner.View()))
	case phaseDone:
		if m.err != nil {
			b.WriteString("  " + errStyle.Render("Xatolik: "+m.err.Error()))
			break
		}
		b.WriteString("  " + okStyle.Render("✓ Tayyor!") + "\n\n")
		b.WriteString(fmt.Sprintf("  %s  %d parol → %s (%s)\n",
			labelStyle.Render("wordlist "), m.WStats.Count, m.cfg.Output, output.HumanSize(m.WStats.Bytes)))
		if m.cfg.WriteUsernames {
			b.WriteString(fmt.Sprintf("  %s %d username → %s (%s)\n",
				labelStyle.Render("usernames"), m.UStats.Count, m.cfg.UsernamesFile, output.HumanSize(m.UStats.Bytes)))
		}
		b.WriteString("\n" + hintStyle.Render("  Enter / Esc = chiqish"))
	}
	return b.String() + "\n"
}

// Run starts the interactive program and returns the final model.
func Run(cfg Config) (Model, error) {
	p := tea.NewProgram(New(cfg))
	fm, err := p.Run()
	if err != nil {
		return Model{}, err
	}
	m, _ := fm.(Model)
	return m, nil
}

func splitList(v string) []string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(v, ",") {
		if s := strings.TrimSpace(part); s != "" {
			out = append(out, s)
		}
	}
	return out
}

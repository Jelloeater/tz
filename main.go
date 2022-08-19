/**
 * This file is part of tz.
 *
 * tz is free software: you can redistribute it and/or modify it under
 * the terms of the GNU General Public License as published by the Free
 * Software Foundation, either version 3 of the License, or (at your
 * option) any later version.
 *
 * tz is distributed in the hope that it will be useful, but WITHOUT
 * ANY WARRANTY; without even the implied warranty of MERCHANTABILITY
 * or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public
 * License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with tz.  If not, see <https://www.gnu.org/licenses/>.
 **/
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
	"github.com/tkuchiki/go-timezone"
)

// CurrentVersion represents the current build version.
const CurrentVersion = "0.6.1"

var (
	term              = termenv.ColorProfile()
	hasDarkBackground = termenv.HasDarkBackground()
)

type tickMsg time.Time

// Send a tickMsg every minute, on the minute.
func tick() tea.Cmd {
	return tea.Every(time.Minute, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type model struct {
	zones       []*Zone
	now         time.Time
	hour        int
	showDates   bool
	interactive bool
}

func (m model) Init() tea.Cmd {
	// If -q flag is passed, send quit message after first render.
	if !m.interactive {
		return tea.Quit
	}

	// Fire initial tick command to begin receiving ticks on the minute.
	return tick()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		case "left", "h":
			if m.hour == 0 {
				m.hour = 23
			} else {
				m.hour--
			}

		case "right", "l":
			if m.hour > 22 {
				m.hour = 0
			} else {
				m.hour++
			}

		case "d":
			m.showDates = !m.showDates
		}

	case tickMsg:
		m.now = time.Time(msg)
		return m, tick()
	}
	return m, nil
}

var clock func() time.Time = time.Now

func main() {
	exitQuick := flag.Bool("q", false, "exit immediately")
	showVersion := flag.Bool("v", false, "show version")
	when := flag.Int64("when", 0, "time in seconds since unix epoch")
	search := flag.String("list", "", "find time zones by name")
	flag.Parse()

	if *showVersion == true {
		fmt.Printf("tz %s\n", CurrentVersion)
		os.Exit(0)
	}

	if *search != "" {
		findZones(strings.ToLower(*search))
		os.Exit(0)
	}

	if *when != 0 {
		t := time.Unix(*when, 0)
		clock = func() time.Time {
			return t
		}
	}

	now := clock()
	config, err := LoadConfig(flag.Args())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %s\n", err)
		os.Exit(2)
	}
	var initialModel = model{
		zones:     config.Zones,
		now:       now,
		hour:      now.Hour(),
		showDates: false,
	}

	initialModel.interactive = !*exitQuick

	p := tea.NewProgram(initialModel)
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func findZones(filter string) {
	t := timezone.New()
	matches := map[string]*timezone.TzInfo{}

	for _, zones := range t.Timezones() {
		for _, name := range zones {
			if !strings.Contains(strings.ToLower(name), filter) {
				continue
			}

			ti, err := t.GetTzInfo(name)
			if err != nil {
				panic(err)
			}
			matches[name] = ti
		}
	}
	for name, ti := range matches {
		fmt.Printf("%5s (%s) :: %s\n",
			ti.ShortStandard(),
			ti.StandardOffsetHHMM(),
			name)
	}
}

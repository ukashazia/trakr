package main

import (
	"flag"
	"fmt"
	"os"
	"time"
	"trakr/trackers"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var refreshInterval *int

type model struct {
	tracker trackers.Tracker

	lastUpdated time.Time
	tickerTime  time.Time
}

var (
	quitStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
	timestampStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Italic(true)
)

func main() {
	model := &model{}
	trackingNumber := flag.String("trackingNumber", "", "Tracking number to query")
	serviceSlug := flag.String("service", "", "Service name for the tracking service")
	refreshInterval := flag.Int("refreshInterval", 0, "Refresh interval in seconds")
	flag.Parse()

	if serviceSlug != nil && *serviceSlug == "" {
		fmt.Println("Service name is required")
		return
	}

	if !serviceNameValid(*serviceSlug) {
		fmt.Printf("Invalid service slug. Valid options are:\n")
		for slug := range trackers.ServiceSlugs {
			fmt.Printf("- %s\n", slug)
		}
		return
	}

	if trackingNumber == nil || *trackingNumber == "" {
		fmt.Println("Please provide a tracking number using the -trackingNumber flag.")
		return
	}

	if refreshInterval == nil {
		refreshInterval = new(int)
		*refreshInterval = 0
	}

	tracker, err := trackers.NewTracker(*trackingNumber, *serviceSlug, time.Duration(*refreshInterval))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
		return
	}

	model.tracker = tracker

	model.lastUpdated = time.Now()
	model.tickerTime = time.Now()

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		return
	}
}

func serviceNameValid(serviceSlug string) bool {
	if _, ok := trackers.ServiceSlugs[serviceSlug]; ok {
		return true
	}
	return false
}

func (m model) Init() tea.Cmd {
	ri := *m.tracker.GetRefreshInterval()
	if ri == 0 {
		return getTrackingInfoCmd(&m)
	} else if ri > 0 {
		return tea.Batch(tickCmd(), fetchTimerCmd(&m), getTrackingInfoCmd(&m))
	}

	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		m.tickerTime = time.Time(msg)
		return m, tickCmd()

	case fetchTimerMsg:
		return m, tea.Batch(fetchTimerCmd(&m), getTrackingInfoCmd(&m))

	case trackers.TrackingInfoMsg:
		m.lastUpdated = time.Now()
		m.tracker.SetTrackingData(trackers.TrackingInfoMsg(msg))
		m.tracker.SetTrackingError(nil)

		if *m.tracker.GetRefreshInterval() == 0 {
			return m, tea.Quit
		}

		return m, nil
	case trackers.TrackingInfoErrorMsg:
		m.tracker.SetTrackingError(&msg)

		return m, tea.Quit

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() string {
	var dataStr, timeStr, quitStr, headingStr string

	headingStr = fmt.Sprintf("Tracking Number: %s\nService: %s\n\n", *m.tracker.GetTrackingNumber(), *m.tracker.GetServiceSlug())

	if m.tracker.GetTrackingError() != nil {
		return fmt.Sprintf("%s\n%s\n", headingStr, *m.tracker.GetTrackingError())
	}

	if *m.tracker.GetRefreshInterval() > 0 {
		timeStr = fmt.Sprintf("Last updated %d seconds ago.\n\n", int(m.tickerTime.Sub(m.lastUpdated).Seconds()))
		quitStr = fmt.Sprint(quitStyle.Render("\nPress q to quit.\n"))
	}

	data := *m.tracker.GetTrackingData()

	for _, track := range data {
		track := track.(map[string]any)
		dataStr += fmt.Sprintf("%s -- %s\n", timestampStyle.Render(track["time"].(string)), track["msgEng"])
	}

	return headingStr + timeStr + dataStr + quitStr
}

type (
	fetchTimerMsg      time.Time
	getTrackingInfoMsg []any

	tickMsg time.Time
)

func tickCmd() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(1 * time.Second)
		return tickMsg(time.Now())
	}
}

func fetchTimerCmd(m *model) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(*m.tracker.GetRefreshInterval() * time.Second)
		return fetchTimerMsg(time.Now())
	}
}

func getTrackingInfoCmd(m *model) tea.Cmd {
	return func() tea.Msg {
		data, err := m.tracker.FetchTrackingInfo()
		if err != nil {
			return err
		}

		return data
	}
}

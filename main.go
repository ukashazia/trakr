package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

var serviceSlugs = []string{
	"speedaf",
}

const refreshInterval = 5 // seconds
const uri = "https://speedaf.com/publicservice/v1/api/express/track/listExpressTrack"

type model struct {
	trackingNumber  string
	serviceSlug     string
	refreshInterval time.Duration
	trackingData    getTrackingInfoMsg
	lastUpdated     time.Time
	tickerTime      time.Time
}

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
		fmt.Printf("Invalid service slug. Valid options are: %v\n", serviceSlugs)
		return
	}

	if trackingNumber == nil || *trackingNumber == "" {
		fmt.Println("Please provide a tracking number using the -trackingNumber flag.")
		return
	}

	model.trackingNumber = *trackingNumber
	model.serviceSlug = *serviceSlug
	model.refreshInterval = time.Duration(*refreshInterval)
	model.lastUpdated = time.Now()
	model.tickerTime = time.Now()

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		return
	}
}

func serviceNameValid(serviceSlug string) bool {
	for _, slug := range serviceSlugs {
		if slug == serviceSlug {
			return true
		}
	}
	return false
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), fetchTimerCmd(&m), getTrackingInfoCmd(&m))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		m.tickerTime = time.Time(msg)
		return m, tickCmd()

	case fetchTimerMsg:
		return m, tea.Batch(fetchTimerCmd(&m), getTrackingInfoCmd(&m))

	case getTrackingInfoMsg:
		m.lastUpdated = time.Now()
		m.trackingData = getTrackingInfoMsg(msg)

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}

	}

	return m, nil
}

func (m model) View() string {
	timeStr := fmt.Sprintf("Last updated %d seconds ago.\n", int(m.tickerTime.Sub(m.lastUpdated).Seconds()))

	var dataStr string

	if m.trackingData == nil {
		dataStr = "Fetching tracking data...\n"
	} else {
		for _, track := range m.trackingData {
			track := track.(map[string]any)
			dataStr += fmt.Sprintf("%s -- %s\n", track["time"], track["msgEng"])
		}
	}

	return timeStr + dataStr + "\nPress q to quit.\n"
}

type (
	fetchTimerMsg           time.Time
	getTrackingInfoMsg      []any
	getTrackingInfoErrorMsg error

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
		time.Sleep(m.refreshInterval * time.Second)
		return fetchTimerMsg(time.Now())
	}
}

func getTrackingInfoCmd(m *model) tea.Cmd {
	return func() tea.Msg {
		payload := map[string]any{
			"mailNoList": []string{m.trackingNumber},
		}

		jsonData, err := json.Marshal(payload)
		if err != nil {
			return getTrackingInfoErrorMsg(err)
		}

		req, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsonData))
		if err != nil {
			return getTrackingInfoErrorMsg(err)
		}

		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}

		resp, err := client.Do(req)
		if err != nil {
			return getTrackingInfoErrorMsg(err)
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return getTrackingInfoErrorMsg(err)
		}

		var response map[string]any
		err = json.Unmarshal(respBody, &response)
		if err != nil {
			return getTrackingInfoErrorMsg(err)
		}

		tracks := response["data"].([]any)[0].(map[string]any)["tracks"].([]any)

		return getTrackingInfoMsg(tracks)
	}
}

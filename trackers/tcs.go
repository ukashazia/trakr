package trackers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Tcs struct {
	GenericTracker
}

func NewTcs(trackingNumber string, refreshInterval time.Duration) *Tcs {
	return &Tcs{
		GenericTracker{
			uri:             "https://www.tcsexpress.com/apibridge",
			trackingNumber:  trackingNumber,
			serviceSlug:     "tcs",
			refreshInterval: refreshInterval,
		},
	}
}

func (m Tcs) FetchTrackingInfo() (TrackingInfoMsg, TrackingInfoErrorMsg) {

	payload := map[string]any{
		"body": map[string]any{
			"consignee": []string{m.trackingNumber},
			"url":       "trackapinew",
			"type":      "GET",
			"payload":   map[string]any{},
			"param":     fmt.Sprintf("consignee=%s", m.trackingNumber),
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, TrackingInfoErrorMsg(err)
	}

	req, err := http.NewRequest("POST", m.uri, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, TrackingInfoErrorMsg(err)
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, TrackingInfoErrorMsg(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, TrackingInfoErrorMsg(err)
	}

	var response map[string]any
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return nil, TrackingInfoErrorMsg(err)
	}

	checkpoints := response["responseData"].(map[string]any)["checkpoints"]
	if checkpoints == nil {
		return nil, TrackingInfoErrorMsg(fmt.Errorf("no data found for tracking number %s", m.trackingNumber))
	}

	tracks := checkpoints.([]any)
	data := make([]any, len(tracks))
	for i, track := range tracks {
		track := track.(map[string]any)

		data[i] = map[string]any{
			"time":   track["datetime"].(string),
			"msgEng": track["status"].(string),
		}
	}

	return TrackingInfoMsg(data), nil
}

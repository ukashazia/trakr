package trackers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type Speedaf struct {
	GenericTracker
}

func NewSpeedaf(trackingNumber string, refreshInterval time.Duration) *Speedaf {
	return &Speedaf{
		GenericTracker{
			uri:             "https://speedaf.com/publicservice/v1/api/express/track/listExpressTrack",
			trackingNumber:  trackingNumber,
			serviceSlug:     "speedaf",
			refreshInterval: refreshInterval,
		},
	}
}

func (m Speedaf) FetchTrackingInfo() (TrackingInfoMsg, TrackingInfoErrorMsg) {

	payload := map[string]any{
		"mailNoList": []string{m.trackingNumber},
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

	tracks := response["data"].([]any)[0].(map[string]any)["tracks"].([]any)

	return TrackingInfoMsg(tracks), nil
}

package trackers

import (
	"fmt"
	"time"
)

var ServiceSlugs = map[string]any{
	"speedaf": nil,
	"tcs":     nil,
}

type (
	TrackingInfoMsg      []any
	TrackingInfoErrorMsg error
)

type GenericTracker struct {
	uri             string
	trackingNumber  string
	serviceSlug     string
	refreshInterval time.Duration
	trackingData    TrackingInfoMsg
	trackingError   *TrackingInfoErrorMsg
}

type Tracker interface {
	FetchTrackingInfo() (TrackingInfoMsg, TrackingInfoErrorMsg)

	GetURI() *string
	GetTrackingNumber() *string
	GetServiceSlug() *string
	GetRefreshInterval() *time.Duration
	GetTrackingData() *TrackingInfoMsg
	GetTrackingError() *TrackingInfoErrorMsg

	SetTrackingData(data TrackingInfoMsg)
	SetTrackingError(err *TrackingInfoErrorMsg)
}

func NewTracker(trackingNumber, serviceSlug string, refreshInterval time.Duration) (Tracker, error) {

	switch serviceSlug {
	case "speedaf":
		return NewSpeedaf(trackingNumber, refreshInterval), nil
	case "tcs":
		return NewTcs(trackingNumber, refreshInterval), nil
	}

	return nil, fmt.Errorf("unsupported service slug: %s", serviceSlug)
}

func (s GenericTracker) FetchTrackingInfo() (TrackingInfoMsg, TrackingInfoErrorMsg) {
	return nil, nil
}

func (s GenericTracker) GetURI() *string {
	return &s.uri
}

func (s GenericTracker) GetTrackingNumber() *string {
	return &s.trackingNumber
}

func (s GenericTracker) GetServiceSlug() *string {
	return &s.serviceSlug
}

func (s GenericTracker) GetRefreshInterval() *time.Duration {
	return &s.refreshInterval
}

func (s GenericTracker) GetTrackingData() *TrackingInfoMsg {
	return &s.trackingData
}
func (s GenericTracker) GetTrackingError() *TrackingInfoErrorMsg {
	return s.trackingError
}

func (s *GenericTracker) SetTrackingData(data TrackingInfoMsg) {
	s.trackingData = data
}

func (s *GenericTracker) SetTrackingError(err *TrackingInfoErrorMsg) {
	s.trackingError = err
}

package handlers

import (
	"chrisgross-ctrl-project/internal/services"
)

type ActivityHubAdapter struct {
	hub *ActivityHub
}

func NewActivityHubAdapter(hub *ActivityHub) *ActivityHubAdapter {
	return &ActivityHubAdapter{hub: hub}
}

func (a *ActivityHubAdapter) BroadcastEvent(event services.ActivityEventData) {
	if a.hub == nil {
		return
	}
	
	hubEvent := ActivityEvent{
		Type:       event.Type,
		SessionID:  event.SessionID,
		UserID:     event.UserID,
		UserEmail:  event.UserEmail,
		PropertyID: event.PropertyID,
		Details:    event.Details,
		Timestamp:  event.Timestamp,
		Score:      event.Score,
		EventData:  event.EventData,
	}
	
	a.hub.BroadcastEvent(hubEvent)
}

func (a *ActivityHubAdapter) BroadcastActiveCount(count int) {
	if a.hub == nil {
		return
	}
	a.hub.BroadcastActiveCount(count)
}

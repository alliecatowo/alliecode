package tasks

import "time"

func newStatusEvent(index uint64, status Status, message string, at time.Time) Event {
	return Event{Index: index, Type: "status", Status: status, Message: message, CreatedAt: at}
}

func newCreatedEvent(index uint64, at time.Time) Event {
	return Event{Index: index, Type: "created", Status: StatusRunning, CreatedAt: at}
}

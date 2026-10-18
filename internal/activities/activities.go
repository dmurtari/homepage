package activities

import "time"

type Activity struct {
	Id           string       `json:"id"`
	ActivityType string       `json:"activityType"`
	StartTime    time.Time    `json:"startTime"`
	EndTime      time.Time    `json:"endTime"`
	Duration     float64      `json:"duration"`
	Route        []RoutePoint `json:"route"`
}

type RoutePoint struct {
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Elevation float64   `json:"elevation"`
	Timestamp time.Time `json:"timestamp"`
}

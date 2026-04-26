package domain

import "time"

type StatusPageAnnouncementKind string

const (
	StatusPageAnnouncementInfo        StatusPageAnnouncementKind = "info"
	StatusPageAnnouncementMaintenance StatusPageAnnouncementKind = "maintenance"
	StatusPageAnnouncementIncident    StatusPageAnnouncementKind = "incident"
	StatusPageAnnouncementResolved    StatusPageAnnouncementKind = "resolved"
)

type StatusPage struct {
	ID            string                   `json:"id"`
	Slug          string                   `json:"slug"`
	Title         string                   `json:"title"`
	Description   string                   `json:"description"`
	Enabled       bool                     `json:"enabled"`
	Services      []StatusPageService      `json:"services,omitempty"`
	Announcements []StatusPageAnnouncement `json:"announcements,omitempty"`
	CreatedAt     time.Time                `json:"createdAt"`
	UpdatedAt     time.Time                `json:"updatedAt"`
}

type StatusPageListItem struct {
	ID                string       `json:"id"`
	Slug              string       `json:"slug"`
	Title             string       `json:"title"`
	Description       string       `json:"description"`
	Enabled           bool         `json:"enabled"`
	ServiceCount      int          `json:"serviceCount"`
	AnnouncementCount int          `json:"announcementCount"`
	OverallStatus     HealthStatus `json:"overallStatus"`
	CreatedAt         time.Time    `json:"createdAt"`
	UpdatedAt         time.Time    `json:"updatedAt"`
}

type StatusPageService struct {
	StatusPageID  string       `json:"statusPageId"`
	ServiceID     string       `json:"serviceId"`
	SortOrder     int          `json:"sortOrder"`
	DisplayName   string       `json:"displayName"`
	ServiceName   string       `json:"serviceName"`
	Status        HealthStatus `json:"status"`
	LastCheckedAt time.Time    `json:"lastCheckedAt"`
	LatestCheck   *CheckResult `json:"latestCheck,omitempty"`
}

type StatusPageAnnouncement struct {
	ID           string                     `json:"id"`
	StatusPageID string                     `json:"statusPageId"`
	Kind         StatusPageAnnouncementKind `json:"kind"`
	Title        string                     `json:"title"`
	Message      string                     `json:"message"`
	StartsAt     time.Time                  `json:"startsAt,omitempty"`
	EndsAt       time.Time                  `json:"endsAt,omitempty"`
	CreatedAt    time.Time                  `json:"createdAt"`
	UpdatedAt    time.Time                  `json:"updatedAt"`
}

type PublicStatusPage struct {
	Slug          string                         `json:"slug"`
	Title         string                         `json:"title"`
	Description   string                         `json:"description"`
	OverallStatus HealthStatus                   `json:"overallStatus"`
	LastUpdatedAt time.Time                      `json:"lastUpdatedAt"`
	Services      []PublicStatusPageService      `json:"services"`
	Announcements []PublicStatusPageAnnouncement `json:"announcements"`
}

type PublicStatusPageService struct {
	Name          string              `json:"name"`
	Status        HealthStatus        `json:"status"`
	LastCheckedAt time.Time           `json:"lastCheckedAt"`
	LatestCheck   *PublicCheckSummary `json:"latestCheck,omitempty"`
}

type PublicStatusPageAnnouncement struct {
	Kind     StatusPageAnnouncementKind `json:"kind"`
	Title    string                     `json:"title"`
	Message  string                     `json:"message"`
	StartsAt time.Time                  `json:"startsAt,omitempty"`
	EndsAt   time.Time                  `json:"endsAt,omitempty"`
}

type PublicCheckSummary struct {
	Status         HealthStatus `json:"status"`
	LatencyMS      int64        `json:"latencyMs"`
	HTTPStatusCode int          `json:"httpStatusCode,omitempty"`
	CheckedAt      time.Time    `json:"checkedAt"`
	Message        string       `json:"message,omitempty"`
}

type StatusPageInput struct {
	ID          string `json:"id,omitempty"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Enabled     *bool  `json:"enabled,omitempty"`
}

type StatusPageServiceInput struct {
	ServiceID   string `json:"serviceId"`
	SortOrder   *int   `json:"sortOrder,omitempty"`
	DisplayName string `json:"displayName"`
}

type StatusPageAnnouncementInput struct {
	Kind     StatusPageAnnouncementKind `json:"kind"`
	Title    string                     `json:"title"`
	Message  string                     `json:"message"`
	StartsAt time.Time                  `json:"startsAt,omitempty"`
	EndsAt   time.Time                  `json:"endsAt,omitempty"`
}

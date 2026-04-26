package domain

import "time"

type NotificationChannelType string

const (
	NotificationChannelWebhook NotificationChannelType = "webhook"
	NotificationChannelNtfy    NotificationChannelType = "ntfy"
)

type NotificationDeliveryStatus string

const (
	NotificationDeliveryPending NotificationDeliveryStatus = "pending"
	NotificationDeliverySent    NotificationDeliveryStatus = "sent"
	NotificationDeliveryFailed  NotificationDeliveryStatus = "failed"
)

type NotificationEventType string

const (
	NotificationEventServiceHealthChanged     NotificationEventType = "service_health_changed"
	NotificationEventCheckFailed              NotificationEventType = "check_failed"
	NotificationEventCheckRecovered           NotificationEventType = "check_recovered"
	NotificationEventDiscoveredServiceCreated NotificationEventType = "discovered_service_created"
	NotificationEventDeviceCreated            NotificationEventType = "device_created"
	NotificationEventWorkerFailed             NotificationEventType = "worker_failed"
)

const RedactedSecret = "********"

type NotificationChannel struct {
	ID        string                  `json:"id"`
	Name      string                  `json:"name"`
	Type      NotificationChannelType `json:"type"`
	Enabled   bool                    `json:"enabled"`
	Config    map[string]any          `json:"config"`
	CreatedAt time.Time               `json:"createdAt"`
	UpdatedAt time.Time               `json:"updatedAt"`
}

type NotificationRule struct {
	ID         string                `json:"id"`
	Name       string                `json:"name"`
	EventType  NotificationEventType `json:"eventType"`
	Enabled    bool                  `json:"enabled"`
	Filters    map[string]any        `json:"filters"`
	ChannelIDs []string              `json:"channelIds"`
	Channels   []NotificationChannel `json:"channels,omitempty"`
	CreatedAt  time.Time             `json:"createdAt"`
	UpdatedAt  time.Time             `json:"updatedAt"`
}

type NotificationDelivery struct {
	ID          string                     `json:"id"`
	RuleID      string                     `json:"ruleId,omitempty"`
	RuleName    string                     `json:"ruleName,omitempty"`
	ChannelID   string                     `json:"channelId,omitempty"`
	ChannelName string                     `json:"channelName,omitempty"`
	EventType   NotificationEventType      `json:"eventType"`
	Status      NotificationDeliveryStatus `json:"status"`
	Message     string                     `json:"message"`
	AttemptedAt time.Time                  `json:"attemptedAt"`
}

type CheckResultOutcome struct {
	Result                   CheckResult  `json:"result"`
	Check                    ServiceCheck `json:"check"`
	Service                  Service      `json:"service,omitempty"`
	PreviousServiceStatus    HealthStatus `json:"previousServiceStatus,omitempty"`
	CurrentServiceStatus     HealthStatus `json:"currentServiceStatus,omitempty"`
	PreviousCheckStatus      HealthStatus `json:"previousCheckStatus,omitempty"`
	CurrentCheckStatus       HealthStatus `json:"currentCheckStatus,omitempty"`
	ServiceStatusChanged     bool         `json:"serviceStatusChanged"`
	CheckFailedTransition    bool         `json:"checkFailedTransition"`
	CheckRecoveredTransition bool         `json:"checkRecoveredTransition"`
}

type DeviceObservationOutcome struct {
	Device  Device `json:"device"`
	Created bool   `json:"created"`
}

type DiscoveredServiceObservationOutcome struct {
	DiscoveredService DiscoveredService `json:"discoveredService"`
	Created           bool              `json:"created"`
}

type JobRunOutcome struct {
	JobName             string    `json:"jobName"`
	Failed              bool      `json:"failed"`
	Recovered           bool      `json:"recovered"`
	ConsecutiveFailures int       `json:"consecutiveFailures"`
	LastError           string    `json:"lastError,omitempty"`
	AttemptedAt         time.Time `json:"attemptedAt"`
}

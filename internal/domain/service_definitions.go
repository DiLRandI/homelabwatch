package domain

import "time"

type ServiceDefinition struct {
	ID             string                           `json:"id"`
	Key            string                           `json:"key"`
	Name           string                           `json:"name"`
	Icon           string                           `json:"icon,omitempty"`
	Priority       int                              `json:"priority"`
	BuiltIn        bool                             `json:"builtIn"`
	Enabled        bool                             `json:"enabled"`
	Matchers       []ServiceDefinitionMatcher       `json:"matchers,omitempty"`
	CheckTemplates []ServiceDefinitionCheckTemplate `json:"checkTemplates,omitempty"`
	CreatedAt      time.Time                        `json:"createdAt"`
	UpdatedAt      time.Time                        `json:"updatedAt"`
}

type ServiceDefinitionMatcher struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Operator  string    `json:"operator,omitempty"`
	Value     string    `json:"value"`
	Extra     string    `json:"extra,omitempty"`
	Weight    int       `json:"weight"`
	SortOrder int       `json:"sortOrder"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ServiceDefinitionCheckTemplate struct {
	ID                string                  `json:"id"`
	Name              string                  `json:"name"`
	Type              CheckType               `json:"type"`
	Protocol          string                  `json:"protocol,omitempty"`
	AddressSource     ServiceAddressSource    `json:"addressSource,omitempty"`
	HostValue         string                  `json:"hostValue,omitempty"`
	Port              int                     `json:"port,omitempty"`
	Path              string                  `json:"path,omitempty"`
	Method            string                  `json:"method,omitempty"`
	IntervalSeconds   int                     `json:"intervalSeconds"`
	TimeoutSeconds    int                     `json:"timeoutSeconds"`
	ExpectedStatusMin int                     `json:"expectedStatusMin,omitempty"`
	ExpectedStatusMax int                     `json:"expectedStatusMax,omitempty"`
	Enabled           bool                    `json:"enabled"`
	SortOrder         int                     `json:"sortOrder"`
	ConfigSource      HealthCheckConfigSource `json:"configSource,omitempty"`
}

type ServiceDefinitionInput struct {
	ID             string                           `json:"id,omitempty"`
	Key            string                           `json:"key,omitempty"`
	Name           string                           `json:"name"`
	Icon           string                           `json:"icon,omitempty"`
	Priority       int                              `json:"priority"`
	Enabled        bool                             `json:"enabled"`
	Matchers       []ServiceDefinitionMatcher       `json:"matchers"`
	CheckTemplates []ServiceDefinitionCheckTemplate `json:"checkTemplates"`
}

type ServiceDefinitionMatch struct {
	Definition ServiceDefinition `json:"definition"`
	Score      int               `json:"score"`
	Reasons    []string          `json:"reasons,omitempty"`
}

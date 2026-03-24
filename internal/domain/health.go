package domain

import "time"

type EndpointTestInput struct {
	Check         ServiceCheck `json:"check"`
	DiscoverPaths bool         `json:"discoverPaths,omitempty"`
}

type EndpointTestResult struct {
	Check                    ServiceCheck       `json:"check"`
	ResolvedURL              string             `json:"resolvedUrl,omitempty"`
	MatchedServiceDefinition *ServiceDefinition `json:"matchedServiceDefinition,omitempty"`
	Status                   HealthStatus       `json:"status"`
	HTTPStatusCode           int                `json:"httpStatusCode,omitempty"`
	LatencyMS                int64              `json:"latencyMs"`
	ResponseSizeBytes        int64              `json:"responseSizeBytes,omitempty"`
	Message                  string             `json:"message,omitempty"`
	CheckedAt                time.Time          `json:"checkedAt"`
}

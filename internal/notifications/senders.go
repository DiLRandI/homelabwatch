package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/deleema/homelabwatch/internal/domain"
)

func Send(ctx context.Context, channel domain.NotificationChannel, event NotificationEvent) error {
	switch channel.Type {
	case domain.NotificationChannelWebhook:
		return sendWebhook(ctx, channel, event)
	case domain.NotificationChannelNtfy:
		return sendNtfy(ctx, channel, event)
	default:
		return fmt.Errorf("unsupported channel type %s", channel.Type)
	}
}

func sendWebhook(ctx context.Context, channel domain.NotificationChannel, event NotificationEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}
	timeout := timeoutFromConfig(channel.Config)
	requestCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	request, err := http.NewRequestWithContext(requestCtx, http.MethodPost, configString(channel.Config, "url"), bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return fmt.Errorf("webhook returned %s", response.Status)
	}
	return nil
}

func sendNtfy(ctx context.Context, channel domain.NotificationChannel, event NotificationEvent) error {
	serverURL := strings.TrimRight(configString(channel.Config, "serverUrl"), "/")
	topic := strings.Trim(configString(channel.Config, "topic"), "/")
	timeout := timeoutFromConfig(channel.Config)
	requestCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	request, err := http.NewRequestWithContext(requestCtx, http.MethodPost, serverURL+"/"+topic, strings.NewReader(event.Message))
	if err != nil {
		return err
	}
	request.Header.Set("Title", event.Title)
	request.Header.Set("Tags", "homelabwatch")
	if priority := configString(channel.Config, "priority"); priority != "" && priority != "default" {
		request.Header.Set("Priority", priority)
	}
	if token := configString(channel.Config, "token"); token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return fmt.Errorf("ntfy returned %s", response.Status)
	}
	return nil
}

func timeoutFromConfig(config map[string]any) time.Duration {
	value := 10
	switch typed := config["timeoutSeconds"].(type) {
	case int:
		value = typed
	case float64:
		value = int(typed)
	}
	if value <= 0 {
		value = 10
	}
	return time.Duration(value) * time.Second
}

func configString(config map[string]any, key string) string {
	if config == nil || config[key] == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(config[key]))
}

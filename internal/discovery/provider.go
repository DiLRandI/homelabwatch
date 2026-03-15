package discovery

import (
	"context"

	"github.com/deleema/homelabwatch/internal/domain"
)

type Provider interface {
	Name() string
	Discover(context.Context) ([]domain.Observation, error)
}

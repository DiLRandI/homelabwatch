package manual

import (
	"context"

	"github.com/deleema/homelabwatch/internal/domain"
)

type Provider struct{}

func NewProvider() *Provider {
	return &Provider{}
}

func (p *Provider) Name() string {
	return "manual"
}

func (p *Provider) Discover(context.Context) ([]domain.Observation, error) {
	return nil, nil
}

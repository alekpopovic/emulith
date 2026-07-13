package aws

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/alekpopovic/emulith/internal/service"
	"github.com/alekpopovic/emulith/internal/state"
)

type Provider struct {
	gateway  *Gateway
	registry *service.Registry
}

func NewProvider(store *state.Store, logger *slog.Logger, registry *service.Registry) *Provider {
	return &Provider{gateway: NewGateway(store, logger), registry: registry}
}
func (p *Provider) Gateway() *Gateway { return p.gateway }
func (p *Provider) Register(name string, handler Handler) error {
	if _, ok := p.registry.Find("aws", name); ok {
		return fmt.Errorf("duplicate service aws.%s", name)
	}
	p.gateway.handlers[name] = handler
	return p.registry.Register(awsService{name: name})
}

type awsService struct{ name string }

func (s awsService) Provider() string { return "aws" }
func (s awsService) Name() string     { return s.name }
func (s awsService) Health(context.Context) service.HealthStatus {
	return service.HealthStatus{Status: "ok"}
}
func (s awsService) Reset(context.Context) error { return nil }

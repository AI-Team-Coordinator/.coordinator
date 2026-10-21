package app

import (
	"relay/domain/activity"
	"relay/domain/config"
	"relay/domain/connector"
	"relay/domain/gateway"
	"relay/domain/marketplace"
	"relay/domain/poll"
	"relay/infra/settings"
)

// ConnectorInit builds a connector after Runtime and Config exist (secrets from marketplace).
type ConnectorInit func(runtime *config.Runtime, cfg Config) connector.Connector

// Option extends a relay instance (project connectors, marketplace tiles, poll jobs).
type Option func(*instanceOpts)

// ActionRegistrar adds connector-specific on-demand actions (beyond auto poll_now).
type ActionRegistrar func(reg *connector.ActionRegistry, gw *gateway.Gateway, runtime *config.Runtime, settings *settings.RelayConnectors, pollJobs []poll.Job)

type instanceOpts struct {
	connectors      []connector.Connector
	connectorInits  []ConnectorInit
	catalog         []marketplace.CatalogEntry
	pollFactories   []func(*gateway.Gateway, *activity.Recorder) poll.Job
	actionRegs      []ActionRegistrar
}

func WithConnectorInit(init ConnectorInit) Option {
	return func(o *instanceOpts) {
		if init != nil {
			o.connectorInits = append(o.connectorInits, init)
		}
	}
}

func WithConnector(c connector.Connector) Option {
	return func(o *instanceOpts) {
		if c != nil {
			o.connectors = append(o.connectors, c)
		}
	}
}

func WithCatalogEntry(e marketplace.CatalogEntry) Option {
	return func(o *instanceOpts) {
		o.catalog = append(o.catalog, e)
	}
}

func WithPoll(factory func(*gateway.Gateway, *activity.Recorder) poll.Job) Option {
	return func(o *instanceOpts) {
		if factory != nil {
			o.pollFactories = append(o.pollFactories, factory)
		}
	}
}

func WithActions(reg ActionRegistrar) Option {
	return func(o *instanceOpts) {
		if reg != nil {
			o.actionRegs = append(o.actionRegs, reg)
		}
	}
}

func applyOptions(opts []Option) instanceOpts {
	var out instanceOpts
	for _, opt := range opts {
		if opt != nil {
			opt(&out)
		}
	}
	return out
}

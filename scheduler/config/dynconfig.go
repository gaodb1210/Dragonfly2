/*
 *     Copyright 2020 The Dragonfly Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

//go:generate mockgen -destination mocks/dynconfig_mock.go -source dynconfig.go -package mocks

package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/status"

	managerv1 "d7y.io/api/pkg/apis/manager/v1"

	logger "d7y.io/dragonfly/v2/internal/dflog"
	dc "d7y.io/dragonfly/v2/internal/dynconfig"
	"d7y.io/dragonfly/v2/manager/types"
	"d7y.io/dragonfly/v2/pkg/net/ip"
	healthclient "d7y.io/dragonfly/v2/pkg/rpc/health/client"
	managerclient "d7y.io/dragonfly/v2/pkg/rpc/manager/client"
)

var (
	// Cache filename.
	cacheFileName = "scheduler"

	// Notify observer interval.
	watchInterval = 10 * time.Second
)

type DynconfigData struct {
	Scheduler    *managerv1.Scheduler
	Applications []*managerv1.Application
}

type DynconfigInterface interface {
	// GetResolveSeedPeerAddrs returns the dynamic schedulers resolve addrs.
	GetResolveSeedPeerAddrs() ([]resolver.Address, error)

	// GetApplications returns the applications config from manager.
	GetApplications() ([]*managerv1.Application, error)

	// GetSeedPeers returns the dynamic seed peers config from manager.
	GetSeedPeers() ([]*managerv1.SeedPeer, error)

	// GetSchedulerClusterConfig returns the scheduler cluster config.
	GetSchedulerClusterConfig() (types.SchedulerClusterConfig, error)

	// GetSchedulerClusterClientConfig returns the client config.
	GetSchedulerClusterClientConfig() (types.SchedulerClusterClientConfig, error)

	// Get returns the dynamic config from manager.
	Get() (*DynconfigData, error)

	// Refresh refreshes dynconfig in cache.
	Refresh() error

	// Register allows an instance to register itself to listen/observe events.
	Register(Observer)

	// Deregister allows an instance to remove itself from the collection of observers/listeners.
	Deregister(Observer)

	// Notify publishes new events to listeners.
	Notify() error

	// Serve the dynconfig listening service.
	Serve() error

	// Stop the dynconfig listening service.
	Stop() error
}

type Observer interface {
	// OnNotify allows an event to be published to interface implementations.
	OnNotify(*DynconfigData)
}

type dynconfig struct {
	dc.Dynconfig
	observers            map[Observer]struct{}
	done                 chan struct{}
	cachePath            string
	transportCredentials credentials.TransportCredentials
}

// DynconfigOption is a functional option for configuring the dynconfig.
type DynconfigOption func(d *dynconfig) error

// WithTransportCredentials returns a DialOption which configures a connection
// level security credentials (e.g., TLS/SSL).
func WithTransportCredentials(creds credentials.TransportCredentials) DynconfigOption {
	return func(d *dynconfig) error {
		d.transportCredentials = creds
		return nil
	}
}

// NewDynconfig returns a new dynconfig instence.
func NewDynconfig(rawManagerClient managerclient.Client, cacheDir string, cfg *Config, options ...DynconfigOption) (DynconfigInterface, error) {
	cachePath := filepath.Join(cacheDir, cacheFileName)
	d := &dynconfig{
		observers: map[Observer]struct{}{},
		done:      make(chan struct{}),
		cachePath: cachePath,
	}

	for _, opt := range options {
		if err := opt(d); err != nil {
			return nil, err
		}
	}

	if rawManagerClient != nil {
		client, err := dc.New(
			newManagerClient(rawManagerClient, cfg),
			cachePath,
			cfg.DynConfig.RefreshInterval,
		)
		if err != nil {
			return nil, err
		}

		d.Dynconfig = client
	}

	return d, nil
}

// GetResolveSeedPeerAddrs returns the dynamic schedulers resolve addrs.
func (d *dynconfig) GetResolveSeedPeerAddrs() ([]resolver.Address, error) {
	seedPeers, err := d.GetSeedPeers()
	if err != nil {
		return nil, err
	}

	var (
		addrs        = map[string]bool{}
		resolveAddrs []resolver.Address
	)
	for _, seedPeer := range seedPeers {
		ip, ok := ip.FormatIP(seedPeer.Ip)
		if !ok {
			continue
		}

		addr := fmt.Sprintf("%s:%d", ip, seedPeer.Port)
		dialOptions := []grpc.DialOption{}
		if d.transportCredentials != nil {
			dialOptions = append(dialOptions, grpc.WithTransportCredentials(d.transportCredentials))
		} else {
			dialOptions = append(dialOptions, grpc.WithTransportCredentials(insecure.NewCredentials()))
		}

		healthClient, err := healthclient.GetClient(context.Background(), addr, dialOptions...)
		if err != nil {
			logger.Errorf("get health client %s failed: %s", addr, err.Error())
			continue
		}

		if err := healthClient.Check(context.Background(), &healthpb.HealthCheckRequest{}); err != nil {
			logger.Errorf("seed peer address %s is unreachable: %s", addr, err.Error())
			continue
		}

		if addrs[addr] {
			continue
		}

		resolveAddrs = append(resolveAddrs, resolver.Address{
			ServerName: seedPeer.Ip,
			Addr:       addr,
		})
		addrs[addr] = true
	}

	if len(resolveAddrs) == 0 {
		return nil, errors.New("can not found available seed peer addresses")
	}

	return resolveAddrs, nil
}

// GetApplications returns the applications config from manager.
func (d *dynconfig) GetApplications() ([]*managerv1.Application, error) {
	data, err := d.Get()
	if err != nil {
		return nil, err
	}

	return data.Applications, nil
}

// GetSeedPeers returns the dynamic seed peers config from manager.
func (d *dynconfig) GetSeedPeers() ([]*managerv1.SeedPeer, error) {
	data, err := d.Get()
	if err != nil {
		return nil, err
	}

	return data.Scheduler.SeedPeers, nil
}

// GetSchedulerClusterConfig returns the scheduler cluster config.
func (d *dynconfig) GetSchedulerClusterConfig() (types.SchedulerClusterConfig, error) {
	data, err := d.Get()
	if err != nil {
		return types.SchedulerClusterConfig{}, err
	}

	if data.Scheduler.SchedulerCluster == nil {
		return types.SchedulerClusterConfig{}, errors.New("invalid scheduler cluster")
	}

	var config types.SchedulerClusterConfig
	if err := json.Unmarshal(data.Scheduler.SchedulerCluster.Config, &config); err != nil {
		return types.SchedulerClusterConfig{}, err
	}

	return config, nil
}

// GetSchedulerClusterClientConfig returns the client config.
func (d *dynconfig) GetSchedulerClusterClientConfig() (types.SchedulerClusterClientConfig, error) {
	data, err := d.Get()
	if err != nil {
		return types.SchedulerClusterClientConfig{}, err
	}

	if data.Scheduler.SchedulerCluster == nil {
		return types.SchedulerClusterClientConfig{}, errors.New("invalid scheduler cluster")
	}

	var config types.SchedulerClusterClientConfig
	if err := json.Unmarshal(data.Scheduler.SchedulerCluster.ClientConfig, &config); err != nil {
		return types.SchedulerClusterClientConfig{}, err
	}

	return config, nil
}

// Get the dynamic config from manager.
func (d *dynconfig) Get() (*DynconfigData, error) {
	var config DynconfigData
	if err := d.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Refresh refreshes dynconfig in cache.
func (d *dynconfig) Refresh() error {
	if err := d.Dynconfig.Refresh(); err != nil {
		return err
	}

	if err := d.Notify(); err != nil {
		return err
	}

	return nil
}

// Register allows an instance to register itself to listen/observe events.
func (d *dynconfig) Register(l Observer) {
	d.observers[l] = struct{}{}
}

// Deregister allows an instance to remove itself from the collection of observers/listeners.
func (d *dynconfig) Deregister(l Observer) {
	delete(d.observers, l)
}

// Notify publishes new events to listeners.
func (d *dynconfig) Notify() error {
	config, err := d.Get()
	if err != nil {
		return err
	}

	for o := range d.observers {
		o.OnNotify(config)
	}

	return nil
}

// Serve the dynconfig listening service.
func (d *dynconfig) Serve() error {
	if err := d.Notify(); err != nil {
		return err
	}

	tick := time.NewTicker(watchInterval)
	for {
		select {
		case <-tick.C:
			if err := d.Notify(); err != nil {
				logger.Error("dynconfig notify failed", err)
			}
		case <-d.done:
			return nil
		}
	}
}

// Stop the dynconfig listening service.
func (d *dynconfig) Stop() error {
	close(d.done)
	if err := os.Remove(d.cachePath); err != nil {
		return err
	}

	return nil
}

// Manager client for dynconfig.
type managerClient struct {
	managerclient.Client
	config *Config
}

// New the manager client used by dynconfig.
func newManagerClient(client managerclient.Client, cfg *Config) dc.ManagerClient {
	return &managerClient{
		Client: client,
		config: cfg,
	}
}

func (mc *managerClient) Get() (any, error) {
	getSchedulerResp, err := mc.GetScheduler(context.Background(), &managerv1.GetSchedulerRequest{
		SourceType:         managerv1.SourceType_SCHEDULER_SOURCE,
		HostName:           mc.config.Server.Host,
		Ip:                 mc.config.Server.AdvertiseIP,
		SchedulerClusterId: uint64(mc.config.Manager.SchedulerClusterID),
	})
	if err != nil {
		return nil, err
	}

	listApplicationsResp, err := mc.ListApplications(context.Background(), &managerv1.ListApplicationsRequest{
		SourceType: managerv1.SourceType_SCHEDULER_SOURCE,
		HostName:   mc.config.Server.Host,
		Ip:         mc.config.Server.AdvertiseIP,
	})
	if err != nil {
		// TODO Compatible with old version manager.
		if s, ok := status.FromError(err); ok && s.Code() == codes.Unimplemented {
			return DynconfigData{
				Scheduler:    getSchedulerResp,
				Applications: nil,
			}, nil
		}

		return nil, err
	}

	return DynconfigData{
		Scheduler:    getSchedulerResp,
		Applications: listApplicationsResp.Applications,
	}, nil
}

// GetSeedPeerClusterConfigBySeedPeer returns the seed peer cluster config by seed peer.
func GetSeedPeerClusterConfigBySeedPeer(seedPeer *managerv1.SeedPeer) (types.SeedPeerClusterConfig, error) {
	if seedPeer == nil {
		return types.SeedPeerClusterConfig{}, errors.New("invalid seed peer")
	}

	if seedPeer.SeedPeerCluster == nil {
		return types.SeedPeerClusterConfig{}, errors.New("invalid seed peer cluster")
	}

	var config types.SeedPeerClusterConfig
	if err := json.Unmarshal(seedPeer.SeedPeerCluster.Config, &config); err != nil {
		return types.SeedPeerClusterConfig{}, err
	}

	return config, nil
}

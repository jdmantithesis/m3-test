// Copyright (c) 2021  Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package inprocess

import (
	"errors"
	"fmt"
	"net"
	"strconv"

	dbcfg "github.com/m3db/m3/src/cmd/services/m3dbnode/config"
	coordinatorcfg "github.com/m3db/m3/src/cmd/services/m3query/config"
	"github.com/m3db/m3/src/dbnode/discovery"
	"github.com/m3db/m3/src/dbnode/environment"
	"github.com/m3db/m3/src/integration/resources"
	"github.com/m3db/m3/src/x/config/hostid"
	xerrors "github.com/m3db/m3/src/x/errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

// ClusterOptions contains options for spinning up a new M3 cluster
// composed of in-process components.
type ClusterOptions struct {
	// Coordinator contains cluster options for spinning up a coordinator.
	Coordinator CoordinatorClusterOptions
	// DBNode contains cluster options for spinning up dbnodes.
	DBNode DBNodeClusterOptions
}

// DBNodeClusterOptions contains the cluster options for spinning up
// dbnodes.
type DBNodeClusterOptions struct {
	// Config contains the dbnode configuration.
	Config DBNodeClusterConfig
	// RF is the replication factor to use for the cluster.
	RF int
	// NumShards is the number of shards to use for each RF.
	NumShards int
	// NumInstances is the number of dbnode instances per RF.
	NumInstances int
}

// NewDBNodeClusterOptions creates DBNodeClusteOptions with sane defaults.
// DBNode config must still be provided.
func NewDBNodeClusterOptions() DBNodeClusterOptions {
	return DBNodeClusterOptions{
		RF:           1,
		NumShards:    4,
		NumInstances: 1,
	}
}

// Validate validates the DBNodeClusterOptions.
func (d *DBNodeClusterOptions) Validate() error {
	if d.RF < 1 {
		return errors.New("rf must be at least 1")
	}

	if d.NumShards < 1 {
		return errors.New("numShards must be at least 1")
	}

	if d.NumInstances < 1 {
		return errors.New("numInstances must be at least 1")
	}

	return d.Config.Validate()
}

// CoordinatorClusterOptions contains the options for spinning up
// a coordinator.
type CoordinatorClusterOptions struct {
	// Config contains the coordinator configuration.
	Config CoordinatorClusterConfig
}

// Validate validates the CoordinatorClusterOptions.
func (c *CoordinatorClusterOptions) Validate() error {
	return c.Config.Validate()
}

// DBNodeClusterConfig contains the configuration for dbnodes in the
// cluster. Must specify one of the options but not both.
type DBNodeClusterConfig struct {
	// ConfigString contains the configuration as a raw YAML string.
	ConfigString string
	// ConfigObject contains the configuration as an inflated object.
	ConfigObject *dbcfg.Configuration
}

// Validate validates the DBNodeClusterConfig.
func (d *DBNodeClusterConfig) Validate() error {
	if d.ConfigString != "" && d.ConfigObject != nil {
		return errors.New("must specify either ConfigString or ConfigObject, but not both")
	}

	if d.ConfigString == "" && d.ConfigObject == nil {
		return errors.New("ConfigString and ConfigObject cannot both be empty")
	}

	return nil
}

// ToConfig generates a dbcfg.Configuration object from the DBNodeClusterConfig.
func (d *DBNodeClusterConfig) ToConfig() (dbcfg.Configuration, error) {
	if d.ConfigObject != nil {
		return *d.ConfigObject, nil
	}

	var cfg dbcfg.Configuration
	if err := yaml.Unmarshal([]byte(d.ConfigString), &cfg); err != nil {
		return dbcfg.Configuration{}, err
	}

	return cfg, nil
}

// CoordinatorClusterConfig contains the configuration for coordinators in the
// cluster. Must specify one of the options but not both.
type CoordinatorClusterConfig struct {
	// ConfigString contains the configuration as a raw YAML string.
	ConfigString string
	// ConfigObject contains the configuration as an inflated object.
	ConfigObject *coordinatorcfg.Configuration
}

// Validate validates the CoordinatorClusterConfig.
func (c *CoordinatorClusterConfig) Validate() error {
	if c.ConfigString != "" && c.ConfigObject != nil {
		return errors.New("must specify either ConfigString or ConfigObject, but not both")
	}

	if c.ConfigString == "" && c.ConfigObject == nil {
		return errors.New("ConfigString and ConfigObject cannot both be empty")
	}

	return nil
}

// ToConfig generates a coordinatorcfg.Configuration object from the CoordinatorClusterConfig.
func (c *CoordinatorClusterConfig) ToConfig() (coordinatorcfg.Configuration, error) {
	if c.ConfigObject != nil {
		return *c.ConfigObject, nil
	}

	var cfg coordinatorcfg.Configuration
	if err := yaml.Unmarshal([]byte(c.ConfigString), &cfg); err != nil {
		return coordinatorcfg.Configuration{}, err
	}

	return cfg, nil
}

// NewCluster creates a new M3 cluster based on the ClusterOptions provided.
// Expects at least a coordinator and a dbnode config.
func NewCluster(opts ClusterOptions) (resources.M3Resources, error) {
	if err := opts.DBNode.Validate(); err != nil {
		return nil, err
	}

	if err := opts.Coordinator.Validate(); err != nil {
		return nil, err
	}

	logger, err := newLogger()
	if err != nil {
		return nil, err
	}

	var (
		numNodes            = opts.DBNode.RF * opts.DBNode.NumInstances
		generatePortsAndIDs = numNodes > 1
		coord               resources.Coordinator
		nodes               = make(resources.Nodes, 0, numNodes)
		defaultDBNodeOpts   = DBNodeOptions{
			GenerateHostID: generatePortsAndIDs,
			GeneratePorts:  generatePortsAndIDs,
		}
	)

	// TODO(nate): eventually support clients specifying their own discovery stanza.
	// Practically, this should cover 99% of cases.
	//
	// Generate a discovery config with the dbnode using the generated hostID marked as
	// the etcd server (i.e. seed node).
	hostID := uuid.NewString()
	defaultDBNodesCfg, err := opts.DBNode.Config.ToConfig()
	if err != nil {
		return nil, err
	}
	discoveryCfg, envConfig, err := generateDefaultDiscoveryConfig(defaultDBNodesCfg, hostID)

	// Ensure that once we start creating resources, they all get cleaned up even if the function
	// fails half way.
	defer func() {
		if err != nil {
			cleanup(logger, nodes, coord)
		}
	}()

	for i := 0; i < numNodes; i++ {
		cfg := defaultDBNodesCfg
		dbnodeOpts := defaultDBNodeOpts

		if i == 0 {
			// Mark the initial node as the etcd seed node.
			dbnodeOpts.GenerateHostID = false
			cfg.DB.HostID = &hostid.Configuration{
				Resolver: hostid.ConfigResolver,
				Value:    &hostID,
			}
		}
		cfg.DB.Discovery = &discoveryCfg

		var node resources.Node
		node, err = NewDBNode(cfg, dbnodeOpts)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}

	var coordConfig coordinatorcfg.Configuration
	coordConfig, err = opts.Coordinator.Config.ToConfig()
	if err != nil {
		return nil, err
	}
	// TODO(nate): refactor to support having envconfig if no DB.
	coordConfig.Clusters[0].Client.EnvironmentConfig = &envConfig
	coord, err = NewCoordinator(coordConfig, CoordinatorOptions{})
	if err != nil {
		return nil, err
	}

	m3 := NewM3Resources(ResourceOptions{
		Coordinator: coord,
		DBNodes:     nodes,
	})
	if err = resources.SetupCluster(m3, &resources.ClusterOptions{
		ReplicationFactor: int32(opts.DBNode.RF),
		NumShards:         int32(opts.DBNode.NumShards),
	}); err != nil {
		return nil, err
	}

	return m3, nil
}

// generateDefaultDiscoveryConfig handles creating the correct config
// for having an embedded ETCD server with the correct server and
// client configuration.
func generateDefaultDiscoveryConfig(
	cfg dbcfg.Configuration,
	hostID string,
) (discovery.Configuration, environment.Configuration, error) {
	discoveryConfig := cfg.DB.DiscoveryOrDefault()
	envConfig, err := discoveryConfig.EnvironmentConfig(hostID)
	if err != nil {
		return discovery.Configuration{}, environment.Configuration{}, nil
	}

	// TODO(nate): Fix expectations in envconfig for:
	//   - InitialAdvertisePeerUrls
	//	 - AdvertiseClientUrls
	//	 - ListenPeerUrls
	//	 - ListenClientUrls
	// when not using the default ports of 2379 and 2380
	envConfig.SeedNodes.InitialCluster[0].Endpoint =
		fmt.Sprintf("http://0.0.0.0:%d", 2380)
	envConfig.Services[0].Service.ETCDClusters[0].Endpoints = []string{
		net.JoinHostPort("0.0.0.0", strconv.Itoa(2379)),
	}
	configType := discovery.ConfigType
	return discovery.Configuration{
		Type:   &configType,
		Config: &envConfig,
	}, envConfig, nil
}

func cleanup(logger *zap.Logger, nodes resources.Nodes, coord resources.Coordinator) {
	var multiErr xerrors.MultiError
	for _, n := range nodes {
		multiErr = multiErr.Add(n.Close())
	}

	if coord != nil {
		multiErr = multiErr.Add(coord.Close())
	}

	if !multiErr.Empty() {
		logger.Warn("failed closing resources", zap.Error(multiErr.FinalError()))
	}
}

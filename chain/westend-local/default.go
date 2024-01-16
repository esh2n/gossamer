// Copyright 2023 ChainSafe Systems (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package westendlocal

import (
	cfg "github.com/ChainSafe/gossamer/config"
	"github.com/adrg/xdg"
)

const (
	// defaultChainSpec is the default chain spec for the westend local node
	defaultChainSpec = "./chain/westend-local/westend-local-spec-raw.json"
)

// DefaultConfig returns a westend local node configuration
func DefaultConfig() *cfg.Config {
	config := cfg.DefaultConfig()
	config.ChainSpec = defaultChainSpec
	config.Network.NoMDNS = false
	config.RPC.RPCExternal = true
	config.RPC.UnsafeRPC = true
	config.RPC.WSExternal = true
	config.RPC.UnsafeWSExternal = true

	return config
}

// DefaultAliceConfig returns a westend local node configuration
func DefaultAliceConfig() *cfg.Config {
	config := DefaultConfig()
	config.ConfigDir = xdg.ConfigHome + "/westend-local-alice/config"
	config.DataDir = xdg.DataHome + "/westend-local-alice/data"
	config.PrometheusPort = uint32(9856)
	config.Network.Port = 7001
	config.RPC.Port = 8545
	config.RPC.WSPort = 8546
	config.Pprof.ListeningAddress = "localhost:6060"

	return config
}

// DefaultBobConfig returns a westend local node configuration with bob as the authority
func DefaultBobConfig() *cfg.Config {
	config := DefaultConfig()
	config.ConfigDir = xdg.ConfigHome + "/westend-local-bob/config"
	config.DataDir = xdg.DataHome + "/westend-local-bob/data"
	config.PrometheusPort = uint32(9866)
	config.Network.Port = 7011
	config.RPC.Port = 8555
	config.RPC.WSPort = 8556
	config.Pprof.ListeningAddress = "localhost:6070"

	return config
}

// DefaultCharlieConfig returns a westend local node configuration with charlie as the authority
func DefaultCharlieConfig() *cfg.Config {
	config := DefaultConfig()
	config.ConfigDir = xdg.ConfigHome + "/westend-local-charlie/config"
	config.DataDir = xdg.DataHome + "/westend-local-charlie/data"
	config.PrometheusPort = uint32(9876)
	config.Network.Port = 7021
	config.RPC.Port = 8565
	config.RPC.WSPort = 8566
	config.Pprof.ListeningAddress = "localhost:6080"

	return config
}

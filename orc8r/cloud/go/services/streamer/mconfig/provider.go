/*
Copyright (c) Facebook, Inc. and its affiliates.
All rights reserved.

This source code is licensed under the BSD-style license found in the
LICENSE file in the root directory of this source tree.
*/

package mconfig

import (
	"os"

	"magma/orc8r/cloud/go/orc8r"
	"magma/orc8r/cloud/go/protos"
	"magma/orc8r/cloud/go/services/configurator"
	"magma/orc8r/cloud/go/services/magmad"
	"magma/orc8r/cloud/go/services/streamer/mconfig/factory"
	"magma/orc8r/cloud/go/services/streamer/providers"

	"github.com/golang/protobuf/ptypes/any"
)

// GetProvider returns the StreamProvider for on demand mconfigs.
func GetProvider() providers.StreamProvider {
	return &ConfigProvider{}
}

type ConfigProvider struct{}

func (provider *ConfigProvider) GetStreamName() string {
	return "configs"
}

func (provider *ConfigProvider) GetUpdates(gatewayId string, extraArgs *any.Any) ([]*protos.DataUpdate, error) {
	migrated := os.Getenv(orc8r.UseConfiguratorMconfigsEnv)
	if migrated == "1" {
		resp, err := configurator.GetMconfigFor(gatewayId)
		if err != nil {
			return nil, err
		}
		return mconfigToUpdate(resp.Configs, resp.LogicalID)
	} else {
		return buildMconfigLegacy(gatewayId)
	}
}

func buildMconfigLegacy(gatewayId string) ([]*protos.DataUpdate, error) {
	networkId, err := magmad.FindGatewayNetworkId(gatewayId)
	if err != nil {
		return nil, err
	}
	logicalId, err := magmad.FindGatewayId(networkId, gatewayId)
	if err != nil {
		return nil, err
	}

	gatewayConfig, err := factory.CreateMconfig(networkId, logicalId)
	if err != nil {
		return nil, err
	}
	return mconfigToUpdate(gatewayConfig, logicalId)
}

func mconfigToUpdate(configs *protos.GatewayConfigs, key string) ([]*protos.DataUpdate, error) {
	marshaledConfig, err := protos.MarshalIntern(configs)
	if err != nil {
		return nil, err
	}
	update := new(protos.DataUpdate)
	update.Key = key
	update.Value = marshaledConfig
	return []*protos.DataUpdate{update}, nil

}

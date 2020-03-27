package sds

import (
	"fmt"

	config_kumadp "github.com/Kong/kuma/pkg/config/app/kuma-dp"
	sds_auth "github.com/Kong/kuma/pkg/sds/auth"
	sds_vault "github.com/Kong/kuma/pkg/sds/provider/vault"
	sds_server "github.com/Kong/kuma/pkg/sds/server"
	util_xds "github.com/Kong/kuma/pkg/util/xds"
)

func NewVaultSdsServer(dataplane config_kumadp.Dataplane, vaultCfg config_kumadp.Vault, serviceName string) (*grpcServer, error) {
	callbacks := util_xds.CallbacksChain{
		util_xds.LoggingCallbacks{Log: sdsServerLog},
	}

	client, err := sds_vault.NewVaultClient(convertConfig(vaultCfg))
	if err != nil {
		return nil, err
	}
	handler := &dpSdsHandler{
		dpIdentity: sds_auth.Identity{
			Mesh:    dataplane.Mesh,
			Service: serviceName,
		},
		identitySecretProvider: sds_vault.NewIdentityCertProvider(client),
		meshSecretProvider:     sds_vault.NewMeshCaProvider(client),
	}
	return &grpcServer{
		server:  sds_server.NewServer(handler, callbacks, sdsServerLog),
		address: fmt.Sprintf("/tmp/kuma-sds-%s-%s.sock", dataplane.Name, dataplane.Mesh),
	}, nil
}

func convertConfig(dpVaultCfg config_kumadp.Vault) sds_vault.Config {
	return sds_vault.Config{
		Address:      dpVaultCfg.Address,
		AgentAddress: dpVaultCfg.AgentAddress,
		Token:        dpVaultCfg.Token,
		Namespace:    dpVaultCfg.Namespace,
		Tls: sds_vault.TLSConfig{
			CaCertPath:     dpVaultCfg.TLS.CaCertPath,
			CaCertDir:      dpVaultCfg.TLS.CaCertDir,
			ClientCertPath: dpVaultCfg.TLS.ClientCertPath,
			ClientKeyPath:  dpVaultCfg.TLS.ClientKeyPath,
			SkipVerify:     dpVaultCfg.TLS.SkipVerify,
			ServerName:     dpVaultCfg.TLS.ServerName,
		},
	}
}
package connectivity

//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

import (
	bmc "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/bmc20221120"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	sdn "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/sdn20230830"
	vm "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/vm20230313"
	zga "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zga20230706"
)

var ReqClient = "Terraform-latest"

type ZenlayerCloudClient struct {
	SecretKeyId       string
	SecretKeyPassword string
	Domain            string
	Scheme            string
	Timeout           int
	BmcConn           *bmc.Client
	VmConn            *vm.Client
	SdnConn           *sdn.Client
	ZgaConn           *zga.Client
}

func (client *ZenlayerCloudClient) WithSdnClient() *sdn.Client {
	if client.SdnConn != nil {
		return client.SdnConn
	}
	config := client.NewConfig()
	client.SdnConn, _ = sdn.NewClient(config, client.SecretKeyId, client.SecretKeyPassword)
	client.SdnConn.WithRequestClient(ReqClient)
	return client.SdnConn
}

func (client *ZenlayerCloudClient) WithBmcClient() *bmc.Client {
	if client.BmcConn != nil {
		return client.BmcConn
	}
	config := client.NewConfig()
	client.BmcConn, _ = bmc.NewClient(config, client.SecretKeyId, client.SecretKeyPassword)
	client.BmcConn.WithRequestClient(ReqClient)
	return client.BmcConn
}

func (client *ZenlayerCloudClient) WithVmClient() *vm.Client {
	if client.VmConn != nil {
		return client.VmConn
	}
	config := client.NewConfig()
	client.VmConn, _ = vm.NewClient(config, client.SecretKeyId, client.SecretKeyPassword)
	client.VmConn.WithRequestClient(ReqClient)
	return client.VmConn
}

func (client *ZenlayerCloudClient) WithZgaClient() *zga.Client {
	if client.ZgaConn != nil {
		return client.ZgaConn
	}
	config := client.NewConfig()
	client.ZgaConn, _ = zga.NewClient(config, client.SecretKeyId, client.SecretKeyPassword)
	client.ZgaConn.WithRequestClient(ReqClient)
	return client.ZgaConn
}

func (client *ZenlayerCloudClient) NewConfig() *common.Config {
	config := common.NewConfig()
	config.Timeout = client.Timeout
	config.Scheme = client.Scheme
	config.Domain = client.Domain
	return config
}

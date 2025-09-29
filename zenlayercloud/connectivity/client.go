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
	ccs "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/ccs20250901"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	sdn "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/sdn20230830"
	traffic "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/traffic20240326"
	user "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/user20240529"
	vm "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/vm20230313"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20240401"
	zec2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"
	zga "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zga20230706"
	zlb "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zlb20250401"
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
	ZecConn           *zec.Client
	ZecConn2           *zec2.Client
	ZlbConn           *zlb.Client
	tfkConn           *traffic.Client
	usrConn           *user.Client
	CcsConn           *ccs.Client
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

func (client *ZenlayerCloudClient) WithCcsClient() *ccs.Client {
	if client.CcsConn != nil {
		return client.CcsConn
	}
	config := client.NewConfig()
	client.CcsConn, _ = ccs.NewClient(config, client.SecretKeyId, client.SecretKeyPassword)
	client.CcsConn.WithRequestClient(ReqClient)
	return client.CcsConn
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

func (client *ZenlayerCloudClient) WithZecClient() *zec.Client {
	if client.ZecConn != nil {
		return client.ZecConn
	}
	config := client.NewConfig()
	client.ZecConn, _ = zec.NewClient(config, client.SecretKeyId, client.SecretKeyPassword)
	client.ZecConn.WithRequestClient(ReqClient)
	return client.ZecConn
}

func (client *ZenlayerCloudClient) WithZec2Client() *zec2.Client {
	if client.ZecConn2 != nil {
		return client.ZecConn2
	}
	config := client.NewConfig()
	client.ZecConn2, _ = zec2.NewClient(config, client.SecretKeyId, client.SecretKeyPassword)
	client.ZecConn2.WithRequestClient(ReqClient)
	return client.ZecConn2
}

func (client *ZenlayerCloudClient) WithZlbClient() *zlb.Client {
	if client.ZlbConn != nil {
		return client.ZlbConn
	}
	config := client.NewConfig()
	client.ZlbConn, _ = zlb.NewClient(config, client.SecretKeyId, client.SecretKeyPassword)
	client.ZlbConn.WithRequestClient(ReqClient)
	return client.ZlbConn
}

func (client *ZenlayerCloudClient) WithTrafficClient() *traffic.Client {
	if client.tfkConn != nil {
		return client.tfkConn
	}
	config := client.NewConfig()
	client.tfkConn, _ = traffic.NewClient(config, client.SecretKeyId, client.SecretKeyPassword)
	client.tfkConn.WithRequestClient(ReqClient)
	return client.tfkConn
}

func (client *ZenlayerCloudClient) WithUsrClient() *user.Client {
	if client.usrConn != nil {
		return client.usrConn
	}
	config := client.NewConfig()
	client.usrConn, _ = user.NewClient(config, client.SecretKeyId, client.SecretKeyPassword)
	client.usrConn.WithRequestClient(ReqClient)
	return client.usrConn
}

func (client *ZenlayerCloudClient) NewConfig() *common.Config {
	config := common.NewConfig()
	config.Timeout = client.Timeout
	config.Scheme = client.Scheme
	config.Domain = client.Domain
	return config
}

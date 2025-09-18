/*
The ZenlayerCloud provider is used to interact with many resources supported by [ZenlayerCloud](https://console.zenlayer.com).
The provider needs to be configured with the proper credentials before it can be used.

Use the navigation on the left to read about the available resources.

Example Usage

```hcl

terraform {
  required_providers {
	zenlayercloud = {
	  source = "zenlayer/zenlayercloud"
	}
  }
}

# Configure the Zenlayer Cloud Provider
provider "zenlayercloud" {
  access_key_id       =  var.access_key_id
  access_key_password =  var.access_key_password
}

```

Resources List

Provider Data Sources

Zenlayer Virtual Machine(VM)
  Data Source
	zenlayercloud_zones
	zenlayercloud_images
	zenlayercloud_instance_types
	zenlayercloud_security_groups
	zenlayercloud_instance_types
	zenlayercloud_disks
	zenlayercloud_subnets
	zenlayercloud_key_pairs

  Resource
	zenlayercloud_image
	zenlayercloud_security_group
	zenlayercloud_security_group_attachment
	zenlayercloud_security_group_rule
	zenlayercloud_instance
	zenlayercloud_disk
	zenlayercloud_disk_attachment
	zenlayercloud_subnet
	zenlayercloud_key_pair

Bare Metal Cloud(BMC)
  Data Source
	zenlayercloud_bmc_zones
	zenlayercloud_bmc_instance_types
    zenlayercloud_bmc_images
	zenlayercloud_bmc_instances
	zenlayercloud_bmc_eips
	zenlayercloud_bmc_ddos_ips
	zenlayercloud_bmc_vpc_regions
	zenlayercloud_bmc_vpcs
	zenlayercloud_bmc_subnets

  Resource
	zenlayercloud_bmc_instance
	zenlayercloud_bmc_ddos_ip
	zenlayercloud_bmc_ddos_ip_association
	zenlayercloud_bmc_eip
	zenlayercloud_bmc_eip_association
	zenlayercloud_bmc_vpc
	zenlayercloud_bmc_subnet

Cloud Networking(SDN)
  Data Source
	zenlayercloud_sdn_datacenters
	zenlayercloud_sdn_ports
	zenlayercloud_sdn_private_connects
	zenlayercloud_sdn_cloud_regions
  Resource
	zenlayercloud_sdn_port
	zenlayercloud_sdn_private_connect

Zenlayer Global Accelerator(ZGA)

  Data Source
 	zenlayercloud_zga_certificates
	zenlayercloud_zga_origin_regions
	zenlayercloud_zga_accelerate_regions
	zenlayercloud_zga_accelerators

  Resource
	zenlayercloud_zga_certificate
	zenlayercloud_zga_accelerator

Zenlayer Elastic Compute(ZEC)

  Data Source
	zenlayercloud_zec_vpcs
	zenlayercloud_zec_subnets
	zenlayercloud_zec_images
	zenlayercloud_zec_disks
	zenlayercloud_zec_disk_snapshots
	zenlayercloud_zec_disk_snapshot_policies
	zenlayercloud_zec_cidrs
	zenlayercloud_zec_eips
	zenlayercloud_zec_instances
	zenlayercloud_zec_vnics
	zenlayercloud_zec_nat_gateways
	zenlayercloud_zec_border_gateways

  Resource
	zenlayercloud_zec_vpc
	zenlayercloud_zec_security_group
	zenlayercloud_zec_security_group_rule_set
	zenlayercloud_zec_vpc_security_group_attachment
	zenlayercloud_zec_vpc_route
	zenlayercloud_zec_subnet
	zenlayercloud_zec_vnic
	zenlayercloud_zec_vnic_attachment
	zenlayercloud_zec_vnic_ipv4
	zenlayercloud_zec_cidr
	zenlayercloud_zec_eip
	zenlayercloud_zec_eip_association
	zenlayercloud_zec_instance
	zenlayercloud_zec_disk
	zenlayercloud_zec_disk_attachment
	zenlayercloud_zec_disk_snapshot
	zenlayercloud_zec_disk_snapshot_policy
	zenlayercloud_zec_disk_snapshot_policy_attachment

Zenlayer Load Balancing(ZLB)

  Data Source
	zenlayercloud_zlb_regions
	zenlayercloud_zlb_instances
	zenlayercloud_zlb_listeners
	zenlayercloud_zlb_backends

  Resource
	zenlayercloud_zlb_instance
	zenlayercloud_zlb_listener
	zenlayercloud_zlb_backend

Traffic

  Data Source
	zenlayercloud_traffic_bandwidth_cluster_areas
	zenlayercloud_traffic_bandwidth_clusters

  Resource
*/
package zenlayercloud

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
	"context"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/services/traffic"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/services/zec"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/services/zlb"

	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
)

const (
	PROVIDER_SECRET_KEY_ID       = "ZENLAYERCLOUD_ACCESS_KEY_ID"
	PROVIDER_SECRET_KEY_PASSWORD = "ZENLAYERCLOUD_ACCESS_KEY_PASSWORD"
	PROVIDER_CLIENT_TIMEOUT      = "ZENLAYERCLOUD_CLIENT_TIMEOUT"
	PROVIDER_SCHEME              = "ZENLAYERCLOUD_SCHEME"
	PROVIDER_DOMAIN              = "ZENLAYERCLOUD_DOMAIN"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"access_key_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(PROVIDER_SECRET_KEY_ID, os.Getenv(PROVIDER_SECRET_KEY_ID)),
				Description: "Access Key Id",
			},
			"access_key_password": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(PROVIDER_SECRET_KEY_PASSWORD, os.Getenv(PROVIDER_SECRET_KEY_PASSWORD)),
				Description: "Access Key Password",
			},
			"domain": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(PROVIDER_DOMAIN, nil),
				Description: "The root domain of the API request, Default is `console.zenlayer.com`.",
			},
			"scheme": {
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc(PROVIDER_SCHEME, "HTTPS"),
				ValidateFunc: validation.StringInSlice([]string{"HTTP", "HTTPS"}, false),
				Description:  "The scheme of the API request. Valid values: `HTTP` and `HTTPS`. Default is `HTTPS`.",
			},
			"client_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(PROVIDER_CLIENT_TIMEOUT, 600),
				Description: "The maximum timeout of the client request.",
			},
		},
		DataSourcesMap:       dataSourcesMap(),
		ResourcesMap:         resourcesMap(),
		ConfigureContextFunc: providerConfigure,
	}
}

func resourcesMap() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		// bmc product
		"zenlayercloud_bmc_instance":            resourceZenlayerCloudInstance(),
		"zenlayercloud_bmc_eip":                 resourceZenlayerCloudEip(),
		"zenlayercloud_bmc_eip_association":     resourceZenlayerCloudEipAssociationAssociation(),
		"zenlayercloud_bmc_ddos_ip":             resourceZenlayerCloudDDosIp(),
		"zenlayercloud_bmc_ddos_ip_association": resourceZenlayerCloudDdosIpAssociationAssociation(),
		"zenlayercloud_bmc_vpc":                 resourceZenlayerCloudVpc(),
		"zenlayercloud_bmc_subnet":              resourceZenlayerCloudBmcSubnet(),

		// vm product
		"zenlayercloud_image":                     resourceZenlayerCloudVmImage(),
		"zenlayercloud_instance":                  resourceZenlayerCloudVmInstance(),
		"zenlayercloud_security_group":            resourceZenlayerCloudSecurityGroup(),
		"zenlayercloud_security_group_attachment": resourceZenlayerCloudSecurityGroupAttachment(),
		"zenlayercloud_security_group_rule":       resourceZenlayerCloudSecurityGroupRule(),
		"zenlayercloud_subnet":                    resourceZenlayerCloudSubnet(),
		"zenlayercloud_disk":                      resourceZenlayerCloudVmDisk(),
		"zenlayercloud_disk_attachment":           resourceZenlayerCloudVmDiskAttachment(),
		"zenlayercloud_key_pair":                  resourceZenlayerCloudKeyPair(),

		// cloud networking product
		"zenlayercloud_sdn_port":            resourceZenlayerCloudDcPorts(),
		"zenlayercloud_sdn_private_connect": resourceZenlayerCloudPrivateConnect(),

		// zenlayer global accelerator
		"zenlayercloud_zga_certificate": resourceZenlayerCloudCertificate(),
		"zenlayercloud_zga_accelerator": resourceZenlayerCloudAccelerator(),

		// zenlayer zec product
		"zenlayercloud_zec_vpc":                           zec.ResourceZenlayerCloudGlobalVpc(),
		"zenlayercloud_zec_security_group":           	   zec.ResourceZenlayerCloudZecSecurityGroup(),
		"zenlayercloud_zec_security_group_rule_set":       zec.ResourceZenlayerCloudZecSecurityGroupRuleSet(),
		"zenlayercloud_zec_vpc_security_group_attachment": zec.ResourceZenlayerCloudZecVpcSecurityGroupAttachment(),
		"zenlayercloud_zec_vpc_route": 					   zec.ResourceZenlayerCloudGlobalVpcRoute(),
		"zenlayercloud_zec_subnet":                        zec.ResourceZenlayerCloudZecSubnet(),
		"zenlayercloud_zec_vnic":                          zec.ResourceZenlayerCloudZecVNic(),
		"zenlayercloud_zec_vnic_attachment":               zec.ResourceZenlayerCloudZecVNicAttachment(),
		"zenlayercloud_zec_vnic_ipv4":                     zec.ResourceZenlayerCloudZecVNicIPv4(),
		"zenlayercloud_zec_instance":                      zec.ResourceZenlayerCloudZecInstance(),
		"zenlayercloud_zec_cidr": 						   zec.ResourceZenlayerCloudZecCidr(),
		"zenlayercloud_zec_eip":                           zec.ResourceZenlayerCloudZecElasticIP(),
		"zenlayercloud_zec_eip_association":               zec.ResourceZenlayerCloudEipAssociation(),
		"zenlayercloud_zec_disk":                          zec.ResourceZenlayerCloudZecDisk(),
		"zenlayercloud_zec_disk_attachment":               zec.ResourceZenlayerCloudZecDiskAttachment(),
		"zenlayercloud_zec_disk_snapshot":                 zec.ResourceZenlayerCloudZecSnapshot(),
		"zenlayercloud_zec_disk_snapshot_policy":          zec.ResourceZenlayerCloudZecSnapshotPolicy(),
		"zenlayercloud_zec_disk_snapshot_policy_attachment": zec.ResourceZenlayerCloudZecSnapshotPolicyAttachment(),
		"zenlayercloud_zec_border_gateway":                zec.ResourceZenlayerCloudBorderGateway(),
		"zenlayercloud_zec_border_gateway_association":    zec.ResourceZenlayerCloudBorderGatewayAssociation(),
		"zenlayercloud_zec_nat_gateway":                   zec.ResourceZenlayerCloudZecVpcNatGateway(),

		// zenlayer load balancer
		"zenlayercloud_zlb_instance": zlb.ResourceZenlayerCloudZlbInstance(),
		"zenlayercloud_zlb_listener": zlb.ResourceZenlayerCloudZlbListener(),
		"zenlayercloud_zlb_backend":  zlb.ResourceZenlayerCloudZlbBackend(),
	}
}

func dataSourcesMap() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		// bmc product
		"zenlayercloud_bmc_zones":          dataSourceZenlayerCloudBmcZones(),
		"zenlayercloud_bmc_instance_types": dataSourceZenlayerCloudInstanceTypes(),
		"zenlayercloud_bmc_images":         dataSourceZenlayerCloudImages(),
		"zenlayercloud_bmc_instances":      dataSourceZenlayerCloudInstances(),
		"zenlayercloud_bmc_eips":           dataSourceZenlayerCloudEips(),
		"zenlayercloud_bmc_ddos_ips":       dataSourceZenlayerCloudDdosIps(),
		"zenlayercloud_bmc_vpc_regions":    dataSourceZenlayerCloudVpcRegions(),
		"zenlayercloud_bmc_vpcs":           dataSourceZenlayerCloudVpcs(),
		"zenlayercloud_bmc_subnets":        dataSourceZenlayerCloudVpcSubnets(),

		// vm product
		"zenlayercloud_security_groups": dataSourceZenlayerCloudSecurityGroups(),
		"zenlayercloud_zones":           dataSourceZenlayerCloudZones(),
		"zenlayercloud_images":          dataSourceZenlayerCloudVmImages(),
		"zenlayercloud_instance_types":  dataSourceZenlayerCloudVmInstanceTypes(),
		"zenlayercloud_disks":           dataSourceZenlayerCloudDisks(),
		"zenlayercloud_subnets":         dataSourceZenlayerCloudSubnets(),
		"zenlayercloud_key_pairs":       dataSourceZenlayerCloudKeyPairs(),

		// cloud networking product
		"zenlayercloud_sdn_datacenters":      dataSourceZenlayerCloudSdnDatacenters(),
		"zenlayercloud_sdn_ports":            dataSourceZenlayerCloudDcPorts(),
		"zenlayercloud_sdn_private_connects": dataSourceZenlayerCloudSdnPrivateConnects(),
		//"zenlayercloud_sdn_cloud_routers":    dataSourceZenlayerCloudSdnCloudRouters(),
		"zenlayercloud_sdn_cloud_regions": dataSourceZenlayerCloudCloudRegions(),

		// zenlayer global accelerator
		"zenlayercloud_zga_certificates":       dataSourceZenlayerCloudZgaCertificates(),
		"zenlayercloud_zga_origin_regions":     dataSourceZenlayerCloudZgaOriginRegions(),
		"zenlayercloud_zga_accelerate_regions": dataSourceZenlayerCloudZgaAccelerateRegions(),
		"zenlayercloud_zga_accelerators":       dataSourceZenlayerCloudZgaAccelerators(),

		// zenlayer zec product
		"zenlayercloud_zec_images":          zec.DataSourceZenlayerCloudZecImages(),
		"zenlayercloud_zec_vpcs":            zec.DataSourceZenlayerCloudZecVpcs(),
		"zenlayercloud_zec_subnets":         zec.DataSourceZenlayerCloudZecSubnets(),
		"zenlayercloud_zec_border_gateways": zec.DataSourceZenlayerCloudBorderGateways(),
		"zenlayercloud_zec_cidrs": 			 zec.DataSourceZenlayerCloudCidrs(),
		"zenlayercloud_zec_eips":            zec.DataSourceZenlayerCloudEips(),
		"zenlayercloud_zec_disks":           zec.DataSourceZenlayerCloudZecDisks(),
		"zenlayercloud_zec_disk_snapshots":  zec.DataSourceZenlayerCloudZecSnapshots(),
		"zenlayercloud_zec_disk_snapshot_policies":      zec.DataSourceZenlayerCloudZecAutoSnapshotPolicies(),
		"zenlayercloud_zec_nat_gateways":    zec.DataSourceZenlayerCloudZecNatGateway(),
		"zenlayercloud_zec_instances":       zec.DataSourceZenlayerCloudZecInstances(),
		"zenlayercloud_zec_vnics":           zec.DataSourceZenlayerCloudZecVnics(),

		// zenlayer load balancer
		"zenlayercloud_zlb_regions":   zlb.DataSourceZenlayerCloudZlbRegions(),
		"zenlayercloud_zlb_instances": zlb.DataSourceZenlayerCloudZlbInstances(),
		"zenlayercloud_zlb_listeners": zlb.DataSourceZenlayerCloudZlbListeners(),
		"zenlayercloud_zlb_backends":  zlb.DataSourceZenlayerCloudZlbBackends(),

		// zenlayer traffic
		"zenlayercloud_traffic_bandwidth_cluster_areas": traffic.DataSourceZenlayerCloudTrafficBandwidthClusterAreas(),
		"zenlayercloud_traffic_bandwidth_clusters": traffic.DataSourceZenlayerCloudTrafficBandwidthClusters(),

	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (client interface{}, diags diag.Diagnostics) {
	accessKeyId := d.Get("access_key_id").(string)
	accessKeyPassword := d.Get("access_key_password").(string)
	domain := d.Get("domain").(string)
	scheme := d.Get("scheme").(string)
	clientTimeout := d.Get("client_timeout").(int)

	if (accessKeyId != "") && (accessKeyPassword != "") {
		client = &connectivity.ZenlayerCloudClient{
			SecretKeyId:       strings.TrimSpace(accessKeyId),
			SecretKeyPassword: strings.TrimSpace(accessKeyPassword),
			Scheme:            scheme,
			Domain:            domain,
			Timeout:           clientTimeout,
		}
	} else {
		diags = append(diags, diag.Diagnostic{
			Summary: "Missing Credential Value",
			Detail:  "access_key_id or access_key_password is missing.",
		})
		return nil, diags
	}
	return
}

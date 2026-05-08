/*
Use this data source to query the public IPv6 attached to a single vNIC.

Example Usage

```hcl
data "zenlayercloud_zec_vnic_public_ipv6" "demo" {
  nic_id = "1680855999352675875"
}

output "ipv6_rate_limit_mode" {
  value = data.zenlayercloud_zec_vnic_public_ipv6.demo.rate_limit_mode
}
```
*/
package zec

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	zec2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"

	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
)

func DataSourceZenlayerCloudZecVNicPublicIPv6() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudZecVNicPublicIPv6Read,
		Schema: map[string]*schema.Schema{
			"nic_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the vNIC whose public IPv6 should be queried.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			"ipv6_cidr_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The IPv6 CIDR ID associated with this public IPv6.",
			},
			"ipv6_cidr": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The IPv6 CIDR address.",
			},
			"primary_ipv6_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The primary IPv6 address of the vNIC.",
			},
			"internet_charge_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Internet charge type of the public IPv6.",
			},
			"bandwidth": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Public bandwidth limit of the IPv6, measured in Mbps.",
			},
			"rate_limit_mode": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Bandwidth rate limit mode. `LOOSE` or `STRICT`.",
			},
			"traffic_package_size": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "Traffic package size of the IPv6, measured in TB.",
			},
			"bandwidth_cluster_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the associated shared bandwidth cluster, if any.",
			},
			"bandwidth_cluster_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the associated shared bandwidth cluster, if any.",
			},
		},
	}
}

func dataSourceZenlayerCloudZecVNicPublicIPv6Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "data_source.zenlayercloud_zec_vnic_public_ipv6.read")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	nicId := d.Get("nic_id").(string)

	var address *zec2.PublicIpv6CidrAddress
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		address, e = zecService.DescribeNetworkInterfacePublicIPv6ByNicId(ctx, nicId)
		if e != nil {
			return common.RetryError(ctx, e, common.InternalServerError)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	if address == nil || address.Ipv6CidrId == nil {
		return diag.FromErr(fmt.Errorf("public IPv6 not found on vNIC %s", nicId))
	}

	d.SetId(nicId)
	_ = d.Set("ipv6_cidr_id", address.Ipv6CidrId)
	_ = d.Set("ipv6_cidr", address.Ipv6Cidr)
	_ = d.Set("primary_ipv6_address", address.PrimaryIpv6Address)
	_ = d.Set("internet_charge_type", address.InternetChargeType)
	_ = d.Set("bandwidth", address.Bandwidth)
	_ = d.Set("rate_limit_mode", address.RateLimitMode)
	_ = d.Set("traffic_package_size", address.TrafficPackageSize)
	if address.BandwidthCluster != nil {
		_ = d.Set("bandwidth_cluster_id", address.BandwidthCluster.BandwidthClusterId)
		_ = d.Set("bandwidth_cluster_name", address.BandwidthCluster.BandwidthClusterName)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		dump := map[string]interface{}{
			"nic_id":               nicId,
			"ipv6_cidr_id":         address.Ipv6CidrId,
			"ipv6_cidr":            address.Ipv6Cidr,
			"primary_ipv6_address": address.PrimaryIpv6Address,
			"internet_charge_type": address.InternetChargeType,
			"bandwidth":            address.Bandwidth,
			"rate_limit_mode":      address.RateLimitMode,
			"traffic_package_size": address.TrafficPackageSize,
		}
		if address.BandwidthCluster != nil {
			dump["bandwidth_cluster_id"] = address.BandwidthCluster.BandwidthClusterId
			dump["bandwidth_cluster_name"] = address.BandwidthCluster.BandwidthClusterName
		}
		if err := common.WriteToFile(output.(string), dump); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

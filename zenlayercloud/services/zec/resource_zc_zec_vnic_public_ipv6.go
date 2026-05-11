/*
Provides a resource to manage the rate limit mode of an existing public IPv6 address on a vNIC.

This resource does NOT create or delete the public IPv6 address itself; it only adopts
an existing public IPv6 (already attached to a vNIC) into Terraform state and allows
updates to its `rate_limit_mode`. Destroying this resource only removes it from state.

Example Usage

```hcl
resource "zenlayercloud_zec_vnic_public_ipv6" "demo" {
  nic_id          = "xxxxxxxxxxxxxxxxx"
  rate_limit_mode = "LOOSE"
}
```

Import

vNIC public IPv6 rate limit mode can be imported using the vNIC ID, e.g.

```
$ terraform import zenlayercloud_zec_vnic_public_ipv6.demo nic-id
```
*/
package zec

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
)

func ResourceZenlayerCloudZecVNicPublicIPv6() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZecVNicPublicIPv6Create,
		ReadContext:   resourceZenlayerCloudZecVNicPublicIPv6Read,
		UpdateContext: resourceZenlayerCloudZecVNicPublicIPv6Update,
		DeleteContext: resourceZenlayerCloudZecVNicPublicIPv6Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"nic_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the vNIC. The vNIC must already have a public IPv6 attached; this resource does not create it.",
			},
			"rate_limit_mode": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{ZecEipRateLimitModeLoose, ZecEipRateLimitModeStrict}, false),
				Description:  "Bandwidth rate limit mode. Valid values: `LOOSE`, `STRICT`. Only takes effect on public IPv6 with a fixed bandwidth.",
			},
			"ipv6_cidr_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The IPv6 CIDR ID associated with the public IPv6.",
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
			"traffic_package_size": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "Traffic package size of the IPv6, measured in TB.",
			},
		},
	}
}

func resourceZenlayerCloudZecVNicPublicIPv6Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_vnic_public_ipv6.create")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	nicId := d.Get("nic_id").(string)

	address, err := zecService.DescribeNetworkInterfacePublicIPv6ByNicId(ctx, nicId)
	if err != nil {
		return diag.FromErr(err)
	}
	if address == nil || address.Ipv6CidrId == nil {
		return diag.FromErr(fmt.Errorf("public IPv6 not found on vNIC %s", nicId))
	}

	d.SetId(nicId)

	desired := d.Get("rate_limit_mode").(string)
	if address.RateLimitMode == nil || *address.RateLimitMode != desired {
		if err := zecService.ModifyNetworkInterfacePublicIPv6BandwidthLimitMode(ctx, *address.Ipv6CidrId, desired); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceZenlayerCloudZecVNicPublicIPv6Read(ctx, d, meta)
}

func resourceZenlayerCloudZecVNicPublicIPv6Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_vnic_public_ipv6.read")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	nicId := d.Id()
	address, err := zecService.DescribeNetworkInterfacePublicIPv6ByNicId(ctx, nicId)
	if err != nil {
		return diag.FromErr(err)
	}
	if address == nil {
		d.SetId("")
		return nil
	}

	_ = d.Set("nic_id", nicId)
	_ = d.Set("rate_limit_mode", address.RateLimitMode)
	_ = d.Set("ipv6_cidr_id", address.Ipv6CidrId)
	_ = d.Set("ipv6_cidr", address.Ipv6Cidr)
	_ = d.Set("primary_ipv6_address", address.PrimaryIpv6Address)
	_ = d.Set("internet_charge_type", address.InternetChargeType)
	_ = d.Set("bandwidth", address.Bandwidth)
	_ = d.Set("traffic_package_size", address.TrafficPackageSize)

	return nil
}

func resourceZenlayerCloudZecVNicPublicIPv6Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_vnic_public_ipv6.update")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	if d.HasChange("rate_limit_mode") {
		address, err := zecService.DescribeNetworkInterfacePublicIPv6ByNicId(ctx, d.Id())
		if err != nil {
			return diag.FromErr(err)
		}
		if address == nil || address.Ipv6CidrId == nil {
			return diag.FromErr(fmt.Errorf("public IPv6 not found on vNIC %s", d.Id()))
		}
		if err := zecService.ModifyNetworkInterfacePublicIPv6BandwidthLimitMode(ctx, *address.Ipv6CidrId, d.Get("rate_limit_mode").(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceZenlayerCloudZecVNicPublicIPv6Read(ctx, d, meta)
}

func resourceZenlayerCloudZecVNicPublicIPv6Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_vnic_public_ipv6.delete")()
	return nil
}

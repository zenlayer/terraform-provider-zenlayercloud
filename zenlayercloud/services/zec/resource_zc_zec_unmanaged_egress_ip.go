/*
Provides a resource to manage the rate limit mode of an existing unmanaged egress IP.

This resource does NOT create or delete the unmanaged egress IP itself; it only adopts
an existing unmanaged egress IP into Terraform state and allows updates to its
`rate_limit_mode`. Destroying this resource only removes it from state.

Example Usage

```hcl
resource "zenlayercloud_zec_unmanaged_egress_ip" "demo" {
  unmanaged_egress_ip_id = "xxxxxxxxxxxxxxxxx"
  rate_limit_mode        = "LOOSE"
}
```

Import

Unmanaged egress IP rate limit mode can be imported, e.g.

```
$ terraform import zenlayercloud_zec_unmanaged_egress_ip.demo unmanaged-egress-ip-id
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

func ResourceZenlayerCloudZecUnmanagedEgressIp() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZecUnmanagedEgressIpCreate,
		ReadContext:   resourceZenlayerCloudZecUnmanagedEgressIpRead,
		UpdateContext: resourceZenlayerCloudZecUnmanagedEgressIpUpdate,
		DeleteContext: resourceZenlayerCloudZecUnmanagedEgressIpDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"unmanaged_egress_ip_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the unmanaged egress IP. The IP must already exist; this resource does not create it.",
			},
			"rate_limit_mode": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{ZecEipRateLimitModeLoose, ZecEipRateLimitModeStrict}, false),
				Description:  "Bandwidth rate limit mode. Valid values: `LOOSE`, `STRICT`.",
			},
			"ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The public IP address.",
			},
			"region_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The region ID that the unmanaged egress IP locates at.",
			},
			"vpc_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the VPC that the unmanaged egress IP belongs to.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the unmanaged egress IP.",
			},
			"network_line_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Network line type.",
			},
			"internet_charge_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Internet charge type.",
			},
			"bandwidth_cap": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Bandwidth cap, measured in Mbps. Null if there is no fixed bandwidth.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation time of the unmanaged egress IP.",
			},
		},
	}
}

func resourceZenlayerCloudZecUnmanagedEgressIpCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_unmanaged_egress_ip.create")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	id := d.Get("unmanaged_egress_ip_id").(string)

	ipInfo, err := zecService.DescribeUnmanagedEgressIpById(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}
	if ipInfo == nil {
		return diag.FromErr(fmt.Errorf("unmanaged egress IP %s not found", id))
	}

	d.SetId(id)

	desired := d.Get("rate_limit_mode").(string)
	if ipInfo.RateLimitMode == nil || *ipInfo.RateLimitMode != desired {
		if err := zecService.ModifyUnmanagedEgressIpBandwidthLimitMode(ctx, id, desired); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceZenlayerCloudZecUnmanagedEgressIpRead(ctx, d, meta)
}

func resourceZenlayerCloudZecUnmanagedEgressIpRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_unmanaged_egress_ip.read")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	id := d.Id()
	ipInfo, err := zecService.DescribeUnmanagedEgressIpById(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}
	if ipInfo == nil {
		d.SetId("")
		return nil
	}

	_ = d.Set("unmanaged_egress_ip_id", ipInfo.UnmanagedEgressIpId)
	_ = d.Set("rate_limit_mode", ipInfo.RateLimitMode)
	_ = d.Set("ip", ipInfo.Ip)
	_ = d.Set("region_id", ipInfo.RegionId)
	_ = d.Set("vpc_id", ipInfo.VpcId)
	_ = d.Set("status", ipInfo.Status)
	_ = d.Set("network_line_type", ipInfo.NetworkLineType)
	_ = d.Set("internet_charge_type", ipInfo.InternetChargeType)
	_ = d.Set("bandwidth_cap", ipInfo.BandwidthCap)
	_ = d.Set("create_time", ipInfo.CreateTime)

	return nil
}

func resourceZenlayerCloudZecUnmanagedEgressIpUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_unmanaged_egress_ip.update")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	if d.HasChange("rate_limit_mode") {
		if err := zecService.ModifyUnmanagedEgressIpBandwidthLimitMode(ctx, d.Id(), d.Get("rate_limit_mode").(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceZenlayerCloudZecUnmanagedEgressIpRead(ctx, d, meta)
}

func resourceZenlayerCloudZecUnmanagedEgressIpDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_unmanaged_egress_ip.delete")()
	return nil
}

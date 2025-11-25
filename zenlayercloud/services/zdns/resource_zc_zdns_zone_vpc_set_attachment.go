package zdns

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	pvtdns "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zdns20251101"
	"time"
)

func ResourceZenlayerCloudPvtdnsZoneVpcAttachment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudPvtdnsZoneVpcAttachmentCreate,
		ReadContext:   resourceZenlayerCloudPvtdnsZoneVpcAttachmentRead,
		UpdateContext: resourceZenlayerCloudPvtdnsZoneVpcAttachmentUpdate,
		DeleteContext: resourceZenlayerCloudPvtdnsZoneVpcAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the private zone.",
			},
			"vpc_ids": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "The IDs of the VPCs to be attached to the private zone.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceZenlayerCloudPvtdnsZoneVpcAttachmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	defer common2.LogElapsed(ctx, "resource.zenlayercloud_pvtdns_zone_vpc_set_attachment.update")()

	pvtDnsService := ZdnsService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	zoneId := d.Id()

	if d.HasChange("vpc_ids") {

		var zone *pvtdns.PrivateZone
		var errRet error

		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
			zone, errRet = pvtDnsService.DescribePrivateZoneById(ctx, zoneId)
			if errRet != nil {
				return common2.RetryError(ctx, errRet)
			}

			return nil
		})

		if err != nil {
			return diag.FromErr(err)
		}

		oldVpcs := getIdSet(zone.VpcIds)
		newVpcs := getIdSet(common2.ToStringList(d.Get("vpc_ids").(*schema.Set).List()))

		addSet := newVpcs.Difference(oldVpcs)
		removeSet := oldVpcs.Difference(newVpcs)

		if addSet.Len() > 0 {
			request := pvtdns.NewBindPrivateZoneVpcRequest()
			request.ZoneId = common.String(zoneId)
			request.VpcIds = common2.ToStringList(addSet.List())

			err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
				response, err := pvtDnsService.client.WithZDnsClient().BindPrivateZoneVpc(request)
				if err != nil {
					tflog.Info(ctx, "Fail to bind vpcs to zone.", map[string]interface{}{
						"action":  request.GetAction(),
						"request": common2.ToJsonString(request),
						"err":     err.Error(),
					})
					return common2.RetryError(ctx, err)
				}

				tflog.Info(ctx, "Bind vpcs to zone success", map[string]interface{}{
					"action":   request.GetAction(),
					"request":  common2.ToJsonString(request),
					"response": common2.ToJsonString(response),
				})

				return nil
			})
			if err != nil {
				return diag.FromErr(err)
			}
		}

		if removeSet.Len() > 0 {
			request := pvtdns.NewUnbindPrivateZoneVpcRequest()
			request.ZoneId = common.String(zoneId)
			request.VpcIds = common2.ToStringList(removeSet.List())

			err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
				_, errRet := pvtDnsService.client.WithZDnsClient().UnbindPrivateZoneVpc(request)
				if errRet != nil {
					return common2.RetryError(ctx, errRet)
				}
				return nil
			})
			if err != nil {
				return diag.FromErr(err)
			}

		}
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceZenlayerCloudPvtdnsZoneRead(ctx, d, meta)
}

func getIdSet(items []string) *schema.Set {
	rmId := make([]interface{}, 0)
	for _, item := range items {
		rmId = append(rmId, item)
	}
	return schema.NewSet(schema.HashString, rmId)
}

func resourceZenlayerCloudPvtdnsZoneVpcAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_pvtdns_zone_vpc_set_attachment.delete")()

	pvtDnsService := ZdnsService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	zoneId := d.Id()

	var zone *pvtdns.PrivateZone
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		zone, errRet = pvtDnsService.DescribePrivateZoneById(ctx, zoneId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}
	if len(zone.VpcIds) == 0 {
		return nil
	}
	request := pvtdns.NewUnbindPrivateZoneVpcRequest()
	request.ZoneId = common.String(zoneId)
	request.VpcIds = zone.VpcIds

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		_, errRet := pvtDnsService.client.WithZDnsClient().UnbindPrivateZoneVpc(request)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceZenlayerCloudPvtdnsZoneVpcAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	defer common2.LogElapsed(ctx, "resource.zenlayercloud_pvtdns_zone_vpc_set_attachment.create")()

	pvtDnsService := ZdnsService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	zoneId := d.Get("zone_id").(string)
	vpcIds := d.Get("vpc_ids").(*schema.Set).List()

	request := pvtdns.NewBindPrivateZoneVpcRequest()
	request.ZoneId = common.String(zoneId)
	request.VpcIds = common2.ToStringList(vpcIds)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := pvtDnsService.client.WithZDnsClient().BindPrivateZoneVpc(request)
		if err != nil {
			tflog.Info(ctx, "Fail to bind vpcs to zone.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": common2.ToJsonString(request),
				"err":     err.Error(),
			})
			return common2.RetryError(ctx, err)
		}

		tflog.Info(ctx, "Bind vpcs to zone success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  common2.ToJsonString(request),
			"response": common2.ToJsonString(response),
		})

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(zoneId)

	return resourceZenlayerCloudPvtdnsZoneVpcAttachmentRead(ctx, d, meta)
}

func resourceZenlayerCloudPvtdnsZoneVpcAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_pvtdns_zone_vpc_set_attachment.read")()

	var diags diag.Diagnostics

	zoneId := d.Id()

	pvtDnsService := ZdnsService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var zone *pvtdns.PrivateZone
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		zone, errRet = pvtDnsService.DescribePrivateZoneById(ctx, zoneId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if zone == nil {
		d.SetId("")
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Private DNS record doesn't exist",
			Detail:   fmt.Sprintf("The private DNS record %s is not exist", zoneId),
		})
		return diags
	}

	_ = d.Set("zone_id", zoneId)
	_ = d.Set("vpc_ids", zone.VpcIds)

	return diags
}

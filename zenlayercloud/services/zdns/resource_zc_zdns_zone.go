package zdns

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	user "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/user20240529"
	pvtdns "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zdns20251101"
	"time"
)

func ResourceZenlayerCloudPvtdnsZone() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudPvtdnsZoneCreate,
		ReadContext:   resourceZenlayerCloudPvtdnsZoneRead,
		UpdateContext: resourceZenlayerCloudPvtdnsZoneUpdate,
		DeleteContext: resourceZenlayerCloudPvtdnsZoneDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"zone_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the private zone.",
			},
			"proxy_pattern": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"ZONE", "RECURSION"}, false),
				Description: "The recursive DNS proxy setting for subdomains. Default: `ZONE`. Valid values: \n\t- `ZONE`: Disable recursive DNS proxy. When resolving non-existent subdomains under this domain, it directly returns NXDOMAIN, indicating the subdomain does not exist. \n\t- `RECURSION`: Enable recursive DNS proxy. When resolving non-existent subdomains under this domain, it queries the recursive module and responds to the resolution request with the final query result."			},
			"remark": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Remarks.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The resource group id the private zone belongs to, default to Default Resource Group.",
			},
			"resource_group_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The resource group name the private zone belongs to, default to Default Resource Group.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Create time of the private zone.",
			},
		},
	}
}

func resourceZenlayerCloudPvtdnsZoneDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_pvtdns_zone.delete")()

	zoneId := d.Id()

	pvtDnsService := ZdnsService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := pvtDnsService.DeletePrivateZoneById(ctx, zoneId)
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

func resourceZenlayerCloudPvtdnsZoneUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	defer common2.LogElapsed(ctx, "resource.zenlayercloud_pvtdns_zone.update")()

	pvtDnsService := ZdnsService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	zoneId := d.Id()

	if d.HasChanges("remark", "proxy_pattern") {

		request := pvtdns.NewModifyPrivateZoneRequest()
		request.ZoneId = common.String(zoneId)
		request.Remark = common.String(d.Get("remark").(string))
		request.ProxyPattern = common.String(d.Get("proxy_pattern").(string))

		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {

			_, err := pvtDnsService.client.WithZDnsClient().ModifyPrivateZone(request)
			if err != nil {
				return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("resource_group_id") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			request := user.NewAddResourceResourceGroupRequest()
			request.ResourceGroupId = common.String(d.Get("resource_group_id").(string))
			request.Resources = []*string{common.String(zoneId)}

			_, err := meta.(*connectivity.ZenlayerCloudClient).WithUsrClient().AddResourceResourceGroup(request)
			if err != nil {
				return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceZenlayerCloudPvtdnsZoneRead(ctx, d, meta)
}

func resourceZenlayerCloudPvtdnsZoneCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_pvtdns_zone.create")()

	pvtDnsService := ZdnsService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	request := pvtdns.NewAddPrivateZoneRequest()
	request.ZoneName = common.String(d.Get("zone_name").(string))
	request.ProxyPattern = common.String(d.Get("proxy_pattern").(string))

	if v, ok := d.GetOk("remark"); ok {
		request.Remark = common.String(v.(string))
	}

	if v, ok := d.GetOk("resource_group_id"); ok {
		request.ResourceGroupId = common.String(v.(string))
	}

	zoneId := ""

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := pvtDnsService.client.WithZDnsClient().AddPrivateZone(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create private zone.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": common2.ToJsonString(request),
				"err":     err.Error(),
			})
			return common2.RetryError(ctx, err)
		}

		tflog.Info(ctx, "Create private zone gateway success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  common2.ToJsonString(request),
			"response": common2.ToJsonString(response),
		})

		if response.Response.ZoneId == nil {
			err = fmt.Errorf("private zone id is nil")
			return resource.NonRetryableError(err)
		}
		zoneId = *response.Response.ZoneId

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(zoneId)

	return resourceZenlayerCloudPvtdnsZoneRead(ctx, d, meta)
}

func resourceZenlayerCloudPvtdnsZoneRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_pvtdns_zone.read")()

	var diags diag.Diagnostics

	zoneId := d.Id()

	pvtDnsService := ZdnsService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var pz *pvtdns.PrivateZone
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		pz, errRet = pvtDnsService.DescribePrivateZoneById(ctx, zoneId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if pz == nil {
		d.SetId("")
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Private zone doesn't not exist",
			Detail:   fmt.Sprintf("The private zone %s is not exist", zoneId),
		})
		return diags
	}

	_ = d.Set("zone_name", pz.ZoneName)
	_ = d.Set("remark", pz.Remark)
	_ = d.Set("resource_group_id", pz.ResourceGroup.ResourceGroupId)
	_ = d.Set("resource_group_name", pz.ResourceGroup.ResourceGroupName)
	_ = d.Set("create_time", pz.CreateTime)
	_ = d.Set("proxy_pattern", pz.ProxyPattern)

	return diags
}

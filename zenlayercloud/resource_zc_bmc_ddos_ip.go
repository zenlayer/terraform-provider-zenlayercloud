/*
Provides an DDoS IP resource.

Example Usage

```hcl
variable "availability_zone" {
  default = "SEL-A"
}

resource "zenlayercloud_bmc_ddos_ip" "foo" {
  availability_zone = var.availability_zone
}
```

Import

EIP can be imported using the id, e.g.

```
$ terraform import zenlayercloud_bmc_ddos_ip.foo 123123xxxx
```
*/
package zenlayercloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	bmc "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/bmc20221120"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	"time"
)

func resourceZenlayerCloudDDosIp() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudDdosIpCreate,
		ReadContext:   resourceZenlayerCloudDdosIpRead,
		UpdateContext: resourceZenlayerCloudDdosIpUpdate,
		DeleteContext: resourceZenlayerCloudDdosIpDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of zone that the DDoS IP locates at.",
			},
			"charge_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "POSTPAID",
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(BmcChargeTypes, false),
				Description:  "The charge type of DDoS IP. Valid values are `PREPAID`, `POSTPAID`. The default is `POSTPAID`. Note: `PREPAID` DDoS IP may not allow to delete before expired.",
			},
			"charge_prepaid_period": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(1),
				Description:  "The tenancy (time unit is month) of the prepaid DDoS IP, NOTE: it only works when DDoS charge_type is set to `PREPAID`.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The resource group id the DDoS IP belongs to, default to Default Resource Group.",
			},
			"resource_group_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The resource group name the DDoS IP belongs to, default to Default Resource Group.",
			},
			"force_delete": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Indicate whether to force delete the DDoS IP. Default is `false`. If set true, the DDoS IP will be permanently deleted instead of being moved into the recycle bin.",
			},
			"public_ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The DDoS IP address.",
			},
			"ip_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Current status of the DDoS IP.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Create time of the DDoS IP.",
			},
			"expired_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Expired time of the DDoS IP.",
			},
		},
	}
}

func resourceZenlayerCloudDdosIpDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ddosIpId := d.Id()

	forceDelete := d.Get("force_delete").(bool)

	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := bmcService.TerminateDDoSIpAddress(ctx, ddosIpId)
		if errRet != nil {
			return retryError(ctx, errRet)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	notExist := false

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		ddosIp, errRet := bmcService.DescribeDdosIpAddressById(ctx, ddosIpId)
		if errRet != nil {
			return retryError(ctx, errRet, InternalServerError)
		}
		if ddosIp == nil {
			notExist = true
			return nil
		}

		if ddosIp.DdosIpStatus == BmcEipStatusRecycle {
			return nil
		}

		if ddosIp.DdosIpStatus == BmcEipStatusRecycling {
			return resource.RetryableError(fmt.Errorf("DDoS IP (%s) is recycling", ddosIpId))
		}
		return resource.NonRetryableError(fmt.Errorf("ddos status is not recycle, current status %s", ddosIp.DdosIpStatus))
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if notExist || !forceDelete {
		return nil
	}

	tflog.Debug(ctx, "Releasing DDoS IP ...", map[string]interface{}{
		"ddosIp": ddosIpId,
	})

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := bmcService.ReleaseDDoSIpAddressById(ctx, ddosIpId)
		if errRet != nil {

			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if !ok {
				return retryError(ctx, errRet)
			}
			if ee.Code == "INVALID_DDOS_IP_NOT_FOUND" || ee.Code == ResourceNotFound {
				return nil
			}
			return retryError(ctx, errRet, InternalServerError)
		}
		return nil
	})

	return nil
}

func resourceZenlayerCloudDdosIpUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	ddosIpId := d.Id()

	if d.HasChange("resource_group_id") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			err := bmcService.ModifyDdosIpResourceGroup(ctx, ddosIpId, d.Get("resource_group_id").(string))
			if err != nil {
				return retryError(ctx, err, InternalServerError, common.NetworkError)
			}
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceZenlayerCloudDdosIpRead(ctx, d, meta)
}

func resourceZenlayerCloudDdosIpCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	request := bmc.NewAllocateDdosIpAddressesRequest()
	request.ZoneId = d.Get("availability_zone").(string)
	request.DdosIpChargeType = d.Get("charge_type").(string)

	if request.DdosIpChargeType == BmcChargeTypePrepaid {
		request.DdosIpChargePrepaid = &bmc.ChargePrepaid{}

		if period, ok := d.GetOk("charge_prepaid_period"); ok {
			request.DdosIpChargePrepaid.Period = period.(int)
		} else {
			diags = append(diags, diag.Diagnostic{
				Summary: "Missing required argument",
				Detail:  "charge_prepaid_period is missing on prepaid DDoS IP.",
			})
			return diags
		}
	}
	if v, ok := d.GetOk("resource_group_id"); ok {
		request.ResourceGroupId = v.(string)
	}

	ddosIpId := ""

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := meta.(*connectivity.ZenlayerCloudClient).WithBmcClient().AllocateDdosIpAddresses(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create DDoS IP address.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": toJsonString(request),
				"err":     err.Error(),
			})
			return retryError(ctx, err)
		}

		tflog.Info(ctx, "Create DDoS IP success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  toJsonString(request),
			"response": toJsonString(response),
		})

		if len(response.Response.DdosIdSet) < 1 {
			err = fmt.Errorf("DDoS IP id is nil")
			return resource.NonRetryableError(err)
		}
		ddosIpId = *response.Response.DdosIdSet[0]

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(ddosIpId)

	return resourceZenlayerCloudDdosIpRead(ctx, d, meta)
}

func resourceZenlayerCloudDdosIpRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	ddosIpId := d.Id()

	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var ddosIpAddress *bmc.DdosIpAddress
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		ddosIpAddress, errRet = bmcService.DescribeDdosIpAddressById(ctx, ddosIpId)
		if errRet != nil {
			return retryError(ctx, errRet)
		}

		if ddosIpAddress != nil && ipIsOperating(ddosIpAddress.DdosIpStatus) {
			return resource.RetryableError(fmt.Errorf("waiting for ddos %s operation", ddosIpAddress.DdosIpId))
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if ddosIpAddress == nil || ddosIpAddress.DdosIpStatus == BmcEipStatusCreateFailed ||
		ddosIpAddress.DdosIpStatus == BmcEipStatusRecycle {
		d.SetId("")
		tflog.Info(ctx, "DDoS IP not exist or created failed or recycled", map[string]interface{}{
			"ddosIpId": ddosIpId,
		})
		return nil
	}

	// DDoS IP info
	_ = d.Set("availability_zone", ddosIpAddress.ZoneId)
	_ = d.Set("resource_group_id", ddosIpAddress.ResourceGroupId)
	_ = d.Set("resource_group_name", ddosIpAddress.ResourceGroupName)
	_ = d.Set("charge_type", ddosIpAddress.DdosIpChargeType)
	_ = d.Set("public_ip", ddosIpAddress.IpAddress)
	if ddosIpAddress.DdosIpChargeType == BmcChargeTypePrepaid {
		_ = d.Set("charge_prepaid_period", ddosIpAddress.Period)
	}
	_ = d.Set("ip_status", ddosIpAddress.DdosIpStatus)
	_ = d.Set("create_time", ddosIpAddress.CreateTime)
	_ = d.Set("expired_time", ddosIpAddress.ExpiredTime)

	return diags

}

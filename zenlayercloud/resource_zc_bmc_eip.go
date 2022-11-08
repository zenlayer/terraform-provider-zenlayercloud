/*
Provides an EIP resource.

Example Usage

```hcl
variable "availability_zone" {
  default = "SEL-A"
}

resource "zenlayercloud_bmc_eip" "foo" {
  availability_zone = var.availability_zone
}
```

Import

EIP can be imported using the id, e.g.

```
$ terraform import zenlayercloud_bmc_eip.foo 123123xxxx
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

func resourceZenlayerCloudEip() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudEipCreate,
		ReadContext:   resourceZenlayerCloudEipRead,
		UpdateContext: resourceZenlayerCloudEipUpdate,
		DeleteContext: resourceZenlayerCloudEipDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of zone that the EIP locates at.",
			},
			"eip_charge_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      BmcChargeTypePostpaid,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(BmcChargeTypes, false),
				Description:  "The charge type of EIP. Valid values are `PREPAID`, `POSTPAID`. The default is `POSTPAID`. Note: `PREPAID` EIP may not allow to delete before expired.",
			},
			"eip_charge_prepaid_period": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1),
				ForceNew:     true,
				Description:  "The tenancy (time unit is month) of the prepaid EIP, NOTE: it only works when eip_charge_type is set to `PREPAID`.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The resource group id the EIP belongs to, default to Default Resource Group.",
			},
			"resource_group_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The resource group name the EIP belongs to, default to Default Resource Group.",
			},
			"force_delete": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Indicate whether to force delete the EIP. Default is `false`. If set true, the EIP will be permanently deleted instead of being moved into the recycle bin.",
			},
			"public_ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The EIP address.",
			},
			"eip_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Current status of the EIP.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Create time of the EIP.",
			},
			"expired_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Expired time of the EIP.",
			},
		},
	}
}

func resourceZenlayerCloudEipDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	eipId := d.Id()

	forceDelete := d.Get("force_delete").(bool)

	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := bmcService.TerminateEipAddress(ctx, eipId)
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
		eip, errRet := bmcService.DescribeEipAddressById(ctx, eipId)
		if errRet != nil {
			return retryError(ctx, errRet, InternalServerError)
		}
		if eip == nil {
			notExist = true
			return nil
		}

		if eip.EipStatus == BmcEipStatusRecycle {
			//in recycling
			return nil
		}
		if eip.EipStatus == BmcEipStatusRecycling {
			return resource.RetryableError(fmt.Errorf("eip (%s) is recycling", eipId))
		}

		return resource.NonRetryableError(fmt.Errorf("eip status is not recycle, current status %s", eip.EipStatus))
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if notExist || !forceDelete {
		return nil
	}

	tflog.Debug(ctx, "Releasing EIP ...", map[string]interface{}{
		"eipId": eipId,
	})

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := bmcService.ReleaseEipAddressById(ctx, eipId)
		if errRet != nil {

			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if !ok {
				return retryError(ctx, errRet)
			}
			if ee.Code == "INVALID_EIP_NOT_FOUND" || ee.Code == "OPERATION_FAILED_RESOURCE_NOT_FOUND" {
				// EIP doesn't exist
				return nil
			}
			return retryError(ctx, errRet, InternalServerError)
		}
		return nil
	})

	return nil
}

func resourceZenlayerCloudEipUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	eipId := d.Id()

	if d.HasChange("resource_group_id") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			err := bmcService.ModifyEipResourceGroup(ctx, eipId, d.Get("resource_group_id").(string))
			if err != nil {
				return retryError(ctx, err, InternalServerError, common.NetworkError)
			}
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceZenlayerCloudEipRead(ctx, d, meta)
}

func resourceZenlayerCloudEipCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	request := bmc.NewAllocateEipAddressesRequest()
	request.ZoneId = d.Get("availability_zone").(string)
	request.EipChargeType = d.Get("eip_charge_type").(string)

	if request.EipChargeType == BmcChargeTypePrepaid {
		request.EipChargePrepaid = &bmc.ChargePrepaid{}

		if period, ok := d.GetOk("eip_charge_prepaid_period"); ok {
			request.EipChargePrepaid.Period = period.(int)
		} else {
			diags = append(diags, diag.Diagnostic{
				Summary: "Missing required argument",
				Detail:  "eip_charge_prepaid_period is missing on prepaid EIP instance.",
			})
			return diags
		}
	}
	if v, ok := d.GetOk("resource_group_id"); ok {
		request.ResourceGroupId = v.(string)
	}

	eipId := ""

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := bmcService.client.WithBmcClient().AllocateEipAddresses(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create eip address.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": toJsonString(request),
				"err":     err.Error(),
			})
			return retryError(ctx, err)
		}

		tflog.Info(ctx, "Create EIP success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  toJsonString(request),
			"response": toJsonString(response),
		})

		if len(response.Response.EipIdSet) < 1 {
			err = fmt.Errorf("eip id is nil")
			return resource.NonRetryableError(err)
		}
		eipId = *response.Response.EipIdSet[0]

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	stateConf := BuildEipState(bmcService, eipId, ctx, d)

	eipState, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for bmc eip (%s) to be created: %v", d.Id(), err)
	}

	if eipState == nil {
		return diag.Errorf("associate eip (%s) to instance failed as ip not found", eipId)
	}

	if eipState.(*bmc.EipAddress).EipStatus != BmcEipStatusAvailable {
		return diag.Errorf("associate eip (%s) failed, current status: %s", eipId, eipState.(*bmc.EipAddress).EipStatus)
	}

	d.SetId(eipId)

	return resourceZenlayerCloudEipRead(ctx, d, meta)
}

func resourceZenlayerCloudEipRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	eipId := d.Id()

	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var eipAddress *bmc.EipAddress
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		eipAddress, errRet = bmcService.DescribeEipAddressById(ctx, eipId)
		if errRet != nil {
			return retryError(ctx, errRet)
		}

		if eipAddress != nil && ipIsOperating(eipAddress.EipStatus) {
			return resource.RetryableError(fmt.Errorf("waiting for eip %s operation", eipAddress.EipId))
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if eipAddress == nil || eipAddress.EipStatus == BmcEipStatusCreateFailed ||
		eipAddress.EipStatus == BmcEipStatusRecycle {
		d.SetId("")
		tflog.Info(ctx, "eip not exist or created failed or recycled", map[string]interface{}{
			"eipId": eipId,
		})
		return nil
	}

	// eip info
	_ = d.Set("availability_zone", eipAddress.ZoneId)
	_ = d.Set("resource_group_id", eipAddress.ResourceGroupId)
	_ = d.Set("resource_group_name", eipAddress.ResourceGroupName)
	_ = d.Set("eip_charge_type", eipAddress.EipChargeType)
	_ = d.Set("public_ip", eipAddress.IpAddress)
	if eipAddress.EipChargeType == BmcChargeTypePrepaid {
		_ = d.Set("eip_charge_prepaid_period", eipAddress.Period)
	}
	_ = d.Set("eip_status", eipAddress.EipStatus)
	_ = d.Set("create_time", eipAddress.CreateTime)
	_ = d.Set("expired_time", eipAddress.ExpiredTime)

	return diags

}

func ipIsOperating(status string) bool {
	return IsContains(EipOperatingStatus, status)
}

func BuildEipState(bmcService BmcService, eipId string, ctx context.Context, d *schema.ResourceData) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending: []string{
			BmcEipStatusAssociating,
			BmcEipStatusCreating,
			BmcEipStatusUnAssociating,
		},
		Target: []string{
			BmcEipStatusCreateFailed,
			BmcEipStatusAssociated,
			BmcEipStatusAvailable,
		},
		Refresh:        bmcService.InstanceEipStateRefreshFunc(ctx, eipId),
		Timeout:        d.Timeout(schema.TimeoutRead) - time.Minute,
		Delay:          10 * time.Second,
		MinTimeout:     5 * time.Second,
		NotFoundChecks: 3,
	}
}

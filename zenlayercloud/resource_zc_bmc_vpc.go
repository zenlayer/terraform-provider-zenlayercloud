/*
Provide a resource to create a VPC.

Example Usage

```hcl
data "zenlayercloud_bmc_vpc_regions" "default_region" {

}

resource "zenlayercloud_bmc_vpc" "foo" {
  region 	 = data.zenlayercloud_bmc_vpc_regions.default_region.regions.0.id
  cidr_block = "10.0.0.0/26"
}
```

Import

Vpc instance can be imported, e.g.

```
$ terraform import zenlayercloud_bmc_vpc.test vpc-id
```
*/
package zenlayercloud

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
	bmc "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/bmc20221120"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	"time"
)

func resourceZenlayerCloudVpc() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudVpcCreate,
		ReadContext:   resourceZenlayerCloudVpcRead,
		UpdateContext: resourceZenlayerCloudVpcUpdate,
		DeleteContext: resourceZenlayerCloudVpcDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"region": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of region that the vpc locates at.",
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "Terraform-VPC",
				ValidateFunc: validation.StringLenBetween(1, 64),
				Description:  "The name of the vpc.",
			},
			"cidr_block": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: common2.ValidateCIDRNetworkAddress,
				ForceNew:     true,
				Description:  "A network address block which should be a subnet of the three internal network segments (10.0.0.0/16, 172.16.0.0/12 and 192.168.0.0/16).",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The resource group id the vpc belongs to, default to ID of Default Resource Group.",
			},
			"resource_group_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The resource group name the vpc belongs to, default to Default Resource Group.",
			},
			"vpc_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Current status of the vpc.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Create time of the vpc.",
			},
		},
	}
}

func resourceZenlayerCloudVpcDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcId := d.Id()

	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		filter := &SubnetFilter{
			VpcId: vpcId,
		}
		subnets, errRet := bmcService.DescribeSubnets(ctx, filter)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		associateSubnetsCount := len(subnets)
		if associateSubnetsCount == 0 {
			return nil
		}
		return resource.RetryableError(fmt.Errorf("vpc %s still bind %d subnets", vpcId, associateSubnetsCount))
	})
	if err != nil {
		diag.FromErr(err)
	}

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := bmcService.DeleteVpc(ctx, vpcId)
		if errRet != nil {
			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if !ok {
				return common2.RetryError(ctx, errRet, common2.InternalServerError)
			}
			if ee.Code == common2.ResourceNotFound {
				// vpc doesn't exist
				return nil
			}
			return resource.NonRetryableError(errRet)
		}

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{
			BmcVpcStatusDeleting,
		},
		Target:         []string{},
		Refresh:        bmcService.VpcStateRefreshFunc(ctx, vpcId, []string{}),
		Timeout:        d.Timeout(schema.TimeoutDelete) - time.Minute,
		Delay:          5 * time.Second,
		MinTimeout:     3 * time.Second,
		NotFoundChecks: 3,
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceZenlayerCloudVpcUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	vpcId := d.Id()
	d.Partial(true)
	if d.HasChange("name") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			err := bmcService.ModifyVpcName(ctx, vpcId, d.Get("name").(string))
			if err != nil {
				return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}

	// resource group
	if d.HasChange("resource_group_id") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			err := bmcService.ModifyVpcResourceGroup(ctx, vpcId, d.Get("resource_group_id").(string))
			if err != nil {
				return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.Partial(false)
	return resourceZenlayerCloudVpcRead(ctx, d, meta)
}

func resourceZenlayerCloudVpcCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	request := bmc.NewCreateVpcRequest()
	request.VpcRegionId = d.Get("region").(string)
	request.CidrBlock = d.Get("cidr_block").(string)
	request.VpcName = d.Get("name").(string)
	if v, ok := d.GetOk("resource_group_id"); ok {
		request.ResourceGroupId = v.(string)
	}

	vpcId := ""

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := meta.(*connectivity.ZenlayerCloudClient).WithBmcClient().CreateVpc(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create bmc vpc.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": common2.ToJsonString(request),
				"err":     err.Error(),
			})
			return common2.RetryError(ctx, err)
		}

		tflog.Info(ctx, "Create vpc success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  common2.ToJsonString(request),
			"response": common2.ToJsonString(response),
		})

		if response.Response.VpcId == "" {
			err = fmt.Errorf("vpc id is nil")
			return resource.NonRetryableError(err)
		}
		vpcId = response.Response.VpcId

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(vpcId)

	stateConf := &resource.StateChangeConf{
		Pending: []string{
			BmcVpcStatusDeleting,
			BmcVpcStatusCreating,
		},
		Target: []string{
			BmcVpcStatusCreateAvailable,
		},
		Refresh:        bmcService.VpcStateRefreshFunc(ctx, vpcId, []string{BmcVpcStatusCreateFailed}),
		Timeout:        d.Timeout(schema.TimeoutCreate) - time.Minute,
		Delay:          10 * time.Second,
		MinTimeout:     5 * time.Second,
		NotFoundChecks: 3,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for vpc (%s) to be created: %v", d.Id(), err))
	}

	return resourceZenlayerCloudVpcRead(ctx, d, meta)
}

func resourceZenlayerCloudVpcRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	instanceId := d.Id()

	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var vpc *bmc.VpcInfo
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		vpc, errRet = bmcService.DescribeVpcById(ctx, instanceId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}

		if vpc != nil && vpcIsOperating(vpc.VpcStatus) {
			return resource.RetryableError(fmt.Errorf("waiting for vpc %s operation", vpc.VpcId))
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if vpc == nil || vpc.VpcStatus == BmcVpcStatusCreateFailed {
		d.SetId("")
		tflog.Info(ctx, "vpc not exist or created failed", map[string]interface{}{
			"instanceId": instanceId,
		})
		return nil
	}

	// vpc info
	_ = d.Set("region", vpc.VpcRegionId)
	_ = d.Set("name", vpc.VpcName)
	_ = d.Set("vpc_status", vpc.VpcStatus)
	_ = d.Set("cidr_block", vpc.CidrBlock)
	_ = d.Set("resource_group_id", vpc.ResourceGroupId)
	_ = d.Set("resource_group_name", vpc.ResourceGroupName)
	_ = d.Set("create_time", vpc.CreateTime)

	return diags

}

func vpcIsOperating(status string) bool {
	return common2.IsContains(VpcOperatingStatus, status)
}

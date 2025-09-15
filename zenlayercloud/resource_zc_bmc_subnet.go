/*
Provide a resource to create a VPC subnet.

Example Usage

```hcl
variable "region" {
  default = "SEL1"
}

variable "availability_zone" {
  default = "SEL-A"
}

resource "zenlayercloud_bmc_vpc" "foo" {
  region	 = var.region
  name       = "test-vpc"
  cidr_block = "10.0.0.0/16"
}

resource "zenlayercloud_bmc_subnet" "subnet_with_vpc" {
  availability_zone = var.availability_zone
  name              = "test-subnet"
  vpc_id            = zenlayercloud_bmc_vpc.foo.id
  cidr_block        = "10.0.10.0/24"
}

resource "zenlayercloud_bmc_subnet" "subnet" {
  availability_zone = var.availability_zone
  name              = "test-subnet"
  cidr_block        = "10.0.10.0/24"
}
```

Import

Vpc subnet instance can be imported, e.g.

```
$ terraform import zenlayercloud_bmc_subnet.subnet subnet_id
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
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	bmc "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/bmc20221120"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	"time"
)

func resourceZenlayerCloudBmcSubnet() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudBmcSubnetCreate,
		ReadContext:   resourceZenlayerCloudBmcSubnetRead,
		UpdateContext: resourceZenlayerCloudBmcSubnetUpdate,
		DeleteContext: resourceZenlayerCloudBmcSubnetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of zone that the bmc subnet locates at.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "Terraform-Subnet",
				Description: "The name of the bmc subnet.",
			},
			"cidr_block": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: common2.ValidateCIDRNetworkAddress,
				ForceNew:     true,
				Description:  "A network address block which should be a subnet of the three internal network segments (10.0.0.0/16, 172.16.0.0/12 and 192.168.0.0/16).",
			},
			"vpc_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the VPC to be associated.",
				ForceNew:    true,
			},
			"vpc_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the VPC to be associated.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The resource group id the subnet belongs to, default to Default Resource Group.",
			},
			"resource_group_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The resource group name the subnet belongs to, default to Default Resource Group.",
			},
			"subnet_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Current status of the subnet.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Create time of the subnet.",
			},
		},
	}
}

func resourceZenlayerCloudBmcSubnetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	subnetId := d.Id()

	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	// wait until all instances unbind this subnet
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		subnet, errRet := bmcService.DescribeSubnetById(ctx, subnetId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		associateInstanceCount := len(subnet.SubnetInstanceSet)
		if associateInstanceCount == 0 {
			return nil
		}
		return resource.RetryableError(fmt.Errorf("subnet %s still bind %d instances", subnetId, associateInstanceCount))
	})

	if err != nil {
		return diag.FromErr(err)
	}

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := bmcService.DeleteSubnet(ctx, subnetId)
		if errRet != nil {
			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if !ok {
				return common2.RetryError(ctx, errRet, common2.InternalServerError)
			}
			if ee.Code == common2.ResourceNotFound {
				// vpc doesn't exist
				return nil
			}
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	// wait for subnet deleted
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			BmcSubnetStatusDeleting,
		},
		Target:         []string{},
		Refresh:        bmcService.SubnetStateRefreshFunc(ctx, subnetId, []string{}),
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

func resourceZenlayerCloudBmcSubnetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	subnetId := d.Id()
	d.Partial(true)
	if d.HasChange("name") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			err := bmcService.ModifySubnetName(ctx, subnetId, d.Get("name").(string))
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
			err := bmcService.ModifySubnetResourceGroupById(ctx, subnetId, d.Get("resource_group_id").(string))
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
	return resourceZenlayerCloudBmcSubnetRead(ctx, d, meta)
}

func resourceZenlayerCloudBmcSubnetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	request := bmc.NewCreateSubnetRequest()
	request.ZoneId = d.Get("availability_zone").(string)
	request.CidrBlock = d.Get("cidr_block").(string)
	request.SubnetName = d.Get("name").(string)
	if v, ok := d.GetOk("resource_group_id"); ok {
		request.ResourceGroupId = v.(string)
	}
	if v, ok := d.GetOk("vpc_id"); ok {
		request.VpcId = v.(string)
	}

	subnetId := ""

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := meta.(*connectivity.ZenlayerCloudClient).WithBmcClient().CreateSubnet(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create bmc subnet.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": common2.ToJsonString(request),
				"err":     err.Error(),
			})
			return common2.RetryError(ctx, err)
		}

		tflog.Info(ctx, "Create subnet success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  common2.ToJsonString(request),
			"response": common2.ToJsonString(response),
		})

		if response.Response.SubnetId == "" {
			err = fmt.Errorf("subnet id is nil")
			return resource.NonRetryableError(err)
		}
		subnetId = response.Response.SubnetId

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(subnetId)

	stateConf := &resource.StateChangeConf{
		Pending: []string{
			BmcSubnetStatusPending,
			BmcSubnetStatusCreating,
		},
		Target: []string{
			BmcSubnetStatusAvailable,
		},
		Refresh:        bmcService.SubnetStateRefreshFunc(ctx, subnetId, []string{BmcInstanceStatusCreateFailed}),
		Timeout:        d.Timeout(schema.TimeoutCreate) - time.Minute,
		Delay:          10 * time.Second,
		MinTimeout:     5 * time.Second,
		NotFoundChecks: 3,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for bmc subnet (%s) to be created: %v", d.Id(), err))
	}

	return resourceZenlayerCloudBmcSubnetRead(ctx, d, meta)
}

func resourceZenlayerCloudBmcSubnetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	instanceId := d.Id()

	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var subnet *bmc.Subnet
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		subnet, errRet = bmcService.DescribeSubnetById(ctx, instanceId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}

		if subnet != nil && subnetIsOperating(subnet.SubnetStatus) {
			return resource.RetryableError(fmt.Errorf("waiting for subnet %s operation", subnet.SubnetId))
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if subnet == nil || subnet.SubnetStatus == BmcSubnetStatusCreateFailed {
		d.SetId("")
		tflog.Info(ctx, "subnet not exist or created failed", map[string]interface{}{
			"instanceId": instanceId,
		})
		return nil
	}

	// subnet info
	_ = d.Set("availability_zone", subnet.ZoneId)
	_ = d.Set("name", subnet.SubnetName)
	_ = d.Set("subnet_status", subnet.SubnetStatus)
	_ = d.Set("cidr_block", subnet.CidrBlock)
	_ = d.Set("vpc_id", subnet.VpcId)
	_ = d.Set("vpc_name", subnet.VpcName)
	_ = d.Set("resource_group_id", subnet.ResourceGroupId)
	_ = d.Set("resource_group_name", subnet.ResourceGroupName)
	_ = d.Set("create_time", subnet.CreateTime)

	return diags

}

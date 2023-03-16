/*
Provide a resource to create a subnet.

Example Usage

```hcl

variable "availability_zone" {
  default = "SEL-A"
}

resource "zenlayercloud_subnet" "foo" {
  availability_zone	 = var.availability_zone
  name       = "test-subnet"
  cidr_block = "10.0.0.0/24"
}

```

Import

Subnet instance can be imported, e.g.

```
$ terraform import zenlayercloud_subnet.subnet subnet_id
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
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	vm "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/vm20230313"
	"time"
)

func resourceZenlayerCloudSubnet() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudSubnetCreate,
		ReadContext:   resourceZenlayerCloudSubnetRead,
		UpdateContext: resourceZenlayerCloudSubnetUpdate,
		DeleteContext: resourceZenlayerCloudSubnetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "Terraform-Subnet",
				Description: "The name of the subnet, the default value is 'Terraform-Subnet'.",
			},
			"availability_zone": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of zone that the subnet locates at.",
			},
			"cidr_block": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateCIDRNetworkAddress,
				ForceNew:     true,
				Description:  "A network address block which should be a subnet of the three internal network segments (10.0.0.0/24, 172.16.0.0/24 and 192.168.0.0/24).",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of subnet.",
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

func resourceZenlayerCloudSubnetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	subnetId := d.Id()

	vmService := VmService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	// wait until all instances unbind this subnet
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		subnet, errRet := vmService.DescribeSubnetById(ctx, subnetId)
		if errRet != nil {
			return retryError(ctx, errRet)
		}
		associateInstanceCount := len(subnet.InstanceIdList)
		if associateInstanceCount == 0 {
			return nil
		}
		return resource.NonRetryableError(fmt.Errorf("subnet %s still bind %d instances", subnetId, associateInstanceCount))
	})

	if err != nil {
		return diag.FromErr(err)
	}

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := vmService.DeleteSubnet(ctx, subnetId)
		if errRet != nil {
			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if !ok {
				return retryError(ctx, errRet, InternalServerError)
			}
			if ee.Code == ResourceNotFound {
				// vpc doesn't exist
				return nil
			}
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceZenlayerCloudSubnetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	//var _ diag.Diagnostics
	vmService := VmService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	subnetId := d.Id()
	if d.HasChange("name") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			err := vmService.ModifySubnetName(ctx, subnetId, d.Get("name").(string))
			if err != nil {
				return retryError(ctx, err, InternalServerError, common.NetworkError)
			}
			return nil
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceZenlayerCloudSubnetRead(ctx, d, meta)
}

func resourceZenlayerCloudSubnetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	vmService := VmService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	request := vm.NewCreateSubnetRequest()
	request.ZoneId = d.Get("availability_zone").(string)
	request.CidrBlock = d.Get("cidr_block").(string)
	request.SubnetName = d.Get("name").(string)
	if v, ok := d.GetOk("description"); ok {
		request.SubnetDescription = v.(string)
	}

	subnetId := ""

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := meta.(*connectivity.ZenlayerCloudClient).WithVmClient().CreateSubnet(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create subnet.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": toJsonString(request),
				"err":     err.Error(),
			})
			return retryError(ctx, err)
		}

		tflog.Info(ctx, "Create subnet success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  toJsonString(request),
			"response": toJsonString(response),
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
			VmSubnetStatusCreating,
		},
		Target: []string{
			VmSubnetStatusAvailable,
		},
		Refresh:        vmService.SubnetStateRefreshFunc(ctx, subnetId, []string{VmSubnetStatusFailed}),
		Timeout:        d.Timeout(schema.TimeoutCreate) - time.Minute,
		Delay:          3 * time.Second,
		MinTimeout:     5 * time.Second,
		NotFoundChecks: 3,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for subnet (%s) to be created: %v", d.Id(), err))
	}

	return resourceZenlayerCloudSubnetRead(ctx, d, meta)
}

func resourceZenlayerCloudSubnetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	instanceId := d.Id()

	vmService := VmService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var subnet *vm.SubnetInfo
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		subnet, errRet = vmService.DescribeSubnetById(ctx, instanceId)
		if errRet != nil {
			return retryError(ctx, errRet)
		}

		if subnet != nil && subnetIsOperating(subnet.SubnetStatus) {
			return resource.RetryableError(fmt.Errorf("waiting for subnet %s operation", subnet.SubnetId))
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if subnet == nil || subnet.SubnetStatus == VmSubnetStatusFailed {
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
	_ = d.Set("description", subnet.SubnetDescription)
	_ = d.Set("create_time", subnet.CreateTime)

	return diags
}

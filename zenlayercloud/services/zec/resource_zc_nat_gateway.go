/*
Provide a resource to create a subnet.

# Example Usage

```hcl

	variable "region_shanghai" {
	  default = "asia-east-1"
	}

	resource "zenlayercloud_zec_nat_gateway" "foo" {
	  region_id	 = var.region_shanghai
	  name       = "test-nat"
	  vpc_id = "xxxxxxxxx"
	}

```

# Import

NAT Gateway instance can be imported, e.g.

```
$ terraform import zenlayercloud_zec_nat_gateway.foo nat-gateway-id
```
*/
package zec

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
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20240401"
	"time"
)

func ResourceZenlayerCloudZecVpcNatGateway() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZecVpcNatGatewayCreate,
		ReadContext:   resourceZenlayerCloudZecVpcNatGatewayRead,
		UpdateContext: resourceZenlayerCloudZecVpcNatGatewayUpdate,
		DeleteContext: resourceZenlayerCloudZecVpcNatGatewayDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "Terraform-Nat-Gateway",
				Description: "The name of the NAT gateway, the default value is 'Terraform-Subnet'.",
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The region that the NAT gateway locates at.",
			},
			"vpc_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the VPC to be associated.",
			},
			"subnet_ids": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				ForceNew:    true,
				Description: "IDs of the subnets to be associated. if this value not set",
			},
			"is_all_subnets": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether all the subnets of region is assigned to NAT gateway.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The resource group id the NAT gateway belongs to.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Create time of the NAT gateway.",
			},
		},
	}
}

func resourceZenlayerCloudZecVpcNatGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	subnetId := d.Id()

	ZecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := ZecService.DeleteSubnet(ctx, subnetId)
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

	return nil
}

func resourceZenlayerCloudZecVpcNatGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	//var _ diag.Diagnostics
	//ZecService := ZecService{
	//	client: meta.(*connectivity.ZenlayerCloudClient),
	//}
	//subnetId := d.Id()
	if d.HasChange("name") {
		//err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
		//
		//	err := ZecService.ModifyNatGateway(ctx, subnetId, d.Get("name").(string))
		//	if err != nil {
		//		return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
		//	}
		//	return nil
		//})
		//
		//if err != nil {
		//	return diag.FromErr(err)
		//}
	}

	if d.HasChange("resource_group_id") {
		//err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
		//	err := ZecService.ModifySubnetName(ctx, subnetId, d.Get("name").(string))
		//	if err != nil {
		//		return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
		//	}
		//	return nil
		//})
		//
		//if err != nil {
		//	return diag.FromErr(err)
		//}
	}

	return resourceZenlayerCloudZecVpcNatGatewayRead(ctx, d, meta)
}

func resourceZenlayerCloudZecVpcNatGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	ZecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	request := zec.NewCreateNatGatewayRequest()
	request.RegionId = common.String(d.Get("region_id").(string))
	request.Name = common.String(d.Get("name").(string))
	request.VpcId = common.String(d.Get("vpc_id").(string))

	if v, ok := d.GetOk("subnet_ids"); ok {
		subnetIds := v.(*schema.Set).List()
		if len(subnetIds) > 0 {
			request.SubnetIds = common2.ToStringList(subnetIds)
		}
	}

	if v, ok := d.GetOk("resource_group_id"); ok {
		request.ResourceGroupId = common.String(v.(string))
	}

	natGatewayId := ""

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := ZecService.client.WithZecClient().CreateNatGateway(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create subnet.", map[string]interface{}{
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

		if response.Response.NatGatewayId == nil {
			err = fmt.Errorf("NAT gateway id is nil")
			return resource.NonRetryableError(err)
		}
		natGatewayId = *response.Response.NatGatewayId

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(natGatewayId)

	return resourceZenlayerCloudZecVpcNatGatewayRead(ctx, d, meta)
}

func resourceZenlayerCloudZecVpcNatGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	natGatewayId := d.Id()

	ZecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var natGateway *zec.NatGateway
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		natGateway, errRet = ZecService.DescribeNatGatewayById(ctx, natGatewayId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if natGateway == nil {
		d.SetId("")
		tflog.Info(ctx, "natGateway not exist", map[string]interface{}{
			"natGatewayId": natGatewayId,
		})
		return nil
	}

	// natGateway info
	_ = d.Set("region_id", natGateway.RegionId)
	_ = d.Set("name", natGateway.Name)
	_ = d.Set("vpc_id", natGateway.VpcId)

	// TODO all_subnets
	_ = d.Set("subnet_ids", nil)
	_ = d.Set("resource_group_id", natGateway.ResourceGroupId)
	_ = d.Set("create_time", natGateway.CreateTime)

	return diags
}

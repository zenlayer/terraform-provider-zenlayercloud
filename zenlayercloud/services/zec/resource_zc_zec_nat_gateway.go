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
	user "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/user20240529"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"
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
			"security_group_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of a security group.",
			},
			"subnet_ids": {
				Type:          schema.TypeSet,
				Elem:          &schema.Schema{Type: schema.TypeString},
				Optional:      true,
				ConflictsWith: []string{"is_all_subnets"},
				AtLeastOneOf:  []string{"subnet_ids", "is_all_subnets"},
				Description:   "IDs of the subnets to be associated. The subnets must belong to the specified VPC.",
			},
			"is_all_subnets": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"subnet_ids"},
				AtLeastOneOf:  []string{"subnet_ids", "is_all_subnets"},
				Description:   "Indicates whether all the subnets of region is assigned to NAT gateway. This field is conflict with `subnet_ids`.",
			},
			"enable_icmp_reply": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Indicates whether ICMP replay is enabled. Default is disabled.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The resource group id the NAT gateway belongs to, default to ID of Default Resource Group.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Create time of the NAT gateway.",
			},
			"force_delete": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Indicate whether to force delete the NAT gateway. Default is `true`. If set true, the NAT gateway will be permanently deleted instead of being moved into the recycle bin.",
			},
		},
	}
}

func resourceZenlayerCloudZecVpcNatGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	natGatewayId := d.Id()

	// force_delete: terminate and then delete
	forceDelete := d.Get("force_delete").(bool)

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := zecService.DeleteNatGateway(ctx, natGatewayId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	notExist := false

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		natGateway, errRet := zecService.DescribeNatGatewayById(ctx, natGatewayId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet, common2.InternalServerError)
		}
		if natGateway == nil {
			notExist = true
			return nil
		}

		if *natGateway.Status == NatStatusRecycled {
			//in recycling
			return nil
		}

		if isOperating(*natGateway.Status) {
			return resource.RetryableError(fmt.Errorf("waiting for natGateway %s recycling, current status: %s", natGatewayId, *natGateway.Status))
		}

		return resource.NonRetryableError(fmt.Errorf("natGateway status is not recycle, current status %s", *natGateway.Status))
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if notExist || !forceDelete {
		return nil
	}
	tflog.Debug(ctx, "Releasing NAT gateway ...", map[string]interface{}{
		"natGatewayId": natGatewayId,
	})

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := zecService.DeleteNatGateway(ctx, natGatewayId)
		if errRet != nil {

			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if !ok {
				return common2.RetryError(ctx, errRet)
			}
			if ee.Code == INVALID_NAT_NOT_FOUND || ee.Code == common2.ResourceNotFound {
				return nil
			}
			return common2.RetryError(ctx, errRet, common2.InternalServerError)
		}
		return nil
	})

	if err != nil {
		diag.FromErr(err)
	}
	return nil
}

func resourceZenlayerCloudZecVpcNatGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	natGatewayId := d.Id()

	if d.HasChanges("name", "security_group_id", "subnet_ids", "is_all_subnets", "enable_icmp_reply") {

		request := zec.NewModifyNatGatewayAttributeRequest()
		request.NatGatewayId = common.String(natGatewayId)
		request.Name = common.String(d.Get("name").(string))
		request.SecurityGroupId = common.String(d.Get("security_group_id").(string))
		request.IsAllSubnet = common.Bool(d.Get("is_all_subnets").(bool))
		request.IcmpReplyEnabled = common.Bool(d.Get("enable_icmp_reply").(bool))
		if v, ok := d.GetOk("subnet_ids"); ok {
			request.SubnetIds = common2.ToStringList(v.(*schema.Set).List())
		}

		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {

			_, err := zecService.client.WithZec2Client().ModifyNatGatewayAttribute(request)
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
			request.Resources = []*string{common.String(natGatewayId)}

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

	return resourceZenlayerCloudZecVpcNatGatewayRead(ctx, d, meta)
}

func resourceZenlayerCloudZecVpcNatGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	request := zec.NewCreateNatGatewayRequest()
	request.RegionId = common.String(d.Get("region_id").(string))
	request.Name = common.String(d.Get("name").(string))
	request.VpcId = common.String(d.Get("vpc_id").(string))
	request.SecurityGroupId = common.String(d.Get("security_group_id").(string))

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
		response, err := zecService.client.WithZec2Client().CreateNatGateway(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create NAT gateway.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": common2.ToJsonString(request),
				"err":     err.Error(),
			})
			return common2.RetryError(ctx, err, common2.OperationTimeout)
		}

		tflog.Info(ctx, "Create NAT gateway success", map[string]interface{}{
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

	stateConf := BuildNatGatewayState(zecService, natGatewayId, ctx, d)
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for waiting NAT gateway (%s) to be running: %v", d.Id(), err)
	}

	if v, ok := d.GetOk("enable_icmp_reply"); ok {

		request := zec.NewModifyNatGatewayAttributeRequest()
		request.NatGatewayId = common.String(natGatewayId)
		request.IcmpReplyEnabled = common.Bool(v.(bool))

		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {

			_, err := zecService.client.WithZec2Client().ModifyNatGatewayAttribute(request)
			if err != nil {
				return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}
	return resourceZenlayerCloudZecVpcNatGatewayRead(ctx, d, meta)
}

func BuildNatGatewayState(zecService ZecService, natGatewayId string, ctx context.Context, d *schema.ResourceData) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending: []string{
			NatStatusCreating,
		},
		Target: []string{
			NatStatusRunning,
		},
		Refresh:        zecService.NatStateRefreshFunc(ctx, natGatewayId, []string{NatStatusCreateFailed}),
		Timeout:        d.Timeout(schema.TimeoutRead) - time.Minute,
		Delay:          10 * time.Second,
		MinTimeout:     5 * time.Second,
		NotFoundChecks: 3,
	}
}

func resourceZenlayerCloudZecVpcNatGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	natGatewayId := d.Id()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var natGateway *zec.NatGateway
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		natGateway, errRet = zecService.DescribeNatGatewayById(ctx, natGatewayId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}

		if natGateway != nil && isOperating(*natGateway.Status) {
			return resource.RetryableError(fmt.Errorf("waiting for NAT gateway %s operation", natGatewayId))
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if natGateway == nil {
		d.SetId("")
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "NAT gateway doesn't not exist",
			Detail:   fmt.Sprintf("The NAT gateway %s is not exist", natGatewayId),
		})
		return diags
	}

	// natGateway info
	_ = d.Set("region_id", natGateway.RegionId)
	_ = d.Set("name", natGateway.Name)
	_ = d.Set("vpc_id", natGateway.VpcId)

	_ = d.Set("is_all_subnets", natGateway.IsAllSubnets)
	if natGateway.SubnetIds != nil {
		_ = d.Set("subnet_ids", natGateway.SubnetIds)
	}
	_ = d.Set("subnet_ids", nil)
	_ = d.Set("resource_group_id", natGateway.ResourceGroupId)
	_ = d.Set("security_group_id", natGateway.SecurityGroupId)
	_ = d.Set("create_time", natGateway.CreateTime)
	_ = d.Set("enable_icmp_reply", natGateway.IcmpReplyEnabled)

	return diags
}

func isOperating(status string) bool {
	return common2.IsContains([]string{
		NatStatusCreating,
		NatStatusReleasing}, status)
}

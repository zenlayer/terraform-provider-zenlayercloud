package zec


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
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20240401"
	"time"
)

func ResourceZenlayerCloudZecSecurityGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZecSecurityGroupCreate,
		ReadContext:   resourceZenlayerCloudZecSecurityGroupRead,
		UpdateContext: resourceZenlayerCloudZecSecurityGroupUpdate,
		DeleteContext: resourceZenlayerCloudZecSecurityGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
				Description:  "The name of the security group. The length is 1 to 64 characters. Only letters, numbers, - and periods (.) are supported.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation time of the security group.",
			},
		},
	}
}

func resourceZenlayerCloudZecSecurityGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	securityGroupId := d.Id()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	// wait until all instances unbind this security group
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		sucurityGroup, errRet := zecService.DescribeSecurityGroupById(ctx, securityGroupId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		associateVpcCount := len(sucurityGroup.VpcIds)
		if associateVpcCount == 0 {
			return nil
		}
		return resource.RetryableError(fmt.Errorf("security group %s still have %d vpcs", securityGroupId, associateVpcCount))
	})

	if err != nil {
		return diag.FromErr(err)
	}

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := zecService.DeleteSecurityGroupById(ctx, securityGroupId)
		if errRet != nil {
			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if !ok {
				return common2.RetryError(ctx, errRet, common2.InternalServerError)
			}
			if ee.Code == common2.ResourceNotFound {
				// security group doesn't exist
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

func resourceZenlayerCloudZecSecurityGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	securityGroupId := d.Id()
	if d.HasChanges("name") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			err := zecService.ModifySecurityGroupName(ctx, securityGroupId, d.Get("name").(string))
			if err != nil {
				return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceZenlayerCloudZecSecurityGroupRead(ctx, d, meta)
}

func resourceZenlayerCloudZecSecurityGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	request := zec.NewCreateSecurityGroupRequest()
	request.SecurityGroupName = d.Get("name").(string)

	securityGroupId := ""

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := meta.(*connectivity.ZenlayerCloudClient).WithZecClient().CreateSecurityGroup(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create zec security group.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": common2.ToJsonString(request),
				"err":     err.Error(),
			})
			return common2.RetryError(ctx, err)
		}

		tflog.Info(ctx, "Create zec security group success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  common2.ToJsonString(request),
			"response": common2.ToJsonString(response),
		})

		if response.Response.SecurityGroupId == "" {
			err = fmt.Errorf("zec security group id is nil")
			return resource.NonRetryableError(err)
		}
		securityGroupId = response.Response.SecurityGroupId

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(securityGroupId)

	return resourceZenlayerCloudZecSecurityGroupRead(ctx, d, meta)
}

func resourceZenlayerCloudZecSecurityGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	securityGroupId := d.Id()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var securityGroup *zec.SecurityGroupInfo
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		securityGroup, errRet = zecService.DescribeSecurityGroupById(ctx, securityGroupId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if securityGroup == nil {
		d.SetId("")
		tflog.Info(ctx, "security group not exist", map[string]interface{}{
			"securityGroupId": securityGroupId,
		})
		return nil
	}

	// security group info
	d.SetId(securityGroup.SecurityGroupId)
	_ = d.Set("name", securityGroup.SecurityGroupName)

	return diags

}

package zec

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	zec2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20240401"
	"time"
)

func ResourceZenlayerCloudZecVpcSecurityGroupAttachment() *schema.Resource {

	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZecVpcSecurityGroupAttachmentCreate,
		ReadContext:   resourceZenlayerCloudZecVpcSecurityGroupAttachmentRead,
		DeleteContext: resourceZenlayerCloudZecVpcSecurityGroupAttachmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the VPC.",
			},
			"security_group_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the security group.",
			},
		},
	}
}

func resourceZenlayerCloudZecVpcSecurityGroupAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zec_vpc_security_group_attachment.delete")()

	vmService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	attachment, err := common2.ParseResourceId(d.Id(), 2)
	if err != nil {
		return diag.FromErr(err)
	}

	request := zec2.NewUnAssignSecurityGroupVpcRequest()
	vpcId := attachment[0]
	request.VpcId = vpcId
	request.SecurityGroupId = attachment[1]

	if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		_, errRet := vmService.client.WithZecClient().UnAssignSecurityGroupVpc(request)

		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		return nil
	}); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceZenlayerCloudZecVpcSecurityGroupAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zec_vpc_security_group_attachment.create")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	vpcId := d.Get("vpc_id").(string)
	securityGroupId := d.Get("security_group_id").(string)

	request := zec2.NewAssignSecurityGroupVpcRequest()
	request.VpcId = vpcId
	request.SecurityGroupId = securityGroupId

	if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		_, errRet := zecService.client.WithZecClient().AssignSecurityGroupVpc(request)
		if errRet != nil {
			return common2.RetryError(ctx, errRet, common2.OperationTimeout)
		}
		return nil
	}); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s:%s", vpcId, securityGroupId))

	return resourceZenlayerCloudZecVpcSecurityGroupAttachmentRead(ctx, d, meta)
}

func resourceZenlayerCloudZecVpcSecurityGroupAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zec_vpc_security_group_attachment.read")()

	vNicInstanceId, err := common2.ParseResourceId(d.Id(), 2)
	if err != nil {
		return diag.FromErr(err)
	}
	vpcId := vNicInstanceId[0]

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var vpc *zec2.VpcInfo
	var errRet error

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		vpc, errRet = zecService.DescribeVpcById(ctx, vpcId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		if vpc == nil {
			d.SetId("")
		}

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if vpc == nil || vpc.SecurityGroupId == "" {
		d.SetId("")
		return nil
	}

	_ = d.Set("vpc_id", vpc.VpcId)
	_ = d.Set("security_group_id", vpc.SecurityGroupId)
	return nil
}

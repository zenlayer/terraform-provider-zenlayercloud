package zec

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	zec2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"
	"time"
)

func ResourceZenlayerCloudZecVNicAttachment() *schema.Resource {

	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZecVNicAttachmentCreate,
		ReadContext:   resourceZenlayerCloudZecVNicAttachmentRead,
		DeleteContext: resourceZenlayerCloudZecVNicAttachmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"vnic_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the vNIC.",
			},
			"instance_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of a ZEC instance.",
			},
		},
	}
}

func resourceZenlayerCloudZecVNicAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zec_vnic_attachment.delete")()

	vmService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	attachment, err := common2.ParseResourceId(d.Id(), 2)
	if err != nil {
		return diag.FromErr(err)
	}

	request := zec2.NewDetachNetworkInterfaceRequest()
	vnicId := attachment[0]
	request.NicId = &vnicId

	if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		_, errRet := vmService.client.WithZec2Client().DetachNetworkInterface(request)
		ee, ok := errRet.(*common.ZenlayerCloudSdkError)
		if ok {
			if ee.Code == "OPERATION_DENIED_NIC_NOT_EXIST_INSTANCE" {
				return nil
			}
		}
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		return nil
	}); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceZenlayerCloudZecVNicAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zec_vnic_attachment.create")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	vnicId := d.Get("vnic_id").(string)
	instanceId := d.Get("instance_id").(string)

	request := zec2.NewAttachNetworkInterfaceRequest()
	request.InstanceId = &instanceId
	request.NicId = &vnicId

	if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		_, errRet := zecService.client.WithZec2Client().AttachNetworkInterface(request)
		if errRet != nil {
			return common2.RetryError(ctx, errRet, common2.OperationTimeout)
		}
		return nil
	}); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s:%s", vnicId, instanceId))

	return resourceZenlayerCloudZecVNicAttachmentRead(ctx, d, meta)
}

func resourceZenlayerCloudZecVNicAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zec_vnic_attachment.read")()

	vNicInstanceId, err := common2.ParseResourceId(d.Id(), 2)
	if err != nil {
		return diag.FromErr(err)
	}
	vnicId := vNicInstanceId[0]

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var vnic *zec2.NicInfo
	var errRet error

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		vnic, errRet = zecService.DescribeNicById(ctx, vnicId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if vnic == nil || vnic.InstanceId == nil {
		d.SetId("")
		return nil
	}

	_ = d.Set("disk_id", vnic.NicId)
	_ = d.Set("instance_id", vnic.InstanceId)
	return nil
}

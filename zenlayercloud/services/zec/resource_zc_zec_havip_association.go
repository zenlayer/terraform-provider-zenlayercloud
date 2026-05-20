/*
Provides a resource to bind a ZEC instance to a high-availability virtual IP (HaVip).

Example Usage

```hcl
resource "zenlayercloud_zec_havip_association" "example" {
  ha_vip_id   = "havip-xxxxxxxx"
  instance_id = "vm-xxxxxxxx"
}
```

Import

HaVip association can be imported using the id (ha_vip_id:instance_id), e.g.

```
terraform import zenlayercloud_zec_havip_association.example havip-xxxxxxxx:vm-xxxxxxxx
```
*/
package zec

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	sdkcommon "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"
)

func ResourceZenlayerCloudZecHaVipAssociation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZecHaVipAssociationCreate,
		ReadContext:   resourceZenlayerCloudZecHaVipAssociationRead,
		DeleteContext: resourceZenlayerCloudZecHaVipAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"ha_vip_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the HaVip.",
			},
			"instance_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the instance to associate. The instance's network interface must be in the same subnet as the HaVip.",
			},
		},
	}
}

func resourceZenlayerCloudZecHaVipAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_havip_association.create")()

	haVipId := d.Get("ha_vip_id").(string)
	instanceId := d.Get("instance_id").(string)

	request := zec.NewAssociateHaVipRequest()
	request.HaVipId = sdkcommon.String(haVipId)
	request.InstanceId = sdkcommon.String(instanceId)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := meta.(*connectivity.ZenlayerCloudClient).WithZec2Client().AssociateHaVip(request)
		if err != nil {
			return common.RetryError(ctx, err, common.InternalServerError, sdkcommon.NetworkError)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s:%s", haVipId, instanceId))
	return resourceZenlayerCloudZecHaVipAssociationRead(ctx, d, meta)
}

func resourceZenlayerCloudZecHaVipAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_havip_association.read")()

	parts, err := common.ParseResourceId(d.Id(), 2)
	if err != nil {
		return diag.FromErr(err)
	}
	haVipId := parts[0]
	instanceId := parts[1]

	zecService := ZecService{client: meta.(*connectivity.ZenlayerCloudClient)}

	var haVip *zec.HaVipInfo
	var errRet error

	retryErr := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		haVip, errRet = zecService.DescribeHaVipById(ctx, haVipId)
		if errRet != nil {
			return common.RetryError(ctx, errRet)
		}
		return nil
	})

	if retryErr != nil {
		return diag.FromErr(retryErr)
	}

	if haVip == nil {
		d.SetId("")
		return nil
	}

	found := false
	for _, id := range haVip.AssociatedInstances {
		if id == instanceId {
			found = true
			break
		}
	}
	if !found {
		d.SetId("")
		return nil
	}

	_ = d.Set("ha_vip_id", haVipId)
	_ = d.Set("instance_id", instanceId)
	return nil
}

func resourceZenlayerCloudZecHaVipAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_havip_association.delete")()

	haVipId := d.Get("ha_vip_id").(string)
	instanceId := d.Get("instance_id").(string)

	request := zec.NewUnassociateHaVipRequest()
	request.HaVipId = sdkcommon.String(haVipId)
	request.InstanceId = sdkcommon.String(instanceId)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		_, err := meta.(*connectivity.ZenlayerCloudClient).WithZec2Client().UnassociateHaVip(request)
		if err != nil {
			sdkErr, ok := err.(*sdkcommon.ZenlayerCloudSdkError)
			if ok && (sdkErr.Code == common.ResourceNotFound || sdkErr.Code == INVALID_HAVIP_NOT_FOUND) {
				return nil
			}
			return common.RetryError(ctx, err, common.InternalServerError, sdkcommon.NetworkError)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

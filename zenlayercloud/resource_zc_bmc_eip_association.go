/*
Provides an eip resource associated with BMC instance.

Example Usage

```hcl
resource "zenlayercloud_bmc_eip_association" "foo" {
  eip_id      = "eipxxxxxx"
  instance_id = "instanceIdxxxxxx"
}
```

Import

Eip association can be imported using the id, e.g.

```
$ terraform import zenlayercloud_bmc_eip_association.bar eipIdxxxxxx:instanceIdxxxxxxx
```
*/
package zenlayercloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	bmc "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/bmc20221120"
	"time"
)

func resourceZenlayerCloudEipAssociationAssociation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudEipAssociationCreate,
		ReadContext:   resourceZenlayerCloudEipAssociationRead,
		DeleteContext: resourceZenlayerCloudEipAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"eip_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of EIP.",
			},
			"instance_id": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The instance id going to bind with the EIP.",
			},
		},
	}
}

func resourceZenlayerCloudEipAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	association, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return diag.FromErr(err)
	}

	request := bmc.NewUnassociateEipAddressRequest()
	request.EipId = association[0]

	if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		_, errRet := bmcService.client.WithBmcClient().UnassociateEipAddress(request)
		if errRet != nil {
			return retryError(ctx, errRet)
		}
		return nil
	}); err != nil {
		return diag.FromErr(err)
	}

	stateConf := BuildEipState(bmcService, association[0], ctx, d)

	eipState, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for association (%s) to be deleted", d.Id())
	}
	if eipState.(*bmc.EipAddress).EipStatus != BmcEipStatusAvailable {
		return diag.Errorf("disassociate eip (%s) failed, current status is %s", request.EipId, eipState.(*bmc.EipAddress).EipStatus)
	}
	return nil
}

func resourceZenlayerCloudEipAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	eipId := d.Get("eip_id").(string)
	instanceId := d.Get("instance_id").(string)
	var eip *bmc.EipAddress
	var errRet error

	err := resource.RetryContext(ctx, readRetryTimeout, func() *resource.RetryError {
		eip, errRet = bmcService.DescribeEipAddressById(ctx, eipId)
		if errRet != nil {
			return retryError(ctx, errRet, InternalServerError)
		}
		if eip == nil {
			return resource.NonRetryableError(fmt.Errorf("eip is not found"))
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if eip.InstanceId != instanceId {
		if eip.EipStatus != BmcEipStatusAvailable {
			return diag.FromErr(fmt.Errorf("eip (%s) status is illegal %s", eipId, eip.EipStatus))
		}

		request := bmc.NewAssociateEipAddressRequest()
		request.EipId = eipId
		request.InstanceId = d.Get("instance_id").(string)

		if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
			_, errRet := bmcService.client.WithBmcClient().AssociateEipAddress(request)
			if errRet != nil {
				return retryError(ctx, errRet)
			}
			return nil
		}); err != nil {
			return diag.FromErr(err)
		}
	}
	d.SetId(eipId + ":" + instanceId)

	stateConf := BuildEipState(bmcService, eipId, ctx, d)

	eipState, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for association (%s) to be created: %v", d.Id(), err)
	}
	if eipState.(*bmc.EipAddress).EipStatus == BmcEipStatusAvailable {
		return diag.Errorf("associate instance (%s) to eip (%s) failed", instanceId, eipId)
	}

	return resourceZenlayerCloudEipAssociationRead(ctx, d, meta)
}

func resourceZenlayerCloudEipAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	association, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return diag.FromErr(err)
	}

	var eipAddress *bmc.EipAddress
	var errRet error

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		eipAddress, errRet = bmcService.DescribeEipAddressById(ctx, association[0])
		if errRet != nil {
			return retryError(ctx, errRet)
		}
		if eipAddress == nil {
			d.SetId("")
		}
		if eipAddress != nil && ipIsOperating(eipAddress.EipStatus) {
			return resource.RetryableError(fmt.Errorf("waiting for eip %s operation", eipAddress.EipId))
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("eip_id", association[0])
	_ = d.Set("instance_id", association[1])

	return diags

}

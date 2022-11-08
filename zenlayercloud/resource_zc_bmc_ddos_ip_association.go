/*
Provides an DDoS IP resource associated with BMC instance.

Example Usage

```hcl
resource "zenlayercloud_bmc_ddos_ip_association" "foo" {
  ddos_ip_id      = "ddosIpIdxxxxxx"
  instance_id = "instanceIdxxxxxx"
}
```

Import

DDoS IP association can be imported using the id, e.g.

```
$ terraform import zenlayercloud_bmc_ddos_ip_association.bar ddosIpIdxxxxxx:instanceIdxxxxxxx
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

func resourceZenlayerCloudDdosIpAssociationAssociation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudDdosIpAssociationCreate,
		ReadContext:   resourceZenlayerCloudDdosIpAssociationRead,
		DeleteContext: resourceZenlayerCloudDdosIpAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"ddos_ip_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of DDoS IP.",
			},
			"instance_id": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The instance id going to bind with the DDoS IP.",
			},
		},
	}
}

func resourceZenlayerCloudDdosIpAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	association, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return diag.FromErr(err)
	}

	request := bmc.NewUnassociateDdosIpAddressRequest()
	request.DdosIpId = association[0]

	if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		_, errRet := bmcService.client.WithBmcClient().UnassociateDdosIpAddress(request)
		if errRet != nil {
			return retryError(ctx, errRet)
		}
		return nil
	}); err != nil {
		return diag.FromErr(err)
	}

	stateConf := BuildDdosIpState(bmcService, association[0], ctx, d)

	ddosIpState, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for association (%s) to be deleted: %v", d.Id(), err)
	}
	if ddosIpState == nil {
		return diag.Errorf("disassociate ddos ip (%s) failed as ddos ip not found", request.DdosIpId)
	}
	if ddosIpState.(*bmc.DdosIpAddress).DdosIpStatus != BmcEipStatusAvailable {
		return diag.Errorf("disassociate ddos ip (%s) failed, current status is %s", request.DdosIpId, ddosIpState.(*bmc.DdosIpAddress).DdosIpStatus)
	}
	return nil
}

func resourceZenlayerCloudDdosIpAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	ddosIpId := d.Get("ddos_ip_id").(string)
	instanceId := d.Get("instance_id").(string)
	var ddosIp *bmc.DdosIpAddress
	var errRet error

	err := resource.RetryContext(ctx, readRetryTimeout, func() *resource.RetryError {
		ddosIp, errRet = bmcService.DescribeDdosIpAddressById(ctx, ddosIpId)
		if errRet != nil {
			return retryError(ctx, errRet, InternalServerError)
		}
		if ddosIp == nil {
			return resource.NonRetryableError(fmt.Errorf("ddos ip is not found"))
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if ddosIp.InstanceId != instanceId {
		if ddosIp.DdosIpStatus != BmcEipStatusAvailable {
			return diag.FromErr(fmt.Errorf("ddos ip (%s) status is illegal %s", ddosIpId, ddosIp.DdosIpStatus))
		}

		request := bmc.NewAssociateDdosIpAddressRequest()
		request.DdosIpId = ddosIpId
		request.InstanceId = instanceId

		if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
			_, errRet := bmcService.client.WithBmcClient().AssociateDdosIpAddress(request)
			if errRet != nil {
				return retryError(ctx, errRet)
			}
			return nil
		}); err != nil {
			return diag.FromErr(err)
		}
	}
	d.SetId(ddosIpId + ":" + instanceId)

	stateConf := BuildDdosIpState(bmcService, ddosIpId, ctx, d)

	ddosIpState, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for association (%s) to be created: %v", d.Id(), err)
	}
	if ddosIpState == nil {
		return diag.Errorf("associate instance to ddosIp (%s) failed as ddos ip not found", ddosIpId)
	}

	if ddosIpState.(*bmc.DdosIpAddress).DdosIpStatus == BmcEipStatusAvailable {
		return diag.Errorf("associate instance (%s) to ddosIp (%s) failed", instanceId, ddosIpId)
	}

	return resourceZenlayerCloudDdosIpAssociationRead(ctx, d, meta)
}

func resourceZenlayerCloudDdosIpAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	association, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return diag.FromErr(err)
	}

	var ddosIpAddress *bmc.DdosIpAddress
	var errRet error

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		ddosIpAddress, errRet = bmcService.DescribeDdosIpAddressById(ctx, association[0])
		if errRet != nil {
			return retryError(ctx, errRet)
		}
		if ddosIpAddress == nil {
			d.SetId("")
		}
		if ddosIpAddress != nil && ipIsOperating(ddosIpAddress.DdosIpStatus) {
			return resource.RetryableError(fmt.Errorf("waiting ddos ip %s operation", ddosIpAddress.DdosIpStatus))
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("ddos_ip_id", association[0])
	_ = d.Set("instance_id", association[1])

	return diags

}

func BuildDdosIpState(bmcService BmcService, ddosIpId string, ctx context.Context, d *schema.ResourceData) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending: []string{
			BmcEipStatusAssociating,
			BmcEipStatusCreating,
			BmcEipStatusUnAssociating,
		},
		Target: []string{
			BmcEipStatusCreateFailed,
			BmcEipStatusAssociated,
			BmcEipStatusAvailable,
		},
		Refresh:        bmcService.InstanceDdosIpStateRefreshFunc(ctx, ddosIpId),
		Timeout:        d.Timeout(schema.TimeoutRead) - time.Minute,
		Delay:          10 * time.Second,
		MinTimeout:     5 * time.Second,
		NotFoundChecks: 3,
	}
}

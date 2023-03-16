/*
Provides a resource to create a security group attachment

Example Usage

```hcl
resource "zenlayercloud_security_group_attachment" "foo" {
  security_group_id = "12364246"
  instance_id       = "62343412426423623"
}
```

Import

Security group attachment can be imported using the id, e.g.

```
terraform import zenlayercloud_security_group_attachment.security_group_attachment securityGroupId:instanceId
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
	vm "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/vm20230313"
	"log"
	"strings"
	"time"
)

func resourceZenlayerCloudSecurityGroupAttachment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudSecurityGroupAttachmentCreate,
		ReadContext:   resourceZenlayerCloudSecurityGroupAttachmentRead,
		DeleteContext: resourceZenlayerCloudSecurityGroupAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"security_group_id": {
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
				Description: "The ID of security group.",
			},

			"instance_id": {
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
				Description: "The id of instance.",
			},
		},
	}
}

func resourceZenlayerCloudSecurityGroupAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	idSplit := strings.Split(d.Id(), ":")
	if len(idSplit) != 2 {
		return diag.FromErr(fmt.Errorf("id is broken, %s", d.Id()))
	}

	request := vm.NewUnAssociateSecurityGroupInstanceRequest()
	request.SecurityGroupId = idSplit[0]
	request.InstanceId = idSplit[1]
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := meta.(*connectivity.ZenlayerCloudClient).WithVmClient().UnAssociateSecurityGroupInstance(request)
		if err != nil {
			tflog.Info(ctx, "Fail to unassociate security group instance.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": toJsonString(request),
				"err":     err.Error(),
			})
			return retryError(ctx, err)
		}

		tflog.Info(ctx, "Unassociate security group instance success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  toJsonString(request),
			"response": toJsonString(response),
		})

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceZenlayerCloudSecurityGroupAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	securityGroupId := d.Get("security_group_id").(string)
	instanceId := d.Get("instance_id").(string)

	request := vm.NewAssociateSecurityGroupInstanceRequest()
	request.SecurityGroupId = securityGroupId
	request.InstanceId = instanceId

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := meta.(*connectivity.ZenlayerCloudClient).WithVmClient().AssociateSecurityGroupInstance(request)
		if err != nil {
			tflog.Info(ctx, "Fail to associate security group instance.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": toJsonString(request),
				"err":     err.Error(),
			})
			return retryError(ctx, err)
		}

		tflog.Info(ctx, "Associate security group instance success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  toJsonString(request),
			"response": toJsonString(response),
		})

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(securityGroupId + ":" + instanceId)

	return resourceZenlayerCloudSecurityGroupAttachmentRead(ctx, d, meta)
}

func resourceZenlayerCloudSecurityGroupAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	vmService := VmService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	idSplit := strings.Split(d.Id(), ":")
	if len(idSplit) != 2 {
		return diag.FromErr(fmt.Errorf("id is broken, %s", d.Id()))
	}

	securityGroupId := idSplit[0]
	instanceId := idSplit[1]

	var securityGroup *vm.SecurityGroupInfo
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		securityGroup, errRet = vmService.DescribeSecurityGroupById(ctx, securityGroupId)
		request := vm.NewDescribeSecurityGroupsRequest()
		request.SecurityGroupIds = []string{securityGroupId}
		if errRet != nil {
			return retryError(ctx, errRet)
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

	isAttached := false

	if len(securityGroup.InstanceIds) > 0 {
		for _, rl := range securityGroup.InstanceIds {
			if *rl == instanceId {
				isAttached = true
				break
			}
		}
	}

	if !isAttached {
		d.SetId("")
		log.Printf("The security group get from api does not match with current instance %v", d.Id())
		return nil
	}

	_ = d.Set("instance_id", instanceId)
	_ = d.Set("security_group_id", securityGroupId)

	return diags
}

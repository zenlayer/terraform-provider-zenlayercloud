/*
Provides a resource to create security group.

Example Usage

```hcl
resource "zenlayercloud_security_group" "foo" {
  name       	= "example-name"
  description	= "example purpose"
}
```

Import

Security group can be imported, e.g.

```
$ terraform import zenlayercloud_security_group.security_group security_group_id
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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	vm "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/vm20230313"
	"time"
)

func resourceZenlayerCloudSecurityGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudSecurityGroupCreate,
		ReadContext:   resourceZenlayerCloudSecurityGroupRead,
		UpdateContext: resourceZenlayerCloudSecurityGroupUpdate,
		DeleteContext: resourceZenlayerCloudSecurityGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
				Description:  "The name of the security group.",
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(2, 256),
				Description:  "The name of the security group.",
			},
		},
	}
}

func resourceZenlayerCloudSecurityGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	securityGroupId := d.Id()

	vmService := VmService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	// wait until all instances unbind this security group
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		sucurityGroup, errRet := vmService.DescribeSecurityGroupById(ctx, securityGroupId)
		if errRet != nil {
			return retryError(ctx, errRet)
		}
		associateInstanceCount := len(sucurityGroup.InstanceIds)
		if associateInstanceCount == 0 {
			return nil
		}
		return resource.RetryableError(fmt.Errorf("security group %s still have %d instances", securityGroupId, associateInstanceCount))
	})

	if err != nil {
		return diag.FromErr(err)
	}

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := vmService.DeleteSecurityGroup(ctx, securityGroupId)
		if errRet != nil {
			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if !ok {
				return retryError(ctx, errRet, InternalServerError)
			}
			if ee.Code == ResourceNotFound {
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

func resourceZenlayerCloudSecurityGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	//var _ diag.Diagnostics
	vmService := VmService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	securityGroupId := d.Id()
	if d.HasChanges("name", "description") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			err := vmService.ModifySecurityGroupAttribute(ctx, securityGroupId, d.Get("name").(string), d.Get("description").(string))
			if err != nil {
				return retryError(ctx, err, InternalServerError, common.NetworkError)
			}
			return nil
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceZenlayerCloudSecurityGroupRead(ctx, d, meta)
}

func resourceZenlayerCloudSecurityGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	request := vm.NewCreateSecurityGroupRequest()
	request.SecurityGroupName = d.Get("name").(string)
	request.Description = d.Get("description").(string)

	securityGroupId := ""

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := meta.(*connectivity.ZenlayerCloudClient).WithVmClient().CreateSecurityGroup(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create vm security group.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": toJsonString(request),
				"err":     err.Error(),
			})
			return retryError(ctx, err)
		}

		tflog.Info(ctx, "Create security group success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  toJsonString(request),
			"response": toJsonString(response),
		})

		if response.Response.SecurityGroupId == "" {
			err = fmt.Errorf("security group id is nil")
			return resource.NonRetryableError(err)
		}
		securityGroupId = response.Response.SecurityGroupId

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(securityGroupId)

	return resourceZenlayerCloudSecurityGroupRead(ctx, d, meta)
}

func resourceZenlayerCloudSecurityGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	securityGroupId := d.Id()

	vmService := VmService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var securityGroup *vm.SecurityGroupInfo
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		securityGroup, errRet = vmService.DescribeSecurityGroupById(ctx, securityGroupId)
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

	// security group info
	d.SetId(securityGroup.SecurityGroupId)
	_ = d.Set("name", securityGroup.SecurityGroupName)
	_ = d.Set("description", securityGroup.Description)

	return diags

}

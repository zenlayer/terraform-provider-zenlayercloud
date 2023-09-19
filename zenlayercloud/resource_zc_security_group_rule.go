/*
Provides a resource to create security group rule.

~> **NOTE:** Single security rule is hardly ordered, use zenlayercloud_security_group_rule_set instead.

Example Usage

```hcl

	resource "zenlayercloud_security_group" "foo" {
	  name        = "example-name"
	  description = "example purpose"
	}

	resource "zenlayercloud_security_group_rule" "bar" {
	  security_group_id = zenlayercloud_security_group.foo.id
	  direction         = "ingress"
	  policy            = "accept"
	  cidr_ip    		= "10.0.0.0/16"
	  ip_protocol       = "tcp"
	  port_range        = "80"
	}

```
*/
package zenlayercloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	vm "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/vm20230313"
	"time"
)

func resourceZenlayerCloudSecurityGroupRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudSecurityGroupRuleCreate,
		ReadContext:   resourceZenlayerCloudSecurityGroupRuleRead,
		DeleteContext: resourceZenlayerCloudSecurityGroupRuleDelete,

		Schema: map[string]*schema.Schema{
			"security_group_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the security group to be queried.",
			},
			"direction": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(SecurityGroupRuleDirection, false),
				Description:  "The direction of the rule.",
			},
			"policy": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "accept",
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(SecurityGroupRulePolicy, false),
				Description:  "The policy of the rule, currently only `accept` is supported.",
			},
			"ip_protocol": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringInSlice(SecurityGroupRuleIpProtocol, false),
				Description:  "The protocol of the rule.",
			},
			"port_range": {
				Type:     schema.TypeString,
				Required: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if (old+"/"+old) == new || (new+"/"+new) == old {
						return true
					}
					return old == new
				},
				ForceNew:    true,
				Description: "The port range of the rule.",
			},
			"cidr_ip": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The cidr ip of the rule.",
			},
		},
	}
}

func resourceZenlayerCloudSecurityGroupRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer logElapsed(ctx, "resource.zenlayercloud_security_group_rule.delete")()

	vmService := VmService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	ruleId := d.Id()

	info, ret := parseSecurityGroupRuleId(ruleId)
	if ret != nil {
		return diag.FromErr(ret)
	}

	request := vm.NewRevokeSecurityGroupRulesRequest()
	request.SecurityGroupId = info.SecurityGroupId
	ruleRequest := convertRuleInfo2RuleRequest(info)
	request.RuleInfos = []*vm.RuleInfo{ruleRequest}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		_, errRet := vmService.client.WithVmClient().RevokeSecurityGroupRules(request)
		if errRet != nil {
			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if !ok {
				return retryError(ctx, errRet, InternalServerError)
			}
			if ee.Code == ResourceNotFound {
				// security group rule doesn't exist
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

func resourceZenlayerCloudSecurityGroupRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vmService := VmService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	id := ""

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		info := securityGroupRuleBasicInfo{
			SecurityGroupId: d.Get("security_group_id").(string),
			Direction:       d.Get("direction").(string),
			IpProtocol:      d.Get("ip_protocol").(string),
			PortRange:       d.Get("port_range").(string),
			CidrIp:          d.Get("cidr_ip").(string),
			Policy:          d.Get("policy").(string),
		}

		ruleId, err := vmService.CreateSecurityGroupRule(ctx, info)
		if err != nil {
			return retryError(ctx, err)
		}

		id = ruleId
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)

	return resourceZenlayerCloudSecurityGroupRuleRead(ctx, d, meta)
}

func resourceZenlayerCloudSecurityGroupRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	ruleId := d.Id()

	vmService := VmService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		securityGroupId, rule, err := vmService.DescribeSecurityGroupRule(ruleId)
		if err != nil {
			return retryError(ctx, err)
		}

		if rule == nil {
			d.SetId("")
			return nil
		}

		_ = d.Set("security_group_id", securityGroupId)
		_ = d.Set("direction", rule.Direction)
		_ = d.Set("ip_protocol", rule.IpProtocol)
		_ = d.Set("port_range", rule.PortRange)
		_ = d.Set("cidr_ip", rule.CidrIp)
		_ = d.Set("policy", rule.Policy)

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

package zec

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"
	"time"
)

func ResourceZenlayerCloudZecSecurityGroupRuleSet() *schema.Resource {
	ruleElem := map[string]*schema.Schema{
		"policy": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"accept", "deny"}, false),
			Description:  "Rule policy of security group. Valid values: `accept` and `deny`.",
		},
		"priority": {
			Type:         schema.TypeInt,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.IntBetween(1, 100),
			Description:  "Priority of the security group rule. The smaller the value, the higher the priority. Valid values: `1` to `100`. Default is `1`.",
		},
		"description": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Description of the security group rule.",
		},
		"cidr_block": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "An IP address network or CIDR segment.",
		},
		"protocol": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"tcp", "udp", "gre", "icmpv6", "icmp", "all"}, false),
			Description:  "Type of IP protocol. Valid values: `tcp`, `udp`, `icmp`, `gre`, `icmpv6` and `all`.",
		},
		"port": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Range of the port. The available value can be a single port, or a port range, or `-1` which means all. E.g. `80`, `80,90`, `80-90` or `all`. Note: If the `Protocol` value is set to `all`, the `Port` value needs to be set to `-1`.",
		},
	}
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZecSecurityGroupRuleSetCreate,
		ReadContext:   resourceZenlayerCloudZecSecurityGroupRuleSetRead,
		UpdateContext: resourceZenlayerCloudZecSecurityGroupRuleSetUpdate,
		DeleteContext: resourceZenlayerCloudZecSecurityGroupRuleSetDelete,

		Schema: map[string]*schema.Schema{
			"security_group_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the security group.",
			},
			"ingress": {
				Type:         schema.TypeSet,
				Elem:         &schema.Resource{Schema: ruleElem},
				Optional:     true,
				Description:  "Set of ingress rule.",
			},
			"egress": {
				Type:         schema.TypeSet,
				Elem:         &schema.Resource{Schema: ruleElem},
				Optional:     true,
				Description:  "Set of egress rule.",
			},
		},
	}
}

func resourceZenlayerCloudZecSecurityGroupRuleSetUpdate(ctx context.Context, d *schema.ResourceData, i interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_security_group_rule_set.update")()

	zecService := ZecService{
		client: i.(*connectivity.ZenlayerCloudClient),
	}

	mutableArgs := []string{"ingress", "egress"}
	var needChange = false
	for _, v := range mutableArgs {
		if d.HasChange(v) {
			needChange = true
			break
		}
	}
	if needChange {
		request := zec.NewConfigureSecurityGroupRulesRequest()
		request.SecurityGroupId = common.String(d.Id())
		request.RuleInfos = []*zec.SecurityGroupRuleInfo{}

		if v, ok := d.GetOk("ingress"); ok {
			ingressRules := v.(*schema.Set).List()
			rules := unmarshalSecurityRules(ingressRules, "ingress")
			request.RuleInfos = append(request.RuleInfos, rules...)
		}

		if v, ok := d.GetOk("egress"); ok {
			ingressRules := v.(*schema.Set).List()
			rules := unmarshalSecurityRules(ingressRules, "egress")
			request.RuleInfos = append(request.RuleInfos, rules...)
		}

		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {

			_, err := zecService.client.WithZec2Client().ConfigureSecurityGroupRules(request)
			if err != nil {
				return common2.RetryError(ctx, err)
			}

			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return resourceZenlayerCloudZecSecurityGroupRuleSetRead(ctx, d, i)
}

func resourceZenlayerCloudZecSecurityGroupRuleSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_security_group_rule_set.delete")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	securityGroupId := d.Id()

	request := zec.NewConfigureSecurityGroupRulesRequest()
	request.SecurityGroupId = &securityGroupId
	request.RuleInfos = []*zec.SecurityGroupRuleInfo{}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		_, errRet := zecService.client.WithZec2Client().ConfigureSecurityGroupRules(request)
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

func resourceZenlayerCloudZecSecurityGroupRuleSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_security_group_rule_set.create")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	securityGroupId := d.Get("security_group_id").(string)

	request := zec.NewConfigureSecurityGroupRulesRequest()
	request.SecurityGroupId = common.String(securityGroupId)
	request.RuleInfos = []*zec.SecurityGroupRuleInfo{}

	if v, ok := d.GetOk("ingress"); ok {
		ingressRules := v.(*schema.Set).List()
		rules := unmarshalSecurityRules(ingressRules, "ingress")
		request.RuleInfos = append(request.RuleInfos, rules...)
	}

	if v, ok := d.GetOk("egress"); ok {
		ingressRules := v.(*schema.Set).List()
		rules := unmarshalSecurityRules(ingressRules, "egress")
		request.RuleInfos = append(request.RuleInfos, rules...)
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {

		_, err := zecService.client.WithZec2Client().ConfigureSecurityGroupRules(request)
		if err != nil {
			return common2.RetryError(ctx, err)
		}

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(securityGroupId)

	return resourceZenlayerCloudZecSecurityGroupRuleSetRead(ctx, d, meta)
}

func unmarshalSecurityRules(rules []interface{}, direction string) []*zec.SecurityGroupRuleInfo {
	var result []*zec.SecurityGroupRuleInfo
	for i := range rules {
		rule := rules[i].(map[string]interface{})
		ruleInfo := &zec.SecurityGroupRuleInfo{
			Direction:  &direction,
			Policy:     common.String(rule["policy"].(string)),
			CidrIp:     common.String(rule["cidr_block"].(string)),
			PortRange:  common.String(rule["port"].(string)),
			IpProtocol: common.String(rule["protocol"].(string)),
		}
		desc := rule["description"].(string)
		if desc != "" {
			ruleInfo.Desc = common.String(desc)
		}
		priority := rule["priority"].(int)
		if priority != 0 {
			ruleInfo.Priority = common.Integer(priority)
		}
		result = append(result, ruleInfo)
	}
	return result
}

func resourceZenlayerCloudZecSecurityGroupRuleSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_security_group_rule_set.read")()

	var diags diag.Diagnostics

	securityGroupId := d.Id()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		ingress, egress, err := zecService.DescribeSecurityGroupRules(ctx, securityGroupId)
		if err != nil {
			return common2.RetryError(ctx, err)
		}

		if ingress != nil {
			_ = d.Set("ingress", marshalSecurityRules(ingress))
		}

		if ingress != nil {
			_ = d.Set("egress", marshalSecurityRules(egress))
		}

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func marshalSecurityRules(rules []*zec.SecurityGroupRuleInfo) interface{} {
	result := make([]interface{}, 0, len(rules))
	for i := range rules {
		result = append(result, map[string]interface{}{
			"protocol":    rules[i].IpProtocol,
			"port":        rules[i].PortRange,
			"policy":      rules[i].Policy,
			"priority":    rules[i].Priority,
			"cidr_block":  rules[i].CidrIp,
			"description": rules[i].Desc,
		})
	}
	return result
}


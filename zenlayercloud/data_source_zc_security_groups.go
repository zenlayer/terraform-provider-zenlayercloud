/*
 Use this data source to query detailed information of security groups.

Example Usage

```hcl
data "zenlayercloud_security_groups" "sg1" {
}

data "zenlayercloud_security_groups" "sg2" {
  name = "example_name"
}
```
*/
package zenlayercloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	vm "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/vm20230313"
	"time"
)

func dataSourceZenlayerCloudSecurityGroups() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudSecurityGroupsRead,

		Schema: map[string]*schema.Schema{
			"security_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the security group to be queried..",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the security group to be queried..",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed value
			"security_groups": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of security group. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the security group.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the security group.",
						},
						"instance_ids": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "Instance ids of the security group.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Creation time of the security group.",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Description of the security group.",
						},
						// rule info
						"rule_infos": {
							Type:        schema.TypeList,
							Description: "Rules set of the security.",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"direction": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The direction of the rule.",
									},
									"policy": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The policy of the rule.",
									},
									"ip_protocol": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The protocol of the rule.",
									},
									"port_range": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The port range of the rule.",
									},
									"cidr_ip": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The cidr ip of the rule.",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudSecurityGroupsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "data_source.zenlayercloud_security_groups.read")()

	vmService := VmService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	request := &SecurityGroupFilter{}

	if v, ok := d.GetOk("security_group_id"); ok {
		request.SecurityGroupId = v.(string)
	}
	if v, ok := d.GetOk("name"); ok {
		request.Name = v.(string)
	}

	var securityGroups []*vm.SecurityGroupInfo

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		securityGroups, e = vmService.DescribeSecurityGroupsByFilter(request)
		if e != nil {
			return common.RetryError(ctx, e, common.InternalServerError)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	securityGroupList := make([]map[string]interface{}, 0, len(securityGroups))
	ids := make([]string, 0, len(securityGroups))
	for _, securityGroup := range securityGroups {
		mapping := map[string]interface{}{
			"security_group_id": securityGroup.SecurityGroupId,
			"name":              securityGroup.SecurityGroupName,
			"instance_ids":      securityGroup.InstanceIds,
			"create_time":       securityGroup.CreateTime,
			"description":       securityGroup.Description,
		}
		if securityGroup.RuleInfos != nil && len(securityGroup.RuleInfos) > 0 {
			mapping["rule_infos"] = map2RuleInfo(securityGroup.RuleInfos)
		}
		securityGroupList = append(securityGroupList, mapping)
		ids = append(ids, securityGroup.SecurityGroupId)
	}

	d.SetId(common.DataResourceIdHash(ids))
	err = d.Set("security_groups", securityGroupList)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := common.WriteToFile(output.(string), securityGroupList); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func map2RuleInfo(rules []*vm.RuleInfo) []interface{} {
	var res = make([]interface{}, 0, len(rules))

	for _, rule := range rules {
		m := make(map[string]interface{}, 7)
		m["direction"] = rule.Direction
		m["policy"] = rule.Policy
		m["ip_protocol"] = rule.IpProtocol
		m["port_range"] = rule.PortRange
		m["cidr_ip"] = rule.CidrIp
		res = append(res, m)
	}
	return res
}

type SecurityGroupFilter struct {
	SecurityGroupId string
	Name            string
}

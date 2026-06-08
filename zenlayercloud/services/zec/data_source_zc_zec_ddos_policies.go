package zec

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	common2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	zec2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"
)

func DataSourceZenlayerCloudZecDDoSPolicies() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudZecDDoSPoliciesRead,
		Schema: map[string]*schema.Schema{
			"policy_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Filter by a list of DDoS policy IDs. Maximum 100.",
			},
			"policy_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter by policy name. Fuzzy search is supported.",
			},
			"result": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of DDoS protection policies.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the DDoS policy.",
						},
						"policy_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the DDoS policy.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The time when the DDoS policy was created.",
						},
						"resource_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The resource group ID the policy belongs to.",
						},
						"resource_group_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The resource group name the policy belongs to.",
						},
						"tags": {
							Type:        schema.TypeMap,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Tags associated with the policy.",
						},
					},
				},
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results to a local file.",
			},
		},
	}
}

func dataSourceZenlayerCloudZecDDoSPoliciesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "data_source.zenlayercloud_zec_ddos_policies.read")()

	zecService := ZecService{client: meta.(*connectivity.ZenlayerCloudClient)}

	filter := &DDoSPolicyFilter{}
	if v, ok := d.GetOk("policy_ids"); ok {
		filter.PolicyIds = common.ToStringList(v.(*schema.Set).List())
	}
	if v, ok := d.GetOk("policy_name"); ok {
		filter.PolicyName = v.(string)
	}

	var result []*zec2.PolicyInfo
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		policies, e := zecService.DescribeDDoSPoliciesByFilter(ctx, filter)
		if e != nil {
			return common.RetryError(ctx, e, common.InternalServerError)
		}
		result = policies
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	ids := make([]string, 0, len(result))
	policyList := make([]map[string]interface{}, 0, len(result))

	for _, p := range result {
		if p.PolicyId == nil {
			continue
		}
		mapping := map[string]interface{}{
			"id":                  *p.PolicyId,
			"policy_name":         common2.ToString(p.PolicyName),
			"create_time":         common2.ToString(p.CreateTime),
			"resource_group_id":   common2.ToString(p.ResourceGroupId),
			"resource_group_name": common2.ToString(p.ResourceGroupName),
		}
		if p.Tags != nil && len(p.Tags.Tags) > 0 {
			tagsMap := make(map[string]interface{})
			for _, t := range p.Tags.Tags {
				if t.Key != nil {
					v := ""
					if t.Value != nil {
						v = *t.Value
					}
					tagsMap[*t.Key] = v
				}
			}
			mapping["tags"] = tagsMap
		} else {
			mapping["tags"] = map[string]interface{}{}
		}
		policyList = append(policyList, mapping)
		ids = append(ids, *p.PolicyId)
	}

	d.SetId(common.DataResourceIdHash(ids))
	_ = d.Set("result", policyList)

	if output, ok := d.GetOk("result_output_file"); ok && output.(string) != "" {
		if err := common.WriteToFile(output.(string), policyList); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

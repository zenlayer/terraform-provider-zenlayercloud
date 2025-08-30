package zec

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20240401"
	"regexp"
	"time"
)

func DataSourceZenlayerCloudZecAutoSnapshotPolicies() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudZecAutoSnapshotPoliciesRead,

		Schema: map[string]*schema.Schema{
			"ids": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "IDs of the auto snapshot policy to be queried.",
			},
			"availability_zone": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The availability zone of the auto snapshot policy to be queried.",
			},
			"name_regex": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the auto snapshot policy to be queried.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of resource group grouped auto snapshot policy to be queried.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed value
			"auto_snapshot_policies": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of auto snapshot policy. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the auto snapshot policy.",
						},
						"availability_zone": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The availability zone of the auto snapshot policy.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the auto snapshot policy.",
						},
						"repeat_week_days": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeInt},
							Description: "The days of week when the auto snapshot policy is triggered. Valid values: 1-7.",
						},
						"hours": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeInt},
							Description: "The hours of day when the auto snapshot policy is triggered.",
						},
						"retention_days": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Retention days of the auto snapshot policy.",
						},
						"disk_num": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of disks associated with this auto snapshot policy.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Creation time of the auto snapshot policy.",
						},
						"resource_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of resource group.",
						},
						"resource_group_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of resource group.",
						},
						"disk_ids": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "List of disk IDs associated with this auto snapshot policy.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudZecAutoSnapshotPoliciesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "data_source.zenlayercloud_zec_auto_snapshot_policies.read")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	request := &ZecAutoSnapshotPolicyFilter{}

	if v, ok := d.GetOk("ids"); ok {
		ids := v.(*schema.Set).List()
		if len(ids) > 0 {
			request.AutoSnapshotPolicyIds = common2.ToStringList(ids)
		}
	}

	if v, ok := d.GetOk("availability_zone"); ok {
		request.ZoneId = v.(string)
	}

	var nameRegex *regexp.Regexp
	var errRet error

	if v, ok := d.GetOk("name_regex"); ok {
		nameRegex, errRet = regexp.Compile(v.(string))
		if errRet != nil {
			return diag.Errorf("name_regex format error,%s", errRet.Error())
		}
	}

	if v, ok := d.GetOk("resource_group_id"); ok {
		request.ResourceGroupId = v.(string)
	}

	var result []*zec.AutoSnapshotPolicy

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		result, e = zecService.DescribeAutoSnapshotPolicies(ctx, request)
		if e != nil {
			return common2.RetryError(ctx, e, common2.InternalServerError)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	policyList := make([]map[string]interface{}, 0, len(result))

	ids := make([]string, 0, len(result))
	for _, policy := range result {
		if nameRegex != nil && !nameRegex.MatchString(*policy.AutoSnapshotPolicyName) {
			continue
		}
		mapping := map[string]interface{}{
			"id":                  policy.AutoSnapshotPolicyId,
			"availability_zone":   policy.ZoneId,
			"name":                policy.AutoSnapshotPolicyName,
			"repeat_week_days":    policy.RepeatWeekDays,
			"hours":               policy.Hours,
			"retention_days":      policy.RetentionDays,
			"disk_num":            policy.DiskNum,
			"create_time":         policy.CreateTime,
			"resource_group_id":   policy.ResourceGroup.ResourceGroupId,
			"resource_group_name": policy.ResourceGroup.ResourceGroupName,
			"disk_ids":         policy.DiskIdSet,
		}
		policyList = append(policyList, mapping)
		ids = append(ids, *policy.AutoSnapshotPolicyId)
	}

	d.SetId(common2.DataResourceIdHash(ids))
	err = d.Set("auto_snapshot_policies", policyList)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := common2.WriteToFile(output.(string), policyList); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

type ZecAutoSnapshotPolicyFilter struct {
	AutoSnapshotPolicyIds []string
	ZoneId                string
	ResourceGroupId       string
}

package zec

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20240401"
	"regexp"
	"time"
)

func DataSourceZenlayerCloudZecSnapshots() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudZecSnapshotsRead,

		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The availability zone of the snapshot to be queried.",
			},
			"disk_ids": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "IDs of the disk to be queried.",
			},
			"snapshot_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The type of the snapshot to be queried. Valid values: `Auto`, `Manual`.",
			},
			"ids": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "IDs of the snapshots to be queried.",
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsValidRegExp,
				Description:  "A regex string to apply to the snapshot name.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of resource group grouped snapshot to be queried.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed value
			"snapshots": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of snapshot. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the snapshot.",
						},
						"availability_zone": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The availability zone of snapshot.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the snapshot.",
						},
						"snapshot_type": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The type of the snapshot to be queried. Valid values: `Auto`, `Manual`.",
						},
						"disk_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of disk that the snapshot is created from.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Creation time of the snapshot.",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Status of snapshot. Valid values: `CREATING`, `AVAILABLE`, `FAILED`, `ROLLING_BACK`, `DELETING`.",
						},
						"retention_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Retention time of snapshot.",
						},
						"resource_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of resource group grouped snapshot.",
						},
						"resource_group_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The Name of resource group grouped snapshot.",
						},
						"disk_ability": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the snapshot can be used to create a disk.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudZecSnapshotsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "data_source.zenlayercloud_zec_snapshots.read")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var nameRegex *regexp.Regexp
	request := &ZecSnapshotFilter{}

	if v, ok := d.GetOk("availability_zone"); ok {
		request.ZoneId = v.(string)
	}

	if v, ok := d.GetOk("ids"); ok {
		ids := v.(*schema.Set).List()
		if len(ids) > 0 {
			request.SnapshotIds = common2.ToStringList(ids)
		}
	}

	if v, ok := d.GetOk("disk_ids"); ok {
		ids := v.(*schema.Set).List()
		if len(ids) > 0 {
			request.DiskIds = common2.ToStringList(ids)
		}
	}

	if v, ok := d.GetOk("snapshot_name"); ok {
		request.SnapshotName = v.(string)
	}

	if v, ok := d.GetOk("snapshot_type"); ok {
		request.SnapshotType = v.(string)
	}

	if v, ok := d.GetOk("resource_group_id"); ok {
		request.ResourceGroupId = v.(string)
	}

	if v, ok := d.GetOk("name_regex"); ok {
		var errRet error
		nameRegex, errRet = regexp.Compile(v.(string))
		if errRet != nil {
			return diag.Errorf("name_regex format error,%s", errRet.Error())
		}
	}

	var result []*zec.SnapshotInfo

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		result, e = zecService.DescribeSnapshots(ctx, request)
		if e != nil {
			return common2.RetryError(ctx, e, common2.InternalServerError)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	var snapshots []*zec.SnapshotInfo

	snapshotList := make([]map[string]interface{}, 0, len(snapshots))

	ids := make([]string, 0, len(snapshots))
	for _, snapshot := range result {
		if nameRegex != nil && !nameRegex.MatchString(*snapshot.SnapshotName) {
			continue
		}
		mapping := map[string]interface{}{
			"id":                  snapshot.SnapshotId,
			"availability_zone":   snapshot.ZoneId,
			"name":                snapshot.SnapshotName,
			"snapshot_type":       snapshot.SnapshotType,
			"disk_id":             snapshot.DiskId,
			"create_time":         snapshot.CreateTime,
			"status":              snapshot.Status,
			"resource_group_id":   snapshot.ResourceGroup.ResourceGroupId,
			"resource_group_name": snapshot.ResourceGroup.ResourceGroupName,
			"disk_ability":        snapshot.DiskAbility,
			"retention_time":      snapshot.RetentionTime,
		}
		snapshotList = append(snapshotList, mapping)
		ids = append(ids, *snapshot.SnapshotId)
	}

	d.SetId(common2.DataResourceIdHash(ids))
	err = d.Set("snapshots", snapshotList)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := common2.WriteToFile(output.(string), snapshotList); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

type ZecSnapshotFilter struct {
	SnapshotIds     []string
	DiskIds         []string
	ZoneId          string
	SnapshotName    string
	SnapshotType    string
	ResourceGroupId string
}

package zec

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	zec2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20240401"
	"regexp"
	"time"
)

func DataSourceZenlayerCloudZecDisks() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudZecDisksRead,

		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Zone of the disk to be queried.",
			},
			"ids": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "ids of the disk to be queried.",
			},
			"instance_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Query the disks which attached to the instance.",
			},
			"disk_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"SYSTEM", "DATA"}, false),
				Description:  "Type of the disk. Valid values: `SYSTEM`, `DATA`.",
			},
			"name_regex": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A regex string to apply to the disk list returned.",
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"ATTACHING", "AVAILABLE", "RECYCLED", "IN_USE", "DELETING", "CREATING", "DETACHING"}, false),
				Description:  "Status of disk to be queried.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of resource group grouped disk to be queried.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed value
			"disks": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of disk. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability_zone": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The availability zone of disk.",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the disk.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "name of the disk.",
						},
						"instance_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of instance that the disk attached to.",
						},
						"instance_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of instance that the disk attached to.",
						},
						"disk_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of the disk. Values are: `SYSTEM`, `DATA`.",
						},
						"disk_category": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The category of disk.",
						},
						"disk_size": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Size of the disk.",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Status of disk.",
						},
						"resource_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of resource group grouped disk to be queried.",
						},
						"resource_group_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The Name of resource group grouped disk to be queried.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Creation time of the disk.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudZecDisksRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "data_source.zenlayercloud_zec_disks.read")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var nameRegex *regexp.Regexp
	request := &ZecDiskFilter{}

	if v, ok := d.GetOk("disk_type"); ok {
		request.DiskType = v.(string)
	}

	if v, ok := d.GetOk("ids"); ok {
		ids := v.(*schema.Set).List()
		if len(ids) > 0 {
			request.Ids = common2.ToStringList(ids)
		}
	}

	if v, ok := d.GetOk("availability_zone"); ok {
		request.ZoneId = v.(string)
	}
	if v, ok := d.GetOk("instance_id"); ok {
		request.InstanceId = v.(string)
	}
	if v, ok := d.GetOk("status"); ok {
		request.Status = v.(string)
	}

	if v, ok := d.GetOk("name_regex"); ok {
		var errRet error
		nameRegex, errRet = regexp.Compile(v.(string))
		if errRet != nil {
			return diag.Errorf("name_regex format error,%s", errRet.Error())
		}
	}

	var result []*zec2.DiskInfo

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		result, e = zecService.DescribeDisks(ctx, request)
		if e != nil {
			return common2.RetryError(ctx, e, common2.InternalServerError)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	var disks []*zec2.DiskInfo
	if nameRegex != nil {
		for _, disk := range result {
			if disk.DiskName != "" && nameRegex.MatchString(disk.DiskName) {
				disks = append(disks, disk)
			}
		}
	} else {
		disks = result
	}

	diskList := make([]map[string]interface{}, 0, len(disks))

	ids := make([]string, 0, len(disks))
	for _, disk := range disks {
		mapping := map[string]interface{}{
			"availability_zone":   disk.ZoneId,
			"id":                  disk.DiskId,
			"name":                disk.DiskName,
			"disk_type":           disk.DiskType,
			"disk_size":           disk.DiskSize,
			"instance_id":         disk.InstanceId,
			"instance_name":       disk.InstanceName,
			"disk_category":       disk.DiskCategory,
			"create_time":         disk.CreateTime,
			"status":              disk.DiskStatus,
			"resource_group_id":   disk.ResourceGroupId,
			"resource_group_name": disk.ResourceGroupName,
		}
		diskList = append(diskList, mapping)
		ids = append(ids, disk.DiskId)
	}

	d.SetId(common2.DataResourceIdHash(ids))
	err = d.Set("disks", diskList)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := common2.WriteToFile(output.(string), diskList); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

type ZecDiskFilter struct {
	Ids        []string
	ZoneId     string
	DiskType   string
	DiskName   string
	InstanceId string
	Status     string
}

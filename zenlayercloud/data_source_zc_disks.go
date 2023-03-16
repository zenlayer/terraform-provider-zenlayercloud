/*
 Use this data source to query vm disk information.

Example Usage

```hcl
data "zenlayercloud_disks" "all" {
}

# filter system disk
data "zenlayercloud_disks" "system_disk" {
  disk_type = "SYSTEM"
}

#filter with name regex
data "zenlayercloud_disks" "name_disk" {
  name_regex = "disk20*"
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
	"regexp"
	"time"
)

func dataSourceZenlayerCloudDisks() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudDisksRead,

		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Zone of the disk to be queried.",
			},
			"portable": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether the disk is deleted with instance or not, true means not delete with instance, false otherwise.",
			},
			"id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "id of the disk to be queried.",
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
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"name_regex"},
				Description:   "Fuzzy query with this name.",
			},

			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"ATTACHING", "AVAILABLE", "RECYCLED", "IN_USE", "DELETING", "CREATING", "DETACHING"}, false),
				Description:  "Status of disk to be queried.",
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
						"portable": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the disk is deleted with instance or not, true means not delete with instance, false otherwise.",
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
						"charge_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Charge type of the disk. Values are: `PREPAID`, `POSTPAID`.",
						},
						"disk_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of the disk. Values are: `SYSTEM`, `DATA`.",
						},
						"disk_category": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The category of disk. Values are: cloud_efficiency.",
						},
						"disk_size": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Size of the disk.",
						},
						"period": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The period cycle of the disk. Unit: month.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Creation time of the disk.",
						},
						"expired_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Expired Time of the disk.",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Status of disk.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudDisksRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer logElapsed(ctx, "data_source.zenlayercloud_disks.read")()

	vmService := VmService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var nameRegex *regexp.Regexp
	request := &VmDiskFilter{}

	if v, ok := d.GetOk("disk_type"); ok {
		request.DiskType = v.(string)

	}
	if v, ok := d.GetOk("name"); ok {
		request.DiskName = v.(string)
	}

	if v, ok := d.GetOk("portable"); ok {
		request.Portable = common.Bool(v.(bool))
	}

	if v, ok := d.GetOk("id"); ok {
		request.Id = v.(string)
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

	var result []*vm.DiskInfo

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		result, e = vmService.DescribeDisks(ctx, request)
		if e != nil {
			return retryError(ctx, e, InternalServerError)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	var disks []*vm.DiskInfo
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
			"availability_zone": disk.ZoneId,
			"id":                disk.DiskId,
			"name":              disk.DiskName,
			"portable":          disk.Portable,
			"disk_type":         disk.DiskType,
			"disk_size":         disk.DiskSize,
			"instance_id":       disk.InstanceId,
			"instance_name":     disk.InstanceName,
			"disk_category":     disk.DiskCategory,
			"charge_type":       disk.ChargeType,
			"period":            disk.Period,
			"create_time":       disk.CreateTime,
			"expired_time":      disk.ExpiredTime,
			"status":            disk.DiskStatus,
		}
		diskList = append(diskList, mapping)
		ids = append(ids, disk.DiskId)
	}

	d.SetId(dataResourceIdHash(ids))
	err = d.Set("disks", diskList)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := writeToFile(output.(string), diskList); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

type VmDiskFilter struct {
	Id         string
	ZoneId     string
	DiskType   string
	DiskName   string
	Portable   *bool
	InstanceId string
	Status     string
}

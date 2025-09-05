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

func DataSourceZenlayerCloudZecInstances() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudZecInstancesRead,

		Schema: map[string]*schema.Schema{
			"ids": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "IDs of the ZEC instances to be queried.",
			},
			"availability_zone": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of zone that the bmc instance locates at.",
			},
			"name_regex": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A regex string to filter results by instance name.",
			},
			"image_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The image of the ZEC instance to be queried.",
			},
			"instance_status": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Status of the ZEC instances to be queried.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of resource group that the ZEC instance grouped by.",
			},
			"ipv4_address": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ipv4 address of the ZEC instances to be queried.",
			},
			"ipv6_address": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ipv6 address of the ZEC instances to be queried.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed value
			"instances": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of instances. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the ZEC instances.",
						},
						"instance_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the ZEC instance.",
						},
						"availability_zone": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of zone that the ZEC instance locates at.",
						},
						"instance_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of the ZEC instance.",
						},
						"cpu": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The number of CPU cores of the ZEC instance.",
						},
						"memory": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Memory capacity of the ZEC instance, unit in GiB.",
						},
						"image_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of image to use for the ZEC instance.",
						},
						"image_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The image name to use for the ZEC instance.",
						},
						"nic_network_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The Network card mode for the ZEC instance. Valid values: FailOver,VirtioOnly,VfOnly.",
						},
						"resource_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of resource group that the ZEC instance belongs to.",
						},
						"resource_group_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of resource group that the ZEC instance belongs to.",
						},
						"private_ip_addresses": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "Public Ipv4 addresses of the ZEC instance.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"public_ip_addresses": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "Public Ipv6 addresses of the ZEC instance.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"system_disk_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the system disk.",
						},
						"system_disk_category": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Category of the system disk.",
						},
						"system_disk_size": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Size of the system disk.",
						},
						"data_disks": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of data disk. Each element contains the following attributes:",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"data_disk_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Image ID of the data disk.",
									},
									"data_disk_category": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Category of the data disk.",
									},
									"data_disk_size": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Size of the data disk.",
									},
								},
							},
						},
						"instance_status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Current status of the ZEC instance.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Create time of the ZEC instance.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudZecInstancesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "data_source.zenlayercloud_zec_instances.read")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	filter := &ZecInstancesFilter{}
	if v, ok := d.GetOk("ids"); ok {
		instanceIds := v.(*schema.Set).List()
		if len(instanceIds) > 0 {
			filter.InstancesIds = common2.ToStringList(instanceIds)
		}
	}
	if v, ok := d.GetOk("availability_zone"); ok {
		filter.ZoneId = v.(string)
	}

	var nameRegex *regexp.Regexp
	var errRet error

	if v, ok := d.GetOk("name_regex"); ok {
		nameRegex, errRet = regexp.Compile(v.(string))
		if errRet != nil {
			return diag.Errorf("name_regex format error,%s", errRet.Error())
		}
	}

	if v, ok := d.GetOk("image_id"); ok {
		filter.ImageId = v.(string)
	}

	if v, ok := d.GetOk("instance_status"); ok {
		filter.InstanceStatus = v.(string)
	}

	if v, ok := d.GetOk("resource_group_id"); ok {
		filter.ResourceGroupId = v.(string)
	}

	if v, ok := d.GetOk("ipv4_address"); ok {
		filter.Ipv4 = v.(string)
	}
	if v, ok := d.GetOk("ipv6_address"); ok {
		filter.Ipv6 = v.(string)
	}

	var instances []*zec.InstanceInfo
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		instances, e = zecService.DescribeInstancesByFilter(filter)
		if e != nil {
			return common2.RetryError(ctx, e, common2.InternalServerError)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}
	instanceList := make([]map[string]interface{}, 0, len(instances))
	ids := make([]string, 0, len(instances))
	for _, instance := range instances {
		if nameRegex != nil && !nameRegex.MatchString(instance.InstanceName) {
			continue
		}
		mapping := map[string]interface{}{
			"id":                   instance.InstanceId,
			"instance_name":        instance.InstanceName,
			"availability_zone":    instance.ZoneId,
			"instance_type":        instance.InstanceType,
			"cpu":                  instance.Cpu,
			"memory":               instance.Memory,
			"image_id":             instance.ImageId,
			"image_name":           instance.ImageName,
			"nic_network_type":     instance.NicNetworkType,
			"resource_group_id":    instance.ResourceGroupId,
			"resource_group_name":  instance.ResourceGroupName,
			"system_disk_id":       instance.SystemDisk.DiskId,
			"system_disk_size":     instance.SystemDisk.DiskSize,
			"system_disk_category": instance.SystemDisk.DiskCategory,
			"instance_status":      instance.Status,
			"create_time":          instance.CreateTime,
			"private_ip_addresses": instance.PrivateIpAddresses,
			"public_ip_addresses":  instance.PublicIpAddresses,
		}

		dataDisks := make([]map[string]interface{}, 0, len(instance.DataDisks))
		for _, v := range instance.DataDisks {
			dataDisk := map[string]interface{}{
				"data_disk_category": v.DiskCategory,
				"data_disk_size":     v.DiskSize,
				"data_disk_id":       v.DiskId,
			}

			dataDisks = append(dataDisks, dataDisk)
		}

		mapping["data_disks"] = dataDisks

		instanceList = append(instanceList, mapping)
		ids = append(ids, instance.InstanceId)
	}

	d.SetId(common2.DataResourceIdHash(ids))
	err = d.Set("instances", instanceList)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := common2.WriteToFile(output.(string), instanceList); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

type ZecInstancesFilter struct {
	InstancesIds    []string
	ZoneId          string
	InstanceName    string
	ImageId         string
	InstanceStatus  string
	ResourceGroupId string
	Ipv4            string
	Ipv6            string
}

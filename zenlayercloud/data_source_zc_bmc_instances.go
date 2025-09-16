/*
Use this data source to query bmc instances.

Example Usage

```hcl

data "zenlayercloud_bmc_instances" "foo" {
  availability_zone = "SEL-A"
}

```
*/
package zenlayercloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	bmc "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/bmc20221120"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	"strconv"
)

func dataSourceZenlayerCloudInstances() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudInstancesRead,

		Schema: map[string]*schema.Schema{
			"instance_ids": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "IDs of the instances to be queried.",
			},
			"availability_zone": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of zone that the bmc instance locates at.",
			},
			"instance_type_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Instance type, such as `M6C`.",
			},
			"instance_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the instances to be queried.",
			},
			"hostname": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The hostname of the instance to be queried.",
			},
			"image_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The image of the instance to be queried.",
			},
			"subnet_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of vpc subnetwork.",
			},
			"instance_status": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Status of the instances to be queried.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of resource group that the instance grouped by.",
			},
			"public_ipv4": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The public ipv4 of the instances to be queried.",
			},
			"private_ipv4": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The private ip of the instances to be queried.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed value
			"instance_list": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of instances. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"instance_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the instances.",
						},
						"availability_zone": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of zone that the bmc instance locates at.",
						},
						"instance_type_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of the instance.",
						},
						"image_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of image to use for the instance.",
						},
						"key_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ssh key pair id used for the instance.",
						},
						"image_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The image name to use for the instance.",
						},
						"hostname": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The hostname of the instance.",
						},
						"instance_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the instance.",
						},
						"instance_charge_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The charge type of instance.",
						},
						"instance_charge_prepaid_period": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The tenancy (time unit is month) of the prepaid instance.",
						},
						"resource_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of resource group that the instance belongs to.",
						},
						"resource_group_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of resource group that the instance belongs to.",
						},
						"internet_charge_type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Internet charge type of the instance, Valid values are `ByBandwidth`, `ByTrafficPackage`, `ByInstanceBandwidth95` and `ByClusterBandwidth95`.",
						},
						"internet_max_bandwidth_out": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Maximum outgoing bandwidth to the public network, measured in Mbps (Mega bits per second).",
						},
						"traffic_package_size": {
							Type:        schema.TypeFloat,
							Computed:    true,
							Description: "Traffic package size.",
						},
						"subnet_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The ID of a VPC subnet. If you want to create instances in a VPC network, this parameter must be set.",
						},
						// raid
						"raid_config_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Simple config for instance raid. Modifying will cause the instance reset.",
						},
						"raid_config_custom": {
							Type:        schema.TypeList,
							Description: "Custom config for instance raid. Modifying will cause the instance reset.",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"raid_type": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Simple config for raid.",
									},
									"disk_sequence": {
										Type:        schema.TypeList,
										Computed:    true,
										Elem:        &schema.Schema{Type: schema.TypeInt},
										Description: "The sequence of disk to make raid.",
									},
								},
							},
						},

						// nic
						"nic_wan_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The wan name of the nic. The wan name should be a combination of 4 to 10 characters comprised of letters (case insensitive), numbers. The wan name must start with letter. Modifying will cause the instance reset.",
						},
						"nic_lan_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The lan name of the nic. The lan name should be a combination of 4 to 10 characters comprised of letters (case insensitive), numbers. The lan name must start with letter. Modifying will cause the instance reset.",
						},
						// partition
						"partitions": {
							Type:        schema.TypeList,
							Description: "Partition for the instance. Modifying will cause the instance reset.",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"fs_type": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The type of the partitioned file.",
									},
									"fs_path": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The drive letter(windows) or device name(linux) for the partition.",
									},
									"size": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "The size of the partitioned disk.",
									},
								},
							},
						},
						"public_ipv4_addresses": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "Public Ipv4 addresses of the instance.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"public_ipv6_addresses": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "Public Ipv6 addresses of the instance.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"private_ipv4_addresses": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "Private Ipv4 addresses of the instance.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"instance_status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Current status of the instance.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Create time of the instance.",
						},
						"expired_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Expired time of the instance.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudInstancesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "data_source.zenlayercloud_bmc_instances.read")()

	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	filter := &InstancesFilter{}
	if v, ok := d.GetOk("instance_ids"); ok {
		instanceIds := v.(*schema.Set).List()
		if len(instanceIds) > 0 {
			filter.InstancesIds = common2.ToStringList(instanceIds)
		}
	}
	if v, ok := d.GetOk("availability_zone"); ok {
		filter.ZoneId = common.String(v.(string))
	}
	if v, ok := d.GetOk("instance_type_id"); ok {
		filter.InstanceTypeId = common.String(v.(string))
	}
	if v, ok := d.GetOk("instance_name"); ok {
		filter.InstanceName = common.String(v.(string))
	}
	if v, ok := d.GetOk("hostname"); ok {
		filter.Hostname = common.String(v.(string))
	}
	if v, ok := d.GetOk("image_id"); ok {
		filter.ImageId = common.String(v.(string))
	}
	if v, ok := d.GetOk("subnet_id"); ok {
		filter.SubnetId = common.String(v.(string))
	}
	if v, ok := d.GetOk("instance_status"); ok {
		filter.InstanceStatus = common.String(v.(string))
	}

	if v, ok := d.GetOk("resource_group_id"); ok {
		filter.ResourceGroupId = common.String(v.(string))
	}

	if v, ok := d.GetOk("public_ipv4"); ok {
		filter.PublicIpv4 = common.String(v.(string))
	}
	if v, ok := d.GetOk("private_ipv4"); ok {
		filter.PrivateIpv4 = common.String(v.(string))
	}

	instances, err := bmcService.DescribeInstancesByFilter(filter)
	if err != nil {
		return diag.FromErr(err)
	}
	instanceList := make([]map[string]interface{}, 0, len(instances))
	ids := make([]string, 0, len(instances))
	for _, instance := range instances {
		mapping := map[string]interface{}{
			"instance_id":                instance.InstanceId,
			"instance_name":              instance.InstanceName,
			"hostname":                   instance.Hostname,
			"instance_type_id":           instance.InstanceTypeId,
			"instance_charge_type":       instance.InstanceChargeType,
			"availability_zone":          instance.ZoneId,
			"resource_group_id":          instance.ResourceGroupId,
			"resource_group_name":        instance.ResourceGroupName,
			"image_id":                   instance.ImageId,
			"image_name":                 instance.ImageName,
			"internet_charge_type":       instance.InternetChargeType,
			"internet_max_bandwidth_out": instance.BandwidthOutMbps,
			"instance_status":            instance.InstanceStatus,
			"create_time":                instance.CreateTime,
			"expired_time":               instance.ExpiredTime,
			"private_ipv4_addresses":     instance.PrivateIpAddresses,
			"public_ipv4_addresses":      instance.PublicIpAddresses,
			"key_id":                     instance.KeyId,
		}
		if instance.InstanceChargeType == BmcChargeTypePrepaid {
			mapping["instance_charge_prepaid_period"] = instance.Period
		}
		if instance.InternetChargeType == BmcInternetChargeTypeTrafficPackage {
			mapping["traffic_package_size"] = instance.TrafficPackageSize
		}
		if len(instance.SubnetIds) > 0 {
			mapping["subnet_id"] = instance.SubnetIds[0]
		}
		if instance.RaidConfig != nil && instance.RaidConfig.RaidType != nil {
			mapping["raid_config_type"] = strconv.Itoa(*instance.RaidConfig.RaidType)
		}
		if instance.RaidConfig != nil && len(instance.RaidConfig.CustomRaids) > 0 {
			mapping["raid_config_custom"] = map2RaidConfigCustom(instance.RaidConfig.CustomRaids)
		}
		if instance.Nic != nil {
			mapping["nic_wan_name"] = instance.Nic.WanName
			mapping["nic_lan_name"] = instance.Nic.LanName
		}
		if len(instance.Partitions) > 0 {
			mapping["partitions"] = map2Partitions(instance.Partitions)
		}

		instanceList = append(instanceList, mapping)
		ids = append(ids, instance.InstanceId)
	}

	d.SetId(common2.DataResourceIdHash(ids))
	err = d.Set("instance_list", instanceList)
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

func map2Partitions(partitions []*bmc.Partition) []interface{} {
	var res = make([]interface{}, 0, len(partitions))
	for _, partition := range partitions {
		m := make(map[string]interface{}, 3)
		m["fs_type"] = partition.FsType
		m["fs_path"] = partition.FsPath
		m["size"] = partition.Size
		res = append(res, m)
	}
	return res
}

func map2RaidConfigCustom(raids []*bmc.CustomRaid) []interface{} {
	var res = make([]interface{}, 0, len(raids))

	for _, raid := range raids {
		m := make(map[string]interface{}, 2)
		m["raid_type"] = strconv.Itoa(*raid.RaidType)
		m["disk_sequence"] = raid.DiskSequence
		res = append(res, m)
	}
	return res
}

type InstancesFilter struct {
	InstancesIds    []string
	ZoneId          *string
	InstanceTypeId  *string
	InstanceName    *string
	Hostname        *string
	ImageId         *string
	SubnetId        *string
	InstanceStatus  *string
	ResourceGroupId *string
	PublicIpv4      *string
	PrivateIpv4     *string
}

package zec

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20240401"
	"regexp"
	"time"
)

func DataSourceZenlayerCloudEips() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudEipsRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "IDs of the EIPs to be queried.",
			},
			"region_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The region ID that the elastic IP locates at.",
			},
			"name_regex": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A regex string to apply to the elastic IP list returned.",
			},
			"public_ip_address": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The elastic ipv4 address.",
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"CREATING", "DELETING", "BINDED", "UNBIND", "RECYCLING", "RECYCLED"}, false),
				Description:  "Status of the elastic IP.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Resource group ID.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			"result": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of EIPs. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the EIP.",
						},
						"region_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The region ID that the elastic IP locates at.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the elastic IP.",
						},
						"internet_charge_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Network billing methods.",
						},
						"ip_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Network types of public IPv4.",
						},
						"bandwidth": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Bandwidth. Measured in Mbps.",
						},
						"flow_package_size": {
							Type:        schema.TypeFloat,
							Computed:    true,
							Description: "The Data transfer package. Measured in TB.",
						},
						"public_ip_address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The elastic ipv4 address.",
						},
						"cidr_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "CIDR ID, the elastic ip allocated from.",
						},
						"resource_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Resource group ID.",
						},
						"resource_group_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The Name of resource group.",
						},
						"bandwidth_cluster_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Bandwidth cluster ID.",
						},
						"bandwidth_cluster_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of Bandwidth cluster.",
						},
						"peer_region_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Remote region ID.",
						},
						"private_ip_address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The private address that the EIP attached to. Only valid when the associate type is `NIC`.",
						},
						"associated_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of associated instance that the EIP attached to.",
						},
						"associated_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of associated instance that the EIP attached to. Valid values: `NAT`(for NAT gateway), `NIC`(for virtual NetworkInterface), `LB`(for Load balancer Instance).",
						},
						"bind_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Elastic IP bind type. Effective when the elastic IP is assigned to a vNIC.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Creation time of the elastic IP.",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Status of the elastic IP.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudEipsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "data_source.zenlayercloud_eips.read")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	filter := &EipFilter{}

	if v, ok := d.GetOk("ids"); ok {
		ids := v.(*schema.Set).List()
		if len(ids) > 0 {
			filter.Ids = common.ToStringList(ids)
		}
	}

	if v, ok := d.GetOk("region_id"); ok {
		filter.RegionId = v.(string)
	}

	if v, ok := d.GetOk("public_ip_address"); ok {
		filter.IpAddress = []string{v.(string)}
	}

	if v, ok := d.GetOk("status"); ok {
		filter.Status = v.(string)
	}

	if v, ok := d.GetOk("resource_group_id"); ok {
		filter.ResourceGroupId = v.(string)
	}

	var nameRegex *regexp.Regexp

	if v, ok := d.GetOk("name_regex"); ok {
		imageName := v.(string)
		if imageName != "" {
			reg, err := regexp.Compile(imageName)
			if err != nil {
				return diag.Errorf("image_name_regex format error,%s", err.Error())
			}
			nameRegex = reg
		}
	}

	var eips []*zec.EipInfo
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		eips, e = zecService.DescribeEipsByFilter(ctx, filter)
		if e != nil {
			return common.RetryError(ctx, e, common.InternalServerError)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	ids := make([]string, 0)
	eipList := make([]map[string]interface{}, 0)

	for _, eip := range eips {
		if nameRegex != nil && !nameRegex.MatchString(eip.Name) {
			continue
		}
		mapping := map[string]interface{}{
			"id":                   eip.EipId,
			"region_id":            eip.RegionId,
			"name":                 eip.Name,
			"internet_charge_type": eip.InternetChargeType,
			"ip_type":              eip.EipV4Type,
			"bandwidth":            eip.Bandwidth,
			"flow_package_size":    eip.FlowPackage,
			"cidr_id":              eip.CidrId,
			"resource_group_id":    eip.ResourceGroupId,
			"resource_group_name":  eip.ResourceGroupName,
			"peer_region_id":       eip.PeerRegionId,
			"create_time":          eip.CreateTime,
			"status":               eip.Status,
			"private_ip_address":   eip.PrivateIpAddress,
			"associated_id":        eip.AssociatedId,
			"associated_type":      eip.AssociatedType,
			"bind_type":            eip.BindType,
		}

		if eip.BandwidthCluster != nil {
			mapping["bandwidth_cluster_id"] = eip.BandwidthCluster.BandwidthClusterId
			mapping["bandwidth_cluster_name"] = eip.BandwidthCluster.BandwidthClusterName
		}

		if len(eip.PublicIpAddresses) > 0 {
			mapping["public_ip_address"] = eip.PublicIpAddresses[0]
		}

		eipList = append(eipList, mapping)
		ids = append(ids, eip.EipId)
	}

	d.SetId(common.DataResourceIdHash(ids))
	_ = d.Set("result", eipList)

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := common.WriteToFile(output.(string), eipList); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

type EipFilter struct {
	Ids             []string
	RegionId        string
	IpAddress       []string
	Status          string
	ResourceGroupId string
}

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
)

func DataSourceZenlayerCloudZecVnics() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudZecVnicsRead,

		Schema: map[string]*schema.Schema{
			"ids": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "ID of the vNICs to be queried.",
			},
			"region_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The region that the vNIC locates at.",
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsValidRegExp,
				Description:  "A regex string to apply to the vNIC list returned.",
			},
			"subnet_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the subnet to be queried.",
			},
			"vpc_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the global VPC to be queried.",
			},
			"instance_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the ZEC instances to be queried.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of resource group grouped VPC to be queried.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed value
			"vnics": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of vNICs. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the vNIC.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the vNIC.",
						},
						"region_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The region that the vNIC locates at.",
						},
						"primary": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates whether the IP is primary.",
						},
						"primary_ipv4_address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The primary private IPv4 address of the vNIC.",
						},
						"primary_ipv6_address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The primary IPv6 address of the vNIC.",
						},
						"stack_type": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The stack type of the subnet. Valid values: `IPv4`, `IPv6`, `IPv4_IPv6`",
						},
						"public_ips": {
							Type: schema.TypeSet,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Computed:    true,
							Description: "A set of public IPs. including EIP and public IPv6.",
						},
						"private_ips": {
							Type: schema.TypeSet,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Computed:    true,
							Description: "A set of intranet IPs. including private ipv4 and ipv6.",
						},
						"subnet_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "ID of the subnet.",
						},
						"vpc_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "ID of the global VPC.",
						},
						"instance_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "ID of the ZEC instance.",
						},
						"resource_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The resource group id that the NAT gateway belongs to.",
						},
						"resource_group_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The resource group name that the NAT gateway belongs to.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Create time of the vNIC.",
						},
						"security_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the security group.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudZecVnicsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "data_source.zenlayercloud_zec_vnics.read")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	filter := &ZecNicFilter{}
	if v, ok := d.GetOk("ids"); ok {
		instanceIds := v.(*schema.Set).List()
		if len(instanceIds) > 0 {
			filter.ids = common.ToStringList(instanceIds)
		}
	}
	if v, ok := d.GetOk("region_id"); ok {
		filter.RegionId = v.(string)
	}
	if v, ok := d.GetOk("vpc_id"); ok {
		filter.VpcId = v.(string)
	}
	if v, ok := d.GetOk("resource_group_id"); ok {
		filter.ResourceGroupId = v.(string)
	}
	if v, ok := d.GetOk("subnet_id"); ok {
		filter.SubnetId = v.(string)
	}
	if v, ok := d.GetOk("vpc_id"); ok {
		filter.VpcId = v.(string)
	}
	if v, ok := d.GetOk("instance_id"); ok {
		filter.InstanceId = v.(string)
	}

	var nameRegex *regexp.Regexp
	var errRet error

	if v, ok := d.GetOk("name_regex"); ok {
		nameRegex, errRet = regexp.Compile(v.(string))
		if errRet != nil {
			return diag.Errorf("name_regex format error,%s", errRet.Error())
		}
	}

	var nics []*zec.NicInfo
	err := resource.RetryContext(ctx, common.ReadRetryTimeout, func() *resource.RetryError {
		nics, errRet = zecService.DescribeNics(ctx, filter)
		if errRet != nil {
			return common.RetryError(ctx, errRet, common.InternalServerError, common.ReadTimedOut)
		}
		return nil
	})
	nicList := make([]map[string]interface{}, 0, len(nics))
	ids := make([]string, 0, len(nics))
	for _, nic := range nics {
		if nameRegex != nil && !nameRegex.MatchString(nic.Name) {
			continue
		}

		mapping := map[string]interface{}{
			"id":                   nic.NicId,
			"name":                 nic.Name,
			"region_id":            nic.RegionId,
			"primary":              nic.NicType == "Primary",
			"primary_ipv4_address": nic.PrimaryIpv4,
			"primary_ipv6_address": nic.PrimaryIpv6,
			"public_ips":           nic.PublicIpList,
			"private_ips":          nic.PrivateIpList,
			"subnet_id":            nic.SubnetId,
			"vpc_id":               nic.VpcId,
			"instance_id":          nic.InstanceId,
			"resource_group_id":    nic.ResourceGroup.ResourceGroupId,
			"resource_group_name":  nic.ResourceGroup.ResourceGroupName,
			"create_time":          nic.CreateTime,
			"stack_type":           nic.NicSubnetType,
			// TODO security group
			//"security_group_id":    nic.SecurityGroupId,
		}
		nicList = append(nicList, mapping)
		ids = append(ids, nic.NicId)
	}

	d.SetId(common.DataResourceIdHash(ids))
	err = d.Set("vnics", nicList)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := common.WriteToFile(output.(string), nicList); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

type ZecNicFilter struct {
	ids             []string
	RegionId        string
	VpcId           string
	SubnetId        string
	ResourceGroupId string
	InstanceId      string
}

package zlb

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	zlb "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zlb20250401"
	"regexp"
	"time"
)

func DataSourceZenlayerCloudZlbInstances() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudZlbInstancesRead,

		Schema: map[string]*schema.Schema{
			"ids": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "IDs of the load balancer instances to be queried.",
			},
			"region_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of region that the load balancer instances locates at.",
			},
			"name_regex": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A regex string to filter results by  load balancer instance name.",
			},
			"vpc_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the VPC to be queried.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of resource group that the load balancer instance grouped by.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed value
			"zlbs": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of instances. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"zlb_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the load balancer instances.",
						},
						"zlb_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the load balancer instance.",
						},
						"region_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of region that the load balancer instance locates at.",
						},
						"vpc_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "VPC ID to which the load balance belongs.",
						},
						"resource_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of resource group grouped load balancer instance to be queried.",
						},
						"resource_group_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of resource group that the load balancer instance belongs to.",
						},
						"private_ip_addresses": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "Private virtual Ipv4 addresses of the load balancer instance.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"public_ip_addresses": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "Public IPv4 addresses(EIP) of the load balancer instance.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Current status of the load balancer instance.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Create time of the load balancer instance.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudZlbInstancesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "data_source.zenlayercloud_zlb_instances.read")()

	zlbService := ZlbService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	filter := &LbInstanceFilter{}
	if v, ok := d.GetOk("ids"); ok {
		ids := v.(*schema.Set).List()
		if len(ids) > 0 {
			filter.LbIds = common2.ToStringList(ids)
		}
	}
	if v, ok := d.GetOk("region_id"); ok {
		filter.RegionId = v.(string)
	}

	if v, ok := d.GetOk("resource_group_id"); ok {
		filter.ResourceGroupId = v.(string)
	}

	if v, ok := d.GetOk("vpc_id"); ok {
		filter.VpcId = v.(string)
	}

	var nameRegex *regexp.Regexp
	var errRet error

	if v, ok := d.GetOk("name_regex"); ok {
		nameRegex, errRet = regexp.Compile(v.(string))
		if errRet != nil {
			return diag.Errorf("name_regex format error,%s", errRet.Error())
		}
	}

	var zlbs []*zlb.LoadBalancer

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		zlbs, e =  zlbService.DescribeLbInstancesByFilter(ctx, filter)
		if e != nil {
			return common2.RetryError(ctx, e, common2.InternalServerError)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}
	zlbList := make([]map[string]interface{}, 0, len(zlbs))
	ids := make([]string, 0, len(zlbs))
	for _, zlb := range zlbs {
		if nameRegex != nil && !nameRegex.MatchString(*zlb.LoadBalancerName) {
			continue
		}
		mapping := map[string]interface{}{
			"zlb_id":               zlb.LoadBalancerId,
			"zlb_name":             zlb.LoadBalancerName,
			"region_id":            zlb.RegionId,
			"vpc_id":               zlb.VpcId,
			"resource_group_id":    zlb.ResourceGroup.ResourceGroupId,
			"resource_group_name":  zlb.ResourceGroup.ResourceGroupName,
			"private_ip_addresses": zlb.PrivateIpAddress,
			"public_ip_addresses":  zlb.PublicIpAddress,
			"status":               zlb.Status,
			"create_time":          zlb.CreateTime,
		}

		zlbList = append(zlbList, mapping)
		ids = append(ids, *zlb.LoadBalancerId)
	}

	d.SetId(common2.DataResourceIdHash(ids))
	err = d.Set("zlbs", zlbList)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := common2.WriteToFile(output.(string), zlbList); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

type LbInstanceFilter struct {
	LbIds           []string
	RegionId        string
	VpcId           string
	ResourceGroupId string
}

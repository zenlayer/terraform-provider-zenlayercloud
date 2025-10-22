package zec

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"
	"regexp"
)

func DataSourceZenlayerCloudZecVpcRoutes() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudZecVpcRoutes,

		Schema: map[string]*schema.Schema{
			"ids": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "ID of the route to be queried.",
			},
			"vpc_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the global VPC to be queried.",
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsValidRegExp,
				Description:  "A regex string to apply to the vNIC list returned.",
			},
			"ip_version": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"IPv4", "IPv6"}, true),
				Description:  "IP stack type. Valid values: `IPv4`, `IPv6`.",
			},
			"destination_cidr_block": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Destination address block to be queried.",
			},
			"route_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"RouteTypeStatic", "RouteTypePolicy", "RouteTypeSubnet", "RouteTypeNatGw", "RouteTypeTransit"}, false),
				Description:  "Route type to be queried. Valid values: `RouteTypeStatic`(for static route), `RouteTypePolicy`(for policy route), `RouteTypeSubnet`(for subnet route), `RouteTypeNatGw`(for NAT gateway route), `RouteTypeTransit`(for dynamic route).",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed value
			"routes": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of routes. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the route.",
						},
						"vpc_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "ID of the VPC.",
						},
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(2, 63),
							Description:  "The name of the VPC route. The name should start and end with a number or a letter, containing 2 to 63 characters. Only letters, numbers, - and periods (.) are supported.",
						},
						"destination_cidr_block": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Destination address block.",
						},
						"ip_version": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "IP stack type. Valid values: `IPv4`, `IPv6`.",
						},
						"route_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Route type. Valid values: `RouteTypeStatic`(for static route), `RouteTypePolicy`(for policy route), `RouteTypeSubnet`(for subnet route), `RouteTypeNatGw`(for NAT gateway route), `RouteTypeTransit`(for dynamic route).",
						},
						"source_ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The source IP matched.",
						},
						"priority": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Priority of the route entry. Valid value: from `0` to `65535`.",
						},
						"next_hop_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of next hop instance.",
						},
						"next_hop_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of next hop instance. Valid values: `NIC`(for vNIC), `VPC`(for VPC), `NAT`(for NAT gateway), `ZBG`(for border gateway).",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Creation time of the VPC route.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudZecVpcRoutes(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "data_source.zenlayercloud_zec_routes.read")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	filter := &VpcRouteFilter{}
	if v, ok := d.GetOk("ids"); ok {
		instanceIds := v.(*schema.Set).List()
		if len(instanceIds) > 0 {
			filter.ids = common.ToStringList(instanceIds)
		}
	}
	if v, ok := d.GetOk("destination_cidr_block"); ok {
		filter.cidrBlock = v.(string)
	}
	if v, ok := d.GetOk("vpc_id"); ok {
		filter.vpcId = v.(string)
	}
	if v, ok := d.GetOk("ip_version"); ok {
		filter.ipVersion = v.(string)
	}
	if v, ok := d.GetOk("route_type"); ok {
		filter.routeType = v.(string)
	}

	var nameRegex *regexp.Regexp
	var errRet error

	if v, ok := d.GetOk("name_regex"); ok {
		nameRegex, errRet = regexp.Compile(v.(string))
		if errRet != nil {
			return diag.Errorf("name_regex format error,%s", errRet.Error())
		}
	}

	var routes []*zec.RouteInfo
	err := resource.RetryContext(ctx, common.ReadRetryTimeout, func() *resource.RetryError {
		routes, errRet = zecService.DescribeVpcRoutes(ctx, filter)
		if errRet != nil {
			return common.RetryError(ctx, errRet, common.InternalServerError, common.ReadTimedOut)
		}
		return nil
	})
	routeList := make([]map[string]interface{}, 0, len(routes))
	ids := make([]string, 0, len(routes))
	for _, route := range routes {
		if nameRegex != nil && (route.Name == nil || !nameRegex.MatchString(*route.Name)) {
			continue
		}

		mapping := map[string]interface{}{
			"id":                     route.RouteId,
			"name":                   route.Name,
			"vpc_id":                 route.VpcId,
			"destination_cidr_block": route.DestinationCidrBlock,
			"ip_version":             route.IpVersion,
			"route_type":             route.Type,
			"source_ip":              route.SourceCidrBlock,
			"priority":               route.Priority,
			"next_hop_id":            route.NextHopId,
			"next_hop_type":          route.NextHopType,
			"create_time":            route.CreateTime,
		}
		routeList = append(routeList, mapping)
		ids = append(ids, *route.RouteId)
	}

	d.SetId(common.DataResourceIdHash(ids))
	err = d.Set("routes", routeList)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := common.WriteToFile(output.(string), routeList); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

type VpcRouteFilter struct {
	ids       []string
	routeType string
	vpcId     string
	cidrBlock string
	ipVersion string
}

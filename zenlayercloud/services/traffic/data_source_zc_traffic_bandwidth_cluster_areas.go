package traffic

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	traffic "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/traffic20240326"
	"regexp"
	"time"
)

func DataSourceZenlayerCloudTrafficBandwidthClusterAreas() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudTrafficBandwidthClusterAreasRead,

		Schema: map[string]*schema.Schema{
			"area_code": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Code(ID) of the bandwidth cluster area.",
			},
			"network_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The IP network support to be queried.",
			},
			"name_regex": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A regex string to apply to the name of bandwidth cluster area list returned.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed value
			"areas": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of bandwidth cluster areas. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"area_code": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the bandwidth cluster area.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the bandwidth cluster area.",
						},
						"network_types": {
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
							Description: "IP network type support in the bandwidth cluster. valid values: `BGP`(for BGP network), `Cogent`(for Cogent network),`CN2`(for China Telecom Next Carrier Network), `CMI`(for China Mobile network), `CUG`(for China Unicom network), `CTG`(for China Telecom network).",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudTrafficBandwidthClusterAreasRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "data_source.zenlayercloud_traffic_bandwidth_clusters.read")()

	trafficService := TrafficService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var areaCode string
	if v, ok := d.GetOk("area_code"); ok {
		areaCode = v.(string)
	}

	var networkType string
	if v, ok := d.GetOk("network_type"); ok {
		networkType = v.(string)
	}

	var nameRegex *regexp.Regexp
	if v, ok := d.GetOk("name_regex"); ok {
		var errRet error
		nameRegex, errRet = regexp.Compile(v.(string))
		if errRet != nil {
			return diag.Errorf("name_regex format error,%s", errRet.Error())
		}
	}

	var result []*traffic.BandwidthClusterAreaInfo

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		result, e = trafficService.DescribeBandwidthClusterAreas(ctx)
		if e != nil {
			return common2.RetryError(ctx, e, common2.InternalServerError)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	ids := make([]string, 0)
	filteredAreas := make([]*traffic.BandwidthClusterAreaInfo, 0)

	for _, area := range result {
		if areaCode != "" && *area.AreaCode != areaCode {
			continue
		}

		if networkType != "" {
			found := false
			for _, nt := range area.NetworkTypes {
				if nt == networkType {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		if nameRegex != nil && !nameRegex.MatchString(*area.AreaName) {
			continue
		}

		filteredAreas = append(filteredAreas, area)
		ids = append(ids, *area.AreaCode)
	}

	areaMaps := make([]map[string]interface{}, 0)
	for _, area := range filteredAreas {
		areaMap := map[string]interface{}{
			"area_code":     area.AreaCode,
			"name":          area.AreaName,
			"network_types": area.NetworkTypes,
		}
		areaMaps = append(areaMaps, areaMap)
	}

	d.SetId(common2.DataResourceIdHash(ids))
	if err := d.Set("areas", areaMaps); err != nil {
		return diag.FromErr(err)
	}

	if output, ok := d.GetOk("result_output_file"); ok && output.(string) != "" {
		if err := common2.WriteToFile(output.(string), areaMaps); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}


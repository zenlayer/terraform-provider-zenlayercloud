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

func DataSourceZenlayerCloudTrafficBandwidthClusters() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudTrafficBandwidthClustersRead,

		Schema: map[string]*schema.Schema{
			"ids": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "ids of the bandwidth cluster to be queried.",
			},
			"city_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of city where the bandwidth cluster located.",
			},
			"name_regex": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A regex string to apply to the bandwidth cluster list returned.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed value
			"bandwidth_clusters": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of bandwidth cluster. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the bandwidth cluster.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the bandwidth cluster.",
						},
						"network_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "IP network type. valid values: `BGP`(for BGP network), `Cogent`(for Cogent network),`CN2`(for China Telecom Next Carrier Network), `CMI`(for China Mobile network), `CUG`(for China Unicom network), `CTG`(for China Telecom network).",
						},
						"internet_charge_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Network billing method. valid values: `MonthlyPercent95Bandwidth`(for Monthly Burstable 95th billing method), `DayPeakBandwidth`(for Daily Peak billing method).",
						},
						"commit_bandwidth_mbps": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Bandwidth commitment. Measured in Mbps.",
						},
						"area_code": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The code of area where the bandwidth located.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Creation time of the bandwidth cluster.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudTrafficBandwidthClustersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "data_source.zenlayercloud_traffic_bandwidth_clusters.read")()

	trafficService := TrafficService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var nameRegex *regexp.Regexp
	filter := &TrafficFilter{}

	if v, ok := d.GetOk("ids"); ok {
		ids := v.(*schema.Set).List()
		if len(ids) > 0 {
			filter.Ids = common2.ToStringList(ids)
		}
	}

	if v, ok := d.GetOk("city_name"); ok {
		filter.cityName = v.(string)
	}

	if v, ok := d.GetOk("name_regex"); ok {
		var errRet error
		nameRegex, errRet = regexp.Compile(v.(string))
		if errRet != nil {
			return diag.Errorf("name_regex format error,%s", errRet.Error())
		}
	}

	var result []*traffic.BandwidthClusterInfo

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		result, e = trafficService.DescribeBandwidthClusterByFilter(ctx, filter)
		if e != nil {
			return common2.RetryError(ctx, e, common2.InternalServerError)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	ids := make([]string, 0)

	var bandwidthClusters []map[string]interface{}
	for _, item := range result {
		if nameRegex != nil && !nameRegex.MatchString(*item.BandwidthClusterName) {
			continue
		}
		mapping := map[string]interface{}{
			"id":                    item.BandwidthClusterId,
			"name":                  item.BandwidthClusterName,
			"network_type":          item.NetworkType,
			"internet_charge_type":  item.InternetChargeType,
			"commit_bandwidth_mbps": item.CommitBandwidthMbps,
			"area_code":             item.AreaCode,
			"create_time":           item.CreateTime,
		}
		bandwidthClusters = append(bandwidthClusters, mapping)
		ids = append(ids, *item.BandwidthClusterId)

	}

	d.SetId(common2.DataResourceIdHash(ids))

	if err := d.Set("bandwidth_clusters", bandwidthClusters); err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := common2.WriteToFile(output.(string), bandwidthClusters); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil

}

type TrafficFilter struct {
	Ids        []string
	cityName     string
}

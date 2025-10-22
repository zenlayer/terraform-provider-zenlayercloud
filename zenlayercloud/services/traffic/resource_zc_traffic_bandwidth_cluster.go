package traffic

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	traffic "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/traffic20240326"
	"time"
)

func ResourceZenlayerCloudTrafficBandwidthCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudTrafficBandwidthClusterCreate,
		ReadContext:   resourceZenlayerCloudTrafficBandwidthClusterRead,
		UpdateContext: resourceZenlayerCloudTrafficBandwidthClusterUpdate,
		DeleteContext: resourceZenlayerCloudTrafficBandwidthClusterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"area_code": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The code of area where the bandwidth located.",
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(2, 63),
				Description:  "The name of the bandwidth cluster.",
			},
			"network_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"BGP", "Cogent", "CN2", "CMI", "CUG", "CTG"}, false),
				Description:  "IP network type. The value is required when the billing area for bandwidth cluster is by city. valid values: `BGP`(for BGP network), `Cogent`(for Cogent network),`CN2`(for China Telecom Next Carrier Network), `CMI`(for China Mobile network), `CUG`(for China Unicom network), `CTG`(for China Telecom network).",
			},
			"internet_charge_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"MonthlyPercent95Bandwidth", "DayPeakBandwidth"}, false),
				Description:  "Network billing method. valid values: `MonthlyPercent95Bandwidth`(for Monthly Burstable 95th billing method), `DayPeakBandwidth`(for Daily Peak billing method).",
			},
			"commit_bandwidth_mbps": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Bandwidth commitment. Measured in Mbps. Default value: `0`.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation time of the bandwidth cluster.",
			},
		},
	}
}

func resourceZenlayerCloudTrafficBandwidthClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	bandwidthClusterId := d.Id()

	trafficService := TrafficService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		bandwidthCluster, errRet := trafficService.DescribeBandwidthClusterResourcesById(ctx, bandwidthClusterId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		if bandwidthCluster == nil {
			return nil
		}

		if *bandwidthCluster.TotalCount == 0 {
			return nil
		}

		return resource.RetryableError(fmt.Errorf("bandwidth cluster %s still have %d resources", bandwidthClusterId, *bandwidthCluster.TotalCount))
	})

	if err != nil {
		return diag.FromErr(err)
	}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := trafficService.DeleteBandwidthClusterId(ctx, bandwidthClusterId)
		if errRet != nil {
			ee, ok := err.(*common.ZenlayerCloudSdkError)
			if !ok {
				return common2.RetryError(ctx, errRet)
			}
			if ee.Code == common2.ResourceNotFound {
				return nil
			}
			return common2.RetryError(ctx, errRet, "INVALID_CLUSTER_INSTANCE_ASSOCIATED")
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceZenlayerCloudTrafficBandwidthClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	trafficService := TrafficService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	bandwidthClusterId := d.Id()

	if d.HasChanges("name") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			err := trafficService.ModifyBandwidthClusterName(ctx, bandwidthClusterId, d.Get("name").(string))
			if err != nil {
				return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChanges("commit_bandwidth_mbps") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			err := trafficService.ModifyBandwidthClusterCommitBandwidth(ctx, bandwidthClusterId, d.Get("commit_bandwidth_mbps").(int))
			if err != nil {
				return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceZenlayerCloudTrafficBandwidthClusterRead(ctx, d, meta)
}

func resourceZenlayerCloudTrafficBandwidthClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	request := traffic.NewCreateBandwidthClusterRequest()
	request.Name = common.String(d.Get("name").(string))
	request.AreaCode = common.String(d.Get("area_code").(string))
	request.InternetChargeType = common.String(d.Get("internet_charge_type").(string))

	if v, ok := d.GetOk("network_type"); ok {
		request.NetworkType = common.String(v.(string))
	}

	if v, ok := d.GetOk("commit_bandwidth_mbps"); ok {
		request.CommitBandwidthMbps = common.Integer(v.(int))
	}

	bandwidthClusterId := ""

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := meta.(*connectivity.ZenlayerCloudClient).WithTrafficClient().CreateBandwidthCluster(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create bandwidth cluster.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": common2.ToJsonString(request),
				"err":     err.Error(),
			})
			return common2.RetryError(ctx, err)
		}

		tflog.Info(ctx, "Create bandwidth cluster success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  common2.ToJsonString(request),
			"response": common2.ToJsonString(response),
		})

		if response.Response.BandwidthClusterId == nil {
			err = fmt.Errorf("bandwidht cluster id is nil")
			return resource.NonRetryableError(err)
		}
		bandwidthClusterId = *response.Response.BandwidthClusterId

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(bandwidthClusterId)

	return resourceZenlayerCloudTrafficBandwidthClusterRead(ctx, d, meta)
}

func resourceZenlayerCloudTrafficBandwidthClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	bandwidthClusterId := d.Id()

	trafficService := TrafficService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var bandwidthCluster *traffic.BandwidthClusterInfo
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		bandwidthCluster, errRet = trafficService.DescribeBandwidthClusterById(ctx, bandwidthClusterId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if bandwidthCluster == nil {
		d.SetId("")
		tflog.Info(ctx, "bandwidth cluster not exist", map[string]interface{}{
			"bandwidthClusterId": bandwidthClusterId,
		})
		return nil
	}

	// bandwidth cluster info
	d.SetId(*bandwidthCluster.BandwidthClusterId)
	_ = d.Set("name", bandwidthCluster.BandwidthClusterName)
	_ = d.Set("area_code", bandwidthCluster.AreaCode)
	_ = d.Set("internet_charge_type", bandwidthCluster.InternetChargeType)
	_ = d.Set("network_type", bandwidthCluster.NetworkType)
	_ = d.Set("commit_bandwidth_mbps", bandwidthCluster.CommitBandwidthMbps)
	_ = d.Set("create_time", bandwidthCluster.CreateTime)

	return diags

}

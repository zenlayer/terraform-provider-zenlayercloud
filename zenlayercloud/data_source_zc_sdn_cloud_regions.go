/*
Use this data source to query cloud regions.

Example Usage

```hcl
data "zenlayercloud_sdn_cloud_regions" "google_regions" {
	cloud_type = "GOOGLE"
}
```
*/
package zenlayercloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
)

func dataSourceZenlayerCloudCloudRegions() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudCloudRegionsRead,
		Schema: map[string]*schema.Schema{
			"cloud_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(CLOUD_ENDPOINT_TYPES, false),
				Description:  "The type of the cloud, Valid values: `AWS`, `TENCENT`, `GOOGLE`.",
			},
			"product": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(PRODUCT_TYPES, false),
				Description:  "The product to be queried. Valid values: `PrivateConnect`, `CloudRouter`.",
			},
			"google_pairing_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Google Paring key, which is required when cloud type is `GOOGLE`.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed value
			"region_list": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of cloud region. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cloud_region": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the cloud region.",
						},
						"products": {
							Type:        schema.TypeSet,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "The connect product.",
						},
						"datacenter": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The id of datacenter that can be connect to cloud region.",
						},
						"datacenter_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of datacenter.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudCloudRegionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "data_source.zenlayercloud_sdn_cloud_regions.read")()
	//
	sdnService := SdnService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	cloudType := d.Get("cloud_type").(string)
	regionFilter := CloudRegionFilter{}
	regionFilter.cloudType = cloudType
	if cloudType == POINT_TYPE_GOOGLE {
		v, ok := d.GetOk("google_pairing_key")
		if ok {
			regionFilter.googlePairingKey = common.String(v.(string))
		} else {
			return diag.Errorf("google_paring_key is required for cloud_type `GOOGLE`")
		}
	}

	v, ok := d.GetOk("product")
	if ok {
		regionFilter.product = common.String(v.(string))
	}

	cloudRegions, err := sdnService.DescribeCloudRegions(regionFilter)
	if err != nil {
		return diag.FromErr(err)
	}
	cloudRegionList := make([]map[string]interface{}, 0, len(cloudRegions))
	ids := make([]string, 0, len(cloudRegions))
	for _, cloudRegion := range cloudRegions {
		mapping := map[string]interface{}{
			"cloud_region":    cloudRegion.CloudRegionId,
			"datacenter":      cloudRegion.DataCenter.DcId,
			"products":        cloudRegion.Products,
			"datacenter_name": cloudRegion.DataCenter.DcName,
		}
		cloudRegionList = append(cloudRegionList, mapping)
		ids = append(ids, cloudRegion.CloudRegionId+":"+cloudRegion.DataCenter.DcId)
	}
	d.SetId(common2.DataResourceIdHash(ids))
	err = d.Set("region_list", cloudRegionList)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := common2.WriteToFile(output.(string), cloudRegionList); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

type CloudRegionFilter struct {
	googlePairingKey *string
	product          *string
	cloudType        string
}

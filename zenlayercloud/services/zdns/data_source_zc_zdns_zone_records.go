package zdns

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	pvtdns "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zdns20251101"
	"time"
)

func DataSourceZenlayerCloudPvtdnsRecords() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudPvtdnsRecordsRead,
		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the DNS private zone.",
			},
			"ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "IDs of the records to be queried.",
			},
			"record_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Type of the records to be queried.",
			},
			"record_value": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Value of the records to be queried.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			"records": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of private DNS records. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the private DNS record.",
						},
						"record_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the private DNS record.",
						},
						"record_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of the private DNS record.",
						},
						"record_value": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Value of the private DNS record.",
						},
						"ttl": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "TTL of the private DNS record.",
						},
						"weight": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Weight of the private DNS record.",
						},
						"priority": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Priority of the private DNS record.",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Status of the private DNS record.",
						},
						"line": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The resolver line.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Creation time of the private DNS record.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudPvtdnsRecordsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "data_source.zenlayercloud_pvtdns_records.read")()

	pvtDnsService := ZdnsService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	zoneId := d.Get("zone_id").(string)

	filter := &PrivateRecordFilter{
		ZoneId: zoneId,
	}

	if v, ok := d.GetOk("ids"); ok {
		ids := v.(*schema.Set).List()
		if len(ids) > 0 {
			filter.RecordIds = common.ToStringList(ids)
		}
	}

	if v, ok := d.GetOk("record_type"); ok {
		filter.RecordType = v.(string)
	}

	if v, ok := d.GetOk("record_value"); ok {
		filter.RecordValue = v.(string)
	}

	var records []*pvtdns.PrivateZoneRecord
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		records, e = pvtDnsService.DescribePrivateZoneRecordsByFilter(ctx, filter)
		if e != nil {
			return common.RetryError(ctx, e, common.InternalServerError)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	recordList := make([]map[string]interface{}, 0)
	ids := make([]string, 0)

	for _, record := range records {
		mapping := map[string]interface{}{
			"id":           record.RecordId,
			"record_name":  record.RecordName,
			"record_type":  record.Type,
			"record_value": record.Value,
			"ttl":          record.Ttl,
			"weight":       record.Weight,
			"priority":     record.Priority,
			"status":       record.Status,
			"create_time":  record.CreateTime,
			"line":         record.Line,
		}

		recordList = append(recordList, mapping)
		ids = append(ids, *record.RecordId)
	}

	d.SetId(common.DataResourceIdHash(ids))
	_ = d.Set("records", recordList)

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := common.WriteToFile(output.(string), recordList); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

type PrivateRecordFilter struct {
	ZoneId      string
	RecordIds   []string
	RecordType  string
	RecordValue string
}

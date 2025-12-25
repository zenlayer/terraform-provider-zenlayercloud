package zdns

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	pvtdns "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zdns20251101"
	"time"
)

func ResourceZenlayerCloudPvtdnsRecord() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudPvtdnsRecordCreate,
		ReadContext:   resourceZenlayerCloudPvtdnsRecordRead,
		UpdateContext: resourceZenlayerCloudPvtdnsRecordUpdate,
		DeleteContext: resourceZenlayerCloudPvtdnsRecordDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: customdiff.All(
			weightValidateFunc(),
			priorityValidateFunc(),
		),
		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the private zone.",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"A", "AAAA", "CNAME", "MX", "TXT", "PTR", "SRV"}, false),
				Description: "DNS record type. Valid values: \n\t- `A`: Maps a domain name to an IP address \n\t- `AAAA`: Maps a domain name to an IPv6 address \n\t- `CNAME`: Maps a domain name to another domain name \n\t- `MX`: Maps a domain name to a mail server address \n\t- `TXT`: Text information \n\t- `PTR`: Maps an IP address to a domain name for reverse DNS lookup \n\t- `SRV`: Specifies servers providing specific services (format: [priority] [weight] [port] [target address], e.g., 0 5 5060 sipserver.example.com)."			},
			"record_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the record. such as `www`, `@`.",
			},
			"line": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "The resolver line. Default is `default`. Also valid for specified region, such as `asia-east-1`.",
			},
			"value": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The value of the record.",
			},
			"remark": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Remarks for the record.",
			},
			"priority": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 99),
				Description:  "MX priority, which is required when the record type is `MX`. Range: [1, 99], default: 1.",
			},
			"weight": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 100),
				Description:  "Weight for the record. Only takes effect for type `A` or `AAAA`. Range: [1, 100], default: 1.",
			},
			"ttl": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(5, 86400),
				Description:  "The ttl of the Private Zone Record. Measured in second. Range: [5,86400], default: 60.",
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"Enabled", "Disabled"}, false),
				Description:  "Record status. Valid values: `Enabled`, `Disabled`.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Create time of the record.",
			},
		},
	}
}

func priorityValidateFunc() schema.CustomizeDiffFunc {
	return customdiff.IfValue("type", func(ctx context.Context, value, meta interface{}) bool {
		return value != "MX"
	}, func(ctx context.Context, diff *schema.ResourceDiff, i interface{}) error {

		if _, ok := diff.GetOk("priority"); ok {
			return fmt.Errorf("`priority` can only be set when `type` is `MX`")
		}
		return nil
	})
}

func weightValidateFunc() schema.CustomizeDiffFunc {
	return customdiff.IfValue("weight", func(ctx context.Context, value, meta interface{}) bool {
		return value != 0
	}, func(ctx context.Context, diff *schema.ResourceDiff, i interface{}) error {
		v := diff.Get("type").(string)
		// v is A or AAAA
		if (v != "A" && v != "AAAA") && !diff.GetRawConfig().GetAttr("weight").IsNull() {
			// 只能在 A, AAAA 才能设置weight
			return fmt.Errorf("`weight` can only be set when `type` is `A` or `AAAA`")
		}
		return nil
	})
}

func resourceZenlayerCloudPvtdnsRecordDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_pvtdns_record.delete")()

	id, err := common2.ParseResourceId(d.Id(), 2)
	if err != nil {
		return diag.FromErr(err)
	}
	zoneId := id[0]
	recordId := id[1]

	pvtDnsService := ZdnsService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := pvtDnsService.DeletePrivateDnsRecordById(ctx, zoneId, recordId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceZenlayerCloudPvtdnsRecordUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pvtDnsService := ZdnsService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	id, err := common2.ParseResourceId(d.Id(), 2)
	if err != nil {
		return diag.FromErr(err)
	}
	zoneId := id[0]
	recordId := id[1]

	if d.HasChanges("value", "remark", "priority", "ttl", "weight") {
		request := pvtdns.NewModifyPrivateZoneRecordRequest()
		request.ZoneId = &zoneId
		request.RecordId = common.String(recordId)
		request.Value = common.String(d.Get("value").(string))
		request.Remark = common.String(d.Get("remark").(string))

		if v, ok := d.GetOk("priority"); ok {
			request.Priority = common.Integer(v.(int))
		}

		if v, ok := d.GetOk("ttl"); ok {
			request.Ttl = common.Integer(v.(int))
		}

		if v, ok := d.GetOk("weight"); ok {
			request.Weight = common.Integer(v.(int))
		}

		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			_, err := pvtDnsService.client.WithZDnsClient().ModifyPrivateZoneRecord(request)
			if err != nil {
				return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("status") {
		if v, ok := d.GetOk("status"); ok {
			request := pvtdns.NewModifyPrivateZoneRecordsStatusRequest()
			request.ZoneId = &zoneId
			request.RecordIds = []string{recordId}
			request.Status = common.String(v.(string))

			err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
				_, err := pvtDnsService.client.WithZDnsClient().ModifyPrivateZoneRecordsStatus(request)
				if err != nil {
					tflog.Info(ctx, "Fail to create private zone record.", map[string]interface{}{
						"action":  request.GetAction(),
						"request": common2.ToJsonString(request),
						"err":     err.Error(),
					})
					return common2.RetryError(ctx, err)
				}

				return nil
			})

			if err != nil {
				return diag.FromErr(err)
			}
		}

	}

	return resourceZenlayerCloudPvtdnsRecordRead(ctx, d, meta)
}

func resourceZenlayerCloudPvtdnsRecordCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pvtDnsService := ZdnsService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	zoneId := d.Get("zone_id").(string)

	request := pvtdns.NewAddPrivateZoneRecordRequest()
	request.ZoneId = &zoneId
	request.Type = common.String(d.Get("type").(string))
	request.RecordName = common.String(d.Get("record_name").(string))
	request.Value = common.String(d.Get("value").(string))

	if v, ok := d.GetOk("line"); ok {
		request.Line = common.String(v.(string))
	}

	if v, ok := d.GetOk("remark"); ok {
		request.Remark = common.String(v.(string))
	}
	if v, ok := d.GetOk("priority"); ok {
		request.Priority = common.Integer(v.(int))
	}
	if v, ok := d.GetOk("ttl"); ok {
		request.Ttl = common.Integer(v.(int))
	}
	if v, ok := d.GetOk("weight"); ok {
		request.Weight = common.Integer(v.(int))
	}
	if v, ok := d.GetOk("status"); ok {
		request.Status = common.String(v.(string))
	}

	recordId := ""

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := pvtDnsService.client.WithZDnsClient().AddPrivateZoneRecord(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create private zone record.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": common2.ToJsonString(request),
				"err":     err.Error(),
			})
			return common2.RetryError(ctx, err)
		}

		tflog.Info(ctx, "Create private zone record success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  common2.ToJsonString(request),
			"response": common2.ToJsonString(response),
		})

		if response.Response.RecordId == nil {
			err = fmt.Errorf("private zone record id is nil")
			return resource.NonRetryableError(err)
		}
		recordId = *response.Response.RecordId

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}
	// zoneId:recordId
	d.SetId(fmt.Sprintf("%s:%s", zoneId, recordId))

	return resourceZenlayerCloudPvtdnsRecordRead(ctx, d, meta)
}

func resourceZenlayerCloudPvtdnsRecordRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	id, err := common2.ParseResourceId(d.Id(), 2)
	if err != nil {
		return diag.FromErr(err)
	}
	zoneId := id[0]
	recordId := id[1]

	pvtDnsService := ZdnsService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var record *pvtdns.PrivateZoneRecord
	var errRet error

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		record, errRet = pvtDnsService.DescribePrivateZoneRecordById(ctx, zoneId, recordId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if record == nil {
		d.SetId("")
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Private DNS record doesn't exist",
			Detail:   fmt.Sprintf("The private DNS record %s is not exist", recordId),
		})
		return diags
	}

	_ = d.Set("zone_id", record.ZoneId)
	_ = d.Set("type", record.Type)
	_ = d.Set("record_name", record.RecordName)
	_ = d.Set("value", record.Value)
	_ = d.Set("remark", record.Remark)
	_ = d.Set("priority", record.Priority)
	_ = d.Set("ttl", record.Ttl)
	_ = d.Set("weight", record.Weight)
	_ = d.Set("create_time", record.CreateTime)
	_ = d.Set("status", record.Status)
	_ = d.Set("line", record.Line)

	return diags
}

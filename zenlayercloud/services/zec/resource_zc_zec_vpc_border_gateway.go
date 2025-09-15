package zec

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	common2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20240401"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
)

func ResourceZenlayerCloudBorderGateway() *schema.Resource {

	return &schema.Resource{
		CreateContext: resourceZenlayerCloudBorderGatewayCreate,
		ReadContext:   resourceZenlayerCloudBorderGatewayRead,
		UpdateContext: resourceZenlayerCloudBorderGatewayUpdate,
		DeleteContext: resourceZenlayerCloudBorderGatewayDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the border gateway.",
			},
			"vpc_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "VPC ID that the border gateway belongs to.",
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Region ID of the border gateway.",
			},
			"asn": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Autonomous System Number.",
			},
			"advertised_subnet": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Subnet route advertisement.",
			},
			"advertised_cidrs": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Custom IPv4 CIDR block list.",
			},
			"nat_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of NAT gateway associated.",
			},
			"inter_connect_cidr": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Interconnect IP range.",
			},
			"cloud_router_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Cloud router IDs that border gateway related.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation time of the border gateway.",
			},
		},
	}
}

func resourceZenlayerCloudBorderGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_border_gateway.create")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	request := zec.NewCreateBorderGatewayRequest()
	request.Label = d.Get("name").(string)
	request.VpcId = d.Get("vpc_id").(string)
	request.RegionId = d.Get("region_id").(string)
	request.Asn = d.Get("asn").(int)
	request.AdvertisedSubnet = d.Get("advertised_subnet").(string)

	if v, ok := d.GetOk("advertised_cidrs"); ok {
		cidrs := v.(*schema.Set).List()
		cidrList := make([]string, 0, len(cidrs))
		for _, cidr := range cidrs {
			cidrList = append(cidrList, cidr.(string))
		}
		request.AdvertisedCidrs = cidrList
	}

	zbgId := ""

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := zecService.client.WithZecClient().CreateBorderGateway(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create border gateway.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": common.ToJsonString(request),
				"err":     err.Error(),
			})
			return common.RetryError(ctx, err)
		}

		tflog.Info(ctx, "Create subnet success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  common.ToJsonString(request),
			"response": common.ToJsonString(response),
		})

		if response.Response.ZbgId == "nil" {
			err = fmt.Errorf("border gateway id is nil")
			return resource.NonRetryableError(err)
		}
		zbgId = response.Response.ZbgId

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(zbgId)

	return resourceZenlayerCloudBorderGatewayRead(ctx, d, meta)
}

func resourceZenlayerCloudBorderGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_border_gateway.read")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	borderGateway, err := zecService.DescribeBorderGatewayById(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if borderGateway == nil {
		d.SetId("")
		return nil
	}

	_ = d.Set("zbgId", borderGateway.ZbgId)
	_ = d.Set("name", borderGateway.Name)
	_ = d.Set("vpc_id", borderGateway.VpcId)
	_ = d.Set("region_id", borderGateway.RegionId)
	_ = d.Set("asn", borderGateway.Asn)
	_ = d.Set("inter_connect_cidr", borderGateway.InterConnectCidr)
	_ = d.Set("cloud_router_ids", borderGateway.CloudRouterIds)
	_ = d.Set("advertised_subnet", borderGateway.AdvertisedSubnet)
	_ = d.Set("advertised_cidrs", borderGateway.AdvertisedCidrs)
	_ = d.Set("nat_id", borderGateway)
	_ = d.Set("create_time", borderGateway.CreateTime)
	_ = d.Set("zbgId", borderGateway.ZbgId)

	return nil
}

func resourceZenlayerCloudBorderGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_border_gateway.update")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	request := zec.NewModifyBorderGatewaysAttributeRequest()
	request.ZbgIds = []string{d.Id()}
	update := false

	if d.HasChange("name") {
		request.Name = d.Get("name").(string)
		update = true
	}

	if d.HasChange("advertised_subnet") {
		request.AdvertisedSubnet = d.Get("advertised_subnet").(string)
		update = true
	}

	if d.HasChange("asn") {
		request.Asn = d.Get("asn").(int)
		update = true
	}

	if d.HasChange("advertised_cidrs") {
		if v, ok := d.GetOk("advertised_cidrs"); ok {
			cidrs := v.(*schema.Set).List()
			cidrList := make([]string, 0, len(cidrs))
			for _, cidr := range cidrs {
				cidrList = append(cidrList, cidr.(string))
			}
			request.AdvertisedCidrs = cidrList
		}
		update = true
	}

	if update {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			err := zecService.ModifyBorderGateway(ctx, request)
			if err != nil {
				return common.RetryError(ctx, err, common.InternalServerError, common2.NetworkError)
			}
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceZenlayerCloudBorderGatewayRead(ctx, d, meta)
}

func resourceZenlayerCloudBorderGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_border_gateway.delete")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	err := zecService.DeleteBorderGateway(ctx, d.Id())
	if err != nil {
		ee, ok := err.(*common2.ZenlayerCloudSdkError)
		if !ok {
			return diag.FromErr(err)
		}
		if ee.Code == "INVALID_ZBG_NOT_FOUND" || ee.Code == common.ResourceNotFound {
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}

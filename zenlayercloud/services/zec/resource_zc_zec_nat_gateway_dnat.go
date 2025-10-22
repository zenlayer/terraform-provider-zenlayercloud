package zec

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
	zec2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"
	"time"
)

func ResourceZenlayerCloudZecVpcNatGatewayDnat() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZecVpcNatGatewayDnatCreate,
		ReadContext:   resourceZenlayerCloudZecVpcNatGatewayDnatRead,
		UpdateContext: resourceZenlayerCloudZecVpcNatGatewayDnatUpdate,
		DeleteContext: resourceZenlayerCloudZecVpcNatGatewayDnatDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"nat_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the NAT gateway.",
			},
			"protocol": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"TCP", "UDP", "Any"}, false),
				Description:  "The IP protocol type of the DNAT entry. Valid values: `TCP`, `UDP`, `Any`. If you want to forward all traffic with unchanged ports, please specify the protocol type as `Any` and do not set the internal port and public external port.",
			},
			"private_ip_address": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The private ip address.",
			},
			"private_port": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The internal port or port segment for DNAT rule port forwarding. You can use a hyphen (`-`) to specify a port range, e.g. 80-100. The number of public and private ports must be consistent. The value range is 1-65535.",
			},
			"public_port": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The external public port or port segment for DNAT rule port forwarding. You can use a hyphen (`-`) to specify a port range, e.g. 80-100. The number of public and private ports must be consistent. The value range is 1-65535. If no port is specified, all traffic will be forwarded with the destination port unchanged.",
			},
			"eip_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the public EIP.",
			},
		},
	}
}

func resourceZenlayerCloudZecVpcNatGatewayDnatDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	resourceId, err := common2.ParseResourceId(d.Id(), 2)
	if err != nil {
		return diag.FromErr(err)
	}

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := zecService.DeleteDnatEntry(ctx, resourceId[1])
		if errRet != nil {
			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if !ok {
				return common2.RetryError(ctx, errRet, common2.InternalServerError)
			}
			if ee.Code == common2.ResourceNotFound {
				return nil
			}
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceZenlayerCloudZecVpcNatGatewayDnatUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	//var _ diag.Diagnostics
	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	resourceId, err := common2.ParseResourceId(d.Id(), 2)
	if err != nil {
		return diag.FromErr(err)
	}
	dnatId := resourceId[1]

	if d.HasChanges("private_ip_address", "protocol", "public_port", "private_port", "eip_id") {
		request := zec2.NewModifyDnatEntryRequest()
		request.DnatEntryId = &dnatId
		request.Protocol = common.String(d.Get("protocol").(string))
		request.EipId = common.String(d.Get("eip_id").(string))
		request.PrivateIp = common.String(d.Get("private_ip_address").(string))

		if v, ok := d.GetOk("private_port"); ok {
			request.InternalPort = common.String(v.(string))
		}

		if v, ok := d.GetOk("public_port"); ok {
			request.ListenerPort = common.String(v.(string))
		}

		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {

			_, err := zecService.client.WithZec2Client().ModifyDnatEntry(request)

			// log

			if err != nil {
				return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceZenlayerCloudZecVpcNatGatewayDnatRead(ctx, d, meta)
}

func resourceZenlayerCloudZecVpcNatGatewayDnatCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	request := zec2.NewCreateDnatEntryRequest()
	request.NatGatewayId = common.String(d.Get("nat_gateway_id").(string))
	request.EipId = common.String(d.Get("eip_id").(string))
	request.Protocol = common.String(d.Get("protocol").(string))
	request.PrivateIp = common.String(d.Get("private_ip_address").(string))

	if v, ok := d.GetOk("private_port"); ok {
		request.InternalPort = common.String(v.(string))
	}
	if v, ok := d.GetOk("public_port"); ok {
		request.ListenerPort = common.String(v.(string))
	}

	dnatId := ""

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := zecService.client.WithZec2Client().CreateDnatEntry(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create DNAT entry.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": common2.ToJsonString(request),
				"err":     err.Error(),
			})
			return common2.RetryError(ctx, err)
		}

		tflog.Info(ctx, "Create DNAT entry success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  common2.ToJsonString(request),
			"response": common2.ToJsonString(response),
		})

		if response.Response.DnatEntryId == nil {
			err = fmt.Errorf("dnat entry id is nil")
			return resource.NonRetryableError(err)
		}
		dnatId = *response.Response.DnatEntryId

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s:%s", *request.NatGatewayId, dnatId))

	return resourceZenlayerCloudZecVpcNatGatewayDnatRead(ctx, d, meta)
}

func resourceZenlayerCloudZecVpcNatGatewayDnatRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	resourceId, err := common2.ParseResourceId(d.Id(), 2)
	if err != nil {
		return diag.FromErr(err)
	}

	natGatewayId := resourceId[0]
	dnatId := resourceId[1]

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var natGateway *zec2.DescribeNatGatewayDetailResponseParams
	var errRet error

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		natGateway, errRet = zecService.DescribeNatGatewayDetailById(ctx, natGatewayId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if natGateway == nil {
		d.SetId("")
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "NAT gateway not exist",
			Detail:   fmt.Sprintf("The NAT gateway %s is not exist", natGatewayId),
		})
		return diags
	}
	var dnatEntry *zec2.DnatEntry

	for _, snat := range natGateway.Dnats {
		if *snat.DnatEntryId == dnatId {
			dnatEntry = snat
		}
	}

	if dnatEntry == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "DNat entry not exist",
			Detail:   fmt.Sprintf("The DNat entry %s is not found in NAT gateway %s", dnatId, natGatewayId),
		})
		return diags
	}

	_ = d.Set("eip_id", dnatEntry.EipId)
	_ = d.Set("protocol", dnatEntry.Protocol)
	_ = d.Set("private_ip_address", dnatEntry.PrivateIp)
	_ = d.Set("private_port", dnatEntry.InternalPort)
	_ = d.Set("public_port", dnatEntry.ListenerPort)
	_ = d.Set("nat_gateway_id", natGatewayId)

	return diags
}

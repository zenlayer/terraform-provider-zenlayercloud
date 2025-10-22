package zec

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	zec2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"
	"time"
)

func ResourceZenlayerCloudZecVpcNatGatewaySnat() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZecVpcNatGatewaySnatCreate,
		ReadContext:   resourceZenlayerCloudZecVpcNatGatewaySnatRead,
		UpdateContext: resourceZenlayerCloudZecVpcNatGatewaySnatUpdate,
		DeleteContext: resourceZenlayerCloudZecVpcNatGatewaySnatDelete,

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
			"subnet_ids": {
				Type:          schema.TypeSet,
				Elem:          &schema.Schema{Type: schema.TypeString},
				Optional:      true,
				AtLeastOneOf:  []string{"subnet_ids", "source_cidr_blocks"},
				ConflictsWith: []string{"source_cidr_blocks"},
				Description:   "IDs of the subnets to be associated with the SNAT entry. Cannot be used with `source_cidr_blocks`.",
			},
			"source_cidr_blocks": {
				Type:          schema.TypeSet,
				Elem:          &schema.Schema{Type: schema.TypeString},
				Optional:      true,
				AtLeastOneOf:  []string{"subnet_ids", "source_cidr_blocks"},
				ConflictsWith: []string{"subnet_ids"},
				Description:   "Source CIDR blocks to be associated with the SNAT entry. Cannot be used with `subnet_ids`.",
			},
			"eip_ids": {
				Type:          schema.TypeSet,
				Elem:          &schema.Schema{Type: schema.TypeString},
				Optional:      true,
				AtLeastOneOf:  []string{"eip_ids", "is_all_eip"},
				ConflictsWith: []string{"is_all_eip"},
				Description:   "IDs of the public EIPs to be associated. This field is conflict with `is_all_eip`. This field is conflict with `is_all_eip`.",
			},
			"is_all_eip": {
				Type:          schema.TypeBool,
				Optional:      true,
				AtLeastOneOf:  []string{"eip_ids", "is_all_eip"},
				ConflictsWith: []string{"eip_ids"},
				Description:   "Indicates whether all the EIPs of region is assigned to SNAT entry. This field is conflict with `eip_ids`.",
			},
		},
	}
}

func resourceZenlayerCloudZecVpcNatGatewaySnatDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	resourceId, err := common2.ParseResourceId(d.Id(), 2)
	if err != nil {
		return diag.FromErr(err)
	}

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := zecService.DeleteSnatEntry(ctx, resourceId[1])
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

func resourceZenlayerCloudZecVpcNatGatewaySnatUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	//var _ diag.Diagnostics
	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	resourceId, err := common2.ParseResourceId(d.Id(), 2)
	if err != nil {
		return diag.FromErr(err)
	}
	snatId := resourceId[1]

	if d.HasChanges("eip_ids", "is_all_eip", "subnet_ids", "source_cidr_blocks") {
		request := zec2.NewModifySnatEntryRequest()
		request.SnatEntryId = &snatId
		request.IsAllEip = common.Bool(d.Get("is_all_eip").(bool))
		request.EipIds = common2.ToStringList(d.Get("eip_ids").(*schema.Set).List())
		request.SubnetIds = common2.ToStringList(d.Get("subnet_ids").(*schema.Set).List())
		request.SourceCidrBlocks = common2.ToStringList(d.Get("source_cidr_blocks").(*schema.Set).List())

		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {

			_, err := zecService.client.WithZec2Client().ModifySnatEntry(request)
			if err != nil {
				return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceZenlayerCloudZecVpcNatGatewaySnatRead(ctx, d, meta)
}

func resourceZenlayerCloudZecVpcNatGatewaySnatCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	request := zec2.NewCreateSnatEntryRequest()
	request.NatGatewayId = common.String(d.Get("nat_gateway_id").(string))
	request.EipIds = common2.ToStringList(d.Get("eip_ids").(*schema.Set).List())

	if v, ok := d.GetOk("eip_ids"); ok {
		eipIds := v.(*schema.Set).List()
		if len(eipIds) > 0 {
			request.EipIds = common2.ToStringList(eipIds)
		}
	}

	if v, ok := d.GetOk("subnet_ids"); ok {
		subnetIds := v.(*schema.Set).List()
		if len(subnetIds) > 0 {
			request.SubnetIds = common2.ToStringList(subnetIds)
		}
	}

	if v, ok := d.GetOk("source_cidr_blocks"); ok {
		sourceCidrBlocks := v.(*schema.Set).List()
		if len(sourceCidrBlocks) > 0 {
			request.SourceCidrBlocks = common2.ToStringList(sourceCidrBlocks)
		}
	}

	snatId := ""

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := zecService.client.WithZec2Client().CreateSnatEntry(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create SNAT entry.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": common2.ToJsonString(request),
				"err":     err.Error(),
			})
			return common2.RetryError(ctx, err)
		}

		tflog.Info(ctx, "Create SNAT entry success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  common2.ToJsonString(request),
			"response": common2.ToJsonString(response),
		})

		if response.Response.SnatEntryId == nil {
			err = fmt.Errorf("snat entry id is nil")
			return resource.NonRetryableError(err)
		}
		snatId = *response.Response.SnatEntryId

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s:%s", *request.NatGatewayId, snatId))

	return resourceZenlayerCloudZecVpcNatGatewaySnatRead(ctx, d, meta)
}

func resourceZenlayerCloudZecVpcNatGatewaySnatRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	resourceId, err := common2.ParseResourceId(d.Id(), 2)
	if err != nil {
		return diag.FromErr(err)
	}

	natGatewayId := resourceId[0]
	snatId := resourceId[1]

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
	var snatEntry *zec2.SnatEntry

	for _, snat := range natGateway.Snats {
		if *snat.SnatEntryId == snatId {
			snatEntry = snat
		}
	}

	if snatEntry == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Snat entry not exist",
			Detail:   fmt.Sprintf("The snat %s is not found in NAT gateway %s", snatId, natGatewayId),
		})
		return diags
	}

	_ = d.Set("is_all_eip", snatEntry.IsAllEip)
	if !*snatEntry.IsAllEip {
		_ = d.Set("eip_ids", snatEntry.EipIds)
	}

	if len(snatEntry.SnatSubnets) > 0 {
		var subnetIds []string
		for _, subnet := range snatEntry.SnatSubnets {
			subnetIds = append(subnetIds, *subnet.SubnetId)
		}
		_ = d.Set("subnet_ids", subnetIds)
	} else {
		_ = d.Set("source_cidr_block", snatEntry.Cidrs)
	}
	_ = d.Set("nat_gateway_id", natGatewayId)
	_ = d.Set("nat_gateway_id", natGatewayId)

	return diags
}

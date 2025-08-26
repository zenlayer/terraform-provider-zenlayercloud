package zec

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20240401"
)

func ResourceZenlayerCloudBorderGatewayAssociation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudBorderGatewayAssociationCreate,
		ReadContext:   resourceZenlayerCloudBorderGatewayAssociationRead,
		DeleteContext: resourceZenlayerCloudBorderGatewayAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"zbg_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the border gateway.",
			},
			"nat_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the NAT gateway.",
			},
		},
	}
}

func resourceZenlayerCloudBorderGatewayAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_border_gateway_association.create")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	zbgId := d.Get("zbg_id").(string)
	natId := d.Get("nat_id").(string)

	request := zec.NewAssignBorderGatewayRequest()
	request.ZbgId = zbgId
	request.NatId = natId

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := zecService.client.WithZecClient().AssignBorderGateway(request)
		if err != nil {
			return common.RetryError(ctx, err)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s:%s", zbgId, natId))

	return resourceZenlayerCloudBorderGatewayAssociationRead(ctx, d, meta)
}

func resourceZenlayerCloudBorderGatewayAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_border_gateway_attachment.read")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	zbgId := d.Get("zbg_id").(string)

	borderGateway, err := zecService.DescribeBorderGatewayById(ctx, zbgId)
	if err != nil {
		return diag.FromErr(err)
	}

	if borderGateway == nil {
		d.SetId("")
		return nil
	}

	// Check if the NAT gateway is attached
	if borderGateway.NatId == "" {
		d.SetId("")
		return nil
	}

	_ = d.Set("zbg_id", borderGateway.ZbgId)
	_ = d.Set("nat_id", borderGateway.NatId)

	return nil
}

func resourceZenlayerCloudBorderGatewayAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_border_gateway_attachment.delete")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	zbgId := d.Get("zbg_id").(string)

	request := zec.NewUnassignBorderGatewayRequest()
	request.ZbgId = zbgId

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := zecService.client.WithZecClient().UnassignBorderGateway(request)
		if err != nil {
			return common.RetryError(ctx, err)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

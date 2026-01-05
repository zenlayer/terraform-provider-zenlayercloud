package zec

import (
	"context"
	"fmt"
	"strings"
	"time"

	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"
)

func ResourceZenlayerCloudZecDhcpOptionsSetAttachment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZecDhcpOptionsSetAttachmentCreate,
		ReadContext:   resourceZenlayerCloudZecDhcpOptionsSetAttachmentRead,
		DeleteContext: resourceZenlayerCloudZecDhcpOptionsSetAttachmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"subnet_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Subnet ID.",
			},
			"dhcp_options_set_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "DHCP options set ID.",
			},
		},
	}
}

func resourceZenlayerCloudZecDhcpOptionsSetAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zenlayerCloudClient := meta.(*connectivity.ZenlayerCloudClient)

	request := zec.NewAttachDhcpOptionsSetToSubnetRequest()
	request.DhcpOptionsSetId = common.String(d.Get("dhcp_options_set_id").(string))
	subnetId := d.Get("subnet_id").(string)
	request.SubnetIds = []string{subnetId}

	err := resource.RetryContext(ctx, 5*time.Minute, func() *resource.RetryError {
		_, err := zenlayerCloudClient.WithZec2Client().AttachDhcpOptionsSetToSubnet(request)
		if err != nil {
			return common2.RetryError(ctx, err, common2.OperationTimeout)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("fail to associate dhcp options set after retries: %v", err))
	}

	// The ID will be a combination of subnet_id and dhcp_options_set_id
	d.SetId(fmt.Sprintf("%s:%s", subnetId, *request.DhcpOptionsSetId))

	return resourceZenlayerCloudZecDhcpOptionsSetAttachmentRead(ctx, d, meta)
}

func resourceZenlayerCloudZecDhcpOptionsSetAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zenlayerCloudClient := meta.(*connectivity.ZenlayerCloudClient)
	zecService := &ZecService{client: zenlayerCloudClient}

	// Parse the ID to get dhcp_options_set_id and subnet_id
	subnetId, dhcpOptionsSetId, err := parseDhcpOptionsSetAttachmentId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Check if the subnet exists and has the DHCP options set associated
	subnet, err := zecService.DescribeSubnetById(ctx, subnetId)
	if err != nil {
		return diag.FromErr(err)
	}

	if subnet == nil {
		d.SetId("")
		return nil
	}

	// Verify that the subnet is associated with the DHCP options set
	if common.ToString(subnet.DhcpOptionsSetId) != dhcpOptionsSetId {
		d.SetId("")
		return nil
	}

	_ = d.Set("dhcp_options_set_id", dhcpOptionsSetId)
	_ = d.Set("subnet_id", subnetId)

	return nil
}

func resourceZenlayerCloudZecDhcpOptionsSetAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zenlayerCloudClient := meta.(*connectivity.ZenlayerCloudClient)

	// Parse the ID to get subnet_id and dhcp_options_set_id
	subnetId, _, err := parseDhcpOptionsSetAttachmentId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	request := zec.NewDetachDhcpOptionsSetFromSubnetRequest()
	request.SubnetIds = []string{subnetId}

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		_, errRet := zenlayerCloudClient.WithZec2Client().DetachDhcpOptionsSetFromSubnet(request)
		if errRet != nil {
			return common2.RetryError(ctx, errRet, common2.OperationTimeout)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(fmt.Errorf("fail to disassociate dhcp options set: %v", err))
	}

	return nil
}

func parseDhcpOptionsSetAttachmentId(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid dhcp options set attachment id format: %s, expected dhcp_options_set_id:subnet_id", id)
	}
	return parts[0], parts[1], nil
}

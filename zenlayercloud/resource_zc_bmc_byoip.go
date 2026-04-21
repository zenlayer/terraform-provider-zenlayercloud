/*
Provide a resource to create a BYOIP (Bring Your Own IP) in BMC.

Example Usage

```hcl
resource "zenlayercloud_bmc_byoip" "foo" {
  ip_type                     = "IPv4"
  cidr                        = "203.0.113.0/24"
  asn                         = 65001
  public_virtual_interface_id = "xxxxxxxx"
}
```

Import

BYOIP can be imported using the cidr block ID, e.g.

```
$ terraform import zenlayercloud_bmc_byoip.foo cidr-block-id
```
*/
package zenlayercloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	bmc2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/bmc20260201"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
)

func resourceZenlayerCloudBmcByoip() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudBmcByoipCreate,
		ReadContext:   resourceZenlayerCloudBmcByoipRead,
		DeleteContext: resourceZenlayerCloudBmcByoipDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"ip_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"IPv4", "IPv6"}, false),
				Description:  "IP type. Valid values: `IPv4`, `IPv6`.",
			},
			"cidr": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The announced IPv4 or IPv6 CIDR block.",
			},
			"asn": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "ASN number of the announced CIDR block.",
			},
			"public_virtual_interface_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The unique ID of the public virtual interface (public VLAN).",
			},
			"cidr_block_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the CIDR block.",
			},
			"cidr_block_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of the CIDR block.",
			},
			"zone_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The zone ID that the CIDR block locates at.",
			},
			"gateway": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Gateway address of the CIDR block.",
			},
			"available_ip_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of available IPs in the CIDR block.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Current status of the CIDR block.",
			},
			"charge_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Charge type of the CIDR block.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The resource group ID the CIDR block belongs to.",
			},
			"resource_group_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The resource group name the CIDR block belongs to.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation time of the CIDR block.",
			},
			"expire_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Expiration time of the CIDR block.",
			},
		},
	}
}

func resourceZenlayerCloudBmcByoipCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*connectivity.ZenlayerCloudClient)

	request := bmc2.NewCreateByoipRequest()
	request.IpType = common.String(byoipIpTypeToAPI(d.Get("ip_type").(string)))
	request.Cidr = common.String(d.Get("cidr").(string))
	if v, ok := d.GetOk("asn"); ok {
		asn := int64(v.(int))
		request.Asn = &asn
	}
	if v, ok := d.GetOk("public_virtual_interface_id"); ok {
		request.PublicVirtualInterfaceId = common.String(v.(string))
	}

	var cidrBlockId string

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := client.WithBmc2Client().CreateByoip(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create BYOIP.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": common2.ToJsonString(request),
				"err":     err.Error(),
			})
			return common2.RetryError(ctx, err)
		}

		tflog.Info(ctx, "Create BYOIP success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  common2.ToJsonString(request),
			"response": common2.ToJsonString(response),
		})

		if response.Response == nil || response.Response.CidrBlockId == nil || *response.Response.CidrBlockId == "" {
			return resource.NonRetryableError(fmt.Errorf("cidr block id is empty in CreateByoip response"))
		}
		cidrBlockId = *response.Response.CidrBlockId
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(cidrBlockId)

	stateConf := &resource.StateChangeConf{
		Pending:        []string{BmcCidrBlockStatusCreating},
		Target:         []string{BmcCidrBlockStatusAvailable},
		Refresh:        bmcCidrBlockStateRefreshFunc(ctx, client, cidrBlockId, []string{BmcCidrBlockStatusFailed}),
		Timeout:        d.Timeout(schema.TimeoutCreate) - time.Minute,
		Delay:          5 * time.Second,
		MinTimeout:     3 * time.Second,
		NotFoundChecks: 3,
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for BYOIP cidr block (%s) to become available: %v", cidrBlockId, err))
	}

	return resourceZenlayerCloudBmcByoipRead(ctx, d, meta)
}

func resourceZenlayerCloudBmcByoipRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*connectivity.ZenlayerCloudClient)
	cidrBlockId := d.Id()

	request := bmc2.NewDescribeCidrBlocksRequest()
	request.CidrBlockIds = []string{cidrBlockId}

	var cidrBlock *bmc2.CidrBlockInfo
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		response, err := client.WithBmc2Client().DescribeCidrBlocks(request)
		if err != nil {
			return common2.RetryError(ctx, err)
		}
		if response.Response != nil && len(response.Response.DataSet) > 0 {
			cidrBlock = response.Response.DataSet[0]
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if cidrBlock == nil {
		d.SetId("")
		tflog.Info(ctx, "BYOIP cidr block not found", map[string]interface{}{
			"cidrBlockId": cidrBlockId,
		})
		return nil
	}

	_ = d.Set("cidr", cidrBlock.CidrBlock)
	_ = d.Set("cidr_block_name", cidrBlock.CidrBlockName)
	_ = d.Set("cidr_block_type", cidrBlock.CidrBlockType)
	_ = d.Set("zone_id", cidrBlock.ZoneId)
	_ = d.Set("gateway", cidrBlock.Gateway)
	_ = d.Set("available_ip_count", cidrBlock.AvailableIpCount)
	_ = d.Set("status", cidrBlock.Status)
	_ = d.Set("charge_type", cidrBlock.ChargeType)
	_ = d.Set("resource_group_id", cidrBlock.ResourceGroupId)
	_ = d.Set("resource_group_name", cidrBlock.ResourceGroupName)
	_ = d.Set("create_time", cidrBlock.CreateTime)
	_ = d.Set("expire_time", cidrBlock.ExpireTime)

	return nil
}

func describeBmcCidrBlockById(ctx context.Context, client *connectivity.ZenlayerCloudClient, cidrBlockId string) (*bmc2.CidrBlockInfo, error) {
	request := bmc2.NewDescribeCidrBlocksRequest()
	request.CidrBlockIds = []string{cidrBlockId}
	response, err := client.WithBmc2Client().DescribeCidrBlocks(request)
	if err != nil {
		return nil, err
	}
	if response.Response != nil && len(response.Response.DataSet) > 0 {
		return response.Response.DataSet[0], nil
	}
	return nil, nil
}

func bmcCidrBlockStateRefreshFunc(ctx context.Context, client *connectivity.ZenlayerCloudClient, cidrBlockId string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cidrBlock, err := describeBmcCidrBlockById(ctx, client, cidrBlockId)
		if err != nil {
			return nil, "", err
		}
		if cidrBlock == nil {
			return nil, "", nil
		}
		status := *cidrBlock.Status
		for _, f := range failStates {
			if status == f {
				return cidrBlock, status, common2.Error("BYOIP cidr block reached failed state: %s", status)
			}
		}
		return cidrBlock, status, nil
	}
}

// byoipIpTypeToAPI converts the user-facing `IPv4`/`IPv6` values to the
// uppercase `IPV4`/`IPV6` values expected by the CreateByoip API.
func byoipIpTypeToAPI(ipType string) string {
	switch ipType {
	case "IPv4":
		return "IPV4"
	case "IPv6":
		return "IPV6"
	default:
		return ipType
	}
}

func resourceZenlayerCloudBmcByoipDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*connectivity.ZenlayerCloudClient)
	cidrBlockId := d.Id()

	request := bmc2.NewTerminateCidrBlockRequest()
	request.CidrBlockId = common.String(cidrBlockId)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := client.WithBmc2Client().TerminateCidrBlock(request)
		if err != nil {
			if sdkErr, ok := err.(*common.ZenlayerCloudSdkError); ok {
				if sdkErr.Code == common2.ResourceNotFound {
					return nil
				}
			}
			return common2.RetryError(ctx, err)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

package zec

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	zec2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"
	"time"
)

func ResourceZenlayerCloudZecVNicIPv4() *schema.Resource {

	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZecVNicIPv4Create,
		ReadContext:   resourceZenlayerCloudZecVNicIPv4Read,
		UpdateContext: resourceZenlayerCloudZecVNicIPv4Update,
		DeleteContext: resourceZenlayerCloudZecVNicIPv4Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"vnic_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the vNIC.",
			},
			"secondary_private_ip_addresses": {
				Type:          schema.TypeSet,
				Elem:          &schema.Schema{Type: schema.TypeString},
				Optional:      true,
				Computed:      true,
				AtLeastOneOf:  []string{"secondary_private_ip_count", "secondary_private_ip_addresses"},
				ConflictsWith: []string{"secondary_private_ip_count"},
				Description:   "Assign specified secondary private ipv4 address. This IP address must be an available IP address within the CIDR block of the subnet to which the vNIC belongs.",
			},
			"secondary_private_ip_count": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				AtLeastOneOf:  []string{"secondary_private_ip_count", "secondary_private_ip_addresses"},
				ConflictsWith: []string{"secondary_private_ip_addresses"},
				Description:   "The number of newly-applied private IP addresses.",
			},
		},
	}
}

func resourceZenlayerCloudZecVNicIPv4Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zec_vnic_ipv4.update")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	vnicId := d.Id()
	d.Partial(true)

	if d.HasChange("secondary_private_ip_addresses") {
		oldPrivateIpAddresses, newPrivateIpAddresses := d.GetChange("secondary_private_ip_addresses")
		oldPrivateIpAddressesSet := oldPrivateIpAddresses.(*schema.Set)
		newPrivateIpAddressesSet := newPrivateIpAddresses.(*schema.Set)

		removed := oldPrivateIpAddressesSet.Difference(newPrivateIpAddressesSet)
		added := newPrivateIpAddressesSet.Difference(oldPrivateIpAddressesSet)

		if removed.Len() > 0 {

			request := zec2.NewUnassignNetworkInterfaceIpv4Request()
			request.NicId = &vnicId
			request.IpAddresses = common2.ToStringList(removed.List())

			if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
			_, errRet := zecService.client.WithZec2Client().UnassignNetworkInterfaceIpv4(request)
				if errRet != nil {
					return common2.RetryError(ctx, errRet)
				}
				return nil
			}); err != nil {
				return diag.FromErr(err)
			}
		}
		if added.Len() > 0 {

			request := zec2.NewBatchAssignNetworkInterfaceIpv4Request()
			request.IpAddresses = common2.ToStringList(added.List())
			request.NicId = &vnicId

			if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
			_, errRet := zecService.client.WithZec2Client().BatchAssignNetworkInterfaceIpv4(request)
				if errRet != nil {
					return common2.RetryError(ctx, errRet)
				}
				return nil
			}); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("secondary_private_ip_count") {
		privateIpList := common2.ToStringList(d.Get("secondary_private_ip_addresses").(*schema.Set).List())

		oldIpsCount, newIpsCount := d.GetChange("secondary_private_ip_count")
		if oldIpsCount != nil && newIpsCount != nil && newIpsCount != len(privateIpList) {
			diff := newIpsCount.(int) - oldIpsCount.(int)
			if diff > 0 {

				request := zec2.NewBatchAssignNetworkInterfaceIpv4Request()
				request.IpAddressCount = &diff
				request.NicId = &vnicId

				if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
					_, errRet := zecService.client.WithZec2Client().BatchAssignNetworkInterfaceIpv4(request)
					if errRet != nil {
						return common2.RetryError(ctx, errRet)
					}
					return nil
				}); err != nil {
					return diag.FromErr(err)
				}
			}

			if diff < 0 {
				diff *= -1
				unAssignIps := privateIpList[:diff]

				request := zec2.NewUnassignNetworkInterfaceIpv4Request()
				request.NicId = &vnicId
				request.IpAddresses = unAssignIps

				if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
					_, errRet := zecService.client.WithZec2Client().UnassignNetworkInterfaceIpv4(request)
					if errRet != nil {
						return common2.RetryError(ctx, errRet)
					}
					return nil
				}); err != nil {
					return diag.FromErr(err)
				}

			}
		}
	}

	d.Partial(false)
	return nil
}

func resourceZenlayerCloudZecVNicIPv4Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zec_vnic_attachment.delete")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	vnicId := d.Id()

	var vnic *zec2.NicInfo
	var err error
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		vnic, err = zecService.DescribeNicById(ctx, vnicId)
		if err != nil {
			return common2.RetryError(ctx, err)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	request := zec2.NewUnassignNetworkInterfaceIpv4Request()
	request.NicId = &vnicId
	request.IpAddresses = vnic.SecondaryIpv4s

	if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		_, errRet := zecService.client.WithZec2Client().UnassignNetworkInterfaceIpv4(request)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		return nil
	}); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceZenlayerCloudZecVNicIPv4Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zec_vnic_attachment.create")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	vnicId := d.Get("vnic_id").(string)

	request := zec2.NewBatchAssignNetworkInterfaceIpv4Request()
	if v, ok := d.GetOk("secondary_private_ip_addresses"); ok {
		ipAddresses := v.(*schema.Set).List()
		if len(ipAddresses) > 0 {
			request.IpAddresses = common2.ToStringList(ipAddresses)
		}
	}
	if v, ok := d.GetOk("secondary_private_ip_count"); ok {
		request.IpAddressCount = common.Integer(v.(int))
	}

	request.NicId = &vnicId

	if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		_, errRet := zecService.client.WithZec2Client().BatchAssignNetworkInterfaceIpv4(request)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		return nil
	}); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(vnicId)

	return resourceZenlayerCloudZecVNicIPv4Read(ctx, d, meta)
}

func resourceZenlayerCloudZecVNicIPv4Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zec_vnic_ipv4.read")()

	vnicId := d.Id()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var vnic *zec2.NicInfo
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		vnic, errRet = zecService.DescribeNicById(ctx, vnicId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if vnic == nil || len(vnic.SecondaryIpv4s) == 0 {
		d.SetId("")
		return nil
	}

	_ = d.Set("secondary_private_ip_count", len(vnic.SecondaryIpv4s))
	_ = d.Set("secondary_private_ip_addresses", vnic.SecondaryIpv4s)
	return nil
}

package zec

import (
	"context"
	"fmt"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/services/zrm"
	user "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/user20240529"
	"strconv"
	"time"

	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"

	common2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"
)

func ResourceZenlayerCloudZecDhcpOptionsSet() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZecDhcpOptionsSetCreate,
		ReadContext:   resourceZenlayerCloudZecDhcpOptionsSetRead,
		UpdateContext: resourceZenlayerCloudZecDhcpOptionsSetUpdate,
		DeleteContext: resourceZenlayerCloudZecDhcpOptionsSetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the DHCP options set.",
			},
			"domain_name_servers": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "IPv4 DNS server IP, up to 4 IPv4 addresses, separated by commas.",
			},
			"ipv6_domain_name_servers": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "IPv6 DNS server IP, up to 4 IPv6 addresses.",
				//ValidateFunc: validation.StringMatch(regexp.MustCompile(`^$|^(([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4})(,(([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4})){0,3}$`), "ipv6DomainNameServers must be up to 4 IPv6 addresses separated by commas"),
			},
			"lease_time": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "IPv4 lease time, measured in hour. Value range: 24h~1176h, 87600h(3650d)~175200h(7300d), default value: 24.",
				//ValidateFunc: validation.StringMatch(regexp.MustCompile(`^(1d|(?:[2-9]d|(?:[1-3][0-9]|4[0-9])d))$|^(?:(?:36[5-9][0-9]|[37][0-9]{3}|4[0-9]{3}|5[0-9]{3}|6[0-9]{3}|7[0-2][0-9]{2}|7300)d)$`), "leaseTime must be between 1d~49d or 3650d~7300d"),
			},
			"ipv6_lease_time": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "IPv6 lease time, measured in hour. Value range: 24h~1176h, 87600h(3650d)~175200h(7300d). default value: 24.",
				//ValidateFunc: validation.StringMatch(regexp.MustCompile(`^(1d|(?:[2-9]d|(?:[1-3][0-9]|4[0-9])d))$|^(?:(?:36[5-9][0-9]|[37][0-9]{3}|4[0-9]{3}|5[0-9]{3}|6[0-9]{3}|7[0-2][0-9]{2}|7300)d)$`), "ipv6LeaseTime must be between 1d~49d or 3650d~7300d"),
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "ID of the resource group to which the DHCP options set belongs.",
			},
			"resource_group_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Name of resource group the DHCP options set belongs to.",
			},
			"tags": {
				Type:        schema.TypeMap,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "Tags of the DHCP options set.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the DHCP options set.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation time of the DHCP options set.",
			},
		},
	}
}

func resourceZenlayerCloudZecDhcpOptionsSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zenlayerCloudClient := meta.(*connectivity.ZenlayerCloudClient)

	request := zec.NewCreateDhcpOptionsSetRequest()
	request.DhcpOptionsSetName = common2.String(d.Get("name").(string))

	domainNameServers := d.Get("domain_name_servers").(string)
	request.DomainNameServers = common2.String(domainNameServers)

	if v, ok := d.GetOk("ipv6_domain_name_servers"); ok {
		request.Ipv6DomainNameServers = common2.String(v.(string))
	}

	if leaseTime, ok := d.GetOk("lease_time"); ok {
		request.LeaseTime = common2.String(strconv.Itoa(leaseTime.(int)))
	}

	if ipv6LeaseTime, ok := d.GetOk("ipv6_lease_time"); ok {
		request.Ipv6LeaseTime = common2.String(strconv.Itoa(ipv6LeaseTime.(int)))
	}

	if description, ok := d.GetOk("description"); ok {
		request.Description = common2.String(description.(string))
	}

	// 设置资源组
	if resourceGroupId, ok := d.GetOk("resource_group_id"); ok {
		request.ResourceGroupId = common2.String(resourceGroupId.(string))
	}

	if tags := common.GetTags(d, "tags"); len(tags) > 0 {
		request.Tags = &zec.TagAssociation{}
		for k, v := range tags {
			tmpKey := k
			tmpValue := v
			request.Tags.Tags = append(request.Tags.Tags, &zec.Tag{
				Key:   &tmpKey,
				Value: &tmpValue,
			})
		}
	}
	if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := zenlayerCloudClient.WithZec2Client().CreateDhcpOptionsSet(request)
		if err != nil {
			return common.RetryError(ctx, err)
		}
		d.SetId(*response.Response.DhcpOptionsSetId)
		return nil
	}); err != nil {
		return diag.FromErr(fmt.Errorf("fail to create dhcp options set: %v", err))
	}

	return resourceZenlayerCloudZecDhcpOptionsSetRead(ctx, d, meta)
}

func resourceZenlayerCloudZecDhcpOptionsSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zenlayerCloudClient := meta.(*connectivity.ZenlayerCloudClient)
	zecService := &ZecService{client: zenlayerCloudClient}

	dhcpOptionsSet, err := zecService.DescribeDhcpOptionsSetById(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if dhcpOptionsSet == nil {
		d.SetId("")
		return nil
	}

	_ = d.Set("name", dhcpOptionsSet.DhcpOptionsSetName)
	_ = d.Set("domain_name_servers", dhcpOptionsSet.DomainNameServers)
	_ = d.Set("ipv6_domain_name_servers", dhcpOptionsSet.Ipv6DomainNameServers)
	if dhcpOptionsSet.LeaseTime != nil {
		if leaseTime, err := strconv.Atoi(*dhcpOptionsSet.LeaseTime); err == nil {
			_ = d.Set("lease_time", leaseTime)
		}
	}
	if dhcpOptionsSet.Ipv6LeaseTime != nil {
		if ipv6LeaseTime, err := strconv.Atoi(*dhcpOptionsSet.Ipv6LeaseTime); err == nil {
			_ = d.Set("ipv6_lease_time", ipv6LeaseTime)
		}
	}
	_ = d.Set("description", dhcpOptionsSet.Description)
	_ = d.Set("create_time", dhcpOptionsSet.CreateTime)
	_ = d.Set("resource_group_id", dhcpOptionsSet.ResourceGroupId)
	_ = d.Set("resource_group_name", dhcpOptionsSet.ResourceGroupName)

	// 读取标签
	tagMap, errRet := common.TagsToMap(dhcpOptionsSet.Tags)
	if errRet != nil {
		return diag.FromErr(errRet)
	}
	_ = d.Set("tags", tagMap)

	return nil
}

func resourceZenlayerCloudZecDhcpOptionsSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zenlayerCloudClient := meta.(*connectivity.ZenlayerCloudClient)

	request := zec.NewModifyDhcpOptionsSetAttributesRequest()
	dhcpOptionsSetId := d.Id()
	request.DhcpOptionsSetId = common2.String(dhcpOptionsSetId)

	if d.HasChange("domain_name_servers") {
		request.DomainNameServers = common2.String(d.Get("domain_name_servers").(string))
	}

	if d.HasChange("ipv6_domain_name_servers") {
		request.Ipv6DomainNameServers = common2.String(d.Get("ipv6_domain_name_servers").(string))
	}

	if d.HasChange("lease_time") {
		request.LeaseTime = common2.String(strconv.Itoa(d.Get("lease_time").(int)))
	}

	if d.HasChange("ipv6_lease_time") {
		if v, ok := d.GetOk("ipv6_lease_time"); ok {
			request.Ipv6LeaseTime = common2.String(strconv.Itoa(v.(int)))
		} else {
			request.Ipv6LeaseTime = common2.String("")
		}

	}

	if d.HasChange("description") {
		request.Description = common2.String(d.Get("description").(string))
	}

	if d.HasChange("name") {
		request.DhcpOptionsSetName = common2.String(d.Get("name").(string))
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		_, errRet := zenlayerCloudClient.WithZec2Client().ModifyDhcpOptionsSetAttributes(request)
		if errRet != nil {
			return common.RetryError(ctx, errRet, common.OperationTimeout)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("fail to modify dhcp options set: %v", err))
	}

	// 更新标签
	if d.HasChange("tags") {
		zrmService := zrm.NewZrmService(zenlayerCloudClient)
		if err := zrmService.ModifyResourceTags(ctx, d, dhcpOptionsSetId); err != nil {
			return diag.FromErr(fmt.Errorf("fail to update tags for dhcp options set: %v", err))
		}
	}

	if d.HasChange("resource_group_id") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			request := user.NewAddResourceResourceGroupRequest()
			request.ResourceGroupId = common2.String(d.Get("resource_group_id").(string))
			request.Resources = []string{dhcpOptionsSetId}

			_, err := meta.(*connectivity.ZenlayerCloudClient).WithUsrClient().AddResourceResourceGroup(request)
			if err != nil {
				return common.RetryError(ctx, err, common.InternalServerError, common2.NetworkError, common.OperationTimeout)
			}
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceZenlayerCloudZecDhcpOptionsSetRead(ctx, d, meta)
}

func resourceZenlayerCloudZecDhcpOptionsSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zenlayerCloudClient := meta.(*connectivity.ZenlayerCloudClient)
	zecService := &ZecService{client: zenlayerCloudClient}

	request := zec.NewDeleteDhcpOptionsSetRequest()
	request.DhcpOptionsSetId = common2.String(d.Id())

	_, err := zenlayerCloudClient.WithZec2Client().DeleteDhcpOptionsSet(request)
	if err != nil {
		return diag.FromErr(fmt.Errorf("fail to delete dhcp options set: %v", err))
	}

	// Wait for the DHCP options set to be deleted
	return diag.FromErr(resource.RetryContext(ctx, 5*time.Minute, func() *resource.RetryError {
		dhcpOptionsSet, err := zecService.DescribeDhcpOptionsSetById(ctx, d.Id())
		if err != nil {
			return resource.NonRetryableError(err)
		}
		if dhcpOptionsSet == nil {
			return nil
		}
		return resource.RetryableError(fmt.Errorf("dhcp options set %s still exists", d.Id()))
	}))
}

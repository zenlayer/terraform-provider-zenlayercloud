package zec

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/services/zrm"
	common2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
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
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "1d",
				Description:  "IPv4 lease time, default unit is day: d. Value range: 1d~49d, 3650d~7300d, default value: 1d.",
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^(1d|(?:[2-9]d|(?:[1-3][0-9]|4[0-9])d))$|^(?:(?:36[5-9][0-9]|[37][0-9]{3}|4[0-9]{3}|5[0-9]{3}|6[0-9]{3}|7[0-2][0-9]{2}|7300)d)$`), "leaseTime must be between 1d~49d or 3650d~7300d"),
			},
			"ipv6_lease_time": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "1d",
				Description:  "IPv6 lease time, default unit is day: d. Value range: 1d~49d, 3650d~7300d, default value: 1d.",
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^(1d|(?:[2-9]d|(?:[1-3][0-9]|4[0-9])d))$|^(?:(?:36[5-9][0-9]|[37][0-9]{3}|4[0-9]{3}|5[0-9]{3}|6[0-9]{3}|7[0-2][0-9]{2}|7300)d)$`), "ipv6LeaseTime must be between 1d~49d or 3650d~7300d"),
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the DHCP options set.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the resource group to which the DHCP options set belongs.",
			},
			"tags": {
				Type:        schema.TypeMap,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "Tags of the DHCP options set.",
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
	zecService := &ZecService{client: zenlayerCloudClient}

	request := zec.NewCreateDhcpOptionsSetRequest()
	request.DhcpOptionsSetName = common2.String(d.Get("name").(string))

	domainNameServers := d.Get("domain_name_servers").(string)
	request.DomainNameServers = common2.String(domainNameServers)

	if v, ok := d.GetOk("ipv6_domain_name_servers"); ok {
		request.Ipv6DomainNameServers = common2.String(v.(string))
	}

	if leaseTime, ok := d.GetOk("lease_time"); ok {
		request.LeaseTime = common2.String(leaseTime.(string))
	}

	if ipv6LeaseTime, ok := d.GetOk("ipv6_lease_time"); ok {
		request.Ipv6LeaseTime = common2.String(ipv6LeaseTime.(string))
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

	response, err := zenlayerCloudClient.WithZec2Client().CreateDhcpOptionsSet(request)
	if err != nil {
		return diag.FromErr(fmt.Errorf("fail to create dhcp options set: %v", err))
	}

	d.SetId(*response.Response.DhcpOptionsSetId)

	// Wait for the DHCP options set to be available
	if _, err = zecService.DescribeDhcpOptionsSetById(ctx, d.Id()); err != nil {
		return diag.FromErr(err)
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
	_ = d.Set("lease_time", dhcpOptionsSet.LeaseTime)
	_ = d.Set("ipv6_lease_time", dhcpOptionsSet.Ipv6LeaseTime)
	_ = d.Set("description", dhcpOptionsSet.Description)
	_ = d.Set("create_time", dhcpOptionsSet.CreateTime)
	_ = d.Set("resource_group_id", dhcpOptionsSet.ResourceGroupId)

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
	request.DhcpOptionsSetId = common2.String(d.Id())

	if d.HasChange("domain_name_servers") {
		request.DomainNameServers = common2.String(d.Get("domain_name_servers").(string))
	}

	if d.HasChange("ipv6_domain_name_servers") {
		request.Ipv6DomainNameServers = common2.String(d.Get("ipv6_domain_name_servers").(string))
	}

	if d.HasChange("lease_time") {
		request.LeaseTime = common2.String(d.Get("lease_time").(string))
	}

	if d.HasChange("ipv6_lease_time") {
		request.Ipv6LeaseTime = common2.String(d.Get("ipv6_lease_time").(string))
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
		if err := zrmService.ModifyResourceTags(ctx,d, d.Id()); err != nil {
			return diag.FromErr(fmt.Errorf("fail to update tags for dhcp options set: %v", err))
		}
	}

	// 更新资源组 - 注意：这里可能需要使用专门的API，因为ModifyDhcpOptionsSetAttributesRequest可能不支持ResourceGroupId
	// 目前SDK中可能没有专门的API来更新DHCP选项集的资源组，需要根据实际情况调整

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

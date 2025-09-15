package zec

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	common2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20240401"
)

func ResourceZenlayerCloudEipAssociation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudEipAssociationCreate,
		ReadContext:   resourceZenlayerCloudEipAssociationRead,
		DeleteContext: resourceZenlayerCloudEipAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: customdiff.All(
			nonNicValidFunc(),
		),
		Schema: map[string]*schema.Schema{
			"eip_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the elastic IP.",
			},
			"associated_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the instance to associate with the EIP.",
			},
			"associated_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"NAT", "NIC", "LB"}, false),
				Description:  "Type of the associated instance. Valid values: LB(Load balancer.), NIC(vNic), NAT(NAT gateway).",
			},
			"private_ip_address": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsIPv4Address,
				Description:  "Private IP address of the instance. Required if associated_type is `Nic`.",
			},
			"bind_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				// TODO 支持更新
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"FullNat", "Passthrough"}, false),
				Description:  "Elastic IP bind type. Effective when the elastic IP is assigned to a vNIC.",
			},
		},
	}
}

func nonNicValidFunc() schema.CustomizeDiffFunc {
	return customdiff.IfValue("associated_type", func(ctx context.Context, value, meta interface{}) bool {
		return value != "NIC"
	}, func(ctx context.Context, diff *schema.ResourceDiff, i interface{}) error {
		if _, ok := diff.GetOk("bind_type"); ok {
			return errors.New("`bind_type` is only available for `NIC`")
		}
		if _, ok := diff.GetOk("private_ip_address"); ok {
			return errors.New("`private_ip_address` is only available for `NIC`")
		}
		return nil
	})
}

func resourceZenlayerCloudEipAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_eip_association.create")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	eipId := d.Get("eip_id").(string)
	instanceId := d.Get("associated_id").(string)
	instanceType := d.Get("associated_type").(string)
	// 根据 instanceType 映射到对应的 API 字段
	var loadBalancerId, nicId, natId string
	switch instanceType {
	case "LB":
		loadBalancerId = instanceId
	case "NIC":
		nicId = instanceId
	case "NAT":
		natId = instanceId
	}

	// 如果 instanceType 是 Nic，则 lan_ip 必须提供
	if instanceType == "Nic" && d.Get("private_ip_address").(string) == "" {
		return diag.FromErr(fmt.Errorf("private_ip_address is required when associated_type is Nic"))
	}

	bindType := d.Get("bind_type").(string)

	request := zec.NewAssociateEipAddressRequest()
	request.EipIds = []string{eipId}
	request.LoadBalancerId = &loadBalancerId
	request.NicId = &nicId
	request.LanIp = common2.String(d.Get("private_ip_address").(string))
	request.NatId = &natId
	request.BindType = &bindType

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := zecService.client.WithZecClient().AssociateEipAddress(request)
		if err != nil {
			return common.RetryError(ctx, err)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", eipId, instanceId, instanceType))

	return resourceZenlayerCloudEipAssociationRead(ctx, d, meta)
}

func resourceZenlayerCloudEipAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_eip_association.read")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	association, err := common.ParseResourceId(d.Id(), 3)

	eipId := association[0]

	// 查询 EIP 信息以确认是否已绑定
	eip, err := zecService.DescribeEipById(ctx, eipId)
	if err != nil {
		return diag.FromErr(err)
	}

	if eip == nil {
		d.SetId("")
		return nil
	}

	// 验证绑定的实例 ID
	if eip.AssociatedId == nil {
		d.SetId("")
		return nil
	}

	// 设置属性
	_ = d.Set("eip_id", eip.EipId)
	_ = d.Set("associated_id", eip.AssociatedId)
	_ = d.Set("associated_type", eip.AssociatedType)
	_ = d.Set("private_ip_address", eip.PrivateIpAddress)
	_ = d.Set("bind_type", eip.BindType)

	return nil
}

func resourceZenlayerCloudEipAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_eip_association.delete")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	eipId := d.Get("eip_id").(string)

	request := zec.NewUnassociateEipAddressRequest()
	request.EipIds = []string{eipId}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := zecService.client.WithZecClient().UnassociateEipAddress(request)
		if err != nil {
			return common.RetryError(ctx, err)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

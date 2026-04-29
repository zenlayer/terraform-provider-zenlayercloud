package zec

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	sdkcommon "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"
)

func ResourceZenlayerCloudZecQosPolicyGroupMember() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZecQosPolicyGroupMemberCreate,
		ReadContext:   resourceZenlayerCloudZecQosPolicyGroupMemberRead,
		DeleteContext: resourceZenlayerCloudZecQosPolicyGroupMemberDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"qos_policy_group_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the QoS policy group.",
			},
			"resource_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The resource ID of the member (EIP, IPv6 or UNMANAGED egress IP console UUID).",
			},
			"ip_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"Eip", "Ipv6", "UnmanagedEgressIp"}, false),
				Description:  "The IP type of the member. Valid values: Eip(elastic ip), Ipv6, UnmanagedEgressIp(for unmanaged egress ip).",
			},
		},
	}
}

func resourceZenlayerCloudZecQosPolicyGroupMemberCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_qos_policy_group_member.create")()

	groupId := d.Get("qos_policy_group_id").(string)
	resourceId := d.Get("resource_id").(string)
	ipType := d.Get("ip_type").(string)

	request := zec.NewAddQosPolicyGroupMembersRequest()
	request.QosPolicyGroupId = sdkcommon.String(groupId)
	request.Members = []*zec.QosPolicyGroupMember{
		{
			ResourceId: sdkcommon.String(resourceId),
			IpType:     sdkcommon.String(ipType),
		},
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := meta.(*connectivity.ZenlayerCloudClient).WithZec2Client().AddQosPolicyGroupMembers(request)
		if err != nil {
			return common.RetryError(ctx, err, common.InternalServerError)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s:%s", groupId, resourceId))
	return resourceZenlayerCloudZecQosPolicyGroupMemberRead(ctx, d, meta)
}

func resourceZenlayerCloudZecQosPolicyGroupMemberRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_qos_policy_group_member.read")()

	groupId, resourceId, err := parseQosPolicyGroupMemberId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var found bool
	retryErr := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		group, errRet := zecService.DescribeQosPolicyGroupById(ctx, groupId)
		if errRet != nil {
			return common.RetryError(ctx, errRet)
		}
		if group == nil {
			return nil
		}
		for _, m := range group.Members {
			if sdkcommon.ToString(m.ResourceId) == resourceId {
				_ = d.Set("qos_policy_group_id", groupId)
				_ = d.Set("resource_id", resourceId)
				_ = d.Set("ip_type", sdkcommon.ToString(m.IpType))
				found = true
				break
			}
		}
		return nil
	})
	if retryErr != nil {
		return diag.FromErr(retryErr)
	}

	if !found {
		d.SetId("")
		tflog.Info(ctx, "QoS policy group member does not exist", map[string]interface{}{
			"groupId":    groupId,
			"resourceId": resourceId,
		})
	}

	return nil
}

func resourceZenlayerCloudZecQosPolicyGroupMemberDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_qos_policy_group_member.delete")()

	groupId, resourceId, err := parseQosPolicyGroupMemberId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	request := zec.NewRemoveQosPolicyGroupMembersRequest()
	request.QosPolicyGroupId = sdkcommon.String(groupId)
	request.ResourceIds = []string{resourceId}

	retryErr := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		_, err := meta.(*connectivity.ZenlayerCloudClient).WithZec2Client().RemoveQosPolicyGroupMembers(request)
		if err != nil {
			if sdkErr, ok := err.(*sdkcommon.ZenlayerCloudSdkError); ok {
				if sdkErr.Code == common.ResourceNotFound {
					return nil
				}
			}
			return common.RetryError(ctx, err, common.InternalServerError)
		}
		return nil
	})
	if retryErr != nil {
		return diag.FromErr(retryErr)
	}

	return nil
}

func parseQosPolicyGroupMemberId(id string) (groupId, resourceId string, err error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		err = fmt.Errorf("invalid QoS policy group member ID format: %q, expected <groupId>:<resourceId>", id)
		return
	}
	return parts[0], parts[1], nil
}

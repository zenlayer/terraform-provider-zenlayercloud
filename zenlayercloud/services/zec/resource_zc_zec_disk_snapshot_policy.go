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
	user "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/user20240529"
	zec2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20240401"
	"time"
)

func ResourceZenlayerCloudZecSnapshotPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZecSnapshotPolicyCreate,
		ReadContext:   resourceZenlayerCloudZecSnapshotPolicyRead,
		UpdateContext: resourceZenlayerCloudZecSnapshotPolicyUpdate,
		DeleteContext: resourceZenlayerCloudZecSnapshotPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "Terraform-Snapshot-Policy",
				ValidateFunc: validation.StringLenBetween(2, 63),
				Description:  "The name of the snapshot policy. The name should start and end with a number or a letter, containing 2 to 63 characters. Only letters, numbers, - and periods (.) are supported.",
			},
			"availability_zone": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The availability zone of snapshot policy.",
			},
			"repeat_week_days": {
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
				Description: "The days of week when the auto snapshot policy is triggered. Valid values: `1` to `7`. 1: Monday, 2: Tuesday ~ 7: Sunday.",
			},
			"retention_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.Any(validation.IntBetween(1, 65535), validation.IntInSlice([]int{-1})),
				Description:  "The retention days of the auto snapshot policy. Valid values: `1` to `65535` or `-1` for no expired.",
			},
			"hours": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
				ValidateFunc: func(i interface{}, s string) ([]string, []error) {
					hours := i.(*schema.Set).List()
					for _, v := range hours {
						if h, ok := v.(int); ok {
							if h < 0 || h > 23 {
								return nil, []error{fmt.Errorf("hour value %d is not in range 0-23", h)}
							}
						} else {
							return nil, []error{fmt.Errorf("hour value %v is not an integer", v)}
						}
					}
					return nil, nil
				},
				Description: "The hours of day when the auto snapshot policy is triggered. The time zone of hour is `UTC+0`. Valid values: from `0` to `23`.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation time of the snapshot policy.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The ID of resource group grouped snapshot policy.",
			},
			"resource_group_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Name of resource group grouped snapshot policy.",
			},
		},
	}
}

func resourceZenlayerCloudZecSnapshotPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	policyId := d.Id()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := zecService.DeleteSnapshotPolicy(ctx, policyId)
		if errRet != nil {
			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if !ok {
				return common2.RetryError(ctx, errRet)
			} else if ee.Code == "INVALID_AUTO_SNAPSHOT_POLICY_NOT_FOUND" || ee.Code == common2.ResourceNotFound {
				return nil
			}
			return common2.RetryError(ctx, errRet)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceZenlayerCloudZecSnapshotPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	snapId := d.Id()

	if d.HasChanges("name", "repeat_week_days", "retention_days", "hours") {

		request := zec2.NewModifyAutoSnapshotPolicyRequest()
		request.AutoSnapshotPolicyName = common.String(d.Get("name").(string))
		request.RetentionDays = common.Integer(d.Get("retention_days").(int))
		request.AutoSnapshotPolicyId = common.String(snapId)
		if d.HasChange("repeat_week_days") {
			if v, ok := d.GetOk("repeat_week_days"); ok {
				ids := v.(*schema.Set).List()
				if len(ids) > 0 {
					request.RepeatWeekDays = common2.ToIntList(ids)
				}
			}
		}

		if d.HasChange("hours") {
			if v, ok := d.GetOk("hours"); ok {
				ids := v.(*schema.Set).List()
				if len(ids) > 0 {
					request.Hours = common2.ToIntList(ids)
				}
			}
		}

		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {

			_, err := zecService.client.WithZecClient().ModifyAutoSnapshotPolicy(request)

			if err != nil {
				return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("resource_group_id") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			request := user.NewAddResourceResourceGroupRequest()
			request.ResourceGroupId = common.String(d.Get("resource_group_id").(string))
			request.Resources = []*string{common.String(snapId)}

			_, err := meta.(*connectivity.ZenlayerCloudClient).WithUsrClient().AddResourceResourceGroup(request)
			if err != nil {
				return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}


	return resourceZenlayerCloudZecSnapshotPolicyRead(ctx, d, meta)
}

func resourceZenlayerCloudZecSnapshotPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	request := zec2.NewCreateAutoSnapshotPolicyRequest()
	request.AutoSnapshotPolicyName = common.String(d.Get("name").(string))
	request.ZoneId = common.String(d.Get("availability_zone").(string))
	request.RetentionDays = common.Integer(d.Get("retention_days").(int))

	if v, ok := d.GetOk("resource_group_id"); ok {
		request.ResourceGroupId = common.String(v.(string))
	}

	// Handle repeat_week_days
	weekDays := d.Get("repeat_week_days").(*schema.Set).List()
	weekDaysInt := make([]int, 0)
	for _, day := range weekDays {
		weekDaysInt = append(weekDaysInt, day.(int))
	}
	request.RepeatWeekDays = weekDaysInt

	// Handle hours
	hours := d.Get("hours").(*schema.Set).List()
	hoursInt := make([]int, 0)
	for _, hour := range hours {
		hourInt := hour.(int)
		hoursInt = append(hoursInt, hourInt)
	}
	request.Hours = hoursInt

	var response *zec2.CreateAutoSnapshotPolicyResponse
	var err error

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate)-time.Minute, func() *resource.RetryError {
		response, err = zecService.client.WithZecClient().CreateAutoSnapshotPolicy(request)
		if err != nil {
			return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if response == nil || response.Response == nil || response.Response.AutoSnapshotPolicyId == nil {
		return diag.Errorf("failed to create snapshot policy: response is invalid")
	}

	d.SetId(*response.Response.AutoSnapshotPolicyId)

	return resourceZenlayerCloudZecSnapshotPolicyRead(ctx, d, meta)
}


func resourceZenlayerCloudZecSnapshotPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	snapshotPolicyId := d.Id()

	vmService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var snapshot *zec2.AutoSnapshotPolicy
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		snapshot, errRet = vmService.DescribeSnapshotPolicyById(ctx, snapshotPolicyId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if snapshot == nil{
		d.SetId("")
		tflog.Info(ctx, "snapshot policy not exist", map[string]interface{}{
			"snapshotPolicyId": snapshotPolicyId,
		})
		return nil
	}

	// snapshot policy
	_ = d.Set("name", snapshot.AutoSnapshotPolicyName)
	_ = d.Set("availability_zone", snapshot.ZoneId)
	_ = d.Set("retention_days", snapshot.RetentionDays)
	_ = d.Set("create_time", snapshot.CreateTime)
	_ = d.Set("repeat_week_days",  snapshot.RepeatWeekDays )
	_ = d.Set("hours", snapshot.Hours)
	_ = d.Set("resource_group_id", snapshot.ResourceGroup.ResourceGroupId)
	_ = d.Set("resource_group_name", snapshot.ResourceGroup.ResourceGroupName)

	return diags

}

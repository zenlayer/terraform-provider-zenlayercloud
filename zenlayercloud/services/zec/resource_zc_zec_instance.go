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
	vm "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/vm20230313"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20240401"
	"time"
)

func ResourceZenlayerCloudZecInstance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZecInstanceCreate,
		ReadContext:   resourceZenlayerCloudZecInstanceRead,
		UpdateContext: resourceZenlayerCloudZecInstanceUpdate,
		DeleteContext: resourceZenlayerCloudZecInstanceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(common2.VmCreateTimeout),
			Update: schema.DefaultTimeout(common2.VmUpdateTimeout),
		},
		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of zone that the ZEC instance locates at. such as `asia-southeast-1a`.",
			},
			"instance_type": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The type of the ZEC instance. such as `z2a.cpu.4`.",
			},
			"cpu": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of CPU cores of the ZEC instance.",
			},
			"memory": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Memory capacity of the ZEC instance, unit in GiB.",
			},
			"image_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The image to use for the ZEC instance. Changing `image_id` will cause the ZEC instance reset.",
			},
			"image_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The image name to use for the ZEC instance.",
			},
			"instance_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "Terraform-Instance",
				ValidateFunc: validation.StringLenBetween(2, 63),
				Description:  "The name of the ZEC instance. The minimum length of instance name is `2`. The max length of instance_name is 63, and default value is `Terraform-ZEC-Instance`.",
			},
			"password": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				AtLeastOneOf: []string{"password", "key_id"},
				ValidateFunc: validation.All(validation.StringLenBetween(8, 16)),
				Description:  "Password for the ZEC instance. The max length of password is 16.",
			},
			"key_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "The key pair id to use for the ZEC instance. Changing `key_id` will cause the ZEC instance reset.",
				ConflictsWith: []string{"password"},
				AtLeastOneOf:  []string{"password", "key_id"},
			},
			//"internet_charge_type": {
			//	Type:         schema.TypeString,
			//	Required:     true,
			//	ForceNew:     true,
			//	ValidateFunc: validation.StringInSlice(VmInternetChargeTypes, false),
			//	Description:  "Internet charge type of the ZEC instance, Valid values are `ByBandwidth`, `ByTrafficPackage`, `ByInstanceBandwidth95` and `ByClusterBandwidth95`. This value currently not support to change.",
			//},
			//"internet_max_bandwidth_out": {
			//	Type:         schema.TypeInt,
			//	Optional:     true,
			//	Computed:     true,
			//	ValidateFunc: validation.IntAtLeast(1),
			//	Description:  "Maximum outgoing bandwidth to the public network, measured in Mbps (Mega bits per second).",
			//},
			//"traffic_package_size": {
			//	Type:        schema.TypeFloat,
			//	Optional:    true,
			//	Description: "Traffic package size. Only valid when the charge type of instance is `ByTrafficPackage` and the ZEC instance charge type is `PREPAID`.",
			//},
			"subnet_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of a VPC subnet.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The resource group id the ZEC instance belongs to, default to Default Resource Group.",
			},
			"resource_group_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The resource group name the ZEC instance belongs to, default to Default Resource Group.",
			},
			"system_disk_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the system disk.",
			},
			"system_disk_category": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Category of the system disk.",
			},
			"system_disk_size": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "Size of the system disk. unit is GB. If modified, the ZEC instance may force stop.",
			},
			"time_zone": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Time zone of instance. such as `America/Los_Angeles`. Changing `time_zone` will cause the ZEC instance reset.",
			},
			"disable_qga_agent": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Indicate whether to disable QEMU Guest Agent (QGA). QGA is enabled by default. Changing `disable_qga_agent` will cause the ZEC instance reset.",
			},
			"enable_ip_forwarding": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Indicate whether to enable IP forwarding. IP forwarding is disabled by default.",
			},
			"force_delete": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Indicate whether to force delete the ZEC instance. Default is `true`. If set true, the ZEC instance will be permanently deleted instead of being moved into the recycle bin.",
			},
			"public_ip_addresses": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Public Ip addresses of the ZEC instance.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"private_ip_addresses": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Private Ip addresses of the ZEC instance.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"running_flag": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Set instance to running or stop. Default value is true, the instance will shutdown when this flag is false.",
			},
			"instance_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Current status of the ZEC instance.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Create time of the ZEC instance.",
			},
		},
	}
}

func resourceZenlayerCloudZecInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceId := d.Id()

	// force_delete: terminate and then delete
	forceDelete := d.Get("force_delete").(bool)

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := zecService.DeleteInstance(ctx, instanceId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	notExist := false

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		instance, errRet := zecService.DescribeInstanceById(ctx, instanceId)
		if errRet != nil {
			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if !ok {
				return common2.RetryError(ctx, errRet)
			} else if ee.Code == "INVALID_INSTANCE_NOT_FOUND" || ee.Code == common2.ResourceNotFound {
				notExist = true
				return nil
			}
			return common2.RetryError(ctx, errRet)
		}

		if instance == nil {
			notExist = true
			return nil
		}

		if instance.Status == ZecInstanceStatusRecycle {
			//in recycle bin
			return nil
		}

		if instanceIsOperating(instance.Status) {
			return resource.RetryableError(fmt.Errorf("waiting for instance %s recycling, current status: %s", instance.InstanceId, instance.Status))
		}

		return resource.NonRetryableError(fmt.Errorf("vm instance status is not recycle, current status %s", instance.Status))
	})

	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Debug(ctx, "Delete Instance ...", map[string]interface{}{
		"notExist":    notExist,
		"forceDelete": forceDelete,
	})

	if notExist || !forceDelete {
		return nil
	}

	tflog.Debug(ctx, "Releasing Instance ...", map[string]interface{}{
		"instanceId": instanceId,
	})

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := zecService.DeleteInstance(ctx, instanceId)
		if errRet != nil {

			//check InvalidInstanceState.Terminating
			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if !ok {
				return common2.RetryError(ctx, errRet)
			}
			if ee.Code == "INVALID_INSTANCE_NOT_FOUND" {
				// instance doesn't exist
				return nil
			}
			return common2.RetryError(ctx, errRet, common2.InternalServerError)
		}
		return nil
	})

	return nil
}

func resourceZenlayerCloudZecInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	instanceId := d.Id()
	d.Partial(true)

	if d.HasChange("instance_name") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			request := zec.NewModifyInstancesAttributeRequest()
			request.InstanceIds = []string{instanceId}
			request.InstanceName = d.Get("instance_name").(string)

			response, err := zecService.client.WithZecClient().ModifyInstancesAttribute(request)
			defer common2.LogApiRequest(ctx, "ModifyInstancesAttribute", request, response, err)

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
			request.Resources = []*string{common.String(instanceId)}

			_, err := zecService.client.WithUsrClient().AddResourceResourceGroup(request)
			if err != nil {
				return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	//if d.HasChange("internet_max_bandwidth_out") {
	//	if v, ok := d.GetOk("internet_max_bandwidth_out"); ok {
	//
	//		err := zecService.updateInstanceInternetMaxBandwidthOut(ctx, instanceId, v.(int))
	//
	//		if err != nil {
	//			return diag.FromErr(err)
	//		}
	//		err = waitVmNetworkStatusOK(ctx, zecService, d, instanceId, &vmInternetBandwidthOutCondition{
	//			InternetBandwidthOut: v.(int),
	//		})
	//		if err != nil {
	//			return diag.FromErr(err)
	//		}
	//	}
	//}

	//if d.HasChange("traffic_package_size") {
	//	if v, ok := d.GetOk("traffic_package_size"); ok {
	//		err := zecService.updateInstanceTrafficPackageSize(ctx, instanceId, v.(float64))
	//		if err != nil {
	//			return diag.FromErr(err)
	//		}
	//		err = waitVmNetworkStatusOK(ctx, zecService, d, instanceId, &vmTrafficPackageCondition{
	//			TargetPackageSize: v.(float64),
	//		})
	//		if err != nil {
	//			return diag.FromErr(err)
	//		}
	//	}
	//}

	if d.HasChange("enable_ip_forwarding") {
		_, newValue := d.GetChange("enable_ip_forwarding")

		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			err := zecService.switchInstanceIpForwarding(ctx, instanceId, newValue.(bool))
			if err != nil {
				return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})

		if err != nil {
			return diag.FromErr(err)
		}

	}

	if d.HasChanges("image_id", "key_id", "time_zone", "disable_qga_agent") {
		err := zecService.shutdownInstance(ctx, instanceId)
		if err != nil {
			return nil
		}

		err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			instance, errRet := zecService.DescribeInstanceById(ctx, instanceId)
			if errRet != nil {
				return common2.RetryError(ctx, errRet, common2.InternalServerError)
			}

			if instance.Status == ZecInstanceStatusStopped {
				return nil
			}

			if instanceIsOperating(instance.Status) {
				return resource.RetryableError(fmt.Errorf("waiting for instance %s stopping, current status: %s", instance.InstanceId, instance.Status))
			}

			return resource.NonRetryableError(fmt.Errorf("vm instance status is not stopped, current status %s", instance.Status))
		})

		request := zec.NewResetInstanceRequest()
		request.InstanceId = d.Id()
		if v, ok := d.GetOk("image_id"); ok {
			request.ImageId = v.(string)
		}

		if v, ok := d.GetOk("key_id"); ok {
			request.KeyId = v.(string)
		}

		if v, ok := d.GetOk("time_zone"); ok {
			request.Timezone = v.(string)
		}

		if v, ok := d.GetOk("disable_qga_agent"); ok {
			request.EnableAgent = !v.(bool)
		} else {
			request.EnableAgent = true
		}

		err = zecService.resetInstance(ctx, request)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("password") {
		_, newValue := d.GetChange("password")
		if newValue != "" {
			err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
				err := zecService.resetInstancePassword(ctx, instanceId, d.Get("password").(string))
				if err != nil {
					return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
				}
				return nil
			})
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("running_flag") {
		running := d.Get("running_flag").(bool)
		if running {
			err := zecService.StartInstance(ctx, instanceId)
			if err != nil {
				return diag.FromErr(err)
			}

			err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
				instance, errRet := zecService.DescribeInstanceById(ctx, instanceId)
				if errRet != nil {
					return common2.RetryError(ctx, errRet, common2.InternalServerError)
				}
				if instance.Status == ZecInstanceStatusRunning {
					return nil
				}
				return resource.RetryableError(fmt.Errorf("zec instance status is %s, retry...", instance.Status))

			})
			if err != nil {
				return diag.FromErr(err)
			}
		} else {

			err := zecService.shutdownInstance(ctx, instanceId)
			if err != nil {
				return diag.FromErr(err)
			}

			err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
				instance, errRet := zecService.DescribeInstanceById(ctx, instanceId)
				if errRet != nil {
					return common2.RetryError(ctx, errRet, common2.InternalServerError)
				}
				if instance.Status == ZecInstanceStatusStopped {
					return nil
				}
				return resource.RetryableError(fmt.Errorf("zec instance status is %s, retry...", instance.Status))

			})
			if err != nil {
				return diag.FromErr(err)
			}

		}
	}

	d.Partial(false)

	return resourceZenlayerCloudZecInstanceRead(ctx, d, meta)
}

type vmTrafficPackageCondition struct {
	TargetPackageSize float64
}

func (t *vmTrafficPackageCondition) matchFail(status *vm.DescribeInstanceInternetStatusResponseParams) bool {
	return status.ModifiedTrafficPackageStatus == "Enable" && t.TargetPackageSize != *status.TrafficPackageSize
}

func (t *vmTrafficPackageCondition) matchOk(status *vm.DescribeInstanceInternetStatusResponseParams) bool {
	return status.ModifiedTrafficPackageStatus == "Enable" && t.TargetPackageSize == *status.TrafficPackageSize
}

type vmInternetBandwidthOutCondition struct {
	InternetBandwidthOut int
}

func (t *vmInternetBandwidthOutCondition) matchFail(status *vm.DescribeInstanceInternetStatusResponseParams) bool {
	return status.ModifiedBandwidthStatus == "Enable" && t.InternetBandwidthOut != *status.InternetMaxBandwidthOut
}

func (t *vmInternetBandwidthOutCondition) matchOk(status *vm.DescribeInstanceInternetStatusResponseParams) bool {
	return status.ModifiedBandwidthStatus == "Enable" && t.InternetBandwidthOut == *status.InternetMaxBandwidthOut
}

func resourceZenlayerCloudZecInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	request := zec.NewCreateZecInstancesRequest()
	request.InstanceCount = 1
	request.ZoneId = d.Get("availability_zone").(string)

	request.InstanceType = d.Get("instance_type").(string)
	request.ImageId = d.Get("image_id").(string)
	request.InstanceName = d.Get("instance_name").(string)

	if v, ok := d.GetOk("password"); ok {
		request.Password = v.(string)
	}
	if v, ok := d.GetOk("key_id"); ok {
		request.KeyId = v.(string)
	}
	system := &zec.SystemDisk{}
	system.DiskSize = d.Get("system_disk_size").(int)

	if v, ok := d.GetOk("system_disk_category"); ok {
		system.DiskCategory = v.(string)
	}
	request.SystemDisk = system

	//request.InternetChargeType = d.Get("internet_charge_type").(string)
	//if request.InternetChargeType == VmInternetChargeTypeTrafficPackage && request.InstanceChargeType == VmChargeTypePrepaid {
	//	if v, ok := d.GetOk("traffic_package_size"); ok {
	//		request.TrafficPackageSize = common.Float64(v.(float64))
	//	} else {
	//		diags = append(diags, diag.Diagnostic{
	//			Summary: "Missing required argument",
	//			Detail:  "traffic_package_size is missing with `ByTrafficPackage` instance.",
	//		})
	//		return diags
	//	}
	//}
	//if v, ok := d.GetOk("internet_max_bandwidth_out"); ok {
	//	request.InternetMaxBandwidthOut = v.(int)
	//}
	request.SubnetId = d.Get("subnet_id").(string)

	if v, ok := d.GetOk("resource_group_id"); ok {
		request.ResourceGroupId = v.(string)
	}

	if v, ok := d.GetOk("time_zone"); ok {
		request.TimeZone = v.(string)
	}

	if v, ok := d.GetOk("disable_qga_agent"); ok {
		request.EnableAgent = !v.(bool)
	} else {
		request.EnableAgent = true
	}

	if v, ok := d.GetOk("enable_ip_forwarding"); ok {
		request.EnableIpForward = v.(bool)
	}

	instanceId := ""

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := meta.(*connectivity.ZenlayerCloudClient).WithZecClient().CreateZecInstances(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create zec instance.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": common2.ToJsonString(request),
				"err":     err.Error(),
			})
			return common2.RetryError(ctx, err)
		}

		tflog.Info(ctx, "Create zec instance success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  common2.ToJsonString(request),
			"response": common2.ToJsonString(response),
		})

		if len(response.Response.InstanceIdSet) < 1 {
			err = fmt.Errorf("instance id is nil")
			return resource.NonRetryableError(err)
		}
		instanceId = *response.Response.InstanceIdSet[0]

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(instanceId)

	stateConf := &resource.StateChangeConf{
		Pending: []string{
			ZecInstanceStatusPending,
			ZecInstanceStatusDeloying,
		},
		Target: []string{
			ZecInstanceStatusRunning,
		},
		Refresh:        zecService.InstanceStateRefreshFunc(ctx, instanceId, []string{ZecInstanceStatusCreateFailed}),
		Timeout:        d.Timeout(schema.TimeoutCreate) - time.Minute,
		Delay:          10 * time.Second,
		MinTimeout:     5 * time.Second,
		NotFoundChecks: 3,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for zec instance (%s) to be created: %v", d.Id(), err))
	}

	return resourceZenlayerCloudZecInstanceRead(ctx, d, meta)
}

func resourceZenlayerCloudZecInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	instanceId := d.Id()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var instance *zec.InstanceInfo
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		instance, errRet = zecService.DescribeInstanceById(ctx, instanceId)
		if errRet != nil {
			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if !ok {
				return common2.RetryError(ctx, errRet)
			} else if ee.Code == "INVALID_INSTANCE_NOT_FOUND" || ee.Code == common2.ResourceNotFound {
				return nil
			}
			return common2.RetryError(ctx, errRet)
		}
		if instance != nil && instanceIsOperating(instance.Status) {
			return resource.RetryableError(fmt.Errorf("waiting for zec instance %s operation, current status: %s", instance.InstanceId, instance.Status))
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if instance == nil || instance.Status == ZecInstanceStatusCreateFailed ||
		instance.Status == ZecInstanceStatusRecycle {
		d.SetId("")
		tflog.Info(ctx, "instance not exist or created failed or recycled", map[string]interface{}{
			"instanceId": instanceId,
		})
		return nil
	}

	// instance info
	_ = d.Set("availability_zone", instance.ZoneId)
	//_ = d.Set("instance_charge_type", instance.InstanceChargeType)
	//if instance.InstanceChargeType == VmChargeTypePrepaid {
	//	_ = d.Set("instance_charge_prepaid_period", instance.Period)
	//}
	//_ = d.Set("instance_type", instance.InstanceType)
	_ = d.Set("key_id", instance.KeyId)
	_ = d.Set("image_id", instance.ImageId)
	_ = d.Set("image_name", instance.ImageName)
	_ = d.Set("instance_name", instance.InstanceName)
	//_ = d.Set("internet_charge_type", instance.InternetChargeType)
	//_ = d.Set("internet_max_bandwidth_out", instance.InternetMaxBandwidthOut)
	//if instance.InternetChargeType == VmInternetChargeTypeTrafficPackage && instance.InstanceChargeType == VmChargeTypePrepaid {
	//	_ = d.Set("traffic_package_size", *instance.TrafficPackageSize)
	//}
	_ = d.Set("subnet_id", instance.SubnetId)
	_ = d.Set("resource_group_id", instance.ResourceGroupId)
	_ = d.Set("resource_group_name", instance.ResourceGroupName)
	_ = d.Set("public_ip_addresses", instance.PublicIpAddresses)
	_ = d.Set("private_ip_addresses", instance.PrivateIpAddresses)
	_ = d.Set("instance_status", instance.Status)
	if instance.SystemDisk != nil {
		_ = d.Set("system_disk_id", instance.SystemDisk.DiskId)
		_ = d.Set("system_disk_size", instance.SystemDisk.DiskSize)
		_ = d.Set("system_disk_category", instance.SystemDisk.DiskCategory)
	}
	_ = d.Set("create_time", instance.CreateTime)
	_ = d.Set("time_zone", instance.TimeZone)
	_ = d.Set("running_flag", instance.Status == ZecInstanceStatusRunning)

	_ = d.Set("disable_qga_agent", !*instance.EnableAgent)
	_ = d.Set("enable_ip_forwarding", *instance.EnableIpForward)

	return diags

}

func instanceIsOperating(status string) bool {
	return common2.IsContains(InstanceOperatingStatus, status)
}

/*
Provides a instance resource.

~> **NOTE:** You can launch an instance for a private network via specifying parameter `subnet_id`.

~> **NOTE:** At present, 'PREPAID' instance cannot be deleted and must wait it to be outdated and released automatically.

Example Usage

```hcl

data "zenlayercloud_zones" "default" {

}

data "zenlayercloud_instance_types" "default" {
  availability_zone = data.zenlayercloud_zones.default.zones.0.id
}

# Get a centos image which also supported to install on given instance type
data "zenlayercloud_images" "default" {
  availability_zone = data.zenlayercloud_zones.default.zones.0.id
  category          = "CentOS"
}

resource "zenlayercloud_subnet" "default" {
  name              = "test-subnet"
  cidr_block        = "10.0.10.0/24"
}

# Create a web server
resource "zenlayercloud_instance" "web" {
  availability_zone    = data.zenlayercloud_zones.default.zones.0.id
  image_id             = data.zenlayercloud_images.default.images.0.image_id
  internet_charge_type = "ByBandwidth"
  instance_type        = data.zenlayercloud_instance_types.default.instance_types.0.id
  password             = "Example~123"
  instance_name        = "web"
  subnet_id            = zenlayercloud_subnet.default.id
  system_disk_size     = 100
}
```

Import

Instance can be imported using the id, e.g.

```
terraform import zenlayercloud_instance.foo 123123xxx
```
*/
package zenlayercloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	vm "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/vm20230313"
	"time"
)

func resourceZenlayerCloudVmInstance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudVmInstanceCreate,
		ReadContext:   resourceZenlayerCloudVmInstanceRead,
		UpdateContext: resourceZenlayerCloudVmInstanceUpdate,
		DeleteContext: resourceZenlayerCloudVmInstanceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(vmCreateTimeout),
			Update: schema.DefaultTimeout(vmUpdateTimeout),
		},
		CustomizeDiff: customdiff.All(
			vmInternetMaxBandwidthOutForceNew(),
			vmTrafficPackageSizeForceNew(),
			vmTrafficPackageSizeValidFunc(),
			vmTrafficPackageSizeForPostPaidFunc(),
		),
		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of zone that the instance locates at.",
			},
			"instance_charge_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "POSTPAID",
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(VmChargeTypes, false),
				Description:  "The charge type of instance. Valid values are `PREPAID`, `POSTPAID`. The default is `POSTPAID`. Note: `PREPAID` instance may not allow to delete before expired.",
			},
			"instance_charge_prepaid_period": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1),
				Description:  "The tenancy (time unit is month) of the prepaid instance, NOTE: it only works when instance_charge_type is set to `PREPAID`.",
			},
			"instance_type": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The type of the instance.",
			},

			"image_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The image to use for the instance. Changing `image_id` will cause the instance reset.",
			},
			"image_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The image name to use for the instance.",
			},
			"instance_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "Terraform-Instance",
				ValidateFunc: validation.StringLenBetween(2, 64),
				Description:  "The name of the instance. The max length of instance_name is 64, and default value is `Terraform-Instance`.",
			},
			"password": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.All(validation.StringLenBetween(8, 16)),
				Description:  "Password for the instance. The max length of password is 16.",
			},
			"key_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "The key pair id to use for the instance. Changing `key_id` will cause the instance reset.",
				ConflictsWith: []string{"password"},
			},
			"internet_charge_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(VmInternetChargeTypes, false),
				Description:  "Internet charge type of the instance, Valid values are `ByBandwidth`, `ByTrafficPackage`, `ByInstanceBandwidth95` and `ByClusterBandwidth95`. This value currently not support to change.",
			},
			"internet_max_bandwidth_out": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(1),
				Description:  "Maximum outgoing bandwidth to the public network, measured in Mbps (Mega bits per second).",
			},
			"traffic_package_size": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Description: "Traffic package size. Only valid when the charge type of instance is `ByTrafficPackage` and the instance charge type is `PREPAID`.",
			},
			"subnet_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The ID of a VPC subnet. If you want to create instances in a VPC network, this parameter must be set.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The resource group id the instance belongs to, default to Default Resource Group.",
			},
			"resource_group_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The resource group name the instance belongs to, default to Default Resource Group.",
			},
			"system_disk_size": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "Size of the system disk. unit is GB. If modified, the instance may force stop.",
			},
			"force_delete": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Indicate whether to force delete the instance. Default is `false`. If set true, the instance will be permanently deleted instead of being moved into the recycle bin.",
			},
			"public_ip_addresses": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Public Ip addresses of the instance.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"private_ip_addresses": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Private Ip addresses of the instance.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"instance_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Current status of the instance.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Create time of the instance.",
			},
			"expired_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Expired time of the instance.",
			},
		},
	}
}

func vmTrafficPackageSizeForceNew() schema.CustomizeDiffFunc {
	return customdiff.ForceNewIf("traffic_package_size", forceNewIfVmTrafficPackageSizeDowngradeForPrepaidInstance)
}
func forceNewIfVmTrafficPackageSizeDowngradeForPrepaidInstance(_ context.Context, d *schema.ResourceDiff, _ interface{}) bool {
	if change := d.HasChange("traffic_package_size"); !change {
		return false
	}

	internetChargeType := d.Get("internet_charge_type").(string)
	if internetChargeType != VmInternetChargeTypeTrafficPackage {
		return false
	}

	chargeType := d.Get("instance_charge_type").(string)
	if chargeType == VmChargeTypePostpaid {
		return false
	}
	oldValue, newValue := d.GetChange("traffic_package_size")
	return oldValue.(float64) > newValue.(float64)
}
func vmInternetMaxBandwidthOutForceNew() schema.CustomizeDiffFunc {
	return customdiff.ForceNewIf("internet_max_bandwidth_out", forceNewIfVmBandwidthDowngradeForPrepaidInstance)
}

func vmTrafficPackageSizeValidFunc() schema.CustomizeDiffFunc {
	return customdiff.IfValue("internet_charge_type", func(ctx context.Context, value, meta interface{}) bool {
		return value != VmInternetChargeTypeTrafficPackage
	}, func(ctx context.Context, diff *schema.ResourceDiff, i interface{}) error {
		if _, ok := diff.GetOk("traffic_package_size"); ok {
			return fmt.Errorf("traffic_package_size can't be set as the internet charge type of instance is not `%s`", VmInternetChargeTypeTrafficPackage)
		}
		return nil
	})
}

func vmTrafficPackageSizeForPostPaidFunc() schema.CustomizeDiffFunc {

	return customdiff.If(func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) bool {
		internetType := d.Get("internet_charge_type")
		chargeType := d.Get("instance_charge_type")
		return internetType == VmInternetChargeTypeTrafficPackage && chargeType == VmChargeTypePostpaid

	}, func(ctx context.Context, diff *schema.ResourceDiff, i interface{}) error {
		if _, ok := diff.GetOk("traffic_package_size"); ok {
			return fmt.Errorf("traffic_package_size can't be set for post paid instance with internet type `%s`", VmInternetChargeTypeTrafficPackage)
		}
		return nil
	})
}

func forceNewIfVmBandwidthDowngradeForPrepaidInstance(_ context.Context, d *schema.ResourceDiff, _ interface{}) bool {
	if change := d.HasChange("internet_max_bandwidth_out"); !change {
		return false
	}

	internetChargeType := d.Get("internet_charge_type").(string)
	if internetChargeType != VmInternetChargeTypeBandwidth {
		return false
	}

	chargeType := d.Get("instance_charge_type").(string)
	if chargeType == VmChargeTypePostpaid {
		return false
	}
	oldValue, newValue := d.GetChange("internet_max_bandwidth_out")
	return oldValue.(int) > newValue.(int)
}

func resourceZenlayerCloudVmInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceId := d.Id()

	// force_delete: terminate and then delete
	forceDelete := d.Get("force_delete").(bool)

	vmService := VmService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := vmService.DeleteInstance(ctx, instanceId)
		if errRet != nil {
			return retryError(ctx, errRet)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	notExist := false

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		instance, errRet := vmService.DescribeInstanceById(ctx, instanceId)
		if errRet != nil {
			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if !ok {
				return retryError(ctx, errRet)
			} else if ee.Code == "INVALID_INSTANCE_NOT_FOUND" || ee.Code == ResourceNotFound {
				notExist = true
				return nil
			}
			return retryError(ctx, errRet)
		}

		if instance == nil {
			notExist = true
			return nil
		}

		if instance.InstanceStatus == VmInstanceStatusRecycle {
			//in recycling
			return nil
		}

		if vmInstanceIsOperating(instance.InstanceStatus) {
			return resource.RetryableError(fmt.Errorf("waiting for instance %s recycling, current status: %s", instance.InstanceId, instance.InstanceStatus))
		}

		return resource.NonRetryableError(fmt.Errorf("vm instance status is not recycle, current status %s", instance.InstanceStatus))
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if notExist || !forceDelete {
		return nil
	}
	tflog.Debug(ctx, "Releasing Instance ...", map[string]interface{}{
		"instanceId": instanceId,
	})

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := vmService.DestroyInstance(ctx, instanceId)
		if errRet != nil {

			//check InvalidInstanceState.Terminating
			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if !ok {
				return retryError(ctx, errRet)
			}
			if ee.Code == "INVALID_INSTANCE_NOT_FOUND" {
				// instance doesn't exist
				return nil
			}
			return retryError(ctx, errRet, InternalServerError)
		}
		return nil
	})

	return nil
}

func resourceZenlayerCloudVmInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vmService := VmService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	instanceId := d.Id()
	d.Partial(true)

	if d.HasChange("instance_name") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			err := vmService.ModifyInstanceName(ctx, instanceId, d.Get("instance_name").(string))
			if err != nil {
				return retryError(ctx, err, InternalServerError, common.NetworkError)
			}
			return nil
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("resource_group_id") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			err := vmService.ModifyInstanceResourceGroup(ctx, instanceId, d.Get("resource_group_id").(string))
			if err != nil {
				return retryError(ctx, err, InternalServerError, common.NetworkError)
			}
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("internet_max_bandwidth_out") {
		if v, ok := d.GetOk("internet_max_bandwidth_out"); ok {

			err := vmService.updateInstanceInternetMaxBandwidthOut(ctx, instanceId, v.(int))

			if err != nil {
				return diag.FromErr(err)
			}
			err = waitVmNetworkStatusOK(ctx, vmService, d, instanceId, &vmInternetBandwidthOutCondition{
				InternetBandwidthOut: v.(int),
			})
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("traffic_package_size") {
		if v, ok := d.GetOk("traffic_package_size"); ok {
			err := vmService.updateInstanceTrafficPackageSize(ctx, instanceId, v.(float64))
			if err != nil {
				return diag.FromErr(err)
			}
			err = waitVmNetworkStatusOK(ctx, vmService, d, instanceId, &vmTrafficPackageCondition{
				TargetPackageSize: v.(float64),
			})
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("password") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			err := vmService.resetInstancePassword(ctx, instanceId, d.Get("password").(string))
			if err != nil {
				return retryError(ctx, err, InternalServerError, common.NetworkError)
			}
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChanges("image_id", "key_id") {
		err := vmService.shutdownInstance(ctx, instanceId)
		if err != nil {
			return nil
		}

		err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			instance, errRet := vmService.DescribeInstanceById(ctx, instanceId)
			if errRet != nil {
				return retryError(ctx, errRet, InternalServerError)
			}

			if instance.InstanceStatus == VmInstanceStatusStopped {
				return nil
			}

			if vmInstanceIsOperating(instance.InstanceStatus) {
				return resource.RetryableError(fmt.Errorf("waiting for instance %s stopping, current status: %s", instance.InstanceId, instance.InstanceStatus))
			}

			return resource.NonRetryableError(fmt.Errorf("vm instance status is not stopped, current status %s", instance.InstanceStatus))
		})

		request := vm.NewResetInstanceRequest()
		request.InstanceId = d.Id()
		if v, ok := d.GetOk("image_id"); ok {
			request.ImageId = v.(string)
		}

		if v, ok := d.GetOk("key_id"); ok {
			request.KeyId = v.(string)
		}

		err = vmService.resetInstance(ctx, request)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	d.Partial(false)

	return resourceZenlayerCloudVmInstanceRead(ctx, d, meta)
}

func waitVmNetworkStatusOK(ctx context.Context, vmService VmService, d *schema.ResourceData, instanceId string, condition VmNetworkStateCondition) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			VmNetworkStatusPending,
		},
		Target: []string{
			VmNetworkStatusOK,
		},
		Refresh:        vmService.InstanceNetworkStateRefreshFunc(ctx, instanceId, condition),
		Timeout:        d.Timeout(schema.TimeoutCreate) - time.Minute,
		Delay:          10 * time.Second,
		MinTimeout:     5 * time.Second,
		NotFoundChecks: 3,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("error waiting for vm instance (%s) internet changed: %v", d.Id(), err)
	}
	return nil
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

func resourceZenlayerCloudVmInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	vmService := VmService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	request := vm.NewCreateInstancesRequest()
	request.InstanceCount = 1
	request.ZoneId = d.Get("availability_zone").(string)
	request.InstanceChargeType = d.Get("instance_charge_type").(string)

	if request.InstanceChargeType == VmChargeTypePrepaid {
		request.InstanceChargePrepaid = &vm.ChargePrepaid{}

		if period, ok := d.GetOk("instance_charge_prepaid_period"); ok {
			request.InstanceChargePrepaid.Period = period.(int)
		} else {
			diags = append(diags, diag.Diagnostic{
				Summary: "Missing required argument",
				Detail:  "instance_charge_prepaid_period is missing on prepaid instance.",
			})
			return diags
		}
	}
	request.InstanceType = d.Get("instance_type").(string)
	if v, ok := d.GetOk("image_id"); ok {
		request.ImageId = v.(string)
	}
	if v, ok := d.GetOk("instance_name"); ok {
		request.InstanceName = v.(string)
	}
	if v, ok := d.GetOk("password"); ok {
		request.Password = v.(string)
	}
	if v, ok := d.GetOk("key_id"); ok {
		request.KeyId = v.(string)
	}

	request.InternetChargeType = d.Get("internet_charge_type").(string)
	if request.InternetChargeType == VmInternetChargeTypeTrafficPackage && request.InstanceChargeType == VmChargeTypePrepaid {
		if v, ok := d.GetOk("traffic_package_size"); ok {
			request.TrafficPackageSize = common.Float64(v.(float64))
		} else {
			diags = append(diags, diag.Diagnostic{
				Summary: "Missing required argument",
				Detail:  "traffic_package_size is missing with `ByTrafficPackage` instance.",
			})
			return diags
		}
	}
	if v, ok := d.GetOk("internet_max_bandwidth_out"); ok {
		request.InternetMaxBandwidthOut = v.(int)
	}
	if v, ok := d.GetOk("subnet_id"); ok {
		request.SubnetId = v.(string)
	}
	if v, ok := d.GetOk("resource_group_id"); ok {
		request.ResourceGroupId = v.(string)
	}
	if v, ok := d.GetOk("system_disk_size"); ok {
		request.SystemDisk = &vm.SystemDisk{}
		request.SystemDisk.DiskSize = v.(int)
	}

	instanceId := ""

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := meta.(*connectivity.ZenlayerCloudClient).WithVmClient().CreateInstances(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create vm instance.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": toJsonString(request),
				"err":     err.Error(),
			})
			return retryError(ctx, err)
		}

		tflog.Info(ctx, "Create vm instance success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  toJsonString(request),
			"response": toJsonString(response),
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
			VmInstanceStatusPending,
			VmInstanceStatusDeloying,
			VmInstanceStatusRebuilding,
		},
		Target: []string{
			VmInstanceStatusRunning,
		},
		Refresh:        vmService.InstanceStateRefreshFunc(ctx, instanceId, []string{VmInstanceStatusCreateFailed}),
		Timeout:        d.Timeout(schema.TimeoutCreate) - time.Minute,
		Delay:          10 * time.Second,
		MinTimeout:     5 * time.Second,
		NotFoundChecks: 3,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for vm instance (%s) to be created: %v", d.Id(), err))
	}

	return resourceZenlayerCloudVmInstanceRead(ctx, d, meta)
}

func resourceZenlayerCloudVmInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	instanceId := d.Id()

	vmService := VmService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var instance *vm.InstanceInfo
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		instance, errRet = vmService.DescribeInstanceById(ctx, instanceId)
		if errRet != nil {
			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if !ok {
				return retryError(ctx, errRet)
			} else if ee.Code == "INVALID_INSTANCE_NOT_FOUND" || ee.Code == ResourceNotFound {
				return nil
			}
			return retryError(ctx, errRet)
		}
		if instance != nil && vmInstanceIsOperating(instance.InstanceStatus) {
			return resource.RetryableError(fmt.Errorf("waiting for instance %s operation, current status: %s", instance.InstanceId, instance.InstanceStatus))
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if instance == nil || instance.InstanceStatus == VmInstanceStatusCreateFailed ||
		instance.InstanceStatus == VmInstanceStatusRecycle {
		d.SetId("")
		tflog.Info(ctx, "instance not exist or created failed or recycled", map[string]interface{}{
			"instanceId": instanceId,
		})
		return nil
	}

	// instance info
	_ = d.Set("availability_zone", instance.ZoneId)
	_ = d.Set("instance_charge_type", instance.InstanceChargeType)
	if instance.InstanceChargeType == VmChargeTypePrepaid {
		_ = d.Set("instance_charge_prepaid_period", instance.Period)
	}
	_ = d.Set("instance_type", instance.InstanceType)
	_ = d.Set("key_id", instance.KeyId)
	_ = d.Set("image_id", instance.ImageId)
	_ = d.Set("image_name", instance.ImageName)
	_ = d.Set("instance_name", instance.InstanceName)
	_ = d.Set("internet_charge_type", instance.InternetChargeType)
	_ = d.Set("internet_max_bandwidth_out", instance.InternetMaxBandwidthOut)
	if instance.InternetChargeType == VmInternetChargeTypeTrafficPackage && instance.InstanceChargeType == VmChargeTypePrepaid {
		_ = d.Set("traffic_package_size", *instance.TrafficPackageSize)
	}
	_ = d.Set("subnet_id", instance.SubnetId)
	_ = d.Set("resource_group_id", instance.ResourceGroupId)
	_ = d.Set("resource_group_name", instance.ResourceGroupName)
	_ = d.Set("public_ip_addresses", instance.PublicIpAddresses)
	_ = d.Set("private_ip_addresses", instance.PrivateIpAddresses)
	_ = d.Set("instance_status", instance.InstanceStatus)
	if instance.SystemDisk != nil {
		_ = d.Set("system_disk_size", instance.SystemDisk.DiskSize)
	}
	_ = d.Set("create_time", instance.CreateTime)
	_ = d.Set("expired_time", instance.ExpiredTime)

	return diags

}

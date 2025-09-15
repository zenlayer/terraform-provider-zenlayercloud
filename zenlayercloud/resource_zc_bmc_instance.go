/*
Provides a BMC instance resource.

~> **NOTE:** You can launch an BMC instance for a private network via specifying parameter `subnet_id`.

~> **NOTE:** At present, 'PREPAID' instance cannot be deleted and must wait it to be outdated and released automatically.

Example Usage

```hcl

data "zenlayercloud_bmc_zones" "default" {

}

data "zenlayercloud_bmc_instance_types" "default" {
  availability_zone = data.zenlayercloud_bmc_zones.default.zones.0.id
}

# Get a centos image which also supported to install on given instance type
data "zenlayercloud_bmc_images" "default" {
  catalog          = "centos"
  instance_type_id = data.zenlayercloud_bmc_instance_types.default.instance_types.0.id
}

resource "zenlayercloud_bmc_subnet" "default" {
  availability_zone = data.zenlayercloud_bmc_zones.default.zones.0.id
  name              = "test-subnet"
  cidr_block        = "10.0.10.0/24"
}

# Create a web server
resource "zenlayercloud_bmc_instance" "web" {
  availability_zone    = data.zenlayercloud_bmc_zones.default.zones.0.id
  image_id             = data.zenlayercloud_bmc_images.default.images.0.image_id
  internet_charge_type = "ByBandwidth"
  instance_type_id     = data.zenlayercloud_bmc_instance_types.default.instance_types.0.id
  password             = "Example~123"
  instance_name        = "web"
  subnet_id            =  zenlayercloud_bmc_subnet.default.id
}
```

Import

BMC instance can be imported using the id, e.g.

```
terraform import zenlayercloud_bmc_instance.foo 123123xxx
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
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	bmc "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/bmc20221120"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	"strconv"
	"time"
)

func resourceZenlayerCloudInstance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudInstanceCreate,
		ReadContext:   resourceZenlayerCloudInstanceRead,
		UpdateContext: resourceZenlayerCloudInstanceUpdate,
		DeleteContext: resourceZenlayerCloudInstanceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(common2.BmcCreateTimeout),
			Update: schema.DefaultTimeout(common2.BmcUpdateTimeout),
		},
		CustomizeDiff: customdiff.All(
			internetMaxBandwidthOutForceNew(),
			trafficPackageSizeForceNew(),
			trafficPackageSizeValidFunc(),
			trafficPackageSizeForPostPaidFunc(),
		),
		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of zone that the bmc instance locates at.",
			},
			"instance_type_id": {
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
			"hostname": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "Terraform-Instance",
				ValidateFunc: validation.StringLenBetween(2, 64),
				Description:  "The hostname of the instance. The name should be a combination of 2 to 64 characters comprised of letters (case insensitive), numbers, hyphens (-) and Period (.), and the name must be start with letter. The default value is `Terraform-Instance`. Modifying will cause the instance reset.",
			},
			"instance_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "Terraform-Instance",
				ValidateFunc: validation.StringLenBetween(2, 64),
				Description:  "The name of the instance. The max length of instance_name is 64, and default value is `Terraform-Instance`.",
			},
			"instance_charge_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "POSTPAID",
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(BmcChargeTypes, false),
				Description:  "The charge type of instance. Valid values are `PREPAID`, `POSTPAID`. The default is `POSTPAID`. Note: `PREPAID` instance may not allow to delete before expired.",
			},
			"instance_charge_prepaid_period": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1),
				Description:  "The tenancy (time unit is month) of the prepaid instance, NOTE: it only works when instance_charge_type is set to `PREPAID`.",
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
			"password": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.All(validation.StringLenBetween(8, 16)),
				Description:  "Password for the instance. The max length of password is 16. Modifying will cause the instance reset.",
			},
			"ssh_keys": {
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"password"},
				Description:   "The ssh keys to use for the instance. The max number of ssh keys is 5. Modifying will cause the instance reset.",
				Set:           schema.HashString,
				Elem:          &schema.Schema{Type: schema.TypeString},
			},
			"internet_charge_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(BmcInternetChargeTypes, false),
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
				Description: "The ID of a VPC subnet. If you want to create instances in a VPC network, this parameter must be set.",
			},
			// raid
			"raid_config_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Simple config for instance raid. Modifying will cause the instance reset.",
				ValidateFunc: validation.StringInSlice([]string{"0", "1", "5", "10"}, false),
			},
			"raid_config_custom": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"raid_config_type"},
				Description:   "Custom config for instance raid. Modifying will cause the instance reset.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"raid_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"0", "1", "5", "10"}, false),
							Description:  "Simple config for raid.",
						},
						"disk_sequence": {
							Type:        schema.TypeList,
							Required:    true,
							Elem:        &schema.Schema{Type: schema.TypeInt},
							Description: "The sequence of disk to make raid.",
						},
					},
				},
			},
			"force_delete": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Indicate whether to force delete the instance. Default is `false`. If set true, the instance will be permanently deleted instead of being moved into the recycle bin.",
			},
			// nic
			"nic_wan_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(4, 10),
				Description:  "The wan name of the nic. The wan name should be a combination of 4 to 10 characters comprised of letters (case insensitive), numbers. The wan name must start with letter. Modifying will cause the instance reset.",
			},
			"nic_lan_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(4, 10),
				Description:  "The lan name of the nic. The lan name should be a combination of 4 to 10 characters comprised of letters (case insensitive), numbers. The lan name must start with letter. Modifying will cause the instance reset.",
			},
			// partition
			"partitions": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Partition for the instance. Modifying will cause the instance reset.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"fs_type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The type of the partitioned file.",
						},
						"fs_path": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The drive letter(windows) or device name(linux) for the partition.",
						},
						"size": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(1),
							Description:  "The size of the partitioned disk.",
						},
					},
				},
			},
			"primary_ipv4_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Primary Ipv4 address of the instance.",
			},
			"public_ipv4_addresses": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Public Ipv4 addresses bind to the instance.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"public_ipv6_addresses": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Public Ipv6 addresses of the instance.",
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

func trafficPackageSizeForceNew() schema.CustomizeDiffFunc {
	return customdiff.ForceNewIf("traffic_package_size", forceNewIfTrafficPackageSizeDowngradeForPrepaidInstance)
}
func forceNewIfTrafficPackageSizeDowngradeForPrepaidInstance(_ context.Context, d *schema.ResourceDiff, _ interface{}) bool {
	if change := d.HasChange("traffic_package_size"); !change {
		return false
	}

	internetChargeType := d.Get("internet_charge_type").(string)
	if internetChargeType != BmcInternetChargeTypeTrafficPackage {
		return false
	}

	chargeType := d.Get("instance_charge_type").(string)
	if chargeType == BmcChargeTypePostpaid {
		return false
	}
	oldValue, newValue := d.GetChange("traffic_package_size")
	return oldValue.(float64) > newValue.(float64)
}
func internetMaxBandwidthOutForceNew() schema.CustomizeDiffFunc {
	return customdiff.ForceNewIf("internet_max_bandwidth_out", forceNewIfBandwidthDowngradeForPrepaidInstance)
}

func trafficPackageSizeValidFunc() schema.CustomizeDiffFunc {
	return customdiff.IfValue("internet_charge_type", func(ctx context.Context, value, meta interface{}) bool {
		return value != BmcInternetChargeTypeTrafficPackage
	}, func(ctx context.Context, diff *schema.ResourceDiff, i interface{}) error {
		if _, ok := diff.GetOk("traffic_package_size"); ok {
			return fmt.Errorf("traffic_package_size can't be set as the internet charge type of instance is not `ByTrafficPackage`")
		}
		return nil
	})
}

func trafficPackageSizeForPostPaidFunc() schema.CustomizeDiffFunc {

	return customdiff.If(func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) bool {
		internetType := d.Get("internet_charge_type")
		chargeType := d.Get("instance_charge_type")
		return internetType == BmcInternetChargeTypeTrafficPackage && chargeType == BmcChargeTypePostpaid

	}, func(ctx context.Context, diff *schema.ResourceDiff, i interface{}) error {
		if _, ok := diff.GetOk("traffic_package_size"); ok {
			return fmt.Errorf("traffic_package_size can't be set for post paid instance with internet type `%s`", BmcInternetChargeTypeTrafficPackage)
		}
		return nil
	})
}

func forceNewIfBandwidthDowngradeForPrepaidInstance(_ context.Context, d *schema.ResourceDiff, _ interface{}) bool {
	if change := d.HasChange("internet_max_bandwidth_out"); !change {
		return false
	}

	internetChargeType := d.Get("internet_charge_type").(string)
	if internetChargeType != BmcInternetChargeTypeBandwidth {
		return false
	}

	chargeType := d.Get("instance_charge_type").(string)
	if chargeType == BmcChargeTypePostpaid {
		return false
	}
	oldValue, newValue := d.GetChange("internet_max_bandwidth_out")
	return oldValue.(int) > newValue.(int)
}

func resourceZenlayerCloudInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceId := d.Id()

	// force_delete: terminate and then delete
	forceDelete := d.Get("force_delete").(bool)

	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := bmcService.DeleteInstance(ctx, instanceId)
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
		instance, errRet := bmcService.DescribeInstanceById(ctx, instanceId)

		if errRet != nil {
			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if !ok {
				return common2.RetryError(ctx, errRet, common2.InternalServerError)
			} else if ee.Code == "INVALID_INSTANCE_NOT_FOUND" || ee.Code == common2.ResourceNotFound {
				notExist = true
				return nil
			}
			return common2.RetryError(ctx, errRet, common2.InternalServerError)
		}
		if instance == nil {
			notExist = true
			return nil
		}

		if instance.InstanceStatus == BmcInstanceStatusRecycle {
			//in recycling
			return nil
		}

		return resource.NonRetryableError(fmt.Errorf("bmc instance status is not recycle, current status %s", instance.InstanceStatus))
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
		errRet := bmcService.DestroyInstance(ctx, instanceId)
		if errRet != nil {

			//check InvalidInstanceState.Terminating
			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if !ok {
				return common2.RetryError(ctx, errRet)
			}
			if ee.Code == "INVALID_INSTANCE_NOT_FOUND" || ee.Code == common2.ResourceNotFound {
				// instance doesn't exist
				return nil
			}
			return common2.RetryError(ctx, errRet, common2.InternalServerError)
		}
		return nil
	})

	return nil
}

func resourceZenlayerCloudInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	//var _ diag.Diagnostics
	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	instanceId := d.Id()
	d.Partial(true)

	if d.HasChange("instance_name") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			err := bmcService.ModifyInstanceName(ctx, instanceId, d.Get("instance_name").(string))
			if err != nil {
				return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("subnet_id") {
		o, n := d.GetChange("subnet_id")
		if o.(string) != "" {
			err := bmcService.DisassociateSubnetInstance(ctx, instanceId, o.(string))
			if err != nil {
				return diag.FromErr(err)
			}
			err = waitSubnetChangeOk(ctx, bmcService, d, instanceId, n.(string), InstanceSubnetStatusNotBind)
			if err != nil {
				return diag.FromErr(err)
			}
		}
		err := bmcService.AssociateSubnetInstance(ctx, instanceId, n.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		err = waitSubnetChangeOk(ctx, bmcService, d, instanceId, n.(string), BmcSubnetInstanceStatusBound)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("resource_group_id") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			err := bmcService.ModifyInstanceResourceGroup(ctx, instanceId, d.Get("resource_group_id").(string))
			if err != nil {
				return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("internet_max_bandwidth_out") {
		if v, ok := d.GetOk("internet_max_bandwidth_out"); ok {

			err := bmcService.updateInstanceInternetMaxBandwidthOut(ctx, instanceId, v.(int))

			if err != nil {
				return diag.FromErr(err)
			}
			err = waitNetworkStatusOK(ctx, bmcService, d, instanceId, &internetBandwidthOutCondition{
				InternetBandwidthOut: v.(int),
			})
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("traffic_package_size") {
		if v, ok := d.GetOk("traffic_package_size"); ok {
			err := bmcService.updateInstanceTrafficPackageSize(ctx, instanceId, v.(float64))
			if err != nil {
				return diag.FromErr(err)
			}
			err = waitNetworkStatusOK(ctx, bmcService, d, instanceId, &trafficPackageCondition{
				TargetPackageSize: v.(float64),
			})
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	// need to reinstall the bmc instance
	if d.HasChanges("hostname", "password", "ssh_keys", "image_id", "partitions", "raid_config_type", "raid_config_custom", "nic_lan_name", "nic_wan_name") {

		request := bmc.NewReinstallInstanceRequest()
		request.InstanceId = d.Id()
		if v, ok := d.GetOk("hostname"); ok {
			request.Hostname = v.(string)
		}
		if v, ok := d.GetOk("password"); ok {
			request.Password = v.(string)
		}
		if v, ok := d.GetOk("ssh_keys"); ok {
			sshKeys := v.(*schema.Set).List()
			if len(sshKeys) > 0 {
				request.SshKeys = common2.ToStringList(sshKeys)
			}
		}
		if v, ok := d.GetOk("image_id"); ok {
			request.ImageId = v.(string)
		}
		// nic
		if v, ok := d.GetOk("nic_wan_name"); ok {
			request.Nic = &bmc.Nic{
				WanName: v.(string),
			}
		}

		if v, ok := d.GetOk("nic_lan_name"); ok {
			if request.Nic == nil {
				request.Nic = &bmc.Nic{}
			}
			request.Nic.LanName = v.(string)
		}
		// raid
		if v, ok := d.GetOk("raid_config_type"); ok {
			raidType, err := strconv.Atoi(v.(string))
			if err != nil {
				return diag.FromErr(err)
			}
			request.RaidConfig = &bmc.RaidConfig{
				RaidType: common.Integer(raidType),
			}
		}

		if v, ok := d.GetOk("raid_config_custom"); ok {
			raidCustom := v.([]interface{})
			raidParams := make([]*bmc.CustomRaid, 0, len(raidCustom))

			for i := range raidCustom {
				raidTypeKey := fmt.Sprintf("raid_config_custom.%d.raid_type", i)
				diskSequenceKey := fmt.Sprintf("raid_config_custom.%d.disk_sequence", i)

				raidType, err := strconv.Atoi(d.Get(raidTypeKey).(string))
				if err != nil {
					return diag.FromErr(err)
				}
				diskSequence := d.Get(diskSequenceKey).([]interface{})

				raidParams = append(raidParams, &bmc.CustomRaid{
					RaidType:     common.Integer(raidType),
					DiskSequence: common2.ToIntList(diskSequence),
				})
			}
			request.RaidConfig = &bmc.RaidConfig{
				CustomRaids: raidParams,
			}
		}

		if v, ok := d.GetOk("partitions"); ok {
			partitions := v.([]interface{})
			partitionValue := make([]*bmc.Partition, 0, len(partitions))

			for i := range partitions {
				fsTypeKey := fmt.Sprintf("partitions.%d.fs_type", i)
				fsPathKey := fmt.Sprintf("partitions.%d.fs_path", i)
				sizeKey := fmt.Sprintf("partitions.%d.size", i)

				partitionValue = append(partitionValue, &bmc.Partition{
					FsType: d.Get(fsTypeKey).(string),
					FsPath: d.Get(fsPathKey).(string),
					Size:   d.Get(sizeKey).(int),
				})
			}
			request.Partitions = partitionValue
		}

		err := bmcService.reinstallInstance(ctx, request)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	d.Partial(false)

	return resourceZenlayerCloudInstanceRead(ctx, d, meta)
}

func waitSubnetChangeOk(ctx context.Context, bmcService BmcService, d *schema.ResourceData, instanceId string, subnetId string, targetStatus string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			BmcSubnetInstanceStatusBinding,
			BmcSubnetInstanceStatusUnbinding,
		},
		Target: []string{
			targetStatus,
		},
		Refresh:        bmcService.InstanceSubnetStateRefreshFunc(ctx, instanceId, subnetId),
		Timeout:        d.Timeout(schema.TimeoutUpdate) - time.Minute,
		Delay:          5 * time.Second,
		MinTimeout:     5 * time.Second,
		NotFoundChecks: 3,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("error waiting for bmc instance (%s) to join subnet: %v", d.Id(), err)
	}
	return nil

}

func waitNetworkStatusOK(ctx context.Context, bmcService BmcService, d *schema.ResourceData, instanceId string, condition NetworkStateCondition) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			NetworkStatusPending,
		},
		Target: []string{
			NetworkStatusOK,
		},
		Refresh:        bmcService.InstanceNetworkStateRefreshFunc(ctx, instanceId, condition),
		Timeout:        d.Timeout(schema.TimeoutCreate) - time.Minute,
		Delay:          10 * time.Second,
		MinTimeout:     5 * time.Second,
		NotFoundChecks: 3,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("error waiting for bmc instance (%s) internet changed: %v", d.Id(), err)
	}
	return nil
}

type trafficPackageCondition struct {
	TargetPackageSize float64
}

func (t *trafficPackageCondition) matchFail(status *bmc.InstanceInternetStatus) bool {
	return status.ModifiedTrafficPackageStatus == "Enable" && t.TargetPackageSize != *status.TrafficPackageSize
}

func (t *trafficPackageCondition) matchOk(status *bmc.InstanceInternetStatus) bool {
	return status.ModifiedTrafficPackageStatus == "Enable" && t.TargetPackageSize == *status.TrafficPackageSize
}

type internetBandwidthOutCondition struct {
	InternetBandwidthOut int
}

func (t *internetBandwidthOutCondition) matchFail(status *bmc.InstanceInternetStatus) bool {
	return status.ModifiedBandwidthStatus == "Enable" && t.InternetBandwidthOut != *status.InternetMaxBandwidthOut
}

func (t *internetBandwidthOutCondition) matchOk(status *bmc.InstanceInternetStatus) bool {
	return status.ModifiedBandwidthStatus == "Enable" && t.InternetBandwidthOut == *status.InternetMaxBandwidthOut
}

func resourceZenlayerCloudInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	request := bmc.NewCreateInstancesRequest()
	request.ZoneId = d.Get("availability_zone").(string)
	request.InstanceTypeId = d.Get("instance_type_id").(string)
	request.InstanceChargeType = d.Get("instance_charge_type").(string)

	if request.InstanceChargeType == BmcChargeTypePrepaid {
		request.InstanceChargePrepaid = &bmc.ChargePrepaid{}

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
	if v, ok := d.GetOk("instance_name"); ok {
		request.InstanceName = v.(string)
	}
	if v, ok := d.GetOk("hostname"); ok {
		request.Hostname = v.(string)
	}
	if v, ok := d.GetOk("image_id"); ok {
		request.ImageId = v.(string)
	}
	if v, ok := d.GetOk("resource_group_id"); ok {
		request.ResourceGroupId = v.(string)
	}
	if v, ok := d.GetOk("password"); ok {
		request.Password = v.(string)
	}
	if v, ok := d.GetOk("ssh_keys"); ok {
		sshKeys := v.(*schema.Set).List()
		if len(sshKeys) > 0 {
			request.SshKeys = common2.ToStringList(sshKeys)
		}
	}
	request.InternetChargeType = d.Get("internet_charge_type").(string)
	if request.InternetChargeType == BmcInternetChargeTypeTrafficPackage && request.InstanceChargeType == VmChargeTypePrepaid {
		if v, ok := d.GetOk("traffic_package_size"); ok {
			request.TrafficPackageSize = v.(float64)
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
	// raid
	if v, ok := d.GetOk("raid_config_type"); ok {
		raidType, err := strconv.Atoi(v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		request.RaidConfig = &bmc.RaidConfig{
			RaidType: common.Integer(raidType),
		}
	}
	if v, ok := d.GetOk("raid_config_custom"); ok {
		customRaidConfig := v.([]interface{})
		request.RaidConfig = &bmc.RaidConfig{
			CustomRaids: make([]*bmc.CustomRaid, 0, len(customRaidConfig)),
		}
		for _, d := range customRaidConfig {
			value := d.(map[string]interface{})
			raidType, err := strconv.Atoi(value["raid_type"].(string))
			if err != nil {
				return diag.FromErr(err)
			}
			diskSeq := value["disk_sequence"].([]interface{})

			customRaidConf := bmc.CustomRaid{
				RaidType:     &raidType,
				DiskSequence: common2.ToIntList(diskSeq),
			}
			request.RaidConfig.CustomRaids = append(request.RaidConfig.CustomRaids, &customRaidConf)
		}
	}

	// nic
	if v, ok := d.GetOk("nic_wan_name"); ok {
		request.Nic = &bmc.Nic{}
		request.Nic.WanName = v.(string)
	}

	if v, ok := d.GetOk("nic_lan_name"); ok {
		if request.Nic == nil {
			request.Nic = &bmc.Nic{}
		}
		request.Nic.LanName = v.(string)
	}
	// partitions
	if v, ok := d.GetOk("partitions"); ok {
		partitions := v.([]interface{})
		request.Partitions = make([]*bmc.Partition, 0, len(partitions))

		for _, d := range partitions {
			value := d.(map[string]interface{})
			fsType := value["fs_type"].(string)
			fsPath := value["fs_path"].(string)
			size := value["size"].(int)

			partition := bmc.Partition{
				FsType: fsType,
				FsPath: fsPath,
				Size:   size,
			}
			request.Partitions = append(request.Partitions, &partition)
		}
	}

	instanceId := ""

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := meta.(*connectivity.ZenlayerCloudClient).WithBmcClient().CreateInstances(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create bmc instance.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": common2.ToJsonString(request),
				"err":     err.Error(),
			})
			return common2.RetryError(ctx, err)
		}

		tflog.Info(ctx, "Create instance success", map[string]interface{}{
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
			BmcInstanceStatusPending,
			BmcInstanceStatusCreating,
			BmcInstanceStatusInstalling,
		},
		Target: []string{
			BmcInstanceStatusRunning,
		},
		Refresh:        bmcService.InstanceStateRefreshFunc(ctx, instanceId, []string{BmcInstanceStatusCreateFailed, BmcInstanceStatusInstallFailed}),
		Timeout:        d.Timeout(schema.TimeoutCreate) - time.Minute,
		Delay:          10 * time.Second,
		MinTimeout:     5 * time.Second,
		NotFoundChecks: 3,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for bmc instance (%s) to be created: %v", d.Id(), err))
	}

	return resourceZenlayerCloudInstanceRead(ctx, d, meta)
}

func resourceZenlayerCloudInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	instanceId := d.Id()

	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var instance *bmc.InstanceInfo
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		instance, errRet = bmcService.DescribeInstanceById(ctx, instanceId)
		if errRet != nil {
			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if !ok {
				return common2.RetryError(ctx, errRet)
			} else if ee.Code == "INVALID_INSTANCE_NOT_FOUND" || ee.Code == common2.ResourceNotFound {
				return nil
			}
			return common2.RetryError(ctx, errRet)
		}
		if instance != nil && instanceIsOperating(instance.InstanceStatus) {
			return resource.RetryableError(fmt.Errorf("waiting for instance %s operation, current status: %s", instance.InstanceId, instance.InstanceStatus))
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if instance == nil || instance.InstanceStatus == BmcInstanceStatusCreateFailed ||
		instance.InstanceStatus == BmcInstanceStatusRecycle {
		d.SetId("")
		tflog.Info(ctx, "instance not exist or created failed or recycled", map[string]interface{}{
			"instanceId": instanceId,
		})
		return nil
	}

	// instance info
	_ = d.Set("availability_zone", instance.ZoneId)
	_ = d.Set("instance_name", instance.InstanceName)
	_ = d.Set("hostname", instance.Hostname)
	_ = d.Set("image_id", instance.ImageId)
	_ = d.Set("image_name", instance.ImageName)
	_ = d.Set("instance_type_id", instance.InstanceTypeId)
	_ = d.Set("resource_group_id", instance.ResourceGroupId)
	_ = d.Set("resource_group_name", instance.ResourceGroupName)
	_ = d.Set("internet_charge_type", instance.InternetChargeType)
	_ = d.Set("internet_max_bandwidth_out", instance.BandwidthOutMbps)
	_ = d.Set("instance_charge_type", instance.InstanceChargeType)
	if instance.InstanceChargeType == BmcChargeTypePrepaid {
		_ = d.Set("instance_charge_prepaid_period", instance.Period)
	}
	if len(instance.SubnetIds) > 0 {
		_ = d.Set("subnet_id", instance.SubnetIds[0])
	}

	_ = d.Set("primary_ipv4_address", instance.PrimaryPublicIpAddress)
	_ = d.Set("public_ipv4_addresses", instance.PublicIpAddresses)
	_ = d.Set("public_ipv6_addresses", instance.Ipv6Addresses)
	_ = d.Set("private_ip_addresses", instance.PrivateIpAddresses)
	_ = d.Set("instance_status", instance.InstanceStatus)
	_ = d.Set("create_time", instance.CreateTime)
	_ = d.Set("expired_time", instance.ExpiredTime)

	if instance.InternetChargeType == BmcInternetChargeTypeTrafficPackage {
		_ = d.Set("traffic_package_size", *instance.TrafficPackageSize)
	}

	if len(instance.Partitions) > 0 {
		partitionList := make([]map[string]interface{}, 0, len(instance.Partitions))

		for _, instancePartition := range instance.Partitions {
			partition := make(map[string]interface{}, 3)
			partition["fs_type"] = instancePartition.FsType
			partition["fs_path"] = instancePartition.FsPath
			partition["size"] = instancePartition.Size
			partitionList = append(partitionList, partition)
		}
		_ = d.Set("partitions", partitionList)
	}
	if instance.Nic != nil && instance.Nic.WanName != "" {
		_ = d.Set("nic_wan_name", instance.Nic.WanName)
	}
	if instance.Nic != nil && instance.Nic.LanName != "" {
		_ = d.Set("nic_lan_name", instance.Nic.LanName)
	}
	if instance.RaidConfig != nil && instance.RaidConfig.RaidType != nil {
		_ = d.Set("raid_config_type", strconv.Itoa(*instance.RaidConfig.RaidType))
	}
	if instance.RaidConfig != nil && len(instance.RaidConfig.CustomRaids) > 0 {

		customRaids := make([]map[string]interface{}, 0, len(instance.RaidConfig.CustomRaids))

		for _, customRaid := range instance.RaidConfig.CustomRaids {
			raidValue := make(map[string]interface{}, 2)
			raidValue["raid_type"] = strconv.Itoa(*customRaid.RaidType)
			raidValue["disk_sequence"] = customRaid.DiskSequence
			customRaids = append(customRaids, raidValue)
		}
		_ = d.Set("raid_config_custom", customRaids)
	}

	return diags

}

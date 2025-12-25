/*
Use this data source to query information about BMC instance hardware health status.

~> **NOTE:** Different hardware vendors use different starting indices for CPU numbering (some start from 0, others from 1). The attribute names (e.g., cpu0_temp, cpu1_temp, cpu2_temp) retain the original vendor's numbering style.

Example Usage

```hcl

data "zenlayercloud_bmc_instance_health_status" "foo" {
  instance_id = "<instanceId>"
}

```
*/
package zenlayercloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
)

func dataSourceZenlayerCloudInstanceHealthStatus() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudInstanceHealthStatusRead,

		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the instance to query health status.",
			},
			// Computed value
			"cpu_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "CPU status. OK: Normal; WARNING: Abnormal state; UNKNOWN: State detected failed.",
			},
			"disk_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Disk status. OK: Normal; WARNING: Abnormal state; UNKNOWN: State detected failed.",
			},
			"ipmi_ping": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IPMI IP connectivity. OK: ICMP reachable; CRITICAL: ICMP unreachable; UNKNOWN: State detected failed.",
			},
			"ipmi_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IPMI status. OK: ICMP reachable; WARNING: Abnormal state; UNKNOWN: State detected failed.",
			},
			"memory_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Memory status. OK: Normal; WARNING: Abnormal state; UNKNOWN: State detected failed.",
			},
			"psu_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Power Supply status. OK: Normal; WARNING: Abnormal state; UNKNOWN: State detected failed.",
			},
			"wan_port_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "WAN port status of the switch connected to the server's public network port.",
			},
			"fan_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Fan status. OK: Normal. WARNING: Abnormal state. UNKNOWN: State detected failed.",
			},
			"server_brand": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Server supplier brand.",
			},
			"server_model": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Server supplier model.",
			},
			"cpu_temp": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Temperature of a single CPU in specific server models (e.g., Supermicro blade servers). The range is from 0 to 100. The unit is Celsius. Note that a value of 0 is generally not retrievable, and a value of 100 signifies an exceptionally high temperature.",
			},
			"cpu0_temp": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "CPU temperature at index 0. The range is from 0 to 100. The unit is Celsius. If the value is empty, it means the temperature is not retrievable.",
			},
			"cpu1_temp": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "CPU temperature at index 1. The range is from 0 to 100. The unit is Celsius. If the value is empty, it means the temperature is not retrievable.",
			},
			"cpu2_temp": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "CPU temperature at index 2. The range is from 0 to 100. The unit is Celsius. If the value is empty, it means the temperature is not retrievable.",
			},
			"inlet_temp": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Temperature of the air or environment surrounding the server equipment in a data center or server room.",
			},
			"temp_unit": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Temperature unit. Only Celsius is supported, that is Celsius.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
		},
	}
}

func dataSourceZenlayerCloudInstanceHealthStatusRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "data_source.zenlayercloud_bmc_instance_health_status.read")()

	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	instanceId := d.Get("instance_id").(string)

	healthStatus, err := bmcService.DescribeInstanceMonitorHealth(ctx, instanceId)
	if err != nil {
		return diag.FromErr(err)
	}

	if healthStatus == nil {
		return diag.Errorf("instance health status not found for instance: %s", instanceId)
	}

	d.SetId(instanceId)
	_ = d.Set("cpu_status", healthStatus.CpuStatus)
	_ = d.Set("disk_status", healthStatus.DiskStatus)
	_ = d.Set("ipmi_ping", healthStatus.IpmiPing)
	_ = d.Set("ipmi_status", healthStatus.IpmiStatus)
	_ = d.Set("memory_status", healthStatus.MemoryStatus)
	_ = d.Set("psu_status", healthStatus.PsuStatus)
	_ = d.Set("wan_port_status", healthStatus.WanPortStatus)
	_ = d.Set("fan_status", healthStatus.FanStatus)
	_ = d.Set("server_brand", healthStatus.ServerBrand)
	_ = d.Set("server_model", healthStatus.ServerModel)
	_ = d.Set("cpu_temp", healthStatus.CpuTemp)
	_ = d.Set("cpu0_temp", healthStatus.Cpu0Temp)
	_ = d.Set("cpu1_temp", healthStatus.Cpu1Temp)
	_ = d.Set("cpu2_temp", healthStatus.Cpu2Temp)
	_ = d.Set("inlet_temp", healthStatus.InletTemp)
	_ = d.Set("temp_unit", healthStatus.TempUnit)



	tmpList := make([]map[string]interface{}, 0)
	mapping := map[string]interface{}{
		"instance_id":     instanceId,
		"cpu_status":      healthStatus.CpuStatus,
		"disk_status":     healthStatus.DiskStatus,
		"ipmi_ping":       healthStatus.IpmiPing,
		"ipmi_status":     healthStatus.IpmiStatus,
		"memory_status":   healthStatus.MemoryStatus,
		"psu_status":      healthStatus.PsuStatus,
		"wan_port_status": healthStatus.WanPortStatus,
		"fan_status":      healthStatus.FanStatus,
		"server_brand":    healthStatus.ServerBrand,
		"server_model":    healthStatus.ServerModel,
		"cpu_temp":        healthStatus.CpuTemp,
		"cpu0_temp":       healthStatus.Cpu0Temp,
		"cpu1_temp":       healthStatus.Cpu1Temp,
		"cpu2_temp":       healthStatus.Cpu2Temp,
		"inlet_temp":      healthStatus.InletTemp,
		"temp_unit":       healthStatus.TempUnit,
	}
	tmpList = append(tmpList, mapping)

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if e := common2.WriteToFile(output.(string), tmpList); e != nil {
			return diag.FromErr(e)
		}
	}
	return nil
}

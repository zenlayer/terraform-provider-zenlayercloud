---
subcategory: "Bare Metal Cloud(BMC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_bmc_instance_health_status"
sidebar_current: "docs-zenlayercloud-datasource-bmc_instance_health_status"
description: |-
  Use this data source to query information about BMC instance hardware health status.
---

# zenlayercloud_bmc_instance_health_status

Use this data source to query information about BMC instance hardware health status.

~> **NOTE:** Different hardware vendors use different starting indices for CPU numbering (some start from 0, others from 1). The attribute names (e.g., cpu0_temp, cpu1_temp, cpu2_temp) retain the original vendor's numbering style.

## Example Usage

```hcl
data "zenlayercloud_bmc_instance_health_status" "foo" {
  instance_id = "<instanceId>"
}
```

## Argument Reference

The following arguments are supported:

* `instance_id` - (Required, String) ID of the instance to query health status.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `cpu0_temp` - CPU temperature at index 0. The range is from 0 to 100. The unit is Celsius. If the value is empty, it means the temperature is not retrievable.
* `cpu1_temp` - CPU temperature at index 1. The range is from 0 to 100. The unit is Celsius. If the value is empty, it means the temperature is not retrievable.
* `cpu2_temp` - CPU temperature at index 2. The range is from 0 to 100. The unit is Celsius. If the value is empty, it means the temperature is not retrievable.
* `cpu_status` - CPU status. OK: Normal; WARNING: Abnormal state; UNKNOWN: State detected failed.
* `cpu_temp` - Temperature of a single CPU in specific server models (e.g., Supermicro blade servers). The range is from 0 to 100. The unit is Celsius. Note that a value of 0 is generally not retrievable, and a value of 100 signifies an exceptionally high temperature.
* `disk_status` - Disk status. OK: Normal; WARNING: Abnormal state; UNKNOWN: State detected failed.
* `fan_status` - Fan status. OK: Normal. WARNING: Abnormal state. UNKNOWN: State detected failed.
* `inlet_temp` - Temperature of the air or environment surrounding the server equipment in a data center or server room.
* `ipmi_ping` - IPMI IP connectivity. OK: ICMP reachable; CRITICAL: ICMP unreachable; UNKNOWN: State detected failed.
* `ipmi_status` - IPMI status. OK: ICMP reachable; WARNING: Abnormal state; UNKNOWN: State detected failed.
* `memory_status` - Memory status. OK: Normal; WARNING: Abnormal state; UNKNOWN: State detected failed.
* `psu_status` - Power Supply status. OK: Normal; WARNING: Abnormal state; UNKNOWN: State detected failed.
* `server_brand` - Server supplier brand.
* `server_model` - Server supplier model.
* `temp_unit` - Temperature unit. Only Celsius is supported, that is Celsius.
* `wan_port_status` - WAN port status of the switch connected to the server's public network port.



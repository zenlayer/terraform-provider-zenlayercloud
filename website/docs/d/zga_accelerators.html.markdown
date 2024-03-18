---
subcategory: "Zenlayer Global Accelerator(ZGA)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zga_accelerators"
sidebar_current: "docs-zenlayercloud-datasource-zga_accelerators"
description: |-
  Use this data source to get all zga accelerator.
---

# zenlayercloud_zga_accelerators

Use this data source to get all zga accelerator.

## Example Usage

```hcl
data "zenlayercloud_zga_accelerators" "all" {
}
```

## Argument Reference

The following arguments are supported:

* `accelerate_region_id` - (Optional, String) Accelerate region of the accelerator to be queried.
* `accelerator_ids` - (Optional, Set: [`String`]) IDs of the accelerator to be queried.
* `accelerator_name` - (Optional, String) The name of accelerator. The max length of accelerator name is 64.
* `accelerator_status` - (Optional, String) Status of the accelerator to be queried. Valid values are `Accelerating`, `NotAccelerate`, `Deploying`, `StopAccelerate`, `AccelerateFailure`.
* `cname` - (Optional, String) Cname of the accelerator to be queried.
* `domain` - (Optional, String) Domain of the accelerator to be queried.
* `origin_region_id` - (Optional, String) Origin region of the accelerator to be queried.
* `origin` - (Optional, String) Origin of the accelerator to be queried.
* `resource_group_id` - (Optional, String) The ID of resource group that the accelerator grouped by.
* `result_output_file` - (Optional, String) Used to save results.
* `vip` - (Optional, String) Virtual IP of the accelerator to be queried.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `accelerators` - An information list of accelerator. Each element contains the following attributes:
   * `accelerate_regions` - Accelerate region of the accelerator.
      * `accelerate_region_id` - ID of the accelerate region.
      * `accelerate_region_name` - Name of the accelerate region.
      * `accelerate_region_status` - Configuration status of the accelerate region.
      * `bandwidth` - Virtual IP the accelerate region.
      * `vip` - Virtual IP the accelerate region.
   * `accelerator_id` - ID of the accelerator.
   * `accelerator_name` - Name of the accelerator.
   * `accelerator_status` - Status of the accelerator.
   * `accelerator_type` - Type of the accelerator.
   * `access_control` - Access control of the accelerator.
      * `enable` - Whether to enable access control.
      * `rules` - Rules of the access control.
         * `cidr_ip` - The cidr ip of the rule.
         * `directory` - The directory of the rule.
         * `listener` - The listener of the rule.
         * `note` - The note of the rule.
         * `policy` - The policy of the rule.
   * `backup_origin` - Backup endpoint of the origin.
   * `certificate` - Certificate info of the accelerator.
      * `algorithm` - Algorithm of the certificate.
      * `certificate_id` - ID of the certificate.
      * `certificate_label` - Label of the certificate.
      * `common` - Common of the certificate.
      * `create_time` - Upload time of the certificate.
      * `dns_names` - DNS Names of the certificate.
      * `end_time` - Expiration time of the certificate.
      * `expired` - Whether the certificate has expired.
      * `fingerprint` - Md5 fingerprint of the certificate.
      * `issuer` - Issuer of the certificate.
      * `resource_group_id` - The ID of resource group that the instance belongs to.
      * `start_time` - Start time of the certificate.
   * `charge_type` - The charge type of the accelerator.
   * `cname` - Cname of the accelerator.
   * `create_time` - Create time of the accelerator.
   * `domain` - Main domain of the accelerator.
   * `health_check` - Health check of the accelerator.
      * `alarm` - Whether to enable alarm.
      * `enable` - Whether to enable health check.
      * `port` - The port of health check.
   * `l4_listeners` - L4 listeners of the accelerator.
      * `back_port_range` - The Return-to-origin port range of the l4 listener.
      * `back_port` - The Return-to-origin port of the l4 listener.
      * `port_range` - The port range of the l4 listener.
      * `port` - The port of the l4 listener.
      * `protocol` - The protocol of the l4 listener.
   * `l7_listeners` - L7 listeners of the accelerator.
      * `back_port_range` - The Return-to-origin port range of the l7 listener.
      * `back_port` - The Return-to-origin port of the l7 listener.
      * `back_protocol` - The Return-to-origin protocol of the l7 listener.
      * `host` - The Return-to-origin host of the l7 listener.
      * `port_range` - The port range of the l7 listener.
      * `port` - The port of the l7 listener.
      * `protocol` - The protocol of the l7 listener.
   * `origin_region_id` - ID of the orgin region.
   * `origin_region_name` - Name of the orgin region.
   * `origin` - Endpoints of the origin.
   * `protocol_opts` - Protocol opts of the accelerator.
      * `gzip` - Whether to enable gzip.
      * `proxy_protocol` - Whether to enable proxyProtocol.
      * `toa_value` - TOA verison.
      * `toa` - Whether to enable TOA.
      * `websocket` - Whether to enable websocket.
   * `relate_domains` - Relate domains of the accelerator.
   * `resource_group_id` - The ID of resource group that the instance belongs to.



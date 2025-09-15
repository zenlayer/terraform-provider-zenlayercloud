---
subcategory: "Zenlayer Load Balancing(ZLB)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zlb_listeners"
sidebar_current: "docs-zenlayercloud-datasource-zlb_listeners"
description: |-
  Use this data source to query detailed information of ZLB listener
---

# zenlayercloud_zlb_listeners

Use this data source to query detailed information of ZLB listener

## Example Usage

Query all listeners of ZLB instance

```hcl
data "zenlayercloud_zlb_listeners" "all" {
  zlb_id = "<zlbId>"
}
```

Query listeners by listener protocol

```hcl
data "zenlayercloud_zlb_listeners" "foo" {
  zlb_id   = "<zlbId>"
  protocol = "TCP"
}
```

## Argument Reference

The following arguments are supported:

* `zlb_id` - (Required, String) The ID of load balancer that the listeners belong to.
* `ids` - (Optional, Set: [`String`]) IDs of the load balancer listeners to be queried.
* `name_regex` - (Optional, String) A regex string to filter results by listener name.
* `protocol` - (Optional, String) The protocol of listeners to be queried.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `listeners` - An information list of listeners. Each element contains the following attributes:
   * `create_time` - Create time of the listener.
   * `health_check_conn_timeout` - Connection timeout for health check.
   * `health_check_delay_loop` - Interval between health checks. Measured in second.
   * `health_check_delay_try` - Health check delay try time.
   * `health_check_enabled` - Indicates whether health check is enabled.
   * `health_check_http_get_url` - HTTP request URL for health check.
   * `health_check_http_status_code` - HTTP status code for health check.
   * `health_check_port` - Health check port. Defaults to the backend server port.
   * `health_check_retry` - Number of retry attempts for health check.
   * `health_check_type` - Health check protocols.
   * `kind` - Forwarding mode of the listener. Valid values: `DR`, `FNAT`.
   * `listener_id` - ID of the load balancer listener.
   * `listener_name` - The name of the load balancer listener.
   * `port` - The port of listener. Use commas (,) to separate multiple ports. Use a hyphen (-) to define a port range, e.g., 10000-10005.
   * `protocol` - The protocol of listener.
   * `scheduler` - Scheduling algorithm of the listener.



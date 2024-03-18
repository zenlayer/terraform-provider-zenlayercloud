---
subcategory: "Cloud Networking(SDN)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_sdn_cloud_regions"
sidebar_current: "docs-zenlayercloud-datasource-sdn_cloud_regions"
description: |-
  Use this data source to query cloud regions.
---

# zenlayercloud_sdn_cloud_regions

Use this data source to query cloud regions.

## Example Usage

```hcl
data "zenlayercloud_sdn_cloud_regions" "google_regions" {
  cloud_type = "GOOGLE"
}
```

## Argument Reference

The following arguments are supported:

* `cloud_type` - (Required, String) The type of the cloud, Valid values: `AWS`, `TENCENT`, `GOOGLE`.
* `google_pairing_key` - (Optional, String) Google Paring key, which is required when cloud type is `GOOGLE`.
* `product` - (Optional, String) The product to be queried. Valid values: `PrivateConnect`, `CloudRouter`.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `region_list` - An information list of cloud region. Each element contains the following attributes:
  * `cloud_region` - ID of the cloud region.
  * `datacenter_name` - The name of datacenter.
  * `datacenter` - The id of datacenter that can be connect to cloud region.
  * `products` - The connect product.



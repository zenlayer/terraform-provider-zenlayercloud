---
subcategory: "Zenlayer Global Accelerator(ZGA)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zga_certificates"
sidebar_current: "docs-zenlayercloud-datasource-zga_certificates"
description: |-
  Use this data source to get all zga certificates.
---

# zenlayercloud_zga_certificates

Use this data source to get all zga certificates.

## Example Usage

```hcl
data "zenlayercloud_zga_certificates" "all" {
}
```

## Argument Reference

The following arguments are supported:

* `certificate_ids` - (Optional, Set: [`String`]) IDs of the certificates to be queried.
* `certificate_label` - (Optional, String) Label of the certificate to be queried.
* `dns_name` - (Optional, String) DNS Name of the certificate to be queried.
* `expired` - (Optional, Bool) Whether the certificate has expired.
* `resource_group_id` - (Optional, String) The ID of resource group that the certificate grouped by.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `certificates` - An information list of certificate. Each element contains the following attributes:
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



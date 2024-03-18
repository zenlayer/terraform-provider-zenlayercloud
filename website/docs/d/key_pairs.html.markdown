---
subcategory: "Zenlayer Virtual Machine(VM)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_key_pairs"
sidebar_current: "docs-zenlayercloud-datasource-key_pairs"
description: |-
  Use this data source to query SSH key pair list.
---

# zenlayercloud_key_pairs

Use this data source to query SSH key pair list.

## Example Usage

```hcl
data "zenlayercloud_key_pairs" "all" {
}

data "zenlayercloud_key_pairs" "myname" {
  key_name = "myname"
}
```

## Argument Reference

The following arguments are supported:

* `key_ids` - (Optional, Set: [`String`]) IDs of the key pair to be queried.
* `key_name` - (Optional, String) Name of the key pair to be queried.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `key_pairs` - An information list of key pairs. Each element contains the following attributes:
   * `create_time` - Create time of the key pair.
   * `key_description` - Description of the key pair.
   * `key_id` - ID of the key pair, such as `key-xxxxxxxx`.
   * `key_name` - Name of the key pair.
   * `public_key` - Public SSH keys in OpenSSH format, such as `ssh-rsa XXXXXXXXXXXX`.



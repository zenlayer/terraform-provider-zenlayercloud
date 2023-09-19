---
subcategory: "Zenlayer Virtual Machine(VM)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_key_pair"
sidebar_current: "docs-zenlayercloud-resource-key_pair"
description: |-
  Provides a resource to manage key pair.
---

# zenlayercloud_key_pair

Provides a resource to manage key pair.

~> **NOTE:** This request is to import an SSH key pair to be used for later instance login..

~> **NOTE:** A key pair name and several public SSH keys are required.

## Example Usage

```hcl
resource "zenlayercloud_key_pair" "foo" {
  key_name        = "my_key"
  public_key      = "ssh-rsa XXXXXXXXXXXX key"
  key_description = "create a key pair"
}
```

## Argument Reference

The following arguments are supported:

* `key_name` - (Required, String, ForceNew) Key pair name. Up to 32 characters in length are supported, containing letters, digits and special character -_. The names cannot be duplicated.
* `public_key` - (Required, String, ForceNew) Public SSH keys in OpenSSH format. Up to 5 public keys are allowed, separated by pressing ENTER key.
* `key_description` - (Optional, String) Description of key pair.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.



## Import

Key pair can be imported, e.g.

```
$ terraform import zenlayercloud_key_pair.foo key-xxxxxxx
```


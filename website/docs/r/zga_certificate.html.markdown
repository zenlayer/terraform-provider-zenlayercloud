---
subcategory: "Zenlayer Global Accelerator(ZGA)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zga_certificate"
sidebar_current: "docs-zenlayercloud-resource-zga_certificate"
description: |-
  Provides a certificate resource.
---

# zenlayercloud_zga_certificate

Provides a certificate resource.

~> **NOTE:** Modification of the certificate and key is not supported. If you want to change it, you need to create a new certificate.

~> **NOTE:** When the certificate and key are set to empty strings, the Update will not take effect.

## Example Usage

```hcl
resource "zenlayercloud_zga_certificate" "default" {
  certificate = <<EOF

-----BEGIN CERTIFICATE-----
[......] # cert contents
-----END CERTIFICATE-----
EOF

  key = <<EOF

-----BEGIN RSA PRIVATE KEY-----
[......] # key contents
-----END RSA PRIVATE KEY-----
EOF

  label = "certificate"

  lifecycle {
    create_before_destroy = true
  }
}
```

## Argument Reference

The following arguments are supported:

* `certificate` - (Required, String, ForceNew) The content of certificate.
* `key` - (Required, String, ForceNew) The key of the certificate.
* `label` - (Optional, String) The label of the certificate. Modification is not supported.
* `resource_group_id` - (Optional, String) The resource group id the certificate belongs to, default to Default Resource Group. Modification is not supported.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `common` - Common of the certificate.
* `create_time` - Uploaded time of the certificate.
* `end_time` - Expiration time of the certificate.


## Import

Certificate can be imported using the id, e.g.

```
terraform import zenlayercloud_zga_certificate.default certificateId
```


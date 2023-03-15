---
layout: "zenlayercloud"
page_title: "Provider: zenlayercloud"
sidebar_current: "docs-zenlayercloud-index"
description: |- The Zenlayercloud provider is used to interact with many resources supported by Zenlayer. The provider
needs to be configured with the proper credentials before it can be used.
---

# Zenlayer Cloud Provider

The Zenlayer Cloud provider is used to interact with the many resources supported
by [Zenlayer Cloud](https://console.zenlayer.com). The provider needs to be configured with the proper credentials
before it can be used.

Use the navigation on the left to read about the available resources.

## Example Usage

```hcl
terraform {
  required_providers {
    zenlayercloud = {
      source = "zenlayer/zenlayercloud"
    }
  }
}

provider "zenlayercloud" {
  access_key_id       = "your-access-key-id"
  access_key_password = "your-access-key-password"
}

locals {
  availability_zone = data.zenlayercloud_bmc_zones.default.zones.0.name
  instance_type_id  = data.zenlayercloud_bmc_instance_types.default.instance_types.0.instance_type_id
}

data "zenlayercloud_bmc_zones" "default" {

}

data "zenlayercloud_bmc_instance_types" "default" {
  availability_zone = local.availability_zone
}

# Get a centos image which also supported to install on given instance type
data "zenlayercloud_bmc_images" "default" {
  catalog          = "centos"
  instance_type_id = local.instance_type_id
}

resource "zenlayercloud_bmc_subnet" "default" {
  availability_zone = local.availability_zone
  name              = "test-subnet"
  cidr_block        = "10.0.10.0/24"
}


# Create a web server
resource "zenlayercloud_bmc_instance" "web" {
  availability_zone    = local.availability_zone
  image_id             = data.zenlayercloud_bmc_images.default.images.0.image_id
  internet_charge_type = "ByBandwidth"
  instance_type_id     = local.instance_type_id
  password             = "Example~123"
  instance_name        = "web"
  subnet_id            = zenlayercloud_bmc_subnet.default.id
}

```

## Authentication

The Zenlayercloud provider use access key credential for authentication.

### Credential

Credential can be provided by adding `access_key_id`, `access_key_password` in-line in the zenlayercloud provider block:

Usage:

```hcl
provider "zenlayercloud" {
  access_key_id       = "your-access-key-id"
  access_key_password = "your-access-key-password"
}
```

### Environment variables

You can provide your credentials via `ZENLAYERCLOUD_ACCESS_KEY_ID` and `ZENLAYERCLOUD_ACCESS_KEY_PASSWORD`
environment variables, representing your ZenlayerCloud access key id and access key password respectively.

```hcl
provider "zenlayercloud" {

}
```

Usage:

```shell
$ export ZENLAYERCLOUD_ACCESS_KEY_ID="your-access-key-id"
$ export ZENLAYERCLOUD_ACCESS_KEY_PASSWORD="your-access-key-password"
$ terraform plan
```

## Argument Reference

In addition to [generic `provider` arguments](https://www.terraform.io/docs/configuration/providers.html)
(e.g. `alias` and `version`), the following arguments are supported in the Zenlayer Cloud
`provider` block:

* `access_key_id` - This is the ZenlayerCloud access key id. It must be provided, but it can also be sourced from
  the `ZENLAYERCLOUD_ACCESS_KEY_ID` environment variable.

* `access_key_password` - This is the ZenlayerCloud access key password. It must be provided, but it can also be sourced
  from the `ZENLAYERCLOUD_ACCESS_KEY_PASSWORD` environment variable.

* `domain` - (Optional) The root domain of the API request, Default is console.zenlayer.com.

* `protocol` - (Optional) The protocol of the API request. Valid values: HTTP and HTTPS. Default is HTTPS.

* `client_timeout` - (Optional) The maximum timeout in second of the client request. Default to 600.

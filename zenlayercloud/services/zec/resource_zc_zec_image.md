Provides a resource to manage a custom image in Zenlayer Elastic Compute (ZEC).

Creating this resource will make a custom image from an existing ZEC instance.
Updating `image_name` calls `ModifyImagesAttributes` in place.

Example Usage

Create a custom image from an instance

```hcl
resource "zenlayercloud_zec_image" "foo" {
  availability_zone = "SH1A"
  instance_id       = "1660545330971680835"
  image_name        = "my-custom-image"

  tags = {
    env = "test"
  }
}
```

Import an existing custom image to manage its name

```hcl
resource "zenlayercloud_zec_image" "foo" {
  availability_zone = "SH1A"
  image_name        = "sherlock-test009"
}
```

```
$ terraform import zenlayercloud_zec_image.foo <imageId>
```

~> **NOTE:** After import, `availability_zone` must be set in HCL to the zone
where the image lives. The DescribeImages API requires `zoneId` for reads.
`instance_id` is only needed when creating a new image and is ignored on
import.

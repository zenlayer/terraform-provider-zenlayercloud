Provides a resource to manage a custom image in Zenlayer Elastic Compute (ZEC).

Creating this resource will make a custom image from an existing ZEC instance.
Updating `image_name` calls `ModifyImagesAttributes` in place.

Example Usage

Create a custom image from an instance

```hcl
resource "zenlayercloud_zec_image" "foo" {
  instance_id = "1660545330971680835"
  image_name  = "my-custom-image"

  tags = {
    env = "test"
  }
}
```

Import

A custom image can be imported using its id, e.g.

```
$ terraform import zenlayercloud_zec_image.foo <imageId>
```

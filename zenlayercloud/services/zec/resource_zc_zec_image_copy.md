Provides a resource to manage the cross-region distribution of a custom ZEC image.

Declare the full set of regions (including the source region) where the image
should exist; Terraform will call CopyImage for additions and DeleteImageCopy
for removals.

Example Usage

Distribute a custom image to multiple regions

```hcl
resource "zenlayercloud_zec_image_copy" "example" {
  image_id       = zenlayercloud_zec_image.img.id
  region_id_list = ["SHA", "SEL", "FRA"]
}
```

Import

An image-copy resource can be imported using the image id, e.g.

```
$ terraform import zenlayercloud_zec_image_copy.example <imageId>
```

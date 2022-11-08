package zenlayercloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

var testDataImageAll = "data.zenlayercloud_bmc_images.all"
var testDataImagePublic = "data.zenlayercloud_bmc_images.public"

func TestAccZenlayerCloudImagesDataSource(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testProviders(),
		Steps: []resource.TestStep{
			{
				Config: testAccZenlayerCloudDataBmcImagesAll,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZenlayerCloudDataResourceID(testDataImageAll),
					resource.TestCheckResourceAttrSet(testDataImageAll, "images.#"),
					resource.TestCheckResourceAttrSet(testDataImageAll, "images.0.image_id"),
					resource.TestCheckResourceAttrSet(testDataImageAll, "images.0.image_type"),
					resource.TestCheckResourceAttrSet(testDataImageAll, "images.0.image_name"),
					resource.TestCheckResourceAttrSet(testDataImageAll, "images.0.catalog"),
					resource.TestCheckResourceAttrSet(testDataImageAll, "images.0.os_type"),
				),
			},
			{
				Config: testAccZenlayerCloudDataBmcImagesPublic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZenlayerCloudDataResourceID(testDataImagePublic),
					resource.TestCheckResourceAttrSet(testDataImagePublic, "images.#"),
				),
			},
		},
	})
}

const testAccZenlayerCloudDataBmcImagesAll = `
data "zenlayercloud_bmc_images" "all" {
}
`

const testAccZenlayerCloudDataBmcImagesPublic = `
data "zenlayercloud_bmc_images" "public" {
	image_type = "` + ImageTypePublic + `"
}
`

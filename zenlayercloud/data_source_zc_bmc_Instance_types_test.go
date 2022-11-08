package zenlayercloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

var testDataInstanceTypesAll = "data.zenlayercloud_bmc_instance_types.all"
var testDataInstanceTypesExcludeSoldOut = "data.zenlayercloud_bmc_instance_types.public"

func TestAccZenlayerCloudInstanceTypesDataSource_Basic(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testProviders(),
		Steps: []resource.TestStep{
			{
				Config: testAccZenlayerCloudDataBmcInstanceTypesBasic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZenlayerCloudDataResourceID(testDataInstanceTypesAll),
					resource.TestCheckResourceAttrSet(testDataInstanceTypesAll, "images.#"),
					resource.TestCheckResourceAttrSet(testDataInstanceTypesAll, "images.0.image_id"),
					resource.TestCheckResourceAttrSet(testDataInstanceTypesAll, "images.0.image_type"),
					resource.TestCheckResourceAttrSet(testDataInstanceTypesAll, "images.0.image_name"),
					resource.TestCheckResourceAttrSet(testDataInstanceTypesAll, "images.0.catalog"),
					resource.TestCheckResourceAttrSet(testDataInstanceTypesAll, "images.0.os_type"),
				),
			},
			{
				Config: testAccZenlayerCloudDataBmcInstanceTypesExcludeSoldOut,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZenlayerCloudDataResourceID(testDataInstanceTypesExcludeSoldOut),
					resource.TestCheckResourceAttrSet(testDataInstanceTypesExcludeSoldOut, "images.#"),
				),
			},
		},
	})
}

const testAccZenlayerCloudDataBmcInstanceTypesBasic = `
data "zenlayercloud_bmc_instance_types" "all" {
}
`

const testAccZenlayerCloudDataBmcInstanceTypesExcludeSoldOut = `
data "zenlayercloud_bmc_instance_types" "sold" {
	
}
`

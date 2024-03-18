package zenlayercloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var testDataZgaOriginRegionAll = "data.zenlayercloud_zga_origin_regions.all"

func TestAccZenlayerCloudDataZgaOriginRegions_Basic(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testProviders(),
		Steps: []resource.TestStep{
			{
				Config: testAccZenlayerCloudDataZgaOriginRegionsBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZenlayerCloudDataResourceID(testDataZgaOriginRegionAll),
					resource.TestCheckResourceAttrSet(testDataZgaOriginRegionAll, "regions.#"),
					resource.TestCheckResourceAttrSet(testDataZgaOriginRegionAll, "regions.0.id"),
					resource.TestCheckResourceAttrSet(testDataZgaOriginRegionAll, "regions.0.description"),
				),
			},
		},
	})
}

const testAccZenlayerCloudDataZgaOriginRegionsBasic = `
data "zenlayercloud_zga_origin_regions" "all" {
}
`

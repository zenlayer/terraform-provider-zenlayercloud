package zenlayercloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var testDataZgaAccRegionFR = "data.zenlayercloud_zga_accelerate_regions.FR"

func TestAccZenlayerCloudDataZgaAccRegions_Basic(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testProviders(),
		Steps: []resource.TestStep{
			{
				Config: testAccZenlayerCloudDataZgaAccRegionsBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZenlayerCloudDataResourceID(testDataZgaAccRegionFR),
					resource.TestCheckResourceAttrSet(testDataZgaAccRegionFR, "regions.#"),
					resource.TestCheckResourceAttrSet(testDataZgaAccRegionFR, "regions.0.id"),
					resource.TestCheckResourceAttrSet(testDataZgaAccRegionFR, "regions.0.description"),
				),
			},
		},
	})
}

const testAccZenlayerCloudDataZgaAccRegionsBasic = `
data "zenlayercloud_zga_accelerate_regions" "FR" {
	origin_region_id = "FR"
}
`

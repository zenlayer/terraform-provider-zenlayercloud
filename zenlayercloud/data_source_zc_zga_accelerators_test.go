package zenlayercloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var testDataZgaAcceleratorAll = "data.zenlayercloud_zga_accelerators.all"

func TestAccZenlayerCloudDataZgaAccelerator_Basic(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testProviders(),
		Steps: []resource.TestStep{
			{
				Config: testAccZenlayerCloudDataZgaAcceleratorBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZenlayerCloudDataResourceID(testDataZgaAcceleratorAll),
					resource.TestCheckResourceAttrSet(testDataZgaAcceleratorAll, "accelerators.#"),
					resource.TestCheckResourceAttrSet(testDataZgaAcceleratorAll, "accelerators.0.accelerator_id"),
					resource.TestCheckResourceAttrSet(testDataZgaAcceleratorAll, "accelerators.0.accelerate_regions.0.vip"),
				),
			},
		},
	})
}

const testAccZenlayerCloudDataZgaAcceleratorBasic = `
data "zenlayercloud_zga_accelerators" "all" {
}
`

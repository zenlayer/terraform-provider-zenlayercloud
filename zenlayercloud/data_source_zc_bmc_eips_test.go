package zenlayercloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

var testDataEip = "data.zenlayercloud_bmc_eips.test"

func TestAccZenlayerCloudEipDataSource(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testProviders(),
		CheckDestroy:      testAccCheckEipResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccZenlayerCloudDataBmcEip,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZenlayerCloudDataResourceID(testDataEip),
					resource.TestCheckResourceAttr(testDataEip, "list.#", "1"),
					resource.TestCheckResourceAttrSet(testDataEip, "eip_list.0.create_time"),
					resource.TestCheckResourceAttrSet(testDataEip, "eip_list.0.public_ip"),
					resource.TestCheckResourceAttrSet(testDataEip, "eip_list.0.eip_status"),
					resource.TestCheckResourceAttrSet(testDataEip, "eip_list.0.resource_group_id"),
					resource.TestCheckResourceAttrSet(testDataEip, "eip_list.0.resource_group_name"),
				),
			},
		},
	})
}

const testAccZenlayerCloudDataBmcEip = defaultVariable + `
resource "zenlayercloud_bmc_eip" "test_eip" {
	availability_zone = var.availability_zone
}

data "zenlayercloud_bmc_eips" "test" {
	eip_ids = [zenlayercloud_bmc_eip.test_eip.id]
}
`

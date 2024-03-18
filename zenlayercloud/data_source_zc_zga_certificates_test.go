package zenlayercloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var testDataZgaCertificatesAll = "data.zenlayercloud_zga_certificates.all"

func TestAccZenlayerCloudDataZgaCertificates_Basic(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testProviders(),
		Steps: []resource.TestStep{
			{
				Config: testAccZenlayerCloudDataZgaCertificatesBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZenlayerCloudDataResourceID(testDataZgaCertificatesAll),
					resource.TestCheckResourceAttrSet(testDataZgaCertificatesAll, "certificates.#"),
					resource.TestCheckResourceAttrSet(testDataZgaCertificatesAll, "certificates.0.certificate_id"),
					resource.TestCheckResourceAttrSet(testDataZgaCertificatesAll, "certificates.0.common"),
				),
			},
		},
	})
}

const testAccZenlayerCloudDataZgaCertificatesBasic = `
data "zenlayercloud_zga_certificates" "all" {
}
`

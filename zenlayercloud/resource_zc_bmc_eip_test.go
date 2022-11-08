package zenlayercloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"testing"
)

func init() {
	resource.AddTestSweepers("zenlayercloud_bmc_eip", &resource.Sweeper{
		Name: "zenlayercloud_bmc_eip",
		F:    testSweepBmcEip,
	})
}

func testSweepBmcEip(region string) error {
	sharedClient, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("getting zenlayercloud client error: %s", err.Error())
	}
	client := sharedClient.(*connectivity.ZenlayerCloudClient)
	bmcService := BmcService{
		client: client,
	}

	_, err = bmcService.DescribeEipAddressesByFilter(&EipFilter{})
	if err != nil {
		return fmt.Errorf("get eip list error: %s", err.Error())
	}

	//for _, v := range instances {
	//	instanceId := v.InstanceId
	//	instanceName := v.InstanceName
	//	now := time.Now()
	//	//createTime := stringTotime(*v.CreatedTime)
	//	//interval := now.Sub(createTime).Minutes()
	//	//
	// 为了测试，有一些资源是一开始就创建好的，所以，这些资源名称约定以Default或者keep作为开头，

	//	//if strings.HasPrefix(instanceName, keepResource) || strings.HasPrefix(instanceName, defaultResource) {
	//	//	continue
	//	//}
	//	//
	//	//if needProtect == 1 && int64(interval) < 30 {
	//	//	continue
	//	//}
	//
	//	if err = bmcService.DeleteInstance(ctx, instanceId); err != nil {
	//		log.Printf("[ERROR] sweep instance %s error: %s", instanceId, err.Error())
	//	}
	//}
	return err
}

func TestAccZenlayerCloudEipResource_Basic(t *testing.T) {
	id := "zenlayercloud_bmc_eip.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testProviders(),
		CheckDestroy:      testAccCheckEipResourceDestroy,
		Steps: []resource.TestStep{
			// creation Step
			{
				// todo AK
				PreConfig: nil,
				// main.tf
				Config: testAccZenlayerCloudEipBasic,
				Check: resource.ComposeTestCheckFunc(
					// check resource exists
					testAccCheckZenlayerCloudDataResourceID(id),
					// check eip exists
					testAccCheckZenlayerCloudEipExists(id),
					// check attribute
					resource.TestCheckResourceAttr(id, "eip_status", BmcEipStatusAvailable),
					resource.TestCheckResourceAttrSet(id, "public_ip"),
				),
			},
			{
				// Modify Step
				PreConfig: nil,
				// main.tf
				Config: "nil",
				Check: resource.ComposeTestCheckFunc(
					// check eip exists
					testAccCheckZenlayerCloudEipExists(id),
					// check attribute
					resource.TestCheckResourceAttr(id, "instance_status", "RUNNING"),
					resource.TestCheckResourceAttr(id, "xxxx", "xxx"),
					resource.TestCheckResourceAttrSet(id, "xxx"),
				),
			},
		},
	})
}

func testAccCheckZenlayerCloudEipExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx := context.Background()

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("bmc eip %s is not found", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("bmc eip id is not set")
		}

		bmcService := BmcService{
			client: testAccProvider.Meta().(*connectivity.ZenlayerCloudClient),
		}
		eip, err := bmcService.DescribeEipAddressById(ctx, rs.Primary.ID)
		if err != nil {
			err = resource.RetryContext(ctx, readRetryTimeout, func() *resource.RetryError {
				eip, err = bmcService.DescribeEipAddressById(ctx, rs.Primary.ID)
				if err != nil {
					return retryError(ctx, err)
				}
				return nil
			})
		}
		if err != nil {
			return err
		}
		if eip == nil {
			return fmt.Errorf("bmc eip id is not found")
		}
		return nil
	}
}

func testAccCheckEipResourceDestroy(s *terraform.State) error {

	bmcService := BmcService{
		client: testAccProvider.Meta().(*connectivity.ZenlayerCloudClient),
	}

	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "zenlayercloud_bmc_eip" {
			continue
		}

		eip, err := bmcService.DescribeEipAddressById(ctx, rs.Primary.ID)
		if err != nil {
			err = resource.RetryContext(ctx, readRetryTimeout, func() *resource.RetryError {
				eip, err = bmcService.DescribeEipAddressById(ctx, rs.Primary.ID)
				if err != nil {
					return retryError(ctx, err)
				}
				return nil
			})
		}
		if err != nil {
			return err
		}
		if eip != nil && eip.EipStatus != BmcEipStatusRecycle {
			return fmt.Errorf("bmc eip still exists: %s", rs.Primary.ID)
		}
	}
	return nil
}

const testAccZenlayerCloudEipBasic = defaultVariable + `
resource "zenlayercloud_bmc_eip" "foo" {
  availability_zone = var.availability_zone
}
`

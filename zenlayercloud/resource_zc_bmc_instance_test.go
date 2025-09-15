package zenlayercloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"testing"
)

func init() {
	resource.AddTestSweepers("zenlayercloud_bmc_instance", &resource.Sweeper{
		Name: "zenlayercloud_bmc_instance",
		F:    testSweepBmcInstance,
	})
}

func testSweepBmcInstance(region string) error {
	sharedClient, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("getting zenlayercloud client error: %s", err.Error())
	}
	client := sharedClient.(*connectivity.ZenlayerCloudClient)
	bmcService := BmcService{
		client: client,
	}

	_, err = bmcService.DescribeInstancesByFilter(&InstancesFilter{})
	if err != nil {
		return fmt.Errorf("get instance list error: %s", err.Error())
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

func TestAccZenlayerCloudInstanceResource_Basic(t *testing.T) {
	id := "zenlayercloud_bmc_instance.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testProviders(),
		CheckDestroy:      testAccCheckInstanceResourceDestroy,
		Steps: []resource.TestStep{
			// creation Step
			{
				// todo AK
				PreConfig: nil,
				// main.tf
				Config: testAccZenlayerCLoudInstanceBasic,
				Check: resource.ComposeTestCheckFunc(
					// check resource exists
					testAccCheckZenlayerCloudDataResourceID(id),
					// check instance exists
					testAccCheckZenlayerCloudBmcInstanceExists(id),
					// check attribute
					resource.TestCheckResourceAttr(id, "instance_status", "RUNNING"),
					resource.TestCheckResourceAttrSet(id, "instance_name"),
				),
			},
			{
				// Modify Step
				PreConfig: nil,
				// main.tf
				Config: "nil",
				Check: resource.ComposeTestCheckFunc(
					// check instance exists
					testAccCheckZenlayerCloudBmcInstanceExists(id),
					// check attribute
					resource.TestCheckResourceAttr(id, "instance_status", "RUNNING"),
					resource.TestCheckResourceAttr(id, "xxxx", "xxx"),
					resource.TestCheckResourceAttrSet(id, "xxx"),
				),
			},
		},
	})
}

func testAccCheckZenlayerCloudBmcInstanceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx := context.Background()

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("bmc instance %s is not found", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("bmc instance id is not set")
		}

		bmcService := BmcService{
			client: testAccProvider.Meta().(*connectivity.ZenlayerCloudClient),
		}
		instance, err := bmcService.DescribeInstanceById(ctx, rs.Primary.ID)
		if err != nil {
			err = resource.RetryContext(ctx, common.ReadRetryTimeout, func() *resource.RetryError {
				instance, err = bmcService.DescribeInstanceById(ctx, rs.Primary.ID)
				if err != nil {
					return common.RetryError(ctx, err)
				}
				return nil
			})
		}
		if err != nil {
			return err
		}
		if instance == nil {
			return fmt.Errorf("bmc instance id is not found")
		}
		return nil
	}
}
func TestAccZenlayerCloudInstanceResource_WithTrafficPackage(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testProviders(),
		CheckDestroy:      testAccCheckInstanceResourceDestroy,
		Steps: []resource.TestStep{
			{},
		},
	})
}

func TestAccZenlayerCloudInstanceResource_WithSubnet(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testProviders(),
		CheckDestroy:      testAccCheckInstanceResourceDestroy,
		Steps: []resource.TestStep{
			{},
		},
	})
}

func testAccCheckInstanceResourceDestroy(s *terraform.State) error {

	bmcService := BmcService{
		client: testAccProvider.Meta().(*connectivity.ZenlayerCloudClient),
	}

	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "zenlayercloud_bmc_instance" {
			continue
		}

		instance, err := bmcService.DescribeInstanceById(ctx, rs.Primary.ID)
		if err != nil {
			err = resource.RetryContext(ctx, common.ReadRetryTimeout, func() *resource.RetryError {
				instance, err = bmcService.DescribeInstanceById(ctx, rs.Primary.ID)
				if err != nil {
					return common.RetryError(ctx, err)
				}
				return nil
			})
		}
		if err != nil {
			return err
		}
		if instance != nil && instance.InstanceStatus != BmcInstanceStatusRecycle {
			return fmt.Errorf("bmc instance still exists: %s", rs.Primary.ID)
		}
	}
	return nil
}

const testAccZenlayerCLoudInstanceBasic = defaultVariable + `
resource "zenlayercloud_bmc_instance" "foo" {
  hostname             = "abc"
  instance_name        = "Demo-Instance-Create"
  zone_id              = var.zone_id
  image_id             = var.image_id
  instance_type_id     = var.instance_type_id
  instance_charge_type = "POSTPAID"
  count                = var.number
  internet_charge_type = var.internet_charge_type
  password             = var.password
}
`

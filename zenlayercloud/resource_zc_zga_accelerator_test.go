package zenlayercloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	zga "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zga20230706"
)

func TestAccZenlayerCloudZgaAccelerator_Basic(t *testing.T) {
	var (
		v          *zga.AcceleratorInfo
		resourceId = "zenlayercloud_zga_accelerator.default"
		ra         = resourceAttrInit(resourceId, map[string]string{
			"accelerator_status": ZgaAcceleratorStatusAccelerating,
		})
		rc = resourceCheckInitWithDescribeMethod(resourceId, &v, func() interface{} {
			return NewZgaService(testAccProvider.Meta().(*connectivity.ZenlayerCloudClient))
		}, "DescribeAcceleratorById")
		rac           = resourceAttrCheckInit(rc, ra)
		testAccCheck  = rac.resourceAttrMapUpdateSet()
		testAccConfig = resourceTestAccConfigFunc(resourceId, "", func(name string) string {
			return ""
		})
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testProviders(),
		CheckDestroy:      rc.checkResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccConfig(map[string]interface{}{
					"accelerator_name": "accelerator_test",
					"charge_type":      "ByTraffic",
					"domain":           "test.com",
					"relate_domains":   []string{"a.test.com"},
					"origin_region_id": "HK",
					"origin":           []string{"10.10.10.10"},
					"backup_origin":    []string{"10.10.10.14"},
					"accelerate_regions": []map[string]string{
						{"accelerate_region_id": "HK"},
					},
					"l4_listeners": []map[string]interface{}{
						{
							"protocol":        "udp",
							"port_range":      "53/54",
							"back_port_range": "53/54",
						},
						{
							"port":      "80",
							"back_port": "80",
							"protocol":  "tcp",
						},
					},
					"l7_listeners": []map[string]interface{}{
						{
							"port_range":      "8888/8890",
							"back_port_range": "8888/8890",
							"protocol":        "http",
							"back_protocol":   "http",
						},
					},
					"protocol_opts": []map[string]interface{}{
						{
							"websocket": "true",
							"gzip":      "false",
						},
					},
					"access_control": []map[string]interface{}{
						{
							"enable": "true",
							"rules": []map[string]interface{}{
								{
									"listener":  "udp:53/54",
									"directory": "/",
									"policy":    "accept",
									"cidr_ip":   []string{"10.10.10.11/8"},
								},
							},
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZenlayerCloudDataResourceID(resourceId),
					rc.checkResourceExists(),
					testAccCheck(map[string]string{
						"accelerate_regions.#":                      "1",
						"accelerate_regions.0.accelerate_region_id": "HK",
					}),
				),
			},
			{
				// Modify accelerator name
				Config: testAccConfig(map[string]interface{}{
					"accelerator_name": "accelerator_test1",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"accelerator_name": "accelerator_test1",
					}),
				),
			},
			{
				// Modify accelerator domain
				Config: testAccConfig(map[string]interface{}{
					"domain":         "b.test.com",
					"relate_domains": []string{"c.test.com"},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"domain":           "b.test.com",
						"relate_domains.#": "1",
						"relate_domains.0": "c.test.com",
					}),
				),
			},
			{
				// Modify accelerator origin
				Config: testAccConfig(map[string]interface{}{
					"origin":        []string{"10.10.10.11"},
					"backup_origin": []string{"10.10.10.12"},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"origin.#":        "1",
						"origin.0":        "10.10.10.11",
						"backup_origin.#": "1",
						"backup_origin.0": "10.10.10.12",
					}),
				),
			},
			{
				// Modify accelerator accelerate regions
				Config: testAccConfig(map[string]interface{}{
					"accelerate_regions": []map[string]interface{}{
						{
							"accelerate_region_id": "HK",
							"bandwidth":            "10",
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"accelerate_regions.#":                      "1",
						"accelerate_regions.0.accelerate_region_id": "HK",
						"accelerate_regions.0.bandwidth":            "10",
					}),
				),
			},
			{
				// Modify accelerator accelerate listener
				Config: testAccConfig(map[string]interface{}{
					"l4_listeners": []map[string]interface{}{
						{
							"protocol":        "udp",
							"port_range":      "53/54",
							"back_port_range": "53/54",
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"l4_listeners.#": "1",
					}),
				),
			},
			{
				// Modify accelerator protocol_opts
				Config: testAccConfig(map[string]interface{}{
					"protocol_opts": []map[string]interface{}{
						{
							"websocket": "false",
							"gzip":      "true",
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"protocol_opts.#":           "1",
						"protocol_opts.0.websocket": "false",
						"protocol_opts.0.gzip":      "true",
					}),
				),
			},
			{
				// Modify accelerator access control
				Config: testAccConfig(map[string]interface{}{
					"access_control": []map[string]interface{}{
						{
							"enable": "true",
							"rules": []map[string]interface{}{
								{
									"listener":  "http:8888/8890",
									"directory": "/",
									"policy":    "accept",
									"cidr_ip":   []string{"10.10.10.12/8"},
								},
							},
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"access_control.#":                   "1",
						"access_control.0.rules.#":           "1",
						"access_control.0.rules.0.listener":  "http:8888/8890",
						"access_control.0.rules.0.cidr_ip.#": "1",
						"access_control.0.rules.0.cidr_ip.0": "10.10.10.12/8",
					}),
				),
			},
		},
	})
}

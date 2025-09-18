package zec

//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	common2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20240401"
	zec2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"
	"log"
	"math"
	"sync"
)

type ZecService struct {
	client *connectivity.ZenlayerCloudClient
}

func (s *ZecService) DescribeVpcById(ctx context.Context, vpcId string) (*zec.VpcInfo, error) {
	request := zec.NewDescribeVpcsRequest()
	request.VpcIds = []string{vpcId}

	response, err := s.client.WithZecClient().DescribeVpcs(request)
	defer common.LogApiRequest(ctx, "DescribeVpcs", request, response, err)

	if err != nil {
		return nil, err
	} else if len(response.Response.DataSet) == 0 {
		return nil, nil
	}
	return response.Response.DataSet[0], nil
}

func (s *ZecService) DeleteVpc(ctx context.Context, vpcId string) error {
	request := zec.NewDeleteVpcRequest()
	request.VpcId = vpcId
	response, err := s.client.WithZecClient().DeleteVpc(request)
	defer common.LogApiRequest(ctx, "DeleteVpc", request, response, err)
	return err
}

func (s *ZecService) DescribeVpcsByFilter(ctx context.Context, filter *ZecVpcFilter) (vpcs []*zec.VpcInfo, err error) {
	request := convertVpcRequestFilter(filter)

	var limit = 100
	request.PageSize = limit
	request.PageNum = 1
	response, err := s.client.WithZecClient().DescribeVpcs(request)

	if err != nil {
		log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
			request.GetAction(), common.ToJsonString(request), err.Error())
		return
	}
	if response == nil || len(response.Response.DataSet) < 1 {
		return
	}

	vpcs = response.Response.DataSet
	num := int(math.Ceil(float64(response.Response.TotalCount)/float64(limit))) - 1
	if num == 0 {
		return vpcs, nil
	}
	maxConcurrentNum := 50
	g := common.NewGoRoutine(maxConcurrentNum)
	wg := sync.WaitGroup{}

	var vpcList = make([]interface{}, num)

	for i := 0; i < num; i++ {
		wg.Add(1)
		value := i
		goFunc := func() {
			request := convertVpcRequestFilter(filter)

			request.PageNum = value + 2
			request.PageSize = limit

			response, err := s.client.WithZecClient().DescribeVpcs(request)
			if err != nil {
				log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
					request.GetAction(), common.ToJsonString(request), err.Error())
				return
			}
			log.Printf("[DEBUG] Api[%s] success, request body [%s], response body [%s]\n",
				request.GetAction(), common.ToJsonString(request), common.ToJsonString(response))

			vpcList[value] = response.Response.DataSet

			wg.Done()
			log.Printf("[DEBUG] thread %d finished", value)
		}
		g.Run(goFunc)
	}
	wg.Wait()

	log.Printf("[DEBUG] DescribeVpcs request finished")
	for _, v := range vpcList {
		vpcs = append(vpcs, v.([]*zec.VpcInfo)...)
	}
	log.Printf("[DEBUG] transfer global vpc instances finished")
	return
}

func (s *ZecService) ModifyVpcAttribute(ctx context.Context, request *zec.ModifyVpcAttributeRequest) error {

	response, err := s.client.WithZecClient().ModifyVpcAttribute(request)
	common.LogApiRequest(ctx, "ModifyVpcAttribute", request, response, err)
	return err
}

func (s *ZecService) DescribeSubnetById(ctx context.Context, subnetId string) (*zec.SubnetInfo, error) {
	request := zec.NewDescribeSubnetsRequest()
	request.SubnetIds = []string{subnetId}

	response, err := s.client.WithZecClient().DescribeSubnets(request)
	defer common.LogApiRequest(ctx, "DescribeSubnets", request, response, err)

	if err != nil {
		return nil, err
	} else if len(response.Response.DataSet) == 0 {
		return nil, nil
	}
	return response.Response.DataSet[0], nil
}

func (s *ZecService) DeleteSubnet(ctx context.Context, subnetId string) error {
	request := zec.NewDeleteSubnetRequest()
	request.SubnetId = subnetId
	response, err := s.client.WithZecClient().DeleteSubnet(request)
	defer common.LogApiRequest(ctx, "DeleteSubnet", request, response, err)
	return err
}

func (s *ZecService) DescribeDisks(ctx context.Context, filter *ZecDiskFilter) (disks []*zec.DiskInfo, err error) {
	request := convertDiskRequestFilter(filter)

	var limit = 100
	request.PageSize = limit
	request.PageNum = 1
	response, err := s.client.WithZecClient().DescribeDisks(request)

	if err != nil {
		log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
			request.GetAction(), common.ToJsonString(request), err.Error())
		return
	}
	if response == nil || len(response.Response.DataSet) < 1 {
		return
	}

	disks = response.Response.DataSet
	num := int(math.Ceil(float64(response.Response.TotalCount)/float64(limit))) - 1
	if num == 0 {
		return disks, nil
	}
	maxConcurrentNum := 50
	g := common.NewGoRoutine(maxConcurrentNum)
	wg := sync.WaitGroup{}

	var vpcList = make([]interface{}, num)

	for i := 0; i < num; i++ {
		wg.Add(1)
		value := i
		goFunc := func() {
			request := convertDiskRequestFilter(filter)

			request.PageNum = value + 2
			request.PageSize = limit

			response, err := s.client.WithZecClient().DescribeDisks(request)
			if err != nil {
				log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
					request.GetAction(), common.ToJsonString(request), err.Error())
				return
			}
			log.Printf("[DEBUG] Api[%s] success, request body [%s], response body [%s]\n",
				request.GetAction(), common.ToJsonString(request), common.ToJsonString(response))

			vpcList[value] = response.Response.DataSet

			wg.Done()
			log.Printf("[DEBUG] thread %d finished", value)
		}
		g.Run(goFunc)
	}
	wg.Wait()

	log.Printf("[DEBUG] DescribeDisks request finished")
	for _, v := range vpcList {
		disks = append(disks, v.([]*zec.DiskInfo)...)
	}
	log.Printf("[DEBUG] transfer disks finished")
	return
}

func (s *ZecService) DeleteDiskById(ctx context.Context, diskId string) error {
	request := zec.NewReleaseDiskRequest()
	request.DiskId = diskId
	response, err := s.client.WithZecClient().ReleaseDisk(request)
	defer common.LogApiRequest(ctx, "ReleaseDisk", request, response, err)

	if err != nil {
		if sdkError, ok := err.(*common2.ZenlayerCloudSdkError); ok {
			// TODO
			if sdkError.Code == "UNSUPPORTED_OPERATION_DISK_BEING_RECYCLE" {
				return nil
			}
		}
		return err
	}
	return nil
}

func (s *ZecService) DescribeDiskById(ctx context.Context, diskId string) (*zec.DiskInfo, error) {
	request := zec.NewDescribeDisksRequest()
	request.DiskIds = []string{diskId}

	response, err := s.client.WithZecClient().DescribeDisks(request)
	defer common.LogApiRequest(ctx, "DescribeDisks", request, response, err)

	if err != nil {
		return nil, err
	} else if len(response.Response.DataSet) == 0 {
		return nil, nil
	}
	return response.Response.DataSet[0], nil
}

func (s *ZecService) ResizeDisk(ctx context.Context, diskId string, diskSize int) error {
	request := zec.NewResizeDiskRequest()
	request.DiskId = &diskId
	request.DiskSize = &diskSize

	response, err := s.client.WithZecClient().ResizeDisk(request)
	defer common.LogApiRequest(ctx, "ResizeDisk", request, response, err)

	return err
}

func (s *ZecService) DescribeNatGatewayById(ctx context.Context, natGatewayId string) (*zec.NatGateway, error) {
	request := zec.NewDescribeNatGatewaysRequest()
	request.NatGatewayIds = []string{natGatewayId}

	response, err := s.client.WithZecClient().DescribeNatGateways(request)
	defer common.LogApiRequest(ctx, "DescribeNatGateways", request, response, err)

	if err != nil {
		return nil, err
	} else if len(response.Response.DataSet) == 0 {
		return nil, nil
	}
	return response.Response.DataSet[0], nil
}

func (s *ZecService) DescribeNatGateways(ctx context.Context, filter *ZecNatGatewayFilter) (nats []*zec.NatGateway, err error) {
	request := convertNatGatewayRequestFilter(filter)

	var limit = 100
	request.PageSize = common2.Integer(limit)
	request.PageNum = common2.Integer(1)
	response, err := s.client.WithZecClient().DescribeNatGateways(request)

	if err != nil {
		log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
			request.GetAction(), common.ToJsonString(request), err.Error())
		return
	}
	if response == nil || len(response.Response.DataSet) < 1 {
		return
	}

	nats = response.Response.DataSet
	num := int(math.Ceil(float64(*response.Response.TotalCount)/float64(limit))) - 1
	if num == 0 {
		return nats, nil
	}
	maxConcurrentNum := 50
	g := common.NewGoRoutine(maxConcurrentNum)
	wg := sync.WaitGroup{}

	var vpcList = make([]interface{}, num)

	for i := 0; i < num; i++ {
		wg.Add(1)
		value := i
		goFunc := func() {
			request := convertNatGatewayRequestFilter(filter)

			request.PageNum = common2.Integer(value + 2)
			request.PageSize = &limit

			response, err := s.client.WithZecClient().DescribeNatGateways(request)
			if err != nil {
				log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
					request.GetAction(), common.ToJsonString(request), err.Error())
				return
			}
			log.Printf("[DEBUG] Api[%s] success, request body [%s], response body [%s]\n",
				request.GetAction(), common.ToJsonString(request), common.ToJsonString(response))

			vpcList[value] = response.Response.DataSet

			wg.Done()
			log.Printf("[DEBUG] thread %d finished", value)
		}
		g.Run(goFunc)
	}
	wg.Wait()

	log.Printf("[DEBUG] DescribeNatGateways request finished")
	for _, v := range vpcList {
		nats = append(nats, v.([]*zec.NatGateway)...)
	}
	log.Printf("[DEBUG] transfer NAT gateways finished")
	return
}

func (s *ZecService) DescribeInstancesByFilter(filter *ZecInstancesFilter) (instances []*zec2.InstanceInfo, err error) {
	request := convertInstanceRequestFilter(filter)

	var limit = 100
	request.PageSize = &limit
	request.PageNum = common2.Integer(1)
	response, err := s.client.WithZec2Client().DescribeInstances(request)

	if err != nil {
		log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
			request.GetAction(), common.ToJsonString(request), err.Error())
		return
	}
	if response == nil || len(response.Response.DataSet) < 1 {
		return
	}

	instances = response.Response.DataSet
	num := int(math.Ceil(float64(*response.Response.TotalCount)/float64(limit))) - 1
	if num == 0 {
		return instances, nil
	}
	maxConcurrentNum := 50
	g := common.NewGoRoutine(maxConcurrentNum)
	wg := sync.WaitGroup{}

	var instanceList = make([]interface{}, num)

	for i := 0; i < num; i++ {
		wg.Add(1)
		value := i
		goFunc := func() {
			request := convertInstanceRequestFilter(filter)

			request.PageNum = common2.Integer(value + 2)
			request.PageSize = &limit

			response, err := s.client.WithZec2Client().DescribeInstances(request)
			if err != nil {
				log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
					request.GetAction(), common.ToJsonString(request), err.Error())
				return
			}
			log.Printf("[DEBUG] Api[%s] success, request body [%s], response body [%s]\n",
				request.GetAction(), common.ToJsonString(request), common.ToJsonString(response))

			instanceList[value] = response.Response.DataSet

			wg.Done()
			log.Printf("[DEBUG] thread %d finished", value)
		}
		g.Run(goFunc)
	}
	wg.Wait()

	log.Printf("[DEBUG] DescribeInstances request finished")
	for _, v := range instanceList {
		instances = append(instances, v.([]*zec2.InstanceInfo)...)
	}
	log.Printf("[DEBUG] transfer ZEC instances finished")
	return
}

func (s *ZecService) DescribeNics(ctx context.Context, filter *ZecNicFilter) (vnics []*zec2.NicInfo, err error) {

	request := convertVnicRequestFilter(filter)

	var limit = 100
	request.PageSize = common2.Integer(limit)
	request.PageNum = common2.Integer(1)
	response, err := s.client.WithZec2Client().DescribeNetworkInterfaces(request)

	if err != nil {
		log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
			request.GetAction(), common.ToJsonString(request), err.Error())
		return
	}
	if response == nil || len(response.Response.DataSet) < 1 {
		return
	}

	vnics = response.Response.DataSet
	num := int(math.Ceil(float64(*response.Response.TotalCount)/float64(limit))) - 1
	if num == 0 {
		return vnics, nil
	}
	maxConcurrentNum := 50
	g := common.NewGoRoutine(maxConcurrentNum)
	wg := sync.WaitGroup{}

	var vnicList = make([]interface{}, num)

	for i := 0; i < num; i++ {
		wg.Add(1)
		value := i
		goFunc := func() {
			request := convertVnicRequestFilter(filter)

			request.PageNum = common2.Integer(value + 2)
			request.PageSize = common2.Integer(limit)

			response, err := s.client.WithZec2Client().DescribeNetworkInterfaces(request)
			if err != nil {
				log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
					request.GetAction(), common.ToJsonString(request), err.Error())
				return
			}
			log.Printf("[DEBUG] Api[%s] success, request body [%s], response body [%s]\n",
				request.GetAction(), common.ToJsonString(request), common.ToJsonString(response))

			vnicList[value] = response.Response.DataSet

			wg.Done()
			log.Printf("[DEBUG] thread %d finished", value)
		}
		g.Run(goFunc)
	}
	wg.Wait()

	log.Printf("[DEBUG] DescribeNetworkInterfaces request finished")
	for _, v := range vnicList {
		vnics = append(vnics, v.([]*zec2.NicInfo)...)
	}
	log.Printf("[DEBUG] transfer vnics finished")
	return
}

func (s *ZecService) DescribeNicById(ctx context.Context, nicId string) (*zec2.NicInfo, error) {

	request := zec2.NewDescribeNetworkInterfacesRequest()
	request.NicIds = []string{nicId}

	response, err := s.client.WithZec2Client().DescribeNetworkInterfaces(request)
	defer common.LogApiRequest(ctx, "DescribeNetworkInterfaces", request, response, err)

	if err != nil {
		return nil, err
	} else if len(response.Response.DataSet) == 0 {
		return nil, nil
	}
	return response.Response.DataSet[0], nil
}

func (s *ZecService) DiskStateRefreshFunc(ctx context.Context, diskId string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeDiskById(ctx, diskId)
		if err != nil {
			return nil, "", err
		}

		if object == nil {
			// Set this to nil as if we didn't find anything.
			return nil, "", nil
		}
		for _, failState := range failStates {
			if object.DiskStatus == failState {
				return object, object.DiskStatus, common.Error("Failed to reach target status. Last status: %s.", object.DiskStatus)
			}
		}

		return object, object.DiskStatus, nil
	}
}

func (s *ZecService) DescribeBoardGateways(filter *BoarderGatewayFilter) (zbgs []*zec.ZbgInfo, err error) {
	request := convertZbgRequestFilter(filter)

	var limit = 100
	request.PageSize = limit
	request.PageNum = 1
	response, err := s.client.WithZecClient().DescribeBorderGateways(request)

	if err != nil {
		log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
			request.GetAction(), common.ToJsonString(request), err.Error())
		return
	}
	if response == nil || len(response.Response.DataSet) < 1 {
		return
	}

	zbgs = response.Response.DataSet
	num := int(math.Ceil(float64(response.Response.TotalCount)/float64(limit))) - 1
	if num == 0 {
		return zbgs, nil
	}
	maxConcurrentNum := 50
	g := common.NewGoRoutine(maxConcurrentNum)
	wg := sync.WaitGroup{}

	var zbgList = make([]interface{}, num)

	for i := 0; i < num; i++ {
		wg.Add(1)
		value := i
		goFunc := func() {
			request := convertZbgRequestFilter(filter)

			request.PageNum = value + 2
			request.PageSize = limit

			response, err := s.client.WithZecClient().DescribeBorderGateways(request)
			if err != nil {
				log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
					request.GetAction(), common.ToJsonString(request), err.Error())
				return
			}
			log.Printf("[DEBUG] Api[%s] success, request body [%s], response body [%s]\n",
				request.GetAction(), common.ToJsonString(request), common.ToJsonString(response))

			zbgList[value] = response.Response.DataSet

			wg.Done()
			log.Printf("[DEBUG] thread %d finished", value)
		}
		g.Run(goFunc)
	}
	wg.Wait()

	log.Printf("[DEBUG] DescribeBorderGateways request finished")
	for _, v := range zbgList {
		zbgs = append(zbgs, v.([]*zec.ZbgInfo)...)
	}
	log.Printf("[DEBUG] transfer border gateways finished")
	return
}

func (s *ZecService) DescribeBorderGatewayById(ctx context.Context, zbgId string) (*zec.ZbgInfo, error) {
	request := zec.NewDescribeBorderGatewaysRequest()
	request.ZbgIds = []string{zbgId}

	response, err := s.client.WithZecClient().DescribeBorderGateways(request)
	defer common.LogApiRequest(ctx, "DescribeBorderGateways", request, response, err)

	if err != nil {
		return nil, err
	} else if len(response.Response.DataSet) == 0 {
		return nil, nil
	}
	return response.Response.DataSet[0], nil
}

func (s *ZecService) DeleteBorderGateway(ctx context.Context, zbgId string) error {
	request := zec.NewDeleteBorderGatewayRequest()
	request.ZbgId = zbgId
	response, err := s.client.WithZecClient().DeleteBorderGateway(request)
	defer common.LogApiRequest(ctx, "DeleteBorderGateway", request, response, err)
	return err
}

func (s *ZecService) ModifyBorderGateway(ctx context.Context, request *zec.ModifyBorderGatewaysAttributeRequest) error {
	response, err := s.client.WithZecClient().ModifyBorderGatewaysAttribute(request)
	common.LogApiRequest(ctx, "ModifyBorderGatewaysAttribute", request, response, err)
	return err
}

func (s *ZecService) DescribeEipById(ctx context.Context, eipId string) (*zec.EipInfo, error) {
	request := zec.NewDescribeEipsRequest()
	request.EipIds = []string{eipId}

	response, err := s.client.WithZecClient().DescribeEips(request)
	defer common.LogApiRequest(ctx, "DescribeEips", request, response, err)

	if err != nil {
		return nil, err
	} else if len(response.Response.DataSet) == 0 {
		return nil, nil
	}
	return response.Response.DataSet[0], nil
}

func (s *ZecService) DescribeEipsByFilter(ctx context.Context, filter *EipFilter) (eips []*zec.EipInfo, err error) {
	request := convertEipRequestFilter(filter)

	var limit = 100
	request.PageSize = limit
	request.PageNum = 1
	response, err := s.client.WithZecClient().DescribeEips(request)

	if err != nil {
		log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
			request.GetAction(), common.ToJsonString(request), err.Error())
		return
	}
	if response == nil || len(response.Response.DataSet) < 1 {
		return
	}

	eips = response.Response.DataSet
	num := int(math.Ceil(float64(response.Response.TotalCount)/float64(limit))) - 1
	if num == 0 {
		return eips, nil
	}
	maxConcurrentNum := 50
	g := common.NewGoRoutine(maxConcurrentNum)
	wg := sync.WaitGroup{}

	var vpcList = make([]interface{}, num)

	for i := 0; i < num; i++ {
		wg.Add(1)
		value := i
		goFunc := func() {
			request := convertEipRequestFilter(filter)

			request.PageNum = value + 2
			request.PageSize = limit

			response, err := s.client.WithZecClient().DescribeEips(request)
			if err != nil {
				log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
					request.GetAction(), common.ToJsonString(request), err.Error())
				return
			}
			log.Printf("[DEBUG] Api[%s] success, request body [%s], response body [%s]\n",
				request.GetAction(), common.ToJsonString(request), common.ToJsonString(response))

			vpcList[value] = response.Response.DataSet

			wg.Done()
			log.Printf("[DEBUG] thread %d finished", value)
		}
		g.Run(goFunc)
	}
	wg.Wait()

	log.Printf("[DEBUG] DescribeEips request finished")
	for _, v := range vpcList {
		eips = append(eips, v.([]*zec.EipInfo)...)
	}
	log.Printf("[DEBUG] transfer elastic ip instances finished")
	return
}

func (s *ZecService) DeleteInstance(ctx context.Context, instanceId string) error {
	request := zec.NewReleaseInstancesRequest()
	request.InstanceIds = []string{instanceId}
	response, err := s.client.WithZecClient().ReleaseInstances(request)
	defer common.LogApiRequest(ctx, "ReleaseInstances", request, response, err)

	return err
}

func (s *ZecService) DescribeInstanceById(ctx context.Context, instanceId string) (instance *zec2.InstanceInfo, err error) {
	request := zec2.NewDescribeInstancesRequest()
	request.InstanceIds = []string{instanceId}

	response, err := s.client.WithZec2Client().DescribeInstances(request)

	defer common.LogApiRequest(ctx, "DescribeInstances", request, response, err)
	if err != nil {
		return
	}

	if len(response.Response.DataSet) < 1 {
		return
	}
	instance = response.Response.DataSet[0]
	return
}

func (s *ZecService) DescribeSecurityGroupById(ctx context.Context, securityGroupId string) (securityGroup *zec2.SecurityGroupInfo, err error) {

	request := zec2.NewDescribeSecurityGroupsRequest()
	request.SecurityGroupIds = []string{securityGroupId}

	var response *zec2.DescribeSecurityGroupsResponse

	defer common.LogApiRequest(ctx, "DescribeSecurityGroups", request, response, err)

	response, err = s.client.WithZec2Client().DescribeSecurityGroups(request)

	if err != nil {
		return
	}

	if len(response.Response.DataSet) < 1 {
		return
	}
	securityGroup = response.Response.DataSet[0]
	return
}

func (s *ZecService) ModifySubnet(ctx context.Context, subnetId string, name string, cidr string) error {
	request := zec.NewModifySubnetAttributeRequest()
	request.SubnetId = &subnetId
	request.SubnetName = &name
	request.CidrBlock = &cidr

	response, err := s.client.WithZecClient().ModifySubnetAttribute(request)
	common.LogApiRequest(ctx, "ModifySubnetAttribute", request, response, err)
	return err
}

func (s *ZecService) AddSubnetIpv6(ctx context.Context, subnetId string, ipv6Type string) error {
	request := zec.NewModifySubnetStackTypeRequest()
	request.SubnetId = subnetId
	request.StackType = "IPv4_IPv6"
	request.Ipv6Type = ipv6Type

	response, err := s.client.WithZecClient().ModifySubnetStackType(request)
	common.LogApiRequest(ctx, "ModifySubnetStackType", request, response, err)
	return err
}

func (s *ZecService) DescribeSubnetsByFilter(ctx context.Context, filter *SubnetFilter) (subnets []*zec.SubnetInfo, err error) {
	request := convertSubnetRequestFilter(filter)

	var limit = 100
	request.PageSize = limit
	request.PageNum = 1
	response, err := s.client.WithZecClient().DescribeSubnets(request)

	if err != nil {
		log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
			request.GetAction(), common.ToJsonString(request), err.Error())
		return
	}
	if response == nil || len(response.Response.DataSet) < 1 {
		return
	}

	subnets = response.Response.DataSet
	num := int(math.Ceil(float64(response.Response.TotalCount)/float64(limit))) - 1
	if num == 0 {
		return subnets, nil
	}
	maxConcurrentNum := 50
	g := common.NewGoRoutine(maxConcurrentNum)
	wg := sync.WaitGroup{}

	var subnetList = make([]interface{}, num)

	for i := 0; i < num; i++ {
		wg.Add(1)
		value := i
		goFunc := func() {
			request := convertSubnetRequestFilter(filter)

			request.PageNum = value + 2
			request.PageSize = limit

			response, err := s.client.WithZecClient().DescribeSubnets(request)
			if err != nil {
				log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
					request.GetAction(), common.ToJsonString(request), err.Error())
				return
			}
			log.Printf("[DEBUG] Api[%s] success, request body [%s], response body [%s]\n",
				request.GetAction(), common.ToJsonString(request), common.ToJsonString(response))

			subnetList[value] = response.Response.DataSet

			wg.Done()
			log.Printf("[DEBUG] thread %d finished", value)
		}
		g.Run(goFunc)
	}
	wg.Wait()

	log.Printf("[DEBUG] DescribeSubnets request finished")
	for _, v := range subnetList {
		subnets = append(subnets, v.([]*zec.SubnetInfo)...)
	}
	log.Printf("[DEBUG] transfer zec subnets finished")
	return
}

func (s *ZecService) DeleteVnicById(ctx context.Context, nicId string) error {
	request := zec.NewDeleteNetworkInterfaceRequest()
	request.NicId = nicId
	response, err := s.client.WithZecClient().DeleteNetworkInterface(request)
	defer common.LogApiRequest(ctx, "DeleteNetworkInterface", request, response, err)
	return err
}

func (s *ZecService) ModifyVNicAttribute(ctx context.Context, vnicId string, name string, securityGroupId string) error {
	request := zec2.NewModifyNetworkInterfaceAttributeRequest()
	request.NicId = &vnicId
	request.Name = &name
	request.SecurityGroupId = &securityGroupId
	response, err := s.client.WithZec2Client().ModifyNetworkInterfaceAttribute(request)
	common.LogApiRequest(ctx, "ModifyNetworkInterfacesAttribute", request, response, err)
	return err
}

func (s *ZecService) InstanceStateRefreshFunc(ctx context.Context, instanceId string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeInstanceById(ctx, instanceId)
		if err != nil {
			return nil, "", err
		}

		if object == nil {
			// Set this to nil as if we didn't find anything.
			return nil, "", nil
		}
		for _, failState := range failStates {
			if *object.Status == failState {
				return object, *object.Status, common.Error("Failed to reach target status. Last status: %s.", object.Status)
			}
		}

		return object, *object.Status, nil
	}
}

func (s *ZecService) shutdownInstance(ctx context.Context, instanceId string) error {
	request := zec.NewStopInstancesRequest()
	request.InstanceIds = []string{instanceId}
	response, err := s.client.WithZecClient().StopInstances(request)
	common.LogApiRequest(ctx, "ShutdownInstance", request, response, err)
	return err
}

func (s *ZecService) resetInstance(ctx context.Context, request *zec2.ResetInstanceRequest) error {
	response, err := s.client.WithZec2Client().ResetInstance(request)
	common.LogApiRequest(ctx, "ResetInstance", request, response, err)
	return err
}

func (s *ZecService) resetInstancePassword(ctx context.Context, instanceId string, newPassword string) error {
	request := zec.NewResetInstancePasswordRequest()
	request.InstanceId = instanceId
	request.Password = newPassword
	response, err := s.client.WithZecClient().ResetInstancePassword(request)
	defer common.LogApiRequest(ctx, "ResetInstancePassword", request, response, err)
	return err
}

func (s *ZecService) StartInstance(ctx context.Context, instanceId string) error {
	request := zec.NewStartInstancesRequest()
	request.InstanceIds = []string{instanceId}
	response, err := s.client.WithZecClient().StartInstances(request)
	common.LogApiRequest(ctx, "StartInstance", request, response, err)
	return err
}

func (s *ZecService) DescribeImagesByFilter(filter *ImageFilter) (images []*zec.ImageInfo, err error) {
	request := convertImageFilter(filter)
	var limit = 100
	request.PageSize = limit
	request.PageNum = 1
	response, err := s.client.WithZecClient().DescribeImages(request)

	if err != nil {
		log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
			request.GetAction(), common.ToJsonString(request), err.Error())
		return
	}
	if response == nil || len(response.Response.DataSet) < 1 {
		return
	}

	images = response.Response.DataSet
	num := int(math.Ceil(float64(response.Response.TotalCount)/float64(limit))) - 1
	if num == 0 {
		return images, nil
	}
	maxConcurrentNum := 50
	g := common.NewGoRoutine(maxConcurrentNum)
	wg := sync.WaitGroup{}

	var imageSetList = make([]interface{}, num)

	for i := 0; i < num; i++ {
		wg.Add(1)
		value := i
		goFunc := func() {
			request := convertImageFilter(filter)

			request.PageNum = value + 2
			request.PageSize = limit

			response, err := s.client.WithZecClient().DescribeImages(request)
			if err != nil {
				log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
					request.GetAction(), common.ToJsonString(request), err.Error())
				return
			}
			log.Printf("[DEBUG] Api[%s] success, request body [%s], response body [%s]\n",
				request.GetAction(), common.ToJsonString(request), common.ToJsonString(response))

			imageSetList[value] = response.Response.DataSet

			wg.Done()
			log.Printf("[DEBUG] thread %d finished", value)
		}
		g.Run(goFunc)
	}
	wg.Wait()

	log.Printf("[DEBUG] DescribeImages request finished")
	for _, v := range imageSetList {
		images = append(images, v.([]*zec.ImageInfo)...)
	}
	log.Printf("[DEBUG] transfer images finished")
	return
}

func (s *ZecService) ModifySecurityGroupName(ctx context.Context, securityGroupId string, name string) error {
	request := zec.NewModifySecurityGroupsAttributeRequest()
	request.SecurityGroupIds = []string{securityGroupId}
	request.SecurityGroupName = name
	response, err := s.client.WithZecClient().ModifySecurityGroupsAttribute(request)
	defer common.LogApiRequest(ctx, "ModifySecurityGroupsAttribute", request, response, err)
	return err
}

func (s *ZecService) DeleteSecurityGroupById(ctx context.Context, id string) error {
	request := zec.NewDeleteSecurityGroupRequest()
	request.SecurityGroupId = id
	response, err := s.client.WithZecClient().DeleteSecurityGroup(request)
	defer common.LogApiRequest(ctx, "DeleteSecurityGroup", request, response, err)
	return err
}

func (s *ZecService) DeleteVpcRoute(ctx context.Context, routeId string) error {
	request := zec.NewDeleteRouteRequest()
	request.RouteId = routeId
	response, err := s.client.WithZecClient().DeleteRoute(request)
	defer common.LogApiRequest(ctx, "DeleteRoute", request, response, err)
	return err
}

func (s *ZecService) DescribeVpcRouteById(ctx context.Context, routeId string) (*zec.RouteInfo, error) {
	request := zec.NewDescribeRoutesRequest()
	request.RouteIds = []string{routeId}
	response, err := s.client.WithZecClient().DescribeRoutes(request)
	defer common.LogApiRequest(ctx, "DescribeRoutes", request, response, err)
	if err != nil {
		return nil, err
	} else if len(response.Response.DataSet) == 0 {
		return nil, nil
	}
	return response.Response.DataSet[0], nil
}

func (s *ZecService) ModifyRouteAttribute(ctx context.Context, routeId string, name string) error {
	request := zec.NewModifyRouteAttributeRequest()
	request.RouteId = &routeId
	request.Name = &name
	response, err := s.client.WithZecClient().ModifyRouteAttribute(request)
	defer common.LogApiRequest(ctx, "ModifyRouteAttribute", request, response, err)
	return err
}

func (s *ZecService) DescribeSnapshots(ctx context.Context, filter *ZecSnapshotFilter) (snapshots []*zec.SnapshotInfo, err error) {
	request := convertSnapshotFilter(filter)
	var limit = 100
	request.PageSize = &limit
	request.PageNum = common2.Integer(1)
	response, err := s.client.WithZecClient().DescribeSnapshots(request)

	if err != nil {
		log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
			request.GetAction(), common.ToJsonString(request), err.Error())
		return
	}
	if response == nil || len(response.Response.DataSet) < 1 {
		return
	}

	snapshots = response.Response.DataSet
	num := int(math.Ceil(float64(*response.Response.TotalCount)/float64(limit))) - 1
	if num == 0 {
		return snapshots, nil
	}
	maxConcurrentNum := 50
	g := common.NewGoRoutine(maxConcurrentNum)
	wg := sync.WaitGroup{}

	var imageSetList = make([]interface{}, num)

	for i := 0; i < num; i++ {
		wg.Add(1)
		value := i
		goFunc := func() {
			request := convertSnapshotFilter(filter)

			request.PageNum = common2.Integer(value + 2)
			request.PageSize = &limit

			response, err := s.client.WithZecClient().DescribeSnapshots(request)
			if err != nil {
				log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
					request.GetAction(), common.ToJsonString(request), err.Error())
				return
			}
			log.Printf("[DEBUG] Api[%s] success, request body [%s], response body [%s]\n",
				request.GetAction(), common.ToJsonString(request), common.ToJsonString(response))

			imageSetList[value] = response.Response.DataSet

			wg.Done()
			log.Printf("[DEBUG] thread %d finished", value)
		}
		g.Run(goFunc)
	}
	wg.Wait()

	log.Printf("[DEBUG] DescribeSnapshots request finished")
	for _, v := range imageSetList {
		snapshots = append(snapshots, v.([]*zec.SnapshotInfo)...)
	}
	log.Printf("[DEBUG] transfer snapshots finished")
	return
}

func (s *ZecService) DescribeSnapshotById(ctx context.Context, id string) (*zec.SnapshotInfo, error) {
	request := zec.NewDescribeSnapshotsRequest()
	request.SnapshotIds = []string{id}

	response, err := s.client.WithZecClient().DescribeSnapshots(request)
	defer common.LogApiRequest(ctx, "DescribeSnapshots", request, response, err)

	if err != nil {
		return nil, err
	} else if len(response.Response.DataSet) == 0 {
		return nil, nil
	}
	return response.Response.DataSet[0], nil
}

func (s *ZecService) DeleteSnapshot(ctx context.Context, snapshotId string) error {
	request := zec.NewDeleteSnapshotsRequest()
	request.SnapshotIds = []string{snapshotId}
	response, err := s.client.WithZecClient().DeleteSnapshots(request)
	defer common.LogApiRequest(ctx, "DeleteSnapshots", request, response, err)

	if err != nil {
		if sdkError, ok := err.(*common2.ZenlayerCloudSdkError); ok {
			if sdkError.Code == "INVALID_DISK_SNAPSHOT_NOT_FOUND" {
				return nil
			}
		}
		return err
	}
	return nil
}

func (s *ZecService) SnapshotStateRefreshFunc(ctx context.Context, snapshotId string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeSnapshotById(ctx, snapshotId)
		if err != nil {
			return nil, "", err
		}

		if object == nil {
			// Set this to nil as if we didn't find anything.
			return nil, "", nil
		}
		for _, failState := range failStates {
			if *object.Status == failState {
				return object, *object.Status, common.Error("Failed to reach target status. Last status: %s.", object.Status)
			}
		}

		return object, *object.Status, nil
	}
}

func (s *ZecService) DescribeAutoSnapshotPolicies(ctx context.Context, filter *ZecAutoSnapshotPolicyFilter) (policies []*zec.AutoSnapshotPolicy, err error) {

	request := convertSnapshotPolicyFilter(filter)
	var limit = 100
	request.PageSize = &limit
	request.PageNum = common2.Integer(1)
	response, err := s.client.WithZecClient().DescribeAutoSnapshotPolicies(request)

	if err != nil {
		log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
			request.GetAction(), common.ToJsonString(request), err.Error())
		return
	}
	if response == nil || len(response.Response.DataSet) < 1 {
		return
	}

	policies = response.Response.DataSet
	num := int(math.Ceil(float64(*response.Response.TotalCount)/float64(limit))) - 1
	if num == 0 {
		return policies, nil
	}
	maxConcurrentNum := 50
	g := common.NewGoRoutine(maxConcurrentNum)
	wg := sync.WaitGroup{}

	var imageSetList = make([]interface{}, num)

	for i := 0; i < num; i++ {
		wg.Add(1)
		value := i
		goFunc := func() {
			request := convertSnapshotPolicyFilter(filter)

			request.PageNum = common2.Integer(value + 2)
			request.PageSize = &limit

			response, err := s.client.WithZecClient().DescribeAutoSnapshotPolicies(request)
			if err != nil {
				log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
					request.GetAction(), common.ToJsonString(request), err.Error())
				return
			}
			log.Printf("[DEBUG] Api[%s] success, request body [%s], response body [%s]\n",
				request.GetAction(), common.ToJsonString(request), common.ToJsonString(response))

			imageSetList[value] = response.Response.DataSet

			wg.Done()
			log.Printf("[DEBUG] thread %d finished", value)
		}
		g.Run(goFunc)
	}
	wg.Wait()

	log.Printf("[DEBUG] DescribeAutoSnapshotPolicies request finished")
	for _, v := range imageSetList {
		policies = append(policies, v.([]*zec.AutoSnapshotPolicy)...)
	}
	log.Printf("[DEBUG] transfer snapshot policies finished")
	return
}

func (s *ZecService) DeleteSnapshotPolicy(ctx context.Context, autoSnapshotPolicyId string) error {
	request := zec.NewDeleteAutoSnapshotPoliciesRequest()
	request.AutoSnapshotPolicyIds = []string{autoSnapshotPolicyId}
	response, err := s.client.WithZecClient().DeleteAutoSnapshotPolicies(request)
	defer common.LogApiRequest(ctx, "DeleteAutoSnapshotPolicies", request, response, err)
	return err
}

func (s *ZecService) DescribeSnapshotPolicyById(ctx context.Context, autoSnapshotPolicyId string) (*zec.AutoSnapshotPolicy, error) {

	request := zec.NewDescribeAutoSnapshotPoliciesRequest()
	request.AutoSnapshotPolicyIds = []string{autoSnapshotPolicyId}

	response, err := s.client.WithZecClient().DescribeAutoSnapshotPolicies(request)
	defer common.LogApiRequest(ctx, "DescribeSnapshotPolicies", request, response, err)

	if err != nil {
		return nil, err
	} else if len(response.Response.DataSet) == 0 {
		return nil, nil
	}
	return response.Response.DataSet[0], nil
}

func (s *ZecService) switchInstanceIpForwarding(ctx context.Context, id string, ipForwarding bool) error {
	if ipForwarding {
		request := zec.NewStartIpForwardRequest()
		request.InstanceId = id
		response, err := s.client.WithZecClient().StartIpForward(request)
		defer common.LogApiRequest(ctx, "StartIpForward", request, response, err)
		return err
	} else {
		request := zec.NewStopIpForwardRequest()
		request.InstanceId = id
		response, err := s.client.WithZecClient().StopIpForward(request)
		defer common.LogApiRequest(ctx, "StartIpForward", request, response, err)
		return err
	}
}

func (s *ZecService) DescribeSecurityGroupRules(ctx context.Context, securityGroupId string) ([]*zec2.SecurityGroupRuleInfo, []*zec2.SecurityGroupRuleInfo, error) {
	request := zec2.NewDescribeSecurityGroupRuleRequest()
	request.SecurityGroupId = &securityGroupId
	response, err := s.client.WithZec2Client().DescribeSecurityGroupRule(request)
	defer common.LogApiRequest(ctx, "DescribeSecurityGroupRule", request, response, err)
	return response.Response.IngressRuleList, response.Response.EgressRuleList, err
}

func (s *ZecService) DescribeCidrById(ctx context.Context, cidrId string) (*zec2.CidrInfo, error) {
	request := zec2.NewDescribeCidrsRequest()
	request.CidrIds = []string{cidrId}

	response, err := s.client.WithZec2Client().DescribeCidrs(request)
	defer common.LogApiRequest(ctx, "DescribeCidrs", request, response, err)

	if err != nil {
		return nil, err
	} else if len(response.Response.DataSet) == 0 {
		return nil, nil
	}
	return response.Response.DataSet[0], nil
}

func (s *ZecService) DescribeCidrsByFilter(ctx context.Context, filter *CidrFilter) (cidrs []*zec2.CidrInfo, err error) {
	request := convertCidrRequestFilter(filter)

	var limit = 100
	request.PageSize = &limit
	request.PageNum = common2.Integer(1)
	response, err := s.client.WithZec2Client().DescribeCidrs(request)

	if err != nil {
		log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
			request.GetAction(), common.ToJsonString(request), err.Error())
		return
	}
	if response == nil || len(response.Response.DataSet) < 1 {
		return
	}

	cidrs = response.Response.DataSet
	num := int(math.Ceil(float64(*response.Response.TotalCount)/float64(limit))) - 1
	if num == 0 {
		return cidrs, nil
	}
	maxConcurrentNum := 50
	g := common.NewGoRoutine(maxConcurrentNum)
	wg := sync.WaitGroup{}

	var vpcList = make([]interface{}, num)

	for i := 0; i < num; i++ {
		wg.Add(1)
		value := i
		goFunc := func() {
			request := convertCidrRequestFilter(filter)

			request.PageNum = common2.Integer(value + 2)
			request.PageSize = common2.Integer(limit)

			response, err := s.client.WithZec2Client().DescribeCidrs(request)
			if err != nil {
				log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
					request.GetAction(), common.ToJsonString(request), err.Error())
				return
			}
			log.Printf("[DEBUG] Api[%s] success, request body [%s], response body [%s]\n",
				request.GetAction(), common.ToJsonString(request), common.ToJsonString(response))

			vpcList[value] = response.Response.DataSet

			wg.Done()
			log.Printf("[DEBUG] thread %d finished", value)
		}
		g.Run(goFunc)
	}
	wg.Wait()

	log.Printf("[DEBUG] DescribeCidrs request finished")
	for _, v := range vpcList {
		cidrs = append(cidrs, v.([]*zec2.CidrInfo)...)
	}
	log.Printf("[DEBUG] transfer CIDR block instances finished")
	return

}

func convertCidrRequestFilter(filter *CidrFilter) *zec2.DescribeCidrsRequest {
	request := zec2.NewDescribeCidrsRequest()
	if len(filter.Ids) > 0 {
		request.CidrIds = filter.Ids
	}
	if filter.RegionId != "" {
		request.RegionId = &filter.RegionId
	}

	if filter.ResourceGroupId != "" {
		request.ResourceGroupId = &filter.ResourceGroupId
	}
	return request
}

func convertSnapshotPolicyFilter(filter *ZecAutoSnapshotPolicyFilter) *zec.DescribeAutoSnapshotPoliciesRequest {

	request := zec.NewDescribeAutoSnapshotPoliciesRequest()
	if len(filter.AutoSnapshotPolicyIds) > 0 {
		request.AutoSnapshotPolicyIds = filter.AutoSnapshotPolicyIds
	}
	if filter.ZoneId != "" {
		request.ZoneIds = []string{filter.ZoneId}
	}
	if filter.ResourceGroupId != "" {
		request.ResourceGroupId = &filter.ResourceGroupId
	}
	return request
}

func convertSnapshotFilter(filter *ZecSnapshotFilter) *zec.DescribeSnapshotsRequest {
	request := zec.NewDescribeSnapshotsRequest()
	if len(filter.SnapshotIds) > 0 {
		request.SnapshotIds = filter.SnapshotIds
	}
	if filter.SnapshotName != "" {
		request.SnapshotName = &filter.SnapshotName
	}
	if filter.ZoneId != "" {
		request.ZoneId = &filter.ZoneId
	}
	if filter.ResourceGroupId != "" {
		request.ResourceGroupId = &filter.ResourceGroupId
	}
	if filter.SnapshotType != "" {
		request.SnapshotType = &filter.SnapshotType
	}
	if filter.DiskIds != nil && len(filter.DiskIds) > 0 {
		request.DiskIds = filter.DiskIds
	}

	return request
}

func convertImageFilter(filter *ImageFilter) *zec.DescribeImagesRequest {
	request := zec.NewDescribeImagesRequest()
	request.ImageIds = filter.imageIds
	request.ImageType = filter.imageType
	request.Category = filter.category
	request.OsType = filter.osType
	request.ZoneId = filter.zoneId
	return request

}

func convertSubnetRequestFilter(filter *SubnetFilter) *zec.DescribeSubnetsRequest {
	request := zec.NewDescribeSubnetsRequest()
	request.SubnetIds = filter.ids
	request.RegionId = filter.RegionId
	return request
}

func convertEipRequestFilter(filter *EipFilter) *zec.DescribeEipsRequest {
	request := zec.NewDescribeEipsRequest()
	request.EipIds = filter.Ids
	request.RegionId = filter.RegionId
	request.IpAddresses = filter.IpAddress
	request.Status = filter.Status
	request.ResourceGroupId = filter.ResourceGroupId
	request.PrivateIpAddress = &filter.PrivateIpAddress
	request.CidrIds = filter.CidrIds
	request.AssociatedId = &filter.AssociatedId
	return request
}

func convertZbgRequestFilter(filter *BoarderGatewayFilter) *zec.DescribeBorderGatewaysRequest {
	request := zec.NewDescribeBorderGatewaysRequest()
	request.ZbgIds = filter.Ids
	request.RegionId = filter.RegionId
	request.VpcId = filter.VpcId
	return request
}

func convertVnicRequestFilter(filter *ZecNicFilter) *zec2.DescribeNetworkInterfacesRequest {
	request := zec2.NewDescribeNetworkInterfacesRequest()
	if filter.SubnetId != "" {
		request.SubnetId = common2.String(filter.SubnetId)
	}
	if filter.VpcId != "" {
		request.VpcId = common2.String(filter.VpcId)
	}
	if filter.VpcId != "" {
		request.VpcId = common2.String(filter.VpcId)
	}
	request.NicIds = filter.ids

	if filter.RegionId != "" {
		request.RegionId = common2.String(filter.RegionId)
	}
	if filter.ResourceGroupId != "" {
		request.ResourceGroupId = common2.String(filter.ResourceGroupId)
	}

	if filter.InstanceId != "" {
		request.InstanceId = common2.String(filter.InstanceId)
	}
	return request
}

func convertInstanceRequestFilter(filter *ZecInstancesFilter) *zec2.DescribeInstancesRequest {
	request := zec2.NewDescribeInstancesRequest()
	request.InstanceIds = filter.InstancesIds
	if filter.InstanceName != "" {
		request.Name = common2.String(filter.InstanceName)
	}
	if filter.InstanceStatus != "" {
		request.Status = common2.String(filter.InstanceStatus)
	}
	if filter.Ipv6 != "" {
		request.Ipv6Address = common2.String(filter.Ipv6)
	}
	if filter.Ipv4 != "" {
		request.Ipv4Address = common2.String(filter.Ipv4)
	}
	if filter.ResourceGroupId != "" {
		request.ResourceGroupId = common2.String(filter.ResourceGroupId)
	}
	if filter.ZoneId != "" {
		request.ZoneId = common2.String(filter.ZoneId)
	}
	if filter.ImageId != "" {
		request.ImageId = common2.String(filter.ImageId)
	}
	return request
}

func convertNatGatewayRequestFilter(filter *ZecNatGatewayFilter) *zec.DescribeNatGatewaysRequest {
	request := zec.NewDescribeNatGatewaysRequest()
	request.RegionId = common2.String(filter.RegionId)
	request.ResourceGroupId = common2.String(filter.ResourceGroupId)
	if len(filter.Ids) > 0 {
		request.NatGatewayIds = filter.Ids
	}
	request.Name = common2.String(filter.Name)
	return request
}

func convertDiskRequestFilter(filter *ZecDiskFilter) *zec.DescribeDisksRequest {

	request := zec.NewDescribeDisksRequest()
	if len(filter.Ids) > 0 {
		request.DiskIds = filter.Ids
	}
	if filter.Status != "" {
		request.DiskStatus = filter.Status
	}
	if filter.DiskName != "" {
		request.DiskName = filter.DiskName
	}
	if filter.DiskType != "" {
		request.DiskType = filter.DiskType
	}
	if filter.InstanceId != "" {
		request.InstanceId = filter.InstanceId
	}
	if filter.ZoneId != "" {
		request.ZoneId = filter.ZoneId
	}

	return request
}

func convertVpcRequestFilter(filter *ZecVpcFilter) *zec.DescribeVpcsRequest {
	request := zec.NewDescribeVpcsRequest()

	if filter.VpcIds != nil {
		request.VpcIds = filter.VpcIds
	}

	if filter.CidrBlock != nil {
		request.CidrBlock = *filter.CidrBlock
	}
	if filter.ResourceGroupId != nil {
		request.ResourceGroupId = *filter.ResourceGroupId
	}
	return request
}

package zenlayercloud

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
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	bmc "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/bmc20221120"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	"log"
	"math"
	"sync"
)

type BmcService struct {
	client *connectivity.ZenlayerCloudClient
}

func (s *BmcService) DeleteInstance(ctx context.Context, instanceId string) (err error) {
	// 判断已经终止 OPERATION_DENIED_INSTANCE_RECYCLED,
	request := bmc.NewTerminateInstanceRequest()
	request.InstanceId = instanceId
	response, err := s.client.WithBmcClient().TerminateInstance(request)
	defer common2.LogApiRequest(ctx, "TerminateInstance", request, response, err)

	if err != nil {
		if sdkError, ok := err.(*common.ZenlayerCloudSdkError); ok {
			if sdkError.Code == "OPERATION_DENIED_INSTANCE_RECYCLED" {
				return nil
			}
		}
		return
	}
	return
}

func (s *BmcService) DestroyInstance(ctx context.Context, instanceId string) (err error) {
	request := bmc.NewReleaseInstancesRequest()
	request.InstanceIds = []string{instanceId}
	response, err := s.client.WithBmcClient().ReleaseInstances(request)
	defer common2.LogApiRequest(ctx, "ReleaseInstances", request, response, err)
	return
}

func (s *BmcService) DescribeInstanceById(ctx context.Context, instanceId string) (instance *bmc.InstanceInfo, err error) {
	request := bmc.NewDescribeInstancesRequest()
	request.InstanceIds = []string{instanceId}

	response, err := s.client.WithBmcClient().DescribeInstances(request)

	defer common2.LogApiRequest(ctx, "DescribeInstances", request, response, err)
	if err != nil {
		return
	}

	if len(response.Response.DataSet) < 1 {
		return
	}
	instance = response.Response.DataSet[0]
	return
}

func (s *BmcService) InstanceStateRefreshFunc(ctx context.Context, instanceId string, failStates []string) resource.StateRefreshFunc {
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
			if object.InstanceStatus == failState {
				return object, object.InstanceStatus, common2.Error("Failed to reach target status. Last status: %s.", object.InstanceStatus)
			}
		}

		return object, object.InstanceStatus, nil
	}
}

func (s *BmcService) ModifyInstanceName(ctx context.Context, instanceId string, instanceName string) error {
	request := bmc.NewModifyInstancesAttributeRequest()
	request.InstanceIds = []string{instanceId}
	request.InstanceName = instanceName
	response, err := s.client.WithBmcClient().ModifyInstancesAttribute(request)
	defer common2.LogApiRequest(ctx, "ModifyInstancesAttribute", request, response, err)
	return err
}

func (s *BmcService) DescribeInstancesByFilter(instanceFilter *InstancesFilter) (instances []*bmc.InstanceInfo, err error) {
	request := convertRequestForInstanceFilter(instanceFilter)
	var limit = 100
	request.PageSize = limit
	request.PageNum = 1
	response, err := s.client.WithBmcClient().DescribeInstances(request)

	if err != nil {
		log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
			request.GetAction(), common2.ToJsonString(request), err.Error())
		return
	}
	if response == nil || len(response.Response.DataSet) < 1 {
		return
	}

	//total := response.Response.TotalCount
	instances = response.Response.DataSet
	num := int(math.Ceil(float64(response.Response.TotalCount)/float64(limit))) - 1
	if num == 0 {
		return instances, nil
	}
	maxConcurrentNum := 50
	g := common2.NewGoRoutine(maxConcurrentNum)
	wg := sync.WaitGroup{}

	var instanceSetList = make([]interface{}, num)

	for i := 0; i < num; i++ {
		wg.Add(1)
		value := i
		goFunc := func() {
			request := convertRequestForInstanceFilter(instanceFilter)

			request.PageNum = value + 2
			request.PageSize = limit

			response, err := s.client.WithBmcClient().DescribeInstances(request)
			if err != nil {
				log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
					request.GetAction(), common2.ToJsonString(request), err.Error())
				return
			}
			log.Printf("[DEBUG] Api[%s] success, request body [%s], response body [%s]\n",
				request.GetAction(), common2.ToJsonString(request), common2.ToJsonString(response))

			instanceSetList[value] = response.Response.DataSet

			wg.Done()
			log.Printf("[DEBUG] thread %d finished", value)
		}
		g.Run(goFunc)
	}
	wg.Wait()

	log.Printf("[DEBUG] DescribeInstance request finished")
	for _, v := range instanceSetList {
		instances = append(instances, v.([]*bmc.InstanceInfo)...)
	}
	log.Printf("[DEBUG] transfer Instance finished")
	return
}

func (s *BmcService) reinstallInstance(ctx context.Context, request *bmc.ReinstallInstanceRequest) error {

	response, err := s.client.WithBmcClient().ReinstallInstance(request)
	common2.LogApiRequest(ctx, "ReinstallInstance", request, response, err)
	return err
}

func (s *BmcService) DescribeEipAddressesByFilter(filter *EipFilter) (eipAddresses []*bmc.EipAddress, err error) {

	request := convertEipsFilter(filter)
	var limit = 100
	request.PageSize = limit
	request.PageNum = 1
	response, err := s.client.WithBmcClient().DescribeEipAddresses(request)

	if err != nil {
		log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
			request.GetAction(), common2.ToJsonString(request), err.Error())
		return
	}
	if response == nil || len(response.Response.DataSet) < 1 {
		return
	}

	eipAddresses = response.Response.DataSet
	num := int(math.Ceil(float64(response.Response.TotalCount)/float64(limit))) - 1
	if num == 0 {
		return eipAddresses, nil
	}
	maxConcurrentNum := 50
	g := common2.NewGoRoutine(maxConcurrentNum)
	wg := sync.WaitGroup{}

	var eipAddressSetList = make([]interface{}, num)

	for i := 0; i < num; i++ {
		wg.Add(1)
		value := i
		goFunc := func() {
			request := convertEipsFilter(filter)

			request.PageNum = value + 2
			request.PageSize = limit

			response, err := s.client.WithBmcClient().DescribeEipAddresses(request)
			if err != nil {
				log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
					request.GetAction(), common2.ToJsonString(request), err.Error())
				return
			}
			log.Printf("[DEBUG] Api[%s] success, request body [%s], response body [%s]\n",
				request.GetAction(), common2.ToJsonString(request), common2.ToJsonString(response))

			eipAddressSetList[value] = response.Response.DataSet

			wg.Done()
			log.Printf("[DEBUG] thread %d finished", value)
		}
		g.Run(goFunc)
	}
	wg.Wait()

	log.Printf("[DEBUG] DescribeEipAddresses request finished")
	for _, v := range eipAddressSetList {
		eipAddresses = append(eipAddresses, v.([]*bmc.EipAddress)...)
	}
	log.Printf("[DEBUG] transfer eip addresses finished")
	return
}

func (s *BmcService) DescribeDdosIpAddressesByFilter(filter *DDosIpFilter) (ddosIpAddress []*bmc.DdosIpAddress, err error) {
	request := convertDdosIpFilter(filter)
	var limit = 100
	request.PageSize = limit
	request.PageNum = 1
	response, err := s.client.WithBmcClient().DescribeDdosIpAddresses(request)

	if err != nil {
		log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
			request.GetAction(), common2.ToJsonString(request), err.Error())
		return
	}
	if response == nil || len(response.Response.DataSet) < 1 {
		return
	}

	ddosIpAddress = response.Response.DataSet
	num := int(math.Ceil(float64(response.Response.TotalCount)/float64(limit))) - 1
	if num == 0 {
		return ddosIpAddress, nil
	}
	maxConcurrentNum := 50
	g := common2.NewGoRoutine(maxConcurrentNum)
	wg := sync.WaitGroup{}

	var eipAddressSetList = make([]interface{}, num)

	for i := 0; i < num; i++ {
		wg.Add(1)
		value := i
		goFunc := func() {
			request := convertDdosIpFilter(filter)

			request.PageNum = value + 2
			request.PageSize = limit

			response, err := s.client.WithBmcClient().DescribeDdosIpAddresses(request)
			if err != nil {
				log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
					request.GetAction(), common2.ToJsonString(request), err.Error())
				return
			}
			log.Printf("[DEBUG] Api[%s] success, request body [%s], response body [%s]\n",
				request.GetAction(), common2.ToJsonString(request), common2.ToJsonString(response))

			eipAddressSetList[value] = response.Response.DataSet

			wg.Done()
			log.Printf("[DEBUG] thread %d finished", value)
		}
		g.Run(goFunc)
	}
	wg.Wait()

	log.Printf("[DEBUG] DescribeEipAddresses request finished")
	for _, v := range eipAddressSetList {
		ddosIpAddress = append(ddosIpAddress, v.([]*bmc.DdosIpAddress)...)
	}
	log.Printf("[DEBUG] transfer DDoS addresses finished")
	return
}

func (s *BmcService) InstanceDdosIpStateRefreshFunc(ctx context.Context, eipId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		eip, err := s.DescribeDdosIpAddressById(ctx, eipId)
		if err != nil {
			return nil, "", err
		}

		if eip == nil {
			// Set this to nil as if we didn't find anything.
			return nil, "", nil
		}

		return eip, eip.DdosIpStatus, nil
	}
}

func (s *BmcService) DescribeSubnetById(ctx context.Context, subnetId string) (subnet *bmc.Subnet, err error) {
	request := bmc.NewDescribeSubnetsRequest()
	request.SubnetIds = []string{subnetId}

	var response *bmc.DescribeSubnetsResponse

	defer common2.LogApiRequest(ctx, "DescribeSubnets", request, response, err)

	response, err = s.client.WithBmcClient().DescribeSubnets(request)

	if err != nil {
		return
	}

	if len(response.Response.DataSet) < 1 {
		return
	}
	subnet = response.Response.DataSet[0]
	return
}

func (s *BmcService) SubnetStateRefreshFunc(ctx context.Context, subnetId string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeSubnetById(ctx, subnetId)
		if err != nil {
			return nil, "", err
		}

		if object == nil {
			// Set this to nil as if we didn't find anything.
			return nil, "", nil
		}
		for _, failState := range failStates {
			if object.SubnetStatus == failState {
				return object, object.SubnetStatus, common2.Error("Failed to reach target status. Last status: %s.", object.SubnetStatus)
			}
		}

		return object, object.SubnetStatus, nil
	}
}

func (s *BmcService) DeleteSubnet(ctx context.Context, subnetId string) (err error) {
	request := bmc.NewDeleteSubnetRequest()
	request.SubnetId = subnetId
	response, err := s.client.WithBmcClient().DeleteSubnet(request)
	defer common2.LogApiRequest(ctx, "DeleteSubnet", request, response, err)
	return
}

func (s *BmcService) ModifySubnetName(ctx context.Context, subnetId string, subnetName string) error {
	request := bmc.NewModifySubnetsAttributeRequest()
	request.SubnetIds = []string{subnetId}
	request.SubnetName = subnetName
	response, err := s.client.WithBmcClient().ModifySubnetsAttribute(request)
	defer common2.LogApiRequest(ctx, "ModifySubnetsAttribute", request, response, err)
	return err
}

func (s *BmcService) ModifySubnetResourceGroupById(ctx context.Context, subnetId string, resourceGroupId string) error {
	request := bmc.NewModifySubnetsResourceGroupRequest()
	request.SubnetIds = []string{subnetId}
	request.ResourceGroupId = resourceGroupId
	response, err := s.client.WithBmcClient().ModifySubnetsResourceGroup(request)
	common2.LogApiRequest(ctx, "ModifySubnetsResourceGroup", request, response, err)
	return err
}

func (s *BmcService) DescribeSubnets(ctx context.Context, filter *SubnetFilter) (subnets []*bmc.Subnet, err error) {
	request := convertSubnetFilterRequest(filter)

	var limit = 100
	request.PageSize = limit
	request.PageNum = 1
	response, err := s.client.WithBmcClient().DescribeSubnets(request)

	if err != nil {
		log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
			request.GetAction(), common2.ToJsonString(request), err.Error())
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
	g := common2.NewGoRoutine(maxConcurrentNum)
	wg := sync.WaitGroup{}

	var subnetList = make([]interface{}, num)

	for i := 0; i < num; i++ {
		wg.Add(1)
		value := i
		goFunc := func() {
			request := convertSubnetFilterRequest(filter)

			request.PageNum = value + 2
			request.PageSize = limit

			response, err := s.client.WithBmcClient().DescribeSubnets(request)
			if err != nil {
				log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
					request.GetAction(), common2.ToJsonString(request), err.Error())
				return
			}
			log.Printf("[DEBUG] Api[%s] success, request body [%s], response body [%s]\n",
				request.GetAction(), common2.ToJsonString(request), common2.ToJsonString(response))

			subnetList[value] = response.Response.DataSet

			wg.Done()
			log.Printf("[DEBUG] thread %d finished", value)
		}
		g.Run(goFunc)
	}
	wg.Wait()

	log.Printf("[DEBUG] DescribeEipAddresses request finished")
	for _, v := range subnetList {
		subnets = append(subnets, v.([]*bmc.Subnet)...)
	}
	log.Printf("[DEBUG] transfer Subnet finished")
	return
}

func (s *BmcService) DescribeVpcsByFilter(ctx context.Context, filter *VpcFilter) (vpcs []*bmc.VpcInfo, err error) {
	request := convertVpcFilter(filter)
	var limit = 100
	request.PageSize = limit
	request.PageNum = 1
	response, err := s.client.WithBmcClient().DescribeVpcs(request)

	if err != nil {
		log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
			request.GetAction(), common2.ToJsonString(request), err.Error())
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
	g := common2.NewGoRoutine(maxConcurrentNum)
	wg := sync.WaitGroup{}

	var vpcList = make([]interface{}, num)

	for i := 0; i < num; i++ {
		wg.Add(1)
		value := i
		goFunc := func() {
			request := convertVpcFilter(filter)

			request.PageNum = value + 2
			request.PageSize = limit

			response, err := s.client.WithBmcClient().DescribeVpcs(request)
			if err != nil {
				log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
					request.GetAction(), common2.ToJsonString(request), err.Error())
				return
			}
			log.Printf("[DEBUG] Api[%s] success, request body [%s], response body [%s]\n",
				request.GetAction(), common2.ToJsonString(request), common2.ToJsonString(response))

			vpcList[value] = response.Response.DataSet

			wg.Done()
			log.Printf("[DEBUG] thread %d finished", value)
		}
		g.Run(goFunc)
	}
	wg.Wait()

	log.Printf("[DEBUG] DescribeVpcs request finished")
	for _, v := range vpcList {
		vpcs = append(vpcs, v.([]*bmc.VpcInfo)...)
	}
	log.Printf("[DEBUG] transfer vpc instances finished")
	return
}

func (s *BmcService) VpcStateRefreshFunc(ctx context.Context, vpcId string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeVpcById(ctx, vpcId)
		if err != nil {
			return nil, "", err
		}

		if object == nil {
			// Set this to nil as if we didn't find anything.
			return nil, "", nil
		}
		for _, failState := range failStates {
			if object.VpcStatus == failState {
				return object, object.VpcStatus, common2.Error("Failed to reach target status. Last status: %s.", object.VpcStatus)
			}
		}

		return object, object.VpcStatus, nil
	}
}

func (s *BmcService) DescribeVpcById(ctx context.Context, vpcId string) (vpc *bmc.VpcInfo, err error) {
	request := bmc.NewDescribeVpcsRequest()
	request.VpcIds = []string{vpcId}

	response, err := s.client.WithBmcClient().DescribeVpcs(request)

	defer common2.LogApiRequest(ctx, "DescribeVpcs", request, response, err)
	if err != nil {
		return
	}

	if len(response.Response.DataSet) < 1 {
		return
	}
	vpc = response.Response.DataSet[0]
	return

}

func (s *BmcService) DescribeEipAddressById(ctx context.Context, eipId string) (eip *bmc.EipAddress, err error) {
	request := bmc.NewDescribeEipAddressesRequest()
	request.EipIds = []string{eipId}

	response, err := s.client.WithBmcClient().DescribeEipAddresses(request)

	defer common2.LogApiRequest(ctx, "DescribeEipAddresses", request, response, err)
	if err != nil {
		return
	}

	if len(response.Response.DataSet) < 1 {
		return
	}
	eip = response.Response.DataSet[0]
	return
}

func (s *BmcService) TerminateEipAddress(ctx context.Context, eipId string) (err error) {
	request := bmc.NewTerminateEipAddressRequest()
	request.EipId = eipId
	response, err := s.client.WithBmcClient().TerminateEipAddress(request)
	defer common2.LogApiRequest(ctx, "TerminateEipAddress", request, response, err)

	if err != nil {
		if sdkError, ok := err.(*common.ZenlayerCloudSdkError); ok {
			if sdkError.Code == "OPERATION_DENIED_EIP_RECYCLED" {
				return nil
			}
		}
		return
	}
	return
}

func (s *BmcService) ReleaseEipAddressById(ctx context.Context, eipId string) (err error) {
	request := bmc.NewReleaseEipAddressesRequest()
	request.EipIds = []string{eipId}
	response, err := s.client.WithBmcClient().ReleaseEipAddresses(request)
	defer common2.LogApiRequest(ctx, "ReleaseEipAddresses", request, response, err)
	return
}

func (s *BmcService) DescribeDdosIpAddressById(ctx context.Context, ddosIpId string) (eip *bmc.DdosIpAddress, err error) {
	request := bmc.NewDescribeDdosIpAddressesRequest()
	request.DdosIpIds = []string{ddosIpId}

	response, err := s.client.WithBmcClient().DescribeDdosIpAddresses(request)

	defer common2.LogApiRequest(ctx, "DescribeDdosIpAddresses", request, response, err)
	if err != nil {
		return
	}

	if len(response.Response.DataSet) < 1 {
		return
	}
	eip = response.Response.DataSet[0]
	return
}

func (s *BmcService) TerminateDDoSIpAddress(ctx context.Context, ddosIpId string) (err error) {
	request := bmc.NewTerminateDdosIpAddressRequest()
	request.DdosIpId = ddosIpId
	response, err := s.client.WithBmcClient().TerminateDdosIpAddress(request)
	defer common2.LogApiRequest(ctx, "TerminateDdosIpAddress", request, response, err)

	if err != nil {
		if sdkError, ok := err.(*common.ZenlayerCloudSdkError); ok {
			if sdkError.Code == "OPERATION_DENIED_DDOS_IP_RECYCLED" {
				return nil
			}
		}
		return
	}
	return
}

func (s *BmcService) ReleaseDDoSIpAddressById(ctx context.Context, ddosIpId string) (err error) {
	request := bmc.NewReleaseDdosIpAddressesRequest()
	request.DdosIpIds = []string{ddosIpId}
	response, err := s.client.WithBmcClient().ReleaseDdosIpAddresses(request)
	defer common2.LogApiRequest(ctx, "ReleaseDdosIPAddresses", request, response, err)
	return
}

func (s *BmcService) updateInstanceInternetMaxBandwidthOut(ctx context.Context, instanceId string, internetBandwidthOut int) error {
	request := bmc.NewModifyInstanceBandwidthRequest()
	request.InstanceId = instanceId
	request.BandwidthOutMbps = common.Integer(internetBandwidthOut)
	response, err := s.client.WithBmcClient().ModifyInstanceBandwidth(request)
	defer common2.LogApiRequest(ctx, "ModifyInstanceBandwidth", request, response, err)
	if err != nil {
		return err
	}
	return nil
}

func (s *BmcService) ModifyInstanceResourceGroup(ctx context.Context, instanceId string, resourceGroupId string) error {
	request := bmc.NewModifyInstancesResourceGroupRequest()
	request.InstanceIds = []string{instanceId}
	request.ResourceGroupId = resourceGroupId
	response, err := s.client.WithBmcClient().ModifyInstancesResourceGroup(request)
	defer common2.LogApiRequest(ctx, "ModifyInstancesResourceGroup", request, response, err)

	if err != nil {
		return err
	}

	return err
}

func (s *BmcService) DescribeInstanceInternetStatus(ctx context.Context, instanceId string) (*bmc.InstanceInternetStatus, error) {
	request := bmc.NewDescribeInstanceInternetStatusRequest()
	request.InstanceId = instanceId
	status, err := s.client.WithBmcClient().DescribeInstanceInternetStatus(request)
	if err != nil {
		return nil, err
	}
	return status.Response, nil

}

func (s *BmcService) updateInstanceTrafficPackageSize(ctx context.Context, instanceId string, trafficPackageSize float64) error {

	request := bmc.NewModifyInstanceTrafficPackageRequest()
	request.InstanceId = instanceId
	request.TrafficPackageSize = common.Float64(trafficPackageSize)
	response, err := s.client.WithBmcClient().ModifyInstanceTrafficPackage(request)
	defer common2.LogApiRequest(ctx, "ModifyInstanceTrafficPackageSize", request, response, err)
	return err
}

type NetworkStateCondition interface {
	matchFail(status *bmc.InstanceInternetStatus) bool
	matchOk(status *bmc.InstanceInternetStatus) bool
}

const NetworkStatusOK = "OK"
const NetworkStatusFail = "Fail"
const NetworkStatusPending = "Pending"

func (s *BmcService) InstanceNetworkStateRefreshFunc(ctx context.Context, instanceId string, condition NetworkStateCondition) resource.StateRefreshFunc {

	return func() (interface{}, string, error) {
		internetStatus, err := s.DescribeInstanceInternetStatus(ctx, instanceId)
		if err != nil {
			return nil, "", err
		}

		if internetStatus == nil {
			// Set this to nil as if we didn't find anything.
			return nil, "", nil
		}
		if condition.matchFail(internetStatus) {
			return internetStatus, NetworkStatusFail, common2.Error("Failed to reach target status. Last internet status: %v.", internetStatus)
		}
		if condition.matchOk(internetStatus) {
			return internetStatus, NetworkStatusOK, nil
		}
		return internetStatus, NetworkStatusPending, nil
	}
}

func (s *BmcService) AssociateSubnetInstance(ctx context.Context, instanceId string, subnetId string) error {
	request := bmc.NewAssociateSubnetInstancesRequest()
	request.SubnetInstanceList = []*bmc.AssociateSubnetInstanceIpAddress{{
		InstanceId: instanceId,
	}}
	request.SubnetId = subnetId
	_, err := s.client.WithBmcClient().AssociateSubnetInstances(request)
	defer common2.LogApiRequest(ctx, "AssociateSubnetInstances", request, nil, err)
	return err
}

func (s *BmcService) DisassociateSubnetInstance(ctx context.Context, instanceId string, subnetId string) error {
	request := bmc.NewUnAssociateSubnetInstanceRequest()
	request.InstanceId = instanceId
	request.SubnetId = subnetId
	_, err := s.client.WithBmcClient().UnAssociateSubnetInstance(request)
	defer common2.LogApiRequest(ctx, "UnAssociateSubnetInstance", request, nil, err)
	return err
}

const InstanceSubnetStatusNotBind = "NotBind"

func (s *BmcService) InstanceSubnetStateRefreshFunc(ctx context.Context, instanceId string, subnetId string) resource.StateRefreshFunc {

	return func() (interface{}, string, error) {
		subnetInfo, err := s.DescribeSubnetById(ctx, subnetId)
		if err != nil {
			return nil, "", err
		}

		if subnetInfo == nil {
			// Set this to nil as if we didn't find anything.
			return nil, "", nil
		}
		var instanceStatus string
		for _, v := range subnetInfo.SubnetInstanceSet {
			if v.InstanceId == instanceId {
				instanceStatus = v.PrivateIpStatus
				break
			}
		}
		if instanceStatus == "" {
			return subnetInfo, InstanceSubnetStatusNotBind, nil
		}
		return subnetInfo, instanceStatus, nil
	}
}

func (s *BmcService) ModifyDdosIpResourceGroup(ctx context.Context, ddpsIp string, resourceGroupId string) error {
	request := bmc.NewModifyDdosIpAddressesResourceGroupRequest()
	request.DdosIpIds = []string{ddpsIp}
	request.ResourceGroupId = resourceGroupId
	response, err := s.client.WithBmcClient().ModifyDdosIpAddressesResourceGroup(request)
	defer common2.LogApiRequest(ctx, "ModifyDdosIpAddressesResourceGroup", request, response, err)
	return err
}

func (s *BmcService) ModifyEipResourceGroup(ctx context.Context, eipId string, resourceGroupId string) error {
	request := bmc.NewModifyEipAddressesResourceGroupRequest()
	request.EipIds = []string{eipId}
	request.ResourceGroupId = resourceGroupId
	response, err := s.client.WithBmcClient().ModifyEipAddressesResourceGroup(request)
	defer common2.LogApiRequest(ctx, "ModifyEipAddressesResourceGroup", request, response, err)
	return err
}

func (s *BmcService) InstanceEipStateRefreshFunc(ctx context.Context, eipId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		eip, err := s.DescribeEipAddressById(ctx, eipId)
		if err != nil {
			return nil, "", err
		}

		if eip == nil {
			// Set this to nil as if we didn't find anything.
			return nil, "", nil
		}

		return eip, eip.EipStatus, nil
	}
}

func (s *BmcService) ModifyVpcName(ctx context.Context, vpcId string, vpcName string) error {
	request := bmc.NewModifyVpcsAttributeRequest()
	request.VpcIds = []string{vpcId}
	request.VpcName = vpcName
	response, err := s.client.WithBmcClient().ModifyVpcsAttribute(request)
	defer common2.LogApiRequest(ctx, "ModifySubnetsAttribute", request, response, err)

	return err
}

func (s *BmcService) DeleteVpc(ctx context.Context, vpcId string) error {
	request := bmc.NewDeleteVpcRequest()
	request.VpcId = vpcId
	response, err := s.client.WithBmcClient().DeleteVpc(request)
	common2.LogApiRequest(ctx, "DeleteVpc", request, response, err)
	return err

}

func (s *BmcService) ModifyVpcResourceGroup(ctx context.Context, vpcId string, resourceGroupId string) error {
	request := bmc.NewModifyVpcsResourceGroupRequest()
	request.VpcIds = []string{vpcId}
	request.ResourceGroupId = resourceGroupId
	response, err := s.client.WithBmcClient().ModifyVpcsResourceGroup(request)
	common2.LogApiRequest(ctx, "ModifyVpcsResourceGroup", request, response, err)
	return err
}

func (s *BmcService) DescribeZones(ctx context.Context) (zones []*bmc.ZoneInfo, err error) {
	request := bmc.NewDescribeZonesRequest()
	response, err := s.client.WithBmcClient().DescribeZones(request)
	common2.LogApiRequest(ctx, "DescribeZones", request, response, err)
	if err != nil {
		return
	}
	zones = response.Response.ZoneSet
	return

}

func convertVpcFilter(filter *VpcFilter) (request *bmc.DescribeVpcsRequest) {
	request = bmc.NewDescribeVpcsRequest()

	if filter.VpcId != nil {
		request.VpcIds = []string{*filter.VpcId}
	}
	if filter.vpcRegion != nil {
		request.VpcRegionId = *filter.vpcRegion
	}
	if filter.CidrBlock != nil {
		request.CidrBlock = *filter.CidrBlock
	}
	if filter.ResourceGroupId != nil {
		request.ResourceGroupId = *filter.ResourceGroupId
	}
	return
}

func convertDdosIpFilter(filter *DDosIpFilter) (request *bmc.DescribeDdosIpAddressesRequest) {
	request = bmc.NewDescribeDdosIpAddressesRequest()

	if len(filter.IpIds) > 0 {
		request.DdosIpIds = filter.IpIds
	}
	if filter.DdosIpStatus != nil {
		request.DdosIpStatus = *filter.DdosIpStatus
	}
	if filter.Ip != nil {
		request.IpAddress = *filter.Ip
	}
	if filter.ZoneId != nil {
		request.ZoneId = *filter.ZoneId
	}
	if filter.InstanceId != nil {
		request.InstanceId = *filter.InstanceId
	}
	if filter.ResourceGroupId != nil {
		request.ResourceGroupId = *filter.ResourceGroupId
	}
	return
}

func convertSubnetFilterRequest(filter *SubnetFilter) (request *bmc.DescribeSubnetsRequest) {
	request = bmc.NewDescribeSubnetsRequest()

	if filter.SubnetName != "" {
		request.SubnetName = filter.SubnetName
	}
	if filter.SubnetId != "" {
		request.SubnetIds = []string{filter.SubnetId}
	}
	if filter.ZoneId != "" {
		request.ZoneId = filter.ZoneId
	}
	if filter.VpcId != "" {
		request.VpcId = filter.VpcId
	}
	if filter.ResourceGroupId != "" {
		request.ResourceGroupId = filter.ResourceGroupId
	}
	if filter.CidrBlock != "" {
		request.CidrBlock = filter.CidrBlock
	}
	return
}

func convertEipsFilter(filter *EipFilter) (request *bmc.DescribeEipAddressesRequest) {
	request = bmc.NewDescribeEipAddressesRequest()

	if len(filter.EipIds) > 0 {
		request.EipIds = filter.EipIds
	}
	if filter.EipStatus != nil {
		request.EipStatus = *filter.EipStatus
	}
	if filter.Ip != nil {
		request.IpAddress = *filter.Ip
	}
	if filter.ZoneId != nil {
		request.ZoneId = *filter.ZoneId
	}
	if filter.InstanceId != nil {
		request.InstanceId = *filter.InstanceId
	}
	if filter.ResourceGroupId != nil {
		request.ResourceGroupId = *filter.ResourceGroupId
	}
	return
}

func convertRequestForInstanceFilter(filter *InstancesFilter) (request *bmc.DescribeInstancesRequest) {
	request = bmc.NewDescribeInstancesRequest()
	if len(filter.InstancesIds) > 0 {
		request.InstanceIds = filter.InstancesIds
	}
	if filter.InstanceTypeId != nil {
		request.InstanceTypeId = *filter.InstanceTypeId
	}
	if filter.PublicIpv4 != nil {
		request.PublicIpAddresses = []string{*filter.PublicIpv4}
	}
	if filter.ImageId != nil {
		request.ImageId = *filter.ImageId
	}
	if filter.Hostname != nil {
		request.Hostname = *filter.Hostname
	}
	if filter.InstanceName != nil {
		request.InstanceName = *filter.InstanceName
	}
	if filter.SubnetId != nil {
		request.SubnetId = *filter.SubnetId
	}
	if filter.ResourceGroupId != nil {
		request.ResourceGroupId = *filter.ResourceGroupId
	}

	if filter.InstanceStatus != nil {
		request.InstanceStatus = *filter.InstanceStatus
	}
	if filter.ZoneId != nil {
		request.ZoneId = *filter.ZoneId
	}
	if filter.PrivateIpv4 != nil {
		request.PrivateIpAddresses = []string{*filter.PrivateIpv4}
	}
	return
}

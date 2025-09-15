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
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	sdn "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/sdn20230830"
	"log"
	"math"
	"sync"
)

type SdnService struct {
	client *connectivity.ZenlayerCloudClient
}

func (s *SdnService) DeletePortById(ctx context.Context, portId string) (err error) {
	// 判断已经终止 OPERATION_DENIED_INSTANCE_RECYCLED,
	request := sdn.NewTerminatePortRequest()
	request.PortId = portId
	response, err := s.client.WithSdnClient().TerminatePort(request)
	defer common2.LogApiRequest(ctx, "TerminatePort", request, response, err)

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

func (s *SdnService) DestroyPort(ctx context.Context, portId string) (err error) {
	request := sdn.NewDestroyPortRequest()
	request.PortId = portId
	response, err := s.client.WithSdnClient().DestroyPort(request)
	defer common2.LogApiRequest(ctx, "DestroyPort", request, response, err)
	return
}

func (s *SdnService) DescribePortById(ctx context.Context, portId string) (instance *sdn.PortInfo, err error) {
	request := sdn.NewDescribePortsRequest()
	request.PortIds = []string{portId}

	response, err := s.client.WithSdnClient().DescribePorts(request)

	defer common2.LogApiRequest(ctx, "DescribePorts", request, response, err)
	if err != nil {
		return
	}

	if len(response.Response.DataSet) < 1 {
		return
	}
	instance = response.Response.DataSet[0]
	return
}

func (s *SdnService) PortStateRefreshFunc(ctx context.Context, portId string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribePortById(ctx, portId)
		if err != nil {
			return nil, "", err
		}

		if object == nil {
			// Set this to nil as if we didn't find anything.
			return nil, "", nil
		}
		for _, failState := range failStates {
			if object.PortStatus == failState {
				return object, object.PortStatus, common2.Error("Failed to reach target status. Last status: %s.", object.PortStatus)
			}
		}

		return object, object.PortStatus, nil
	}
}

func (s *SdnService) ModifyPort(ctx context.Context, portId, portName, remarks, businessEntityName string) error {
	request := sdn.NewModifyPortAttributeRequest()
	request.PortId = portId
	request.PortName = portName
	request.PortRemarks = remarks
	request.BusinessEntityName = businessEntityName
	response, err := s.client.WithSdnClient().ModifyPortAttribute(request)
	defer common2.LogApiRequest(ctx, "ModifyPortAttribute", request, response, err)
	return err
}

func (s *SdnService) DescribePortsByFilter(portFilter *PortFilter) (instances []*sdn.PortInfo, err error) {
	request := convertRequestForPortFilter(portFilter)
	var limit = 100
	request.PageSize = limit
	request.PageNum = 1
	response, err := s.client.WithSdnClient().DescribePorts(request)

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

	var portSetList = make([]interface{}, num)

	for i := 0; i < num; i++ {
		wg.Add(1)
		value := i
		goFunc := func() {
			request := convertRequestForPortFilter(portFilter)

			request.PageNum = value + 2
			request.PageSize = limit

			response, err := s.client.WithSdnClient().DescribePorts(request)
			if err != nil {
				log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
					request.GetAction(), common2.ToJsonString(request), err.Error())
				return
			}
			log.Printf("[DEBUG] Api[%s] success, request body [%s], response body [%s]\n",
				request.GetAction(), common2.ToJsonString(request), common2.ToJsonString(response))

			portSetList[value] = response.Response.DataSet

			wg.Done()
			log.Printf("[DEBUG] thread %d finished", value)
		}
		g.Run(goFunc)
	}
	wg.Wait()

	log.Printf("[DEBUG] DescribePorts request finished")
	for _, v := range portSetList {
		instances = append(instances, v.([]*sdn.PortInfo)...)
	}
	log.Printf("[DEBUG] transfer port finished")
	return
}

func (s *SdnService) DescribeDatacenters(ctx context.Context) ([]*sdn.DatacenterInfo, error) {
	request := sdn.NewDescribeDatacentersRequest()
	response, err := s.client.WithSdnClient().DescribeDatacenters(request)
	defer common2.LogApiRequest(ctx, "DescribeDatacenters", request, response, err)

	if err != nil {
		return nil, err
	}
	return response.Response.DcSet, nil
}

func (s *SdnService) DescribePrivateConnectsByFilter(filter *PrivateConnectFilter) (privateConnects []*sdn.PrivateConnect, err error) {
	request := convertRequestForPrivateConnectFilter(filter)
	var limit = 100
	request.PageSize = limit
	request.PageNum = 1
	response, err := s.client.WithSdnClient().DescribePrivateConnects(request)

	if err != nil {
		log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
			request.GetAction(), common2.ToJsonString(request), err.Error())
		return
	}
	if response == nil || len(response.Response.DataSet) < 1 {
		return
	}

	//total := response.Response.TotalCount
	privateConnects = response.Response.DataSet
	num := int(math.Ceil(float64(response.Response.TotalCount)/float64(limit))) - 1
	if num == 0 {
		return privateConnects, nil
	}
	maxConcurrentNum := 50
	g := common2.NewGoRoutine(maxConcurrentNum)
	wg := sync.WaitGroup{}

	var portSetList = make([]interface{}, num)

	for i := 0; i < num; i++ {
		wg.Add(1)
		value := i
		goFunc := func() {
			request := convertRequestForPrivateConnectFilter(filter)

			request.PageNum = value + 2
			request.PageSize = limit

			response, err := s.client.WithSdnClient().DescribePrivateConnects(request)
			if err != nil {
				log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
					request.GetAction(), common2.ToJsonString(request), err.Error())
				return
			}
			log.Printf("[DEBUG] Api[%s] success, request body [%s], response body [%s]\n",
				request.GetAction(), common2.ToJsonString(request), common2.ToJsonString(response))

			portSetList[value] = response.Response.DataSet

			wg.Done()
			log.Printf("[DEBUG] thread %d finished", value)
		}
		g.Run(goFunc)
	}
	wg.Wait()

	log.Printf("[DEBUG] DescribePrivateConnects request finished")
	for _, v := range portSetList {
		privateConnects = append(privateConnects, v.([]*sdn.PrivateConnect)...)
	}
	log.Printf("[DEBUG] transfer private connects finished")
	return
}

func (s *SdnService) DescribeCloudRoutersByFilter(filter *CloudRouterFilter) (cloudRouters []*sdn.CloudRouter, err error) {
	request := convertRequestForCloudRouterFilter(filter)
	var limit = 100
	request.PageSize = limit
	request.PageNum = 1
	response, err := s.client.WithSdnClient().DescribeCloudRouters(request)

	if err != nil {
		log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
			request.GetAction(), common2.ToJsonString(request), err.Error())
		return
	}
	if response == nil || len(response.Response.DataSet) < 1 {
		return
	}

	//total := response.Response.TotalCount
	cloudRouters = response.Response.DataSet
	num := int(math.Ceil(float64(response.Response.TotalCount)/float64(limit))) - 1
	if num == 0 {
		return cloudRouters, nil
	}
	maxConcurrentNum := 50
	g := common2.NewGoRoutine(maxConcurrentNum)
	wg := sync.WaitGroup{}

	var cloudRouterList = make([]interface{}, num)

	for i := 0; i < num; i++ {
		wg.Add(1)
		value := i
		goFunc := func() {
			request := convertRequestForCloudRouterFilter(filter)

			request.PageNum = value + 2
			request.PageSize = limit

			response, err := s.client.WithSdnClient().DescribeCloudRouters(request)
			if err != nil {
				log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
					request.GetAction(), common2.ToJsonString(request), err.Error())
				return
			}
			log.Printf("[DEBUG] Api[%s] success, request body [%s], response body [%s]\n",
				request.GetAction(), common2.ToJsonString(request), common2.ToJsonString(response))

			cloudRouterList[value] = response.Response.DataSet

			wg.Done()
			log.Printf("[DEBUG] thread %d finished", value)
		}
		g.Run(goFunc)
	}
	wg.Wait()

	log.Printf("[DEBUG] DescribeCloudRouters request finished")
	for _, v := range cloudRouterList {
		cloudRouters = append(cloudRouters, v.([]*sdn.CloudRouter)...)
	}
	log.Printf("[DEBUG] transfer cloud router finished")
	return
}

func (s *SdnService) ModifyPrivateConnectName(ctx context.Context, connectId string, name string) error {
	request := sdn.NewModifyPrivateConnectsAttributeRequest()
	request.PrivateConnectIds = []string{connectId}
	request.PrivateConnectName = name
	response, err := s.client.WithSdnClient().ModifyPrivateConnectsAttribute(request)
	common2.LogApiRequest(ctx, "ModifyPrivateConnectsAttribute", request, response, err)
	return err
}

func (s *SdnService) ModifyPrivateConnectBandwidth(ctx context.Context, connectId string, bandwidth int) error {
	request := sdn.NewModifyPrivateConnectBandwidthRequest()
	request.PrivateConnectId = connectId
	request.BandwidthMbps = bandwidth
	response, err := s.client.WithSdnClient().ModifyPrivateConnectBandwidth(request)
	common2.LogApiRequest(ctx, "ModifyPrivateConnectBandwidth", request, response, err)
	return err
}

func (s *SdnService) DescribePrivateConnectById(ctx context.Context, connectId string) (connect *sdn.PrivateConnect, err error) {
	request := sdn.NewDescribePrivateConnectsRequest()
	request.PrivateConnectIds = []string{connectId}

	response, err := s.client.WithSdnClient().DescribePrivateConnects(request)

	defer common2.LogApiRequest(ctx, "DescribePrivateConnects", request, response, err)
	if err != nil {
		return
	}

	if len(response.Response.DataSet) < 1 {
		return
	}
	connect = response.Response.DataSet[0]
	return
}

func (s *SdnService) PrivateConnectStateRefreshFunc(ctx context.Context, connectId string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribePrivateConnectById(ctx, connectId)
		if err != nil {
			return nil, "", err
		}

		if object == nil {
			// Set this to nil as if we didn't find anything.
			return nil, "", nil
		}
		for _, failState := range failStates {
			if object.PrivateConnectStatus == failState {
				return object, object.PrivateConnectStatus, common2.Error("Failed to reach target status. Last status: %s.", object.PrivateConnectStatus)
			}
		}

		return object, object.PrivateConnectStatus, nil
	}
}

func (s *SdnService) DeletePrivateConnectById(ctx context.Context, connectId string) (err error) {
	// 判断已经终止 INVALID_PRIVATE_CONNECT_NOT_FOUND,
	request := sdn.NewDeletePrivateConnectRequest()
	request.PrivateConnectId = connectId
	response, err := s.client.WithSdnClient().DeletePrivateConnect(request)
	defer common2.LogApiRequest(ctx, "DeletePrivateConnect", request, response, err)

	if err != nil {
		if sdkError, ok := err.(*common.ZenlayerCloudSdkError); ok {
			if sdkError.Code == "INVALID_PRIVATE_CONNECT_NOT_FOUND" {
				return nil
			}
		}
		return
	}
	return
}

func (s *SdnService) DestroyPrivateConnect(ctx context.Context, connectId string) (err error) {
	request := sdn.NewDestroyPrivateConnectRequest()
	request.PrivateConnectId = connectId
	response, err := s.client.WithSdnClient().DestroyPrivateConnect(request)
	defer common2.LogApiRequest(ctx, "DestroyPrivateConnect", request, response, err)
	return
}

func (s *SdnService) DescribeCloudRegions(filter CloudRegionFilter) ([]*sdn.CloudRegion, error) {
	if filter.cloudType == POINT_TYPE_GOOGLE {
		request := sdn.NewDescribeGoogleRegionsRequest()
		request.PairingKey = *filter.googlePairingKey
		if filter.product != nil {
			request.Product = *filter.product
		}
		response, err := s.client.WithSdnClient().DescribeGoogleRegions(request)
		if err != nil {
			return nil, err
		}
		return response.Response.CloudRegions, nil
	}
	if filter.cloudType == POINT_TYPE_AWS {
		request := sdn.NewDescribeAWSRegionsRequest()
		if filter.product != nil {
			request.Product = *filter.product
		}
		response, err := s.client.WithSdnClient().DescribeAWSRegions(request)
		if err != nil {
			return nil, err
		}
		return response.Response.CloudRegions, nil
	}
	if filter.cloudType == POINT_TYPE_TENCENT {
		request := sdn.NewDescribeTencentRegionsRequest()
		if filter.product != nil {
			request.Product = *filter.product
		}
		response, err := s.client.WithSdnClient().DescribeTencentRegions(request)
		if err != nil {
			return nil, err
		}
		return response.Response.CloudRegions, nil
	}
	return nil, fmt.Errorf("cloud type: %s is not support", filter.cloudType)
}

func convertRequestForCloudRouterFilter(filter *CloudRouterFilter) (request *sdn.DescribeCloudRoutersRequest) {
	request = sdn.NewDescribeCloudRoutersRequest()
	if len(filter.CloudRouterIds) > 0 {
		request.CloudRouterIds = filter.CloudRouterIds
	}
	return
}

func convertRequestForPortFilter(filter *PortFilter) (request *sdn.DescribePortsRequest) {
	request = sdn.NewDescribePortsRequest()
	if len(filter.PortIds) > 0 {
		request.PortIds = filter.PortIds
	}
	if filter.DcId != nil {
		request.DcId = *filter.DcId
	}
	return
}

func convertRequestForPrivateConnectFilter(filter *PrivateConnectFilter) (request *sdn.DescribePrivateConnectsRequest) {
	request = sdn.NewDescribePrivateConnectsRequest()
	if len(filter.ConnectIds) > 0 {
		request.PrivateConnectIds = filter.ConnectIds
	}
	return
}

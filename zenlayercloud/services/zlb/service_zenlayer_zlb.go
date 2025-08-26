package zlb

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
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	common2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	zlb "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zlb20250401"
	"log"
	"math"
	"sync"
)

type ZlbService struct {
	client *connectivity.ZenlayerCloudClient
}

func (s *ZlbService) DescribeLbInstancesByFilter(ctx context.Context, filter *LbInstanceFilter) (lbs []*zlb.LoadBalancer, err error) {
	request := convertLbInstancesRequestFilter(filter)

	var limit = 100
	request.PageSize = common2.Integer(limit)
	request.PageNum = common2.Integer(1)
	response, err := s.client.WithZlbClient().DescribeLoadBalancers(request)

	if err != nil {
		log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
			request.GetAction(), common.ToJsonString(request), err.Error())
		return
	}
	if response == nil || len(response.Response.DataSet) < 1 {
		return
	}

	lbs = response.Response.DataSet
	num := int(math.Ceil(float64(*response.Response.TotalCount)/float64(limit))) - 1
	if num == 0 {
		return lbs, nil
	}
	maxConcurrentNum := 50
	g := common.NewGoRoutine(maxConcurrentNum)
	wg := sync.WaitGroup{}

	var vpcList = make([]interface{}, num)

	for i := 0; i < num; i++ {
		wg.Add(1)
		value := i
		goFunc := func() {
			request := convertLbInstancesRequestFilter(filter)

			request.PageNum = common2.Integer(value + 2)
			request.PageSize = &limit

			response, err := s.client.WithZlbClient().DescribeLoadBalancers(request)
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

	log.Printf("[DEBUG] DescribeLoadBalancers request finished")
	for _, v := range vpcList {
		lbs = append(lbs, v.([]*zlb.LoadBalancer)...)
	}
	log.Printf("[DEBUG] transfer `Load Balancer Instance` finished")
	return
}

func (s *ZlbService) DeleteZlbInstanceById(ctx context.Context, zlbId string) error {

	request := zlb.NewTerminateLoadBalancerRequest()
	request.LoadBalancerId = &zlbId
	response, err := s.client.WithZlbClient().TerminateLoadBalancer(request)
	defer common.LogApiRequest(ctx, "TerminateLoadBalancer", request, response, err)
	return err
}

func (s *ZlbService) DescribeZlbInstanceById(ctx context.Context, id string) (*zlb.LoadBalancer, error) {
	request := zlb.NewDescribeLoadBalancersRequest()
	request.LoadBalancerIds = []string{id}
	response, err := s.client.WithZlbClient().DescribeLoadBalancers(request)
	defer common.LogApiRequest(ctx, "DescribeLoadBalancers", request, response, err)

	if err != nil {
		ee, ok := err.(*common2.ZenlayerCloudSdkError)
		if ok {
			if ee.Code == common.ResourceNotFound {
				// LB doesn't exist
				return nil, nil
			}
		}
		return nil, err
	} else if len(response.Response.DataSet) == 0 {
		return nil, nil
	}
	return response.Response.DataSet[0], nil

}

func (s *ZlbService) DescribeLoadBalancerRegions() (regions []*zlb.Region, err error) {
	response, err := s.client.WithZlbClient().DescribeLoadBalancerRegions(zlb.NewDescribeLoadBalancerRegionsRequest())
	if err != nil {
		log.Printf("[CRITAL] Api[DescribeLoadBalancerRegions] fail, , error[%s]\n", err.Error())
		return
	}
	regions = response.Response.Regions
	return
}

func (s *ZlbService) ModifyLoadBalancerName(ctx context.Context, zlbId string, name string) error {
	request := zlb.NewModifyLoadBalancersAttributeRequest()
	request.LoadBalancerIds = []string{zlbId}
	request.LoadBalancerName = &name
	response, err := s.client.WithZlbClient().ModifyLoadBalancersAttribute(request)
	defer common.LogApiRequest(ctx, "ModifyLoadBalancersAttribute", request, response, err)
	return err
}

func convertLbInstancesRequestFilter(filter *LbInstanceFilter) *zlb.DescribeLoadBalancersRequest {
	request := zlb.NewDescribeLoadBalancersRequest()
	request.VpcId = &filter.VpcId
	request.ResourceGroupId = &filter.ResourceGroupId
	request.LoadBalancerIds = filter.LbIds
	request.RegionId = &filter.RegionId
	return request
}

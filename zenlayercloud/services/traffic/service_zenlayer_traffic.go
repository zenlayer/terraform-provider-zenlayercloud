package traffic

import (
	"context"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	common2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	traffic "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/traffic20240326"
	"log"
	"math"
	"sync"
)

type TrafficService struct {
	client *connectivity.ZenlayerCloudClient
}

func (s TrafficService) DescribeBandwidthClusterByFilter(ctx context.Context, filter *TrafficFilter) (bandwidthClusters []*traffic.BandwidthClusterInfo, err error) {
	request := convertBandwidthClusterRequestFilter(filter)

	var limit = 100
	request.PageSize = common2.Integer(limit)
	request.PageNum = common2.Integer(1)
	response, err := s.client.WithTrafficClient().DescribeBandwidthClusters(request)

	if err != nil {
		log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
			request.GetAction(), common.ToJsonString(request), err.Error())
		return
	}
	if response == nil || len(response.Response.DataSet) < 1 {
		return
	}

	bandwidthClusters = response.Response.DataSet
	num := int(math.Ceil(float64(*response.Response.TotalCount)/float64(limit))) - 1
	if num == 0 {
		return bandwidthClusters, nil
	}
	maxConcurrentNum := 50
	g := common.NewGoRoutine(maxConcurrentNum)
	wg := sync.WaitGroup{}

	var vpcList = make([]interface{}, num)

	for i := 0; i < num; i++ {
		wg.Add(1)
		value := i
		goFunc := func() {
			request := convertBandwidthClusterRequestFilter(filter)

			request.PageNum = common2.Integer(int(value + 2))
			request.PageSize = common2.Integer(limit)

			response, err := s.client.WithTrafficClient().DescribeBandwidthClusters(request)
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

	log.Printf("[DEBUG] DescribeBandwidthClusters request finished")
	for _, v := range vpcList {
		bandwidthClusters = append(bandwidthClusters, v.([]*traffic.BandwidthClusterInfo)...)
	}
	log.Printf("[DEBUG] transfer bandwidth clusters finished")
	return
}

func (s *TrafficService) DescribeBandwidthClusterAreas(ctx context.Context) ([]*traffic.BandwidthClusterAreaInfo, error) {
	request := traffic.NewDescribeBandwidthClusterAreasRequest()
	response, err := s.client.WithTrafficClient().DescribeBandwidthClusterAreas(request)
	defer common.LogApiRequest(ctx, "DescribeBandwidthClusterAreas", request, response, err)

	return response.Response.Areas, err
}

func (s TrafficService) DescribeBandwidthClusterById(ctx context.Context, bandwidthClusterId string) (bandwidthCluster *traffic.BandwidthClusterInfo, err error) {
	request := traffic.NewDescribeBandwidthClustersRequest()
	request.BandwidthClusterIds= []string{bandwidthClusterId}

	var response *traffic.DescribeBandwidthClustersResponse

	defer common.LogApiRequest(ctx, "DescribeBandwidthClusters", request, response, err)

	response, err = s.client.WithTrafficClient().DescribeBandwidthClusters(request)

	if err != nil {
		return
	}

	if len(response.Response.DataSet) < 1 {
		return
	}
	bandwidthCluster = response.Response.DataSet[0]
	return

}

func (s *TrafficService) ModifyBandwidthClusterCommitBandwidth(ctx context.Context, bandwidthClusterId string, bandwidthMbps int) error {
	request := traffic.NewUpdateBandwidthClusterCommitBandwidthRequest()
	request.BandwidthClusterId = common2.String(bandwidthClusterId)
	request.CommitBandwidthMbps = &bandwidthMbps
	response, err := s.client.WithTrafficClient().UpdateBandwidthClusterCommitBandwidth(request)
	defer common.LogApiRequest(ctx, "UpdateBandwidthClusterCommitBandwidth", request, response, err)
	return err
}

func (s *TrafficService) ModifyBandwidthClusterName(ctx context.Context, bandwidthClusterId string, name string) error {
	request := traffic. NewModifyBandwidthClusterAttributeRequest()
	request.BandwidthClusterId = common2.String(bandwidthClusterId)
	request.Name = &name
	response, err := s.client.WithTrafficClient().ModifyBandwidthClusterAttribute(request)
	defer common.LogApiRequest(ctx, "ModifyBandwidthClusterAttribute", request, response, err)
	return err
}

func (s TrafficService) DescribeBandwidthClusterResourcesById(ctx context.Context, bandwidthClusterId string) (*traffic.DescribeBandwidthClusterResourcesResponseParams, error) {
	request := traffic.NewDescribeBandwidthClusterResourcesRequest()
	request.BandwidthClusterId = &bandwidthClusterId
	response, err := s.client.WithTrafficClient().DescribeBandwidthClusterResources(request)
	defer common.LogApiRequest(ctx, "DescribeBandwidthClusterResources", request, response, err)

	return response.Response, err
}

func (s TrafficService) DeleteBandwidthClusterId(ctx context.Context, bandwidthClusterId string) error {
	request := traffic.NewDeleteBandwidthClustersRequest()
	request.BandwidthClusterIds = []string{bandwidthClusterId}

	response, err := s.client.WithTrafficClient().DeleteBandwidthClusters(request)
	defer common.LogApiRequest(ctx, "DeleteBandwidthClusters", request, response, err)
	return err
}

func convertBandwidthClusterRequestFilter(filter *TrafficFilter) *traffic.DescribeBandwidthClustersRequest {
	request := traffic.NewDescribeBandwidthClustersRequest()
	request.BandwidthClusterIds = filter.Ids
	if filter.cityName != "" {
		request.CityName = common2.String(filter.cityName)
	}
	return request
}

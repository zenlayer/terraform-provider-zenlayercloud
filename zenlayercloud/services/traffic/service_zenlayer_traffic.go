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

func (s TrafficService) DescribeSecurityGroupById(ctx context.Context, bandwidthClusterId string) (bandwidthCluster *traffic.BandwidthClusterInfo, err error) {
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

func (s *TrafficService) ModifyBandwidthClusterName(ctx context.Context, bandwidthClusterId string, name string) error {
	//request := traffic.NewUpdateBandwidthClusterCommitBandwidthRequest()
	//request.BandwidthClusterId = common2.String(bandwidthClusterId)
	//request.CommitBandwidthMbps = common2.Integer(0) // Assuming name modification doesn't require bandwidth change
	//
	//// TODO
	//// Note: The API doesn't seem to support name modification directly.
	//// This is a placeholder implementation. You might need to adjust based on actual API capabilities.
	//
	//response, err := s.client.WithTrafficClient().Upda(request)
	//defer common.LogApiRequest(ctx, "UpdateBandwidthClusterCommitBandwidth", request, response, err)
	//
	//return err
	return nil
}

func convertBandwidthClusterRequestFilter(filter *TrafficFilter) *traffic.DescribeBandwidthClustersRequest {
	request := traffic.NewDescribeBandwidthClustersRequest()
	request.BandwidthClusterIds = filter.Ids
	if filter.cityName != "" {
		request.BandwidthClusterName = common2.String(filter.cityName)
	}
	return request
}

package zdns

import (
	"context"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	common2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	zdns "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zdns20251101"
	"log"
	"math"
	"sync"
)

type ZdnsService struct {
	client *connectivity.ZenlayerCloudClient
}

func (s *ZdnsService) DeletePrivateZoneById(ctx context.Context, zoneId string) error{
	request := zdns.NewDeletePrivateZoneRequest()
	request.ZoneId = &zoneId
	response, err := s.client.WithZDnsClient().DeletePrivateZone(request)
	defer common.LogApiRequest(ctx, "DeletePrivateZone", request, response, err)
	return err
}

func (s *ZdnsService) DescribePrivateZoneById(ctx context.Context, id string) (pz *zdns.PrivateZone,err error) {

	request := zdns.NewDescribePrivateZonesRequest()
	request.ZoneIds = []string{ id}
	response, err := s.client.WithZDnsClient().DescribePrivateZones(request)
	defer common.LogApiRequest(ctx, "DescribePrivateZones", request, response, err)

	if err != nil {
		return
	}

	if len(response.Response.DataSet) < 1 {
		return
	}
	pz = response.Response.DataSet[0]
	return
}

func (s *ZdnsService) DeletePrivateDnsRecordById(ctx context.Context, zoneId string, recordId string) error {
	request := zdns.NewDeletePrivateZoneRecordRequest()
	request.ZoneId = &zoneId
	request.RecordIds = []string{ recordId}
	response, err := s.client.WithZDnsClient().DeletePrivateZoneRecord(request)
	defer common.LogApiRequest(ctx, "DeletePrivateZoneRecord", request, response, err)
	return err
}

func (s *ZdnsService) DescribePrivateZoneRecordById(ctx context.Context, zoneId string, recordId string) (record *zdns.PrivateZoneRecord, err error) {
	request := zdns.NewDescribePrivateZoneRecordsRequest()
	request.ZoneId  = &zoneId
	request.RecordIds = []string{recordId}
	response, err := s.client.WithZDnsClient().DescribePrivateZoneRecords(request)
	defer common.LogApiRequest(ctx, "DescribePrivateZoneRecords", request, response, err)

	if err != nil {
		return
	}

	if len(response.Response.DataSet) < 1 {
		return
	}
	record = response.Response.DataSet[0]
	return

}

func (s *ZdnsService) DescribePrivateZonesByFilter(ctx context.Context, filter *PrivateZoneFilter) (pzs []*zdns.PrivateZone,err error) {

	request := convertPrivateZonesRequestFilter(filter)

	var limit = 100
	request.PageSize = common2.Integer(limit)
	request.PageNum = common2.Integer(1)
	response, err := s.client.WithZDnsClient().DescribePrivateZones(request)

	if err != nil {
		log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
			request.GetAction(), common.ToJsonString(request), err.Error())
		return
	}
	if response == nil || len(response.Response.DataSet) < 1 {
		return
	}

	pzs = response.Response.DataSet
	num := int(math.Ceil(float64(*response.Response.TotalCount)/float64(limit))) - 1
	if num == 0 {
		return pzs, nil
	}
	maxConcurrentNum := 50
	g := common.NewGoRoutine(maxConcurrentNum)
	wg := sync.WaitGroup{}

	var vpcList = make([]interface{}, num)

	for i := 0; i < num; i++ {
		wg.Add(1)
		value := i
		goFunc := func() {
			request := convertPrivateZonesRequestFilter(filter)

			request.PageNum = common2.Integer(value + 2)
			request.PageSize = &limit

			response, err := s.client.WithZDnsClient().DescribePrivateZones(request)
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

	log.Printf("[DEBUG] DescribePrivateZones request finished")
	for _, v := range vpcList {
		pzs = append(pzs, v.([]*zdns.PrivateZone)...)
	}
	log.Printf("[DEBUG] transfer `private zones` finished")
	return
}

func (s *ZdnsService) DescribePrivateZoneRecordsByFilter(ctx context.Context, filter *PrivateRecordFilter) (records []*zdns.PrivateZoneRecord,err error) {

	request := convertPrivateZoneRecordsRequestFilter(filter)

	var limit = 100
	request.PageSize = common2.Integer(limit)
	request.PageNum = common2.Integer(1)
	response, err := s.client.WithZDnsClient().DescribePrivateZoneRecords(request)

	if err != nil {
		log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
			request.GetAction(), common.ToJsonString(request), err.Error())
		return
	}
	if response == nil || len(response.Response.DataSet) < 1 {
		return
	}

	records = response.Response.DataSet
	num := int(math.Ceil(float64(*response.Response.TotalCount)/float64(limit))) - 1
	if num == 0 {
		return records, nil
	}
	maxConcurrentNum := 50
	g := common.NewGoRoutine(maxConcurrentNum)
	wg := sync.WaitGroup{}

	var vpcList = make([]interface{}, num)

	for i := 0; i < num; i++ {
		wg.Add(1)
		value := i
		goFunc := func() {
			request := convertPrivateZoneRecordsRequestFilter(filter)

			request.PageNum = common2.Integer(value + 2)
			request.PageSize = &limit

			response, err := s.client.WithZDnsClient().DescribePrivateZoneRecords(request)
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

	log.Printf("[DEBUG] DescribePrivateZones request finished")
	for _, v := range vpcList {
		records = append(records, v.([]*zdns.PrivateZoneRecord)...)
	}
	log.Printf("[DEBUG] transfer `private zones` finished")
	return

}

func convertPrivateZoneRecordsRequestFilter(filter *PrivateRecordFilter) *zdns.DescribePrivateZoneRecordsRequest {
	request := zdns.NewDescribePrivateZoneRecordsRequest()
	request.RecordIds = filter.RecordIds
	request.Type = &filter.RecordType
	request.ZoneId = &filter.ZoneId
	request.Value = &filter.RecordValue
	return request

}

func convertPrivateZonesRequestFilter(filter *PrivateZoneFilter) *zdns.DescribePrivateZonesRequest {
	request := zdns.NewDescribePrivateZonesRequest()
	request.ZoneIds = filter.Ids
	request.ResourceGroupId = &filter.ResourceGroupId
	return request
}

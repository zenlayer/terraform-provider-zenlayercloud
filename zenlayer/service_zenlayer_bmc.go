package zenlayer

import (
        "github.com/zenlayer/terraform-provider-zenlayer/zenlayer/connectivity"
        "github.com/zenlayer/zenlayer-go-sdk/services/bmc"
)

type BmcService struct {
        client *connectivity.ZenlayerClient
}

func (s *BmcService) ListZones() (zones []bmc.Zone, err error){
        request := bmc.CreateListZonesRequest()
        conn, err := s.client.NewBmcClient()
        if err != nil {
                return
        }

        response, err := conn.ListZones(request)
        if err != nil {
                return
        }

        zones = response.Zones
        return
}

func (s *BmcService) ListModels(zoneUuid string) (models []bmc.Model, err error) {
        request := bmc.CreateListModelsRequest(zoneUuid)
        conn, err := s.client.NewBmcClient()
        if err != nil {
                return
        }

        response, err := conn.ListModels(request)
        if err != nil {
                return
        }

        models = response.Models
        return
}

func (s *BmcService) ListOSs(modelUuid string) (oss []bmc.OS, err error) {
        request := bmc.CreateListOSsRequest(modelUuid)
        conn, err := s.client.NewBmcClient()
        if err != nil {
                return
        }

        response, err := conn.ListOSs(request)
        if err != nil {
                return
        }

        oss = response.OSs
        return
}

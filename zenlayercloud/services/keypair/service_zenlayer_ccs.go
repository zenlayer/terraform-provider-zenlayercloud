package keypair

import (
	"context"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	ccs "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/ccs20250901"
	"log"
)

type CcsService struct {
	client *connectivity.ZenlayerCloudClient
}


func (s *CcsService) DescribeKeyPairs(ctx context.Context, request *ccs.DescribeKeyPairsRequest) (keyPairs []*ccs.KeyPair, err error) {
	response, err := s.client.WithCcsClient().DescribeKeyPairs(request)
	common2.LogApiRequest(ctx, "DescribeKeyPairs", request, response, err)
	if err != nil {
		return
	}
	keyPairs = response.Response.DataSet
	return
}

func (s *CcsService) DescribeKeyPairById(ctx context.Context, keyId string) (keyPair *ccs.KeyPair, err error) {
	var request = ccs.NewDescribeKeyPairsRequest()
	request.KeyIds = []string{keyId}
	response, err := s.client.WithCcsClient().DescribeKeyPairs(request)
	defer common2.LogApiRequest(ctx, "DescribeKeyPair", request, response, err)
	if err != nil {
		return nil, err
	}

	if len(response.Response.DataSet) < 1 {
		return
	}
	keyPair = response.Response.DataSet[0]
	return
}

func (s *CcsService) DeleteKeyPair(keyId string) error {
	request := ccs.NewDeleteKeyPairsRequest()
	request.KeyIds = []string{keyId}

	_, err := s.client.WithCcsClient().DeleteKeyPairs(request)
	if err != nil {
		log.Printf("[CRITAL] api[%s] fail, request body [%s], reason[%s]\n",
			request.GetAction(), keyId, err.Error())
		return err
	}

	return nil
}

func (s *CcsService) ModifyKeyPair(ctx context.Context, keyId string, keyDesc *string) error {
	var request = ccs.NewModifyKeyPairAttributeRequest()
	request.KeyId = &keyId
	request.KeyDescription = keyDesc
	response, e := s.client.WithCcsClient().ModifyKeyPairAttribute(request)
	common2.LogApiRequest(ctx, "ModifyKeyPair", request, response, e)
	if e != nil {
		return e
	}
	return nil
}

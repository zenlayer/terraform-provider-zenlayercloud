package zenlayercloud

import (
	"context"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	zga "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zga20230706"
)

type ZgaService struct {
	client *connectivity.ZenlayerCloudClient
}

func NewZgaService(client *connectivity.ZenlayerCloudClient) *ZgaService {
	return &ZgaService{
		client: client,
	}
}

func (s *ZgaService) DescribeOriginRegions(ctx context.Context) ([]zga.Region, error) {
	request := zga.NewDescribeOriginRegionsRequest()
	response, err := s.client.WithZgaClient().DescribeOriginRegions(request)
	common.LogApiRequest(ctx, "DescribeOriginRegions", request, response, err)
	if err != nil {
		return nil, err
	}
	return response.Response.RegionSet, nil
}

func (s *ZgaService) DescribeAccelerateRegions(ctx context.Context, originRegionId string) ([]*zga.Region, error) {
	request := zga.NewDescribeAccelerateRegionsRequest()
	request.OriginRegionId = originRegionId
	response, err := s.client.WithZgaClient().DescribeAccelerateRegions(request)
	common.LogApiRequest(ctx, "DescribeAccelerateRegions", request, response, err)
	if err != nil {
		return nil, err
	}
	return response.Response.RegionSet, nil
}

func (s *ZgaService) DeleteCertificatesById(ctx context.Context, certificateId string) error {
	request := zga.NewDeleteCertificateRequest()
	request.CertificateId = certificateId
	response, err := s.client.WithZgaClient().DeleteCertificate(request)
	common.LogApiRequest(ctx, request.GetAction(), request, response, err)
	if err != nil {
		return err
	}
	return nil
}

func (s *ZgaService) DescribeCertificateById(ctx context.Context, certificateId string) (*zga.CertificateInfo, error) {
	request := zga.NewDescribeCertificatesRequest()
	request.CertificateIds = []string{certificateId}
	response, err := s.client.WithZgaClient().DescribeCertificates(request)
	common.LogApiRequest(ctx, request.GetAction(), request, response, err)
	if err != nil {
		return nil, err
	} else if len(response.Response.DataSet) == 0 {
		return nil, nil
	}
	return response.Response.DataSet[0], nil
}

func (s *ZgaService) DescribeCertificatesByFilter(ctx context.Context, filter *CertificatesFilter) ([]*zga.CertificateInfo, error) {
	queryFunc := func(ctx context.Context, pageNum, pageSize int) (items []*zga.CertificateInfo, total int, err error) {
		request := convertCertificatesFilter(filter)
		request.PageSize = pageSize
		request.PageNum = pageNum
		response, err := s.client.WithZgaClient().DescribeCertificates(request)
		common.LogApiRequest(ctx, request.GetAction(), request, response, err)
		if err != nil {
			return nil, 0, err
		}
		return response.Response.DataSet, response.Response.TotalCount, nil
	}

	result, err := common.QueryAllPaginatedResource(ctx, queryFunc)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func convertCertificatesFilter(filter *CertificatesFilter) *zga.DescribeCertificatesRequest {
	request := zga.NewDescribeCertificatesRequest()
	if len(filter.CertificateIds) > 0 {
		request.CertificateIds = filter.CertificateIds
	}
	if filter.CertificateLabel != "" {
		request.CertificateLabel = filter.CertificateLabel
	}
	if filter.DnsName != "" {
		request.San = filter.DnsName
	}
	if filter.ResourceGroupId != "" {
		request.ResourceGroupId = filter.ResourceGroupId
	}
	if filter.Expired != nil {
		request.Expired = filter.Expired
	}
	return request
}

func (s *ZgaService) AcceleratorStateRefreshFunc(ctx context.Context, acceleratorId string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeAcceleratorById(ctx, acceleratorId)
		if err != nil {
			return nil, "", err
		}

		if object == nil {
			// Set this to nil as if we didn't find anything.
			return nil, "", nil
		}
		for _, failState := range failStates {
			if object.AcceleratorStatus == failState {
				return object, object.AcceleratorStatus, common.Error("Failed to reach target status. Last status: %s.", object.AcceleratorStatus)
			}
		}
		return object, object.AcceleratorStatus, nil
	}
}

func (s *ZgaService) DescribeAcceleratorById(ctx context.Context, acceleratorId string) (*zga.AcceleratorInfo, error) {
	request := zga.NewDescribeAcceleratorsRequest()
	request.AcceleratorIds = []string{acceleratorId}
	response, err := s.client.WithZgaClient().DescribeAccelerators(request)
	common.LogApiRequest(ctx, request.GetAction(), request, response, err)
	if err != nil {
		return nil, err
	} else if len(response.Response.DataSet) == 0 {
		return nil, nil
	}
	return response.Response.DataSet[0], nil
}

func (s *ZgaService) DescribeAcceleratorsByFilter(ctx context.Context, filter *AcceleratorsFilter) ([]*zga.AcceleratorInfo, error) {
	queryFunc := func(ctx context.Context, pageNum, pageSize int) (items []*zga.AcceleratorInfo, total int, err error) {
		request := convertAcceleratorsFilter(filter)
		request.PageSize = pageSize
		request.PageNum = pageNum
		response, err := s.client.WithZgaClient().DescribeAccelerators(request)
		common.LogApiRequest(ctx, request.GetAction(), request, response, err)
		if err != nil {
			return nil, 0, err
		}
		return response.Response.DataSet, response.Response.TotalCount, nil
	}

	result, err := common.QueryAllPaginatedResource(ctx, queryFunc)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func convertAcceleratorsFilter(filter *AcceleratorsFilter) *zga.DescribeAcceleratorsRequest {
	request := zga.NewDescribeAcceleratorsRequest()
	if len(filter.AcceleratorIds) > 0 {
		request.AcceleratorIds = filter.AcceleratorIds
	}
	if filter.AcceleratorName != "" {
		request.AcceleratorName = filter.AcceleratorName
	}
	if filter.AcceleratorStatus != "" {
		request.AcceleratorStatus = filter.AcceleratorStatus
	}
	if filter.AccelerateRegionId != "" {
		request.AccelerateRegionId = filter.AccelerateRegionId
	}
	if filter.Vip != "" {
		request.Vip = filter.Vip
	}
	if filter.Domain != "" {
		request.Domain = filter.Domain
	}
	if filter.Origin != "" {
		request.Origin = filter.Origin
	}
	if filter.OriginRegionId != "" {
		request.OriginRegionId = filter.OriginRegionId
	}
	if filter.Cname != "" {
		request.Cname = filter.Cname
	}
	if filter.ResourceGroupId != "" {
		request.ResourceGroupId = filter.ResourceGroupId
	}
	return request
}

func (s *ZgaService) ModifyAcceleratorAccessControl(ctx context.Context, acceleratorId string, rules []zga.AccessControlRule) error {
	request := zga.NewModifyAcceleratorAccessControlRequest()
	request.AcceleratorId = acceleratorId
	request.AccessControlRules = rules
	response, err := s.client.WithZgaClient().ModifyAcceleratorAccessControl(request)
	common.LogApiRequest(ctx, request.GetAction(), request, response, err)
	if err != nil {
		return err
	}
	return nil
}

func (s *ZgaService) OpenAcceleratorAccessControl(ctx context.Context, acceleratorId string) error {
	request := zga.NewOpenAcceleratorAccessControlRequest()
	request.AcceleratorId = acceleratorId
	response, err := s.client.WithZgaClient().OpenAcceleratorAccessControl(request)
	common.LogApiRequest(ctx, request.GetAction(), request, response, err)
	if err != nil {
		return err
	}
	return nil
}

func (s *ZgaService) CloseAcceleratorAccessControl(ctx context.Context, acceleratorId string) error {
	request := zga.NewCloseAcceleratorAccessControlRequest()
	request.AcceleratorId = acceleratorId
	response, err := s.client.WithZgaClient().CloseAcceleratorAccessControl(request)
	common.LogApiRequest(ctx, request.GetAction(), request, response, err)
	if err != nil {
		return err
	}
	return nil
}

func (s *ZgaService) ModifyAcceleratorName(ctx context.Context, acceleratorId string, name string) error {
	request := zga.NewModifyAcceleratorNameRequest()
	request.AcceleratorId = acceleratorId
	request.AcceleratorName = name
	response, err := s.client.WithZgaClient().ModifyAcceleratorName(request)
	common.LogApiRequest(ctx, request.GetAction(), request, response, err)
	if err != nil {
		return err
	}
	return nil
}

func (s *ZgaService) ModifyAcceleratorCertificateId(ctx context.Context, acceleratorId string, certificateId string) error {
	request := zga.NewModifyAcceleratorCertificateRequest()
	request.AcceleratorId = acceleratorId
	request.CertificateId = certificateId
	response, err := s.client.WithZgaClient().ModifyAcceleratorCertificate(request)
	common.LogApiRequest(ctx, request.GetAction(), request, response, err)
	if err != nil {
		return err
	}
	return nil
}

func (s *ZgaService) ModifyAcceleratorDomain(ctx context.Context, acceleratorId string, domain zga.Domain) error {
	request := zga.NewModifyAcceleratorDomainRequest()
	request.AcceleratorId = acceleratorId
	request.Domain = domain
	response, err := s.client.WithZgaClient().ModifyAcceleratorDomain(request)
	common.LogApiRequest(ctx, request.GetAction(), request, response, err)
	if err != nil {
		return err
	}
	return nil
}

func (s *ZgaService) ModifyAcceleratorOrigin(ctx context.Context, acceleratorId string, origin zga.Origin) error {
	request := zga.NewModifyAcceleratorOriginRequest()
	request.AcceleratorId = acceleratorId
	request.Origin = origin
	response, err := s.client.WithZgaClient().ModifyAcceleratorOrigin(request)
	common.LogApiRequest(ctx, request.GetAction(), request, response, err)
	if err != nil {
		return err
	}
	return nil
}

func (s *ZgaService) ModifyAcceleratorAccRegions(ctx context.Context, acceleratorId string, accRegions []zga.AccelerateRegion) error {
	request := zga.NewModifyAcceleratorAccRegionRequest()
	request.AcceleratorId = acceleratorId
	request.AccelerateRegions = accRegions
	response, err := s.client.WithZgaClient().ModifyAcceleratorAccRegion(request)
	common.LogApiRequest(ctx, request.GetAction(), request, response, err)
	if err != nil {
		return err
	}
	return nil
}

func (s *ZgaService) ModifyAcceleratorListener(ctx context.Context, acceleratorId string, l4Listeners []*zga.AccelerationRuleL4Listener, l7Listener []*zga.AccelerationRuleL7Listener) error {
	request := zga.NewModifyAcceleratorRuleRequest()
	request.AcceleratorId = acceleratorId
	request.L4Listeners = l4Listeners
	request.L7Listeners = l7Listener
	response, err := s.client.WithZgaClient().ModifyAcceleratorRule(request)
	common.LogApiRequest(ctx, request.GetAction(), request, response, err)
	if err != nil {
		return err
	}
	return nil
}

func (s *ZgaService) ModifyAcceleratorProtocolOpts(ctx context.Context, acceleratorId string, protocolOps zga.AccelerationRuleProtocolOpts) error {
	request := zga.NewModifyAcceleratorProtocolOptsRequest()
	request.AcceleratorId = acceleratorId
	request.ProtocolOpts = protocolOps
	response, err := s.client.WithZgaClient().ModifyAcceleratorProtocolOpts(request)
	common.LogApiRequest(ctx, request.GetAction(), request, response, err)
	if err != nil {
		return err
	}
	return nil
}

func (s *ZgaService) ModifyAcceleratorHealthCheck(ctx context.Context, acceleratorId string, healthCheck zga.HealthCheck) error {
	request := zga.NewModifyAcceleratorHealthCheckRequest()
	request.AcceleratorId = acceleratorId
	request.HealthCheck = healthCheck
	response, err := s.client.WithZgaClient().ModifyAcceleratorHealthCheck(request)
	common.LogApiRequest(ctx, request.GetAction(), request, response, err)
	if err != nil {
		return err
	}
	return nil
}

func (s *ZgaService) DeleteAcceleratorById(ctx context.Context, acceleratorId string) error {
	request := zga.NewDeleteAcceleratorRequest()
	request.AcceleratorId = acceleratorId
	response, err := s.client.WithZgaClient().DeleteAccelerator(request)
	common.LogApiRequest(ctx, request.GetAction(), request, response, err)
	if err != nil {
		return err
	}
	return nil
}

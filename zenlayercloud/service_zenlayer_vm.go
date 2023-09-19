package zenlayercloud

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	vm "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/vm20230313"
	"log"
	"math"
	"sync"
)

type VmService struct {
	client *connectivity.ZenlayerCloudClient
}

func (s *VmService) DescribeSecurityGroupsByFilter(filter *SecurityGroupFilter) (securityGroups []*vm.SecurityGroupInfo, err error) {
	request := convertSecurityGroupFilterRequest(filter)

	var limit = 100
	request.PageSize = limit
	request.PageNum = 1
	response, err := s.client.WithVmClient().DescribeSecurityGroups(request)

	if err != nil {
		log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
			request.GetAction(), toJsonString(request), err.Error())
		return
	}
	if response == nil || len(response.Response.DataSet) < 1 {
		return
	}

	securityGroups = response.Response.DataSet
	num := int(math.Ceil(float64(response.Response.TotalCount)/float64(limit))) - 1
	if num == 0 {
		return securityGroups, nil
	}
	maxConcurrentNum := 50
	g := NewGoRoutine(maxConcurrentNum)
	wg := sync.WaitGroup{}

	var securityGroupList = make([]interface{}, num)

	for i := 0; i < num; i++ {
		wg.Add(1)
		value := i
		goFunc := func() {
			request := convertSecurityGroupFilterRequest(filter)

			request.PageNum = value + 2
			request.PageSize = limit

			response, err := s.client.WithVmClient().DescribeSecurityGroups(request)
			if err != nil {
				log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
					request.GetAction(), toJsonString(request), err.Error())
				return
			}
			log.Printf("[DEBUG] Api[%s] success, request body [%s], response body [%s]\n",
				request.GetAction(), toJsonString(request), toJsonString(response))

			securityGroupList[value] = response.Response.DataSet

			wg.Done()
			log.Printf("[DEBUG] thread %d finished", value)
		}
		g.Run(goFunc)
	}
	wg.Wait()

	log.Printf("[DEBUG] DescribeSecurityGroups request finished")
	for _, v := range securityGroupList {
		securityGroups = append(securityGroups, v.([]*vm.SecurityGroupInfo)...)
	}
	log.Printf("[DEBUG] transfer SecurityGroup finished")
	return
}

func (s *VmService) DescribeSecurityGroupById(ctx context.Context, securityGroupId string) (securityGroup *vm.SecurityGroupInfo, err error) {
	request := vm.NewDescribeSecurityGroupsRequest()
	request.SecurityGroupIds = []string{securityGroupId}

	var response *vm.DescribeSecurityGroupsResponse

	defer logApiRequest(ctx, "DescribeSecurityGroups", request, response, err)

	response, err = s.client.WithVmClient().DescribeSecurityGroups(request)

	if err != nil {
		return
	}

	if len(response.Response.DataSet) < 1 {
		return
	}
	securityGroup = response.Response.DataSet[0]
	return
}

func (s *VmService) DeleteSecurityGroup(ctx context.Context, securityGroupId string) (err error) {
	request := vm.NewDeleteSecurityGroupRequest()
	request.SecurityGroupId = securityGroupId
	response, err := s.client.WithVmClient().DeleteSecurityGroup(request)
	defer logApiRequest(ctx, "DeleteSecurityGroup", request, response, err)
	return
}

func (s *VmService) ModifySecurityGroupAttribute(ctx context.Context, securityGroupId string, securityGroupName string, description string) error {
	request := vm.NewModifySecurityGroupsAttributeRequest()
	request.SecurityGroupIds = []string{securityGroupId}
	request.SecurityGroupName = securityGroupName
	request.Description = &description
	response, err := s.client.WithVmClient().ModifySecurityGroupsAttribute(request)
	defer logApiRequest(ctx, "ModifySecurityGroupsAttribute", request, response, err)
	return err
}

func convertSecurityGroupFilterRequest(filter *SecurityGroupFilter) (request *vm.DescribeSecurityGroupsRequest) {
	request = vm.NewDescribeSecurityGroupsRequest()

	if filter.Name != "" {
		request.SecurityGroupName = filter.Name
	}
	if filter.SecurityGroupId != "" {
		request.SecurityGroupIds = []string{filter.SecurityGroupId}
	}
	return
}

func (s *VmService) DescribeSecurityGroupRule(ruleId string) (securityGroupId string, rule *vm.RuleInfo, err error) {
	info, err := parseSecurityGroupRuleId(ruleId)
	if err != nil {
		return
	}

	request := vm.NewDescribeSecurityGroupsRequest()
	request.SecurityGroupIds = []string{info.SecurityGroupId}
	response, err := s.client.WithVmClient().DescribeSecurityGroups(request)
	if err != nil {
		log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
			request.GetAction(), toJsonString(request), err.Error())
		return
	}

	if response == nil || len(response.Response.DataSet) < 1 {
		return
	}

	ruleSet := response.Response.DataSet[0].RuleInfos

	if ruleSet == nil || len(ruleSet) < 1 {
		return
	}

	for _, rl := range ruleSet {
		if compareRuleAndSecurityGroupInfo(rl, info) {
			rule = rl
			break
		}
	}

	if rule == nil {
		log.Printf("[DEBUG]%s can't find security group rule, maybe user modify rules on web console", ruleId)
		return
	}

	return info.SecurityGroupId, rule, nil
}

func (s *VmService) CreateSecurityGroupRule(ctx context.Context, info securityGroupRuleBasicInfo) (ruleId string, err error) {
	request := vm.NewAuthorizeSecurityGroupRuleRequest()
	request.SecurityGroupId = info.SecurityGroupId
	request.IpProtocol = info.IpProtocol
	request.Policy = info.Policy
	request.PortRange = info.PortRange
	request.CidrIp = info.CidrIp
	request.Direction = info.Direction
	response, ret := s.client.WithVmClient().AuthorizeSecurityGroupRule(request)
	if ret != nil {
		tflog.Info(ctx, "Fail to authorize security group rule.", map[string]interface{}{
			"action":  request.GetAction(),
			"request": toJsonString(request),
			"err":     ret.Error(),
		})
		err = ret
	}
	if err != nil {
		return "", err
	}

	tflog.Info(ctx, "Authorize security group rule success", map[string]interface{}{
		"action":   request.GetAction(),
		"request":  toJsonString(request),
		"response": toJsonString(response),
	})

	ruleId, err = buildSecurityGroupRuleId(info)
	if err != nil {
		return "", fmt.Errorf("build rule id error, reason: %v", err)
	}

	return ruleId, err
}

func (s *VmService) DescribeImageById(ctx context.Context, imageId string) (image *vm.DescribeImageResponseParams, err error) {
	var request = vm.NewDescribeImageRequest()
	request.ImageId = imageId
	response, err := s.client.WithVmClient().DescribeImage(request)
	logApiRequest(ctx, "DescribeImage", request, response, err)
	if err != nil {
		return nil, err
	}

	return response.Response, nil
}

func (s *VmService) ModifyImage(ctx context.Context, imageId string, imageName string, imageDesc string) error {
	var request = vm.NewModifyImagesAttributesRequest()
	request.ImageIds = []string{imageId}
	//request.Image = imageName
	request.ImageDescription = imageDesc
	response, e := s.client.WithVmClient().ModifyImagesAttributes(request)
	logApiRequest(ctx, "DescribeImages", request, response, e)
	if e != nil {
		return e
	}
	return nil
}

func (s *VmService) DeleteImage(ctx context.Context, imageId string) error {
	request := vm.NewDeleteImagesRequest()
	request.ImageIds = []string{imageId}

	_, err := s.client.WithVmClient().DeleteImages(request)
	if err != nil {
		log.Printf("[CRITAL] api[%s] fail, request body [%s], reason[%s]\n",
			request.GetAction(), imageId, err.Error())
		return err
	}

	return nil
}

func (s *VmService) DescribeImagesByFilter(filter *VmImageFilter) (images []*vm.ImageInfo, err error) {
	request := convertVmImageFilter(filter)
	var limit = 100
	request.PageSize = limit
	request.PageNum = 1
	response, err := s.client.WithVmClient().DescribeImages(request)

	if err != nil {
		log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
			request.GetAction(), toJsonString(request), err.Error())
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
	g := NewGoRoutine(maxConcurrentNum)
	wg := sync.WaitGroup{}

	var imageSetList = make([]interface{}, num)

	for i := 0; i < num; i++ {
		wg.Add(1)
		value := i
		goFunc := func() {
			request := convertVmImageFilter(filter)

			request.PageNum = value + 2
			request.PageSize = limit

			response, err := s.client.WithVmClient().DescribeImages(request)
			if err != nil {
				log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
					request.GetAction(), toJsonString(request), err.Error())
				return
			}
			log.Printf("[DEBUG] Api[%s] success, request body [%s], response body [%s]\n",
				request.GetAction(), toJsonString(request), toJsonString(response))

			imageSetList[value] = response.Response.DataSet

			wg.Done()
			log.Printf("[DEBUG] thread %d finished", value)
		}
		g.Run(goFunc)
	}
	wg.Wait()

	log.Printf("[DEBUG] DescribeImages request finished")
	for _, v := range imageSetList {
		images = append(images, v.([]*vm.ImageInfo)...)
	}
	log.Printf("[DEBUG] transfer images finished")
	return
}

func (s *VmService) ImageStateRefreshFunc(ctx context.Context, imageId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		image, err := s.DescribeImageById(ctx, imageId)
		if err != nil {
			return nil, "", err
		}

		if image == nil {
			// Set this to nil as if we didn't find anything.
			return nil, "", nil
		}

		return image, image.ImageStatus, nil
	}
}

func (s *VmService) DescribeSubnets(ctx context.Context, filter *VmSubnetFilter) (subnets []*vm.SubnetInfo, err error) {
	request := convertVmSubnetFilter(filter)
	var limit = 100
	request.PageSize = limit
	request.PageNum = 1
	response, err := s.client.WithVmClient().DescribeSubnets(request)

	if err != nil {
		log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
			request.GetAction(), toJsonString(request), err.Error())
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
	g := NewGoRoutine(maxConcurrentNum)
	wg := sync.WaitGroup{}

	var subnetList = make([]interface{}, num)

	for i := 0; i < num; i++ {
		wg.Add(1)
		value := i
		goFunc := func() {
			request := convertVmSubnetFilter(filter)

			request.PageNum = value + 2
			request.PageSize = limit

			response, err := s.client.WithVmClient().DescribeSubnets(request)
			if err != nil {
				log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
					request.GetAction(), toJsonString(request), err.Error())
				return
			}
			log.Printf("[DEBUG] Api[%s] success, request body [%s], response body [%s]\n",
				request.GetAction(), toJsonString(request), toJsonString(response))

			subnetList[value] = response.Response.DataSet

			wg.Done()
			log.Printf("[DEBUG] thread %d finished", value)
		}
		g.Run(goFunc)
	}
	wg.Wait()

	log.Printf("[DEBUG] DescribeSubnets request finished")
	for _, v := range subnetList {
		subnets = append(subnets, v.([]*vm.SubnetInfo)...)
	}
	log.Printf("[DEBUG] transfer subnets finished")
	return
}

func convertVmSubnetFilter(filter *VmSubnetFilter) *vm.DescribeSubnetsRequest {

	request := vm.NewDescribeSubnetsRequest()
	if filter.ZoneId != "" {
		request.ZoneId = filter.ZoneId
	}
	if filter.CidrBlock != "" {
		request.CidrBlock = filter.CidrBlock
	}
	if filter.SubnetId != "" {
		request.SubnetIds = []string{filter.SubnetId}
	}
	if filter.SubnetName != "" {
		request.SubnetName = filter.SubnetName
	}
	return request
}

func (s *VmService) DeleteSubnet(ctx context.Context, subnetId string) (err error) {
	request := vm.NewDeleteSubnetRequest()
	request.SubnetId = subnetId
	response, err := s.client.WithVmClient().DeleteSubnet(request)
	logApiRequest(ctx, "DeleteSubnet", request, response, err)
	return
}

func (s *VmService) ModifySubnetName(ctx context.Context, subnetId string, subnetName string) error {
	request := vm.NewModifySubnetsAttributeRequest()
	request.SubnetIds = []string{subnetId}
	request.SubnetName = subnetName
	response, err := s.client.WithVmClient().ModifySubnetsAttribute(request)
	logApiRequest(ctx, "ModifySubnetsAttribute", request, response, err)
	return err
}

func (s *VmService) DescribeSubnetById(ctx context.Context, subnetId string) (subnet *vm.SubnetInfo, err error) {
	request := vm.NewDescribeSubnetsRequest()
	request.SubnetIds = []string{subnetId}

	var response *vm.DescribeSubnetsResponse

	defer logApiRequest(ctx, "DescribeSubnets", request, response, err)

	response, err = s.client.WithVmClient().DescribeSubnets(request)

	if err != nil {
		return
	}

	if len(response.Response.DataSet) < 1 {
		return
	}
	subnet = response.Response.DataSet[0]
	return
}

func (s *VmService) SubnetStateRefreshFunc(ctx context.Context, subnetId string, failStates []string) resource.StateRefreshFunc {
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
				return object, object.SubnetStatus, Error("Failed to reach target status. Last status: %s.", object.SubnetStatus)
			}
		}

		return object, object.SubnetStatus, nil
	}
}

func (s *VmService) DescribeDisks(ctx context.Context, filter *VmDiskFilter) (disks []*vm.DiskInfo, err error) {
	request := convertDiskFilterRequest(filter)
	var limit = 100
	request.PageSize = limit
	request.PageNum = 1
	response, err := s.client.WithVmClient().DescribeDisks(request)
	defer logApiRequest(ctx, "DescribeDisks", request, response, err)

	if err != nil {
		log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
			request.GetAction(), toJsonString(request), err.Error())
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
	g := NewGoRoutine(maxConcurrentNum)
	wg := sync.WaitGroup{}

	var diskList = make([]interface{}, num)

	for i := 0; i < num; i++ {
		wg.Add(1)
		value := i
		goFunc := func() {
			request := convertDiskFilterRequest(filter)

			request.PageNum = value + 2
			request.PageSize = limit

			response, err := s.client.WithVmClient().DescribeDisks(request)
			if err != nil {
				log.Printf("[CRITAL] Api[%s] fail, request body [%s], error[%s]\n",
					request.GetAction(), toJsonString(request), err.Error())
				return
			}
			log.Printf("[DEBUG] Api[%s] success, request body [%s], response body [%s]\n",
				request.GetAction(), toJsonString(request), toJsonString(response))

			diskList[value] = response.Response.DataSet

			wg.Done()
			log.Printf("[DEBUG] thread %d finished", value)
		}
		g.Run(goFunc)
	}
	wg.Wait()

	log.Printf("[DEBUG] DescribeDisks request finished")
	for _, v := range diskList {
		disks = append(disks, v.([]*vm.DiskInfo)...)
	}
	log.Printf("[DEBUG] transfer disks finished")
	return
}

func (s *VmService) DiskStateRefreshFunc(ctx context.Context, diskId string, failStates []string) resource.StateRefreshFunc {
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
				return object, object.DiskStatus, Error("Failed to reach target status. Last status: %s.", object.DiskStatus)
			}
		}

		return object, object.DiskStatus, nil
	}
}

func (s *VmService) DescribeDiskById(ctx context.Context, diskId string) (*vm.DiskInfo, error) {
	request := vm.NewDescribeDisksRequest()
	request.DiskIds = []string{diskId}

	response, err := s.client.WithVmClient().DescribeDisks(request)

	defer logApiRequest(ctx, "DescribeDisks", request, response, err)
	if err != nil {
		return nil, err
	}

	if len(response.Response.DataSet) < 1 {
		return nil, nil
	}
	diskInfo := response.Response.DataSet[0]
	return diskInfo, nil
}

func (s *VmService) ModifyDiskName(ctx context.Context, diskId string, diskName string) error {
	request := vm.NewModifyDisksAttributesRequest()
	request.DiskIds = []string{diskId}
	request.DiskName = diskName
	response, err := s.client.WithVmClient().ModifyDisksAttributes(request)
	defer logApiRequest(ctx, "ModifyDisksAttributes", request, response, err)
	return err
}

func (s *VmService) ModifyDiskResourceGroupId(ctx context.Context, diskId string, resourceGroupId string) error {
	request := vm.NewModifyDisksResourceGroupRequest()
	request.DiskIds = []string{diskId}
	request.ResourceGroupId = resourceGroupId
	response, err := s.client.WithVmClient().ModifyDisksResourceGroup(request)
	defer logApiRequest(ctx, "ModifyDisksResourceGroup", request, response, err)
	return err
}

func (s *VmService) DeleteDisk(ctx context.Context, diskId string) (err error) {
	request := vm.NewTerminateDiskRequest()
	request.DiskId = diskId
	response, err := s.client.WithVmClient().TerminateDisk(request)
	defer logApiRequest(ctx, "TerminateInstance", request, response, err)

	if err != nil {
		if sdkError, ok := err.(*common.ZenlayerCloudSdkError); ok {
			if sdkError.Code == "UNSUPPORTED_OPERATION_DISK_BEING_RECYCLE" {
				return nil
			}
		}
		return
	}
	return
}

func (s *VmService) ReleaseDisk(ctx context.Context, diskId string) (err error) {
	request := vm.NewReleaseDiskRequest()
	request.DiskId = diskId
	response, err := s.client.WithVmClient().ReleaseDisk(request)
	defer logApiRequest(ctx, "ReleaseDisk", request, response, err)
	return
}

func convertDiskFilterRequest(filter *VmDiskFilter) *vm.DescribeDisksRequest {
	request := vm.NewDescribeDisksRequest()
	if filter.Portable != nil {
		request.Portable = filter.Portable
	}
	if filter.Id != "" {
		request.DiskIds = []string{filter.Id}
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

func convertVmImageFilter(filter *VmImageFilter) *vm.DescribeImagesRequest {
	image := vm.NewDescribeImagesRequest()
	image.ZoneId = filter.zoneId
	image.ImageType = filter.imageType
	image.Category = filter.category
	image.OsType = filter.osType

	if filter.imageId != "" {
		image.ImageIds = []string{filter.imageId}
	}
	return image
}

func (s *VmService) DescribeZones(ctx context.Context) (zones []*vm.ZoneInfo, err error) {
	request := vm.NewDescribeZonesRequest()
	response, err := s.client.WithVmClient().DescribeZones(request)
	logApiRequest(ctx, "DescribeZones", request, response, err)
	if err != nil {
		return
	}
	zones = response.Response.ZoneSet
	return
}

func (s *VmService) DescribeInstanceById(ctx context.Context, instanceId string) (instance *vm.InstanceInfo, err error) {
	request := vm.NewDescribeInstancesRequest()
	request.InstanceIds = []string{instanceId}

	response, err := s.client.WithVmClient().DescribeInstances(request)

	defer logApiRequest(ctx, "DescribeInstances", request, response, err)
	if err != nil {
		return
	}

	if len(response.Response.DataSet) < 1 {
		return
	}
	instance = response.Response.DataSet[0]
	return
}

func (s *VmService) DeleteInstance(ctx context.Context, instanceId string) (err error) {
	// 判断已经终止 OPERATION_DENIED_INSTANCE_RECYCLED,
	request := vm.NewTerminateInstanceRequest()
	request.InstanceId = instanceId
	response, err := s.client.WithVmClient().TerminateInstance(request)
	defer logApiRequest(ctx, "TerminateInstance", request, response, err)

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

func (s *VmService) DestroyInstance(ctx context.Context, instanceId string) (err error) {
	request := vm.NewReleaseInstancesRequest()
	request.InstanceIds = []string{instanceId}
	response, err := s.client.WithVmClient().ReleaseInstances(request)
	defer logApiRequest(ctx, "ReleaseInstances", request, response, err)
	if err != nil {
		return
	}
	return
}

func (s *VmService) InstanceStateRefreshFunc(ctx context.Context, instanceId string, failStates []string) resource.StateRefreshFunc {
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
				return object, object.InstanceStatus, Error("Failed to reach target status. Last status: %s.", object.InstanceStatus)
			}
		}

		return object, object.InstanceStatus, nil
	}
}

func (s *VmService) ModifyInstanceName(ctx context.Context, instanceId string, instanceName string) error {
	request := vm.NewModifyInstancesAttributeRequest()
	request.InstanceIds = []string{instanceId}
	request.InstanceName = instanceName
	response, err := s.client.WithVmClient().ModifyInstancesAttribute(request)
	defer logApiRequest(ctx, "ModifyInstancesAttribute", request, response, err)
	return err
}

func (s *VmService) ModifyInstanceResourceGroup(ctx context.Context, instanceId string, resourceGroupId string) error {
	request := vm.NewModifyInstancesResourceGroupRequest()
	request.InstanceIds = []string{instanceId}
	request.ResourceGroupId = resourceGroupId
	response, err := s.client.WithVmClient().ModifyInstancesResourceGroup(request)
	defer logApiRequest(ctx, "ModifyInstancesResourceGroup", request, response, err)

	if err != nil {
		return err
	}

	return err
}

func (s *VmService) updateInstanceInternetMaxBandwidthOut(ctx context.Context, instanceId string, internetBandwidthOut int) error {
	request := vm.NewModifyInstanceBandwidthRequest()
	request.InstanceId = instanceId
	request.InternetMaxBandwidthOut = internetBandwidthOut
	response, err := s.client.WithVmClient().ModifyInstanceBandwidth(request)
	defer logApiRequest(ctx, "ModifyInstanceBandwidth", request, response, err)
	if err != nil {
		return err
	}
	return nil
}

type VmNetworkStateCondition interface {
	matchFail(status *vm.DescribeInstanceInternetStatusResponseParams) bool
	matchOk(status *vm.DescribeInstanceInternetStatusResponseParams) bool
}

const VmNetworkStatusOK = "OK"
const VmNetworkStatusFail = "Fail"
const VmNetworkStatusPending = "Pending"

func (s *VmService) InstanceNetworkStateRefreshFunc(ctx context.Context, instanceId string, condition VmNetworkStateCondition) resource.StateRefreshFunc {

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
			return internetStatus, VmNetworkStatusFail, Error("Failed to reach target status. Last internet status: %v.", internetStatus)
		}
		if condition.matchOk(internetStatus) {
			return internetStatus, VmNetworkStatusOK, nil
		}
		return nil, VmNetworkStatusPending, nil
	}
}

func (s *VmService) updateInstanceTrafficPackageSize(ctx context.Context, instanceId string, trafficPackageSize float64) error {

	request := vm.NewModifyInstanceTrafficPackageRequest()
	request.InstanceId = instanceId
	request.TrafficPackageSize = &trafficPackageSize
	response, err := s.client.WithVmClient().ModifyInstanceTrafficPackage(request)
	defer logApiRequest(ctx, "ModifyInstanceTrafficPackageSize", request, response, err)
	return err
}

func (s *VmService) resetInstancePassword(ctx context.Context, instanceId string, newPassword string) error {

	request := vm.NewResetInstancesPasswordRequest()
	request.InstanceIds = []string{instanceId}
	request.Password = newPassword
	response, err := s.client.WithVmClient().ResetInstancesPassword(request)
	defer logApiRequest(ctx, "ResetInstancesPassword", request, response, err)
	return err
}

func (s *VmService) resetInstance(ctx context.Context, request *vm.ResetInstanceRequest) error {

	response, err := s.client.WithVmClient().ResetInstance(request)
	logApiRequest(ctx, "ReinstallInstance", request, response, err)
	return err
}

func (s *VmService) shutdownInstance(ctx context.Context, instanceId string) error {
	request := vm.NewStopInstancesRequest()
	request.InstanceIds = []string{instanceId}
	response, err := s.client.WithVmClient().StopInstances(request)
	logApiRequest(ctx, "ShutdownInstance", request, response, err)
	return err
}

func (s *VmService) DescribeInstanceInternetStatus(ctx context.Context, instanceId string) (*vm.DescribeInstanceInternetStatusResponseParams, error) {
	request := vm.NewDescribeInstanceInternetStatusRequest()
	request.InstanceId = instanceId
	status, err := s.client.WithVmClient().DescribeInstanceInternetStatus(request)
	if err != nil {
		return nil, err
	}
	return status.Response, nil
}

func (s *VmService) DescribeKeyPairs(ctx context.Context, request *vm.DescribeKeyPairsRequest) (keyPairs []*vm.KeyPair, err error) {
	response, err := s.client.WithVmClient().DescribeKeyPairs(request)
	logApiRequest(ctx, "DescribeKeyPairs", request, response, err)
	if err != nil {
		return
	}
	keyPairs = response.Response.DataSet
	return
}

func (s *VmService) DescribeKeyPairById(ctx context.Context, keyId string) (keyPair *vm.KeyPair, err error) {
	var request = vm.NewDescribeKeyPairsRequest()
	request.KeyIds = []string{keyId}
	response, err := s.client.WithVmClient().DescribeKeyPairs(request)
	defer logApiRequest(ctx, "DescribeKeyPair", request, response, err)
	if err != nil {
		return nil, err
	}

	if len(response.Response.DataSet) < 1 {
		return
	}
	keyPair = response.Response.DataSet[0]
	return
}

func (s *VmService) DeleteKeyPair(keyId string) error {
	request := vm.NewDeleteKeyPairsRequest()
	request.KeyIds = []string{keyId}

	_, err := s.client.WithVmClient().DeleteKeyPairs(request)
	if err != nil {
		log.Printf("[CRITAL] api[%s] fail, request body [%s], reason[%s]\n",
			request.GetAction(), keyId, err.Error())
		return err
	}

	return nil
}

func (s *VmService) ModifyKeyPair(ctx context.Context, keyId string, keyDesc *string) error {
	var request = vm.NewModifyKeyPairAttributeRequest()
	request.KeyId = keyId
	request.KeyDescription = keyDesc
	response, e := s.client.WithVmClient().ModifyKeyPairAttribute(request)
	logApiRequest(ctx, "ModifyKeyPair", request, response, e)
	if e != nil {
		return e
	}
	return nil
}

func buildSecurityGroupRuleId(info securityGroupRuleBasicInfo) (ruleId string, err error) {
	b, err := json.Marshal(info)
	if err != nil {
		return "", err
	}

	log.Printf("[DEBUG] build rule is %s", string(b))

	return base64.StdEncoding.EncodeToString(b), nil
}

func parseSecurityGroupRuleId(ruleId string) (info securityGroupRuleBasicInfo, errRet error) {
	log.Printf("[DEBUG] parseSecurityGroupRuleId before: %v", ruleId)
	if b, err := base64.StdEncoding.DecodeString(ruleId); err == nil {
		errRet = json.Unmarshal(b, &info)
		return
	}
	log.Printf("[DEBUG] parseSecurityGroupRuleId after: %v", info)
	return
}

func compareRuleAndSecurityGroupInfo(rule *vm.RuleInfo, info securityGroupRuleBasicInfo) bool {
	if rule.Policy != info.Policy {
		return false
	}
	if rule.PortRange != info.PortRange && rule.PortRange != (info.PortRange+"/"+info.PortRange) {
		return false
	}
	if rule.Direction != info.Direction {
		return false
	}
	if rule.IpProtocol != info.IpProtocol {
		return false
	}
	if rule.CidrIp != info.CidrIp {
		return false
	}
	return true
}

func convertRuleInfo2RuleRequest(ruleInfo securityGroupRuleBasicInfo) (request *vm.RuleInfo) {
	request = &vm.RuleInfo{
		Direction:  ruleInfo.Direction,
		Policy:     ruleInfo.Policy,
		IpProtocol: ruleInfo.IpProtocol,
		PortRange:  ruleInfo.PortRange,
		CidrIp:     ruleInfo.CidrIp,
	}
	return
}

type securityGroupRuleBasicInfo struct {
	SecurityGroupId string `json:"security_group_id"`
	Policy          string `json:"policy"`
	CidrIp          string `json:"cidr_ip"`
	IpProtocol      string `json:"ip_protocol"`
	PortRange       string `json:"port_range"`
	Direction       string `json:"direction"`
}

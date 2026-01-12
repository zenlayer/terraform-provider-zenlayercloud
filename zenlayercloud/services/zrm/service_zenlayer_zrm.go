package zrm

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	common2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	zrm "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zrm20251014"
	"time"
)

func NewZrmService(client *connectivity.ZenlayerCloudClient) ZrmService {
	return ZrmService{client: client}
}

type ZrmService struct {
	client *connectivity.ZenlayerCloudClient
}

func (s *ZrmService) ModifyResourceTags(ctx context.Context, d *schema.ResourceData, resourceId string) error {

	addedTags, removedKeys := common.ParseTagChanges(d)

	request := zrm.NewModifyResourceTagsRequest()
	request.ResourceUuid = &resourceId

	// 设置需要添加或更新的标签
	if len(addedTags) > 0 {
		tags := make([]*zrm.Tag, 0, len(addedTags))
		for k, v := range addedTags {
			tagKey := k
			tagValue := v.(string)
			tags = append(tags, &zrm.Tag{
				Key:   &tagKey,
				Value: &tagValue,
			})
		}
		request.ReplaceTags = tags
	}

	// 设置需要删除的标签键
	if len(removedKeys) > 0 {
		request.DeleteTagKeys = removedKeys
	}

	// 发送API请求
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
		response, err := s.client.WithZrmClient().ModifyResourceTags(request)
		defer common.LogApiRequest(ctx, "ModifyResourceTags", request, response, err)

		if err != nil {
			return common.RetryError(ctx, err, common2.NetworkError, common.OperationTimeout)
		}
		return nil
	})

	return err
}

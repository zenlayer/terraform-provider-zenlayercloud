package common

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"reflect"
)

func TagsToMap(tags interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	if tags == nil {
		return result, nil
	}

	v := reflect.ValueOf(tags)

	// 处理指针
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return result, nil
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct or pointer to struct, got %v", v.Kind())
	}

	// 查找 Tags 字段
	tagsField := v.FieldByName("Tags")
	if !tagsField.IsValid() {
		return nil, fmt.Errorf("field 'Tags' not found")
	}

	// 处理指针类型的 Tags 字段
	tagsFieldVal := tagsField
	for tagsFieldVal.Kind() == reflect.Ptr {
		if tagsFieldVal.IsNil() {
			return result, nil
		}
		tagsFieldVal = tagsFieldVal.Elem()
	}

	if tagsFieldVal.Kind() != reflect.Slice {
		return nil, fmt.Errorf("field 'Tags' is not a slice")
	}

	// 遍历标签
	for i := 0; i < tagsFieldVal.Len(); i++ {
		tagItem := tagsFieldVal.Index(i)

		// 处理指针
		tagItemVal := tagItem
		for tagItemVal.Kind() == reflect.Ptr {
			if tagItemVal.IsNil() {
				continue
			}
			tagItemVal = tagItemVal.Elem()
		}

		if tagItemVal.Kind() != reflect.Struct {
			continue
		}

		// 获取 Key
		keyField := tagItemVal.FieldByName("Key")
		if !keyField.IsValid() {
			continue
		}

		keyVal := keyField
		for keyVal.Kind() == reflect.Ptr {
			if keyVal.IsNil() {
				continue
			}
			keyVal = keyVal.Elem()
		}

		if keyVal.Kind() != reflect.String {
			continue
		}

		key := keyVal.String()
		if key == "" {
			continue
		}

		// 获取 Value
		valueField := tagItemVal.FieldByName("Value")
		var value interface{}

		if valueField.IsValid() {
			valueVal := valueField
			for valueVal.Kind() == reflect.Ptr {
				if valueVal.IsNil() {
					break
				}
				valueVal = valueVal.Elem()
			}

			if valueVal.Kind() == reflect.String {
				value = valueVal.String()
			}
		}

		result[key] = value
	}

	return result, nil
}

func ParseTagChanges(d *schema.ResourceData) (map[string]interface{}, []string) {
	oraw, nraw := d.GetChange("tags")
	removedTags := oraw.(map[string]interface{})
	addedTags := nraw.(map[string]interface{})
	// Build the list of what to remove
	removedKeys := make([]string, 0)
	for key, _ := range removedTags {
		_, ok := addedTags[key]
		if !ok {
			// Delete it!
			removedKeys = append(removedKeys, key)
		}
	}

	return addedTags, removedKeys
}

package common


import (
"reflect"
"testing"
)

// 模拟第一个包的结构
type Tags1 struct {
	Tags []*Tag1 `json:"tags,omitempty"`
}

type Tag1 struct {
	Key   *string `json:"key,omitempty"`
	Value *string `json:"value,omitempty"`
}

// 模拟第二个包的结构
type Tags2 struct {
	Tags []*Tag2 `json:"tags,omitempty"`
}

type Tag2 struct {
	Key   *string `json:"key,omitempty"`
	Value *string `json:"value,omitempty"`
}

// 模拟结构相同但字段名不同的结构（用于测试失败情况）
type WrongTags struct {
	NotTags []*Tag1 `json:"not_tags,omitempty"`
}


// 测试V2版本函数（带错误返回）
func TestTagsToMap(t *testing.T) {
	// 测试正常情况
	t.Run("NormalCase", func(t *testing.T) {
		key1 := "env"
		value1 := "test"

		tags := &Tags1{
			Tags: []*Tag1{
				{Key: &key1, Value: &value1},
			},
		}

		result, err := TagsToMap(tags)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := map[string]interface{}{
			"env": "test",
		}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	// 测试非结构体输入
	t.Run("NonStructInput", func(t *testing.T) {
		_, err := TagsToMap("string input")
		if err == nil {
			t.Error("Expected error for non-struct input")
		}
	})

	// 测试缺少Tags字段
	t.Run("V2_MissingTagsField", func(t *testing.T) {
		type NoTags struct {
			Name string
		}

		_, err := TagsToMap(&NoTags{Name: "test"})
		if err == nil {
			t.Error("Expected error for missing Tags field")
		}
	})

	// 测试Tags字段不是切片
	t.Run("TagsNotSlice", func(t *testing.T) {
		type WrongTags struct {
			Tags string
		}

		_, err := TagsToMap(&WrongTags{Tags: "string"})
		if err == nil {
			t.Error("Expected error for non-slice Tags field")
		}
	})
}

// 性能测试
func BenchmarkTagsToMap(b *testing.B) {
	// 准备测试数据
	keys := []string{"key1", "key2", "key3", "key4", "key5"}
	values := []string{"value1", "value2", "value3", "value4", "value5"}

	tags := &Tags1{}
	for i := 0; i < 5; i++ {
		key := keys[i]
		value := values[i]
		tags.Tags = append(tags.Tags, &Tag1{Key: &key, Value: &value})
	}

	// 重置计时器
	b.ResetTimer()

	// 运行基准测试
	for i := 0; i < b.N; i++ {
		_, _ = TagsToMap(tags)
	}
}


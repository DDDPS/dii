package exporter

import (
	"DII/pkg/provider"
	"encoding/json"
	"fmt"

	"github.com/docker/go-units"
	"gopkg.in/yaml.v3"
)

func ExportYaml(imageMeta *provider.ImageMeta) (string, error) {
	if imageMeta == nil {
		return "", fmt.Errorf("imageMeta is nil")
	}

	// yaml.v3 默认不会扁平化嵌套结构体，所以需要先将嵌套结构体序列化为JSON
	jsonData, err := json.Marshal(imageMeta)
	if err != nil {
		return "", err
	}

	var generic interface{}
	// 反序列化JSON到通用结构中，以消除结构体类型的约束并扁平化字段
	if err := json.Unmarshal(jsonData, &generic); err != nil {
		return "", err
	}

	if m, ok := generic.(map[string]interface{}); ok {
		m["Size"] = units.HumanSize(float64(imageMeta.Size))
	}

	cleanedGeneric := cleanNumbers(generic)

	// 将通用结构序列化为 YAML
	yamlData, err := yaml.Marshal(cleanedGeneric)
	if err != nil {
		return "", err
	}

	return string(yamlData), nil
}

// cleanNumbers 递归地遍历 map 和 slice，将整数值的 float64 类型转换为 int64。
// 为了解决 JSON 反序列化到 interface{} 带来的副作用。
func cleanNumbers(val interface{}) interface{} {
	switch v := val.(type) {
	case map[string]interface{}:
		for k, item := range v {
			v[k] = cleanNumbers(item)
		}
		return v
	case []interface{}:
		for i, item := range v {
			v[i] = cleanNumbers(item)
		}
		return v
	case float64:
		if v == float64(int64(v)) {
			return int64(v)
		}
		return v
	default:
		return v
	}
}

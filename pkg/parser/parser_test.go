package parser_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tgagor/template-dockerfiles/pkg/config"
	"github.com/tgagor/template-dockerfiles/pkg/parser"
)

func loadConfig(file string) *config.Config {
	cfg, _ := config.Load(filepath.Join("../../tests", file))
	return cfg
}

func TestCombinationsCase1(t *testing.T) {
	inputs := loadConfig("test-1.yaml").Images

	expected := map[string]interface{}{
		"test-case-1": []map[string]interface{}{
			{
				"alpine":    "3.18",
				"multiword": "kuku ruku",
			},
			{
				"alpine":    "3.19",
				"multiword": "kuku ruku",
			},
			{
				"alpine":    "3.20",
				"multiword": "kuku ruku",
			},
		},
	}

	for image, cfg := range inputs {
		combinations := parser.GenerateVariableCombinations(cfg.Variables)
		assert.Equal(t, expected[image], combinations)
	}
}
func TestCombinationsCase2(t *testing.T) {
	inputs := loadConfig("test-2.yaml").Images

	expected := map[string]interface{}{
		"test-case-2": []map[string]interface{}{
			{
				"alpine": "3.18",
			},
			{
				"alpine": "3.19",
			},
			{
				"alpine": "3.20",
			},
		},
		"test-case-2b": []map[string]interface{}{
			{
				"alpine": "3.20",
			},
		},
	}

	for image, cfg := range inputs {
		combinations := parser.GenerateVariableCombinations(cfg.Variables)
		assert.Equal(t, expected[image], combinations)
	}
}

// redundant
// func TestCombinationsCase3(t *testing.T) {
// 	inputs := loadImagesFromConfig("test-3.yaml")

// 	expected := map[string]interface{}{
// 		"test-case-3": []map[string]interface{}{
// 			{
// 				"alpine": "3.18",
// 			},
// 			{
// 				"alpine": "3.19",
// 			},
// 			{
// 				"alpine": "3.20",
// 			},
// 		},
// 	}

// 	for image, cfg := range inputs {
// 		combinations := parser.GenerateVariableCombinations(cfg.Variables)
// 		assert.Equal(t, expected[image], combinations)
// 	}
// }

func TestCombinationsCase4(t *testing.T) {
	inputs := loadConfig("test-4.yaml").Images

	expected := map[string]interface{}{
		"test-case-4": []map[string]interface{}{
			{}, // one image with empty variables
		},
	}

	for image, cfg := range inputs {
		combinations := parser.GenerateVariableCombinations(cfg.Variables)
		assert.Equal(t, expected[image], combinations)
	}
}

func TestCombinationsCase5(t *testing.T) {
	inputs := loadConfig("test-5.yaml").Images

	expected := map[string]interface{}{
		"test-case-5": []map[string]interface{}{
			{
				"alpine": 3,
			},
		},
	}

	for image, cfg := range inputs {
		combinations := parser.GenerateVariableCombinations(cfg.Variables)
		assert.Equal(t, expected[image], combinations)
	}
}

// redundant
// func TestCombinationsCase6(t *testing.T) {
// 	inputs := loadImagesFromConfig("test-6.yaml")

// 	expected := map[string]interface{}{
// 		"test-case-6": []map[string]interface{}{
// 			{
// 				"alpine": "3.19",
// 			},
// 			{
// 				"alpine": "3.20",
// 			},
// 			{
// 				"alpine": "3.21",
// 			},
// 		},
// 	}

// 	for image, cfg := range inputs {
// 		combinations := parser.GenerateVariableCombinations(cfg.Variables)
// 		assert.Equal(t, expected[image], combinations)
// 	}
// }

func TestCombinationsCase7(t *testing.T) {
	inputs := loadConfig("test-7.yaml").Images

	expected := map[string]interface{}{
		"test-case-7": []map[string]interface{}{
			{
				"alpine":   "3.20",
				"crazy":    map[string]interface{}{"key1": "value1"},
				"timezone": "UTC",
			},
			{
				"alpine":   "3.20",
				"crazy":    map[string]interface{}{"key2": "value2"},
				"timezone": "UTC",
			},
			{
				"alpine":   "3.21",
				"crazy":    map[string]interface{}{"key1": "value1"},
				"timezone": "UTC",
			},
			{
				"alpine":   "3.21",
				"crazy":    map[string]interface{}{"key2": "value2"},
				"timezone": "UTC",
			},
		},
	}

	for image, cfg := range inputs {
		combinations := parser.GenerateVariableCombinations(cfg.Variables)
		assert.Equal(t, expected[image], combinations)
	}
}

func TestCombinationsCase8(t *testing.T) {
	inputs := loadConfig("test-8.yaml").Images

	excluded := []map[string]interface{}{
		{"alpine": "3.19", "tomcat": "11.0.2", "java": 8},
		{"alpine": "3.19", "tomcat": "11.0.2", "java": 11},
		{"alpine": "3.19", "tomcat": "10.1.34", "java": 8},
		{"alpine": "3.20", "tomcat": "11.0.2", "java": 8},
		{"alpine": "3.20", "tomcat": "11.0.2", "java": 11},
		{"alpine": "3.20", "tomcat": "10.1.34", "java": 8},
		{"alpine": "3.21", "tomcat": "11.0.2", "java": 8},
		{"alpine": "3.21", "tomcat": "11.0.2", "java": 11},
		{"alpine": "3.21", "tomcat": "10.1.34", "java": 8},
	}

	// collect all sets
	var combinations [][]map[string]interface{}
	for _, cfg := range inputs {
		set := parser.GenerateVariableCombinations(cfg.Variables)
		combinations = append(combinations, set)
	}

	// check if any matches excluded
	for _, exclude := range excluded {
		assert.NotContains(t, combinations, exclude)
	}
}

func TestConfigSetGenerationCase1(t *testing.T) {

}

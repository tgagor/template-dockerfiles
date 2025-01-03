package image_test

// func getKeys(m map[string]string) []string {
// 	keys := make([]string, 0, len(m))
// 	for k := range m {
// 		keys = append(keys, k)
// 	}
// 	return keys
// }

// FIXME: rewrite those tests for the image package

// func TestConfigSetGenerationCase1(t *testing.T) {
// 	t.Parallel()

// 	cfg := loadConfig("test-1.yaml")

// 	for _, imageName := range cfg.ImageOrder {
// 		combinations := parser.GenerateVariableCombinations(cfg.Images[imageName].Variables)
// 		for _, set := range combinations {
// 			configSet, err := parser.GenerateConfigSet(imageName, cfg, set, &config.Flags{})
// 			require.NotEmpty(t, configSet)
// 			require.NoError(t, err)

// 			assert.Empty(t, configSet["registry"])
// 			assert.Empty(t, configSet["prefix"])
// 			assert.Empty(t, configSet["maintenance"])
// 			assert.Empty(t, configSet["platforms"])
// 			assert.Empty(t, configSet["args"])
// 			assert.NotEmpty(t, configSet["labels"]) // because of default OCI labels
// 			// check example OCI labels
// 			labelKeys := getKeys(configSet["labels"].(map[string]string))
// 			assert.Contains(t, labelKeys, "org.opencontainers.image.created")

// 			assert.NotEmpty(t, configSet["tags"])
// 			assert.Contains(t, configSet["tags"], "test-case-1")

// 			assert.NotEmpty(t, configSet["alpine"])              // one of 3.18/3.19/3.20
// 			assert.Equal(t, "kuku ruku", configSet["multiword"]) // static
// 		}
// 	}
// }

// func TestConfigSetGenerationCase2(t *testing.T) {
// 	t.Parallel()

// 	cfg := loadConfig("test-2.yaml")

// 	for _, imageName := range cfg.ImageOrder {
// 		combinations := parser.GenerateVariableCombinations(cfg.Images[imageName].Variables)
// 		for _, set := range combinations {
// 			configSet, err := parser.GenerateConfigSet(imageName, cfg, set, &config.Flags{})
// 			require.NotEmpty(t, configSet)
// 			require.NoError(t, err)

// 			// check if global labels are populated everywhere
// 			assert.NotEmpty(t, configSet["labels"])
// 			labelKeys := getKeys(configSet["labels"].(map[string]string))
// 			assert.Contains(t, labelKeys, "org.opencontainers.image.created")
// 			assert.Contains(t, labelKeys, "org.opencontainers.image.vendor")
// 			assert.Contains(t, labelKeys, "org.opencontainers.image.licenses")
// 			assert.Contains(t, labelKeys, "org.opencontainers.image.description")

// 			// per image labels should only be where added
// 			if imageName == "test-case-2" {
// 				assert.Contains(t, labelKeys, "org.opencontainers.image.url")
// 				assert.Contains(t, labelKeys, "org.opencontainers.image.documentation")
// 				assert.Contains(t, labelKeys, "org.opencontainers.image.title")
// 				assert.Contains(t, labelKeys, "org.opencontainers.image.description")
// 				assert.Contains(t, labelKeys, "org.opencontainers.image.test-case-2.name")
// 			} else {
// 				assert.NotContains(t, labelKeys, "org.opencontainers.image.url")
// 				assert.NotContains(t, labelKeys, "org.opencontainers.image.documentation")
// 				assert.NotContains(t, labelKeys, "org.opencontainers.image.title")
// 				assert.NotContains(t, labelKeys, "org.opencontainers.image.test-case-2.name")
// 			}

// 		}
// 	}
// }

// func TestConfigSetGenerationCase5(t *testing.T) {
// 	t.Parallel()

// 	cfg := loadConfig("test-5.yaml")

// 	for _, imageName := range cfg.ImageOrder {
// 		combinations := parser.GenerateVariableCombinations(cfg.Images[imageName].Variables)
// 		for _, set := range combinations {
// 			configSet, err := parser.GenerateConfigSet(imageName, cfg, set, &config.Flags{})
// 			require.NotEmpty(t, configSet)
// 			require.NoError(t, err)

// 			assert.NotEmpty(t, configSet["labels"]) // because of default OCI labels
// 			assert.Equal(t, "label", configSet["labels"].(map[string]string)["ugly"])

// 			assert.NotEmpty(t, configSet["tags"])
// 			assert.Contains(t, configSet["tags"], "whatever")

// 			assert.Equal(t, 3, configSet["alpine"])
// 		}
// 	}
// }

// func TestConfigSetGenerationCase6(t *testing.T) {
// 	t.Parallel()

// 	cfg := loadConfig("test-6.yaml")

// 	for _, imageName := range cfg.ImageOrder {
// 		combinations := parser.GenerateVariableCombinations(cfg.Images[imageName].Variables)
// 		for _, set := range combinations {
// 			configSet, err := parser.GenerateConfigSet(imageName, cfg, set, &config.Flags{Engine: "buildx"})
// 			require.NotEmpty(t, configSet)
// 			require.NoError(t, err)

// 			assert.NotEmpty(t, configSet["platforms"])
// 			if imageName == "test-case-6a" {
// 				// two platforms here
// 				platforms := []string{
// 					"linux/amd64",
// 					"linux/arm64",
// 				}
// 				assert.Equal(t, platforms, configSet["platforms"])
// 			}
// 			if imageName == "test-case-6b" {
// 				// just one here because of per image override
// 				platforms := []string{
// 					"linux/amd64",
// 				}
// 				assert.Equal(t, platforms, configSet["platforms"])
// 			}
// 		}
// 	}
// }

// func TestConfigSetGenerationCase6FailWithBadEngine(t *testing.T) {
// 	t.Parallel()

// 	cfg := loadConfig("test-6.yaml")

// 	for _, imageName := range cfg.ImageOrder {
// 		combinations := parser.GenerateVariableCombinations(cfg.Images[imageName].Variables)
// 		for _, set := range combinations {
// 			configSet, err := parser.GenerateConfigSet(imageName, cfg, set, &config.Flags{Build: true})
// 			require.Error(t, err)
// 			require.Empty(t, configSet)
// 		}
// 	}
// }

// Broken assumptions, I generate all config sets and then filter them in the parser.Run
// So this one would always fail, as all config sets are available
// func TestConfigSetGenerationCase8(t *testing.T) {
// 	t.Parallel()

// 	cfg := loadConfig("test-8.yaml")

// 	for _, imageName := range cfg.ImageOrder {
// 		combinations := parser.GenerateVariableCombinations(cfg.Images[imageName].Variables)
// 		for _, set := range combinations {
// 			configSet, err := parser.GenerateConfigSet(imageName, cfg, set, config.Flags{})
// 			require.NotEmpty(t, configSet)
// 			require.NoError(t, err)

// 			if imageName == "test-case-8" {
// 				fmt.Printf("%v", configSet)
// 				// check if our excludes do not match something
// 				assert.Falsef(t, configSet["tomcat"] == "11.0.2" && configSet["java"] == 8, "excluded configuration found!")
// 				assert.Falsef(t, configSet["tomcat"] == "11.0.2" && configSet["java"] == 11, "excluded configuration found!")
// 				assert.Falsef(t, configSet["tomcat"] == "10.1.34" && configSet["java"] == 8, "excluded configuration found!")
// 			}
// 		}
// 	}
// }

// TODO: I hava a lot of tests for proper config sets generation, but not much for proper tags
// maybe I should add test for test-8.yaml where a lot of tags are generated

// func TestConfigSetGenerationCase9(t *testing.T) {
// 	t.Parallel()

// 	cfg := loadConfig("test-9.yaml")

// 	for _, imageName := range cfg.ImageOrder {
// 		combinations := parser.GenerateVariableCombinations(cfg.Images[imageName].Variables)
// 		for _, set := range combinations {
// 			configSet, err := parser.GenerateConfigSet(imageName, cfg, set, &config.Flags{})
// 			require.NotEmpty(t, configSet)
// 			require.NoError(t, err)

// 			assert.NotEmpty(t, configSet["args"]) // because of default OCI labels
// 			assert.Equal(t, configSet["alpine"], configSet["args"].(map[string]string)["BASEIMAGE"])
// 			assert.Equal(t, configSet["timezone"], configSet["args"].(map[string]string)["TIMEZONE"])
// 		}
// 	}
// }

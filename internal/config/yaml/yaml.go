package yaml

import (
	"errors"
	backupComponents "fedob/backup/internal/backupComponent"
	config "fedob/backup/internal/config"
	"fmt"
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"
)

type RawComponentsWrapper struct {
	BackupComponents []yaml.Node `yaml:"components"`
}

type YAMLConfigInterface struct {
	path string
}

func NewYAMLConfigInterface(path string) YAMLConfigInterface {
    return YAMLConfigInterface{path: path}
}

func extractComponent(node *yaml.Node) (backupComponents.BackupComponent, error) {
	var base backupComponents.BaseComponentAttributes
	if err := node.Decode(&base); err != nil {
		return nil, fmt.Errorf("error decoding base component attributes: %w", err)
	}

	constructor, exists := config.BackupComponentRegistry[backupComponents.BackupType(base.Type)]
	if !exists {
		return nil, fmt.Errorf("unknown component type \"%s\" for component %s", base.Type, base.Name)
	}

	componentObject := constructor()
	slog.Debug("Parsing component", "componentType", base.Type, "componentName", base.Name)
	
	if err := node.Decode(componentObject); err != nil {
		return nil, fmt.Errorf("can't decode node into struct of type %T: %w", componentObject, err)
	}
	return componentObject, nil
}

func (y YAMLConfigInterface) LoadConfig() (config.ClientConfig, error) {
	// Array of raw yaml nodes to hold unparsed backup components
	var rawComponentsWrapper RawComponentsWrapper
	//
	var clientConfig config.ClientConfig
	// Root level attributes of config
	var rootLevelConfig config.RootLevelConfig
	// Raw yaml node that holds entire yaml content
	var rawYaml yaml.Node

	if y.path == "" {
		return config.ClientConfig{}, errors.New("Config file path is empty")
	}
	data, err := os.ReadFile(y.path)
	if err != nil {
		return config.ClientConfig{}, fmt.Errorf("Error reading config file: %w", err)
	}
	
	if err := yaml.Unmarshal(data, &rawYaml); err != nil {
		return config.ClientConfig{}, fmt.Errorf("Error unmarshaling YAML: %w", err)
	}

	if err := rawYaml.Decode(&rootLevelConfig); err != nil {
		return config.ClientConfig{}, fmt.Errorf("error parsing root level config: %w", err)
	}
	if err := rawYaml.Decode(&rawComponentsWrapper); err != nil {
		return config.ClientConfig{}, fmt.Errorf("error parsing raw components: %w", err)
	}

	// copy root level config into final config
	clientConfig.RootLevelConfig = rootLevelConfig

	var parsedComponents []backupComponents.BackupComponent

	for i, rawComponent := range rawComponentsWrapper.BackupComponents {
		component, err := extractComponent(&rawComponent)
		if err != nil {
			return config.ClientConfig{}, fmt.Errorf("error parsing component at index %d: %w", i, err)
		}
		parsedComponents = append(parsedComponents, component)
	}
	clientConfig.BackupComponents = parsedComponents
	return clientConfig, nil
}

func (y YAMLConfigInterface) SaveConfig(config config.ClientConfig) error {
	return errors.New("Yaml SaveConfig not implemented")
}

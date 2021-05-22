package models

import (
	"io/ioutil"
	"gopkg.in/yaml.v3"
)


type OnboardClusterConfig struct {
	OnboardingTemplate OnboardingTmpl `yaml:"onboardingTemplate,omitempty"`
}

type OnboardingTmpl struct {
	TeamName 			string 				`yaml:"teamName,omitempty"`
	Namespaces 			[]ConfigNamespace 	`yaml:"namespaces,omitempty"`
	Usernames			[]string 			`yaml:"usernames,omitempty"`
	ProjectDescription 	string 				`yaml:"projectDescription,omitempty"`
	Env 				string 				`yaml:"env,omitempty"`
	Cluster 			string 				`yaml:"cluster,omitempty"`
	PGPKeys			 	[]string 			`yaml:"pgpKeys,omitempty"`
}

type ConfigNamespace struct {
	Name				string 	`yaml:"name,omitempty"`
	Quota				string 	`yaml:"quota,omitempty"`
	EnableMonitoring	bool 	`yaml:"enableMonitoring,omitempty"`
	DisplayName			string 	`yaml:"displayName,omitempty"`
}

func OnboardConfigFromYAMLPath(path string) (OnboardClusterConfig, error) {
	var occ OnboardClusterConfig

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return OnboardClusterConfig{}, err
	}
	if err = yaml.Unmarshal(content, &occ); err != nil {
		return OnboardClusterConfig{}, err
	}

	return occ, nil
}
package models

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"gopkg.in/yaml.v3"
)

// Group represents a group of users.
type Group struct {
	Resource `yaml:",inline"`
	Users    []string
}

// NewGroup creates a new Group object.
func NewGroup(name string) Group {
	if len(name) == 0 {
		log.Fatal("a group requires a name")
	}

	rsrc := Group{
		Resource: Resource{
			APIVersion: "user.openshift.io/v1",
			Kind:       "Group",
			Metadata: Metadata{
				Name: name,
			},
		},
		Users: make([]string, 0),
	}
	return rsrc
}

// GroupFromYAMLPath reads the file at path and returns a
// Group object.
func GroupFromYAMLPath(path string) (Group, error) {
	var group Group

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return Group{}, err
	}

	err = yaml.Unmarshal(content, &group)
	if err != nil {
		return Group{}, err
	}

	return group, nil
}

func (g Group) Contains(user string) bool {
	for _, u := range g.Users {
		if u == user {
			return true
		}
	}
	return false
}
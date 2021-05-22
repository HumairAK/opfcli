package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/operate-first/opfcli/models"
	"github.com/operate-first/opfcli/utils"
	log "github.com/sirupsen/logrus"
)

var validRoles = map[string]bool{
	"admin": true,
	"view":  true,
	"edit":  true,
}

func createNamespace(projectName, projectOwner, projectDescription string, existsOk bool) error {
	appName := config.GetString("app-name")
	path := filepath.Join(repoDirectory, appName, namespacePath, projectName, "namespace.yaml")

	exists, err := utils.PathExists(filepath.Dir(path))
	if err != nil {
		return err
	}

	if exists {
		if existsOk {
			log.Printf("namespace already exists (continuing)")
			return nil
		}
		return fmt.Errorf("namespace %s already exists", projectName)
	}

	ns := models.NewNamespace(projectName, projectOwner, projectDescription)
	nsOut := models.ToYAML(ns)

	log.Printf("writing namespace definition to %s", filepath.Dir(path))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create namespace directory: %w", err)
	}

	err = ioutil.WriteFile(path, nsOut, 0644)
	if err != nil {
		return fmt.Errorf("failed to write namespace file: %w", err)
	}

	err = utils.WriteKustomization(
		filepath.Dir(path),
		[]string{"namespace.yaml"},
		[]string{
			filepath.Join(componentRelPath, "project-admin-rolebindings", projectOwner),
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func createRoleBinding(groupName, roleName string) error {
	appName := config.GetString("app-name")
	bindingName := fmt.Sprintf("project-%s-rolebindings", roleName)
	path := filepath.Join(
		repoDirectory, appName, componentPath,
		bindingName, groupName, "rbac.yaml",
	)

	exists, err := utils.PathExists(filepath.Dir(path))
	if err != nil {
		return err
	}

	if exists {
		log.Printf("rolebinding already exists (continuing)")
		return nil
	}

	rbac := models.NewRoleBinding(
		fmt.Sprintf("namespace-%s-%s", roleName, groupName),
		roleName,
	)
	rbac.AddGroup(groupName)
	rbacOut := models.ToYAML(rbac)

	log.Printf("writing rbac definition to %s", filepath.Dir(path))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create rolebinding directory: %w", err)
	}

	err = ioutil.WriteFile(path, rbacOut, 0644)
	if err != nil {
		return fmt.Errorf("failed to write rbac: %w", err)
	}

	err = utils.WriteComponent(
		filepath.Dir(path),
		[]string{"rbac.yaml"},
	)
	if err != nil {
		return err
	}

	return nil
}

func createAdminRoleBinding(projectName, projectOwner string) error {
	return createRoleBinding(projectOwner, "admin")
}

func createGroup(groupName string, existsOk bool) error {
	appName := config.GetString("app-name")
	path := filepath.Join(repoDirectory, appName, groupPath, groupName, "group.yaml")

	exists, err := utils.PathExists(filepath.Dir(path))
	if err != nil {
		return err
	}

	if exists {
		if existsOk {
			log.Printf("group already exists (continuing)")
			return nil
		}
		return fmt.Errorf("group %s already exists", groupName)
	}

	group := models.NewGroup(groupName)
	groupOut := models.ToYAML(group)

	log.Printf("writing group definition to %s", filepath.Dir(path))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create group directory: %w", err)
	}

	err = ioutil.WriteFile(path, groupOut, 0644)
	if err != nil {
		return fmt.Errorf("failed to write group: %w", err)
	}

	err = utils.WriteKustomization(
		filepath.Dir(path),
		[]string{"group.yaml"},
		nil,
	)
	if err != nil {
		return err
	}

	return nil
}

func addGroupRBAC(projectName, groupName, roleName string) error {
	if !validRoles[roleName] {
		return fmt.Errorf("no such role named %q", roleName)
	}

	appName := config.GetString("app-name")
	bindingName := fmt.Sprintf("project-%s-rolebindings", roleName)

	nsPath := filepath.Join(
		repoDirectory, appName, namespacePath, projectName,
	)

	groupPath := filepath.Join(
		repoDirectory, appName, groupPath, groupName,
	)

	exists, err := utils.PathExists(nsPath)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("namespace %s does not exist", projectName)
	}

	exists, err = utils.PathExists(groupPath)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("group %s does not exist", groupName)
	}

	if err = createRoleBinding(groupName, roleName); err != nil {
		return err
	}

	log.Printf("granting %s role %s on %s", groupName, roleName, projectName)
	err = utils.AddKustomizeComponent(
		nsPath,
		filepath.Join(componentRelPath, bindingName, groupName),
	)
	if err != nil {
		return err
	}

	return nil
}

func addNamespaceToCluster(ns, env, cluster string) error {
	appName := config.GetString("app-name")
	clusterPath := filepath.Join(repoDirectory, appName, overlaysPath, env, cluster)
	nsPath := filepath.Join(appRootRelPath, namespacePath, ns)
	exists, err := utils.PathExists(filepath.Dir(clusterPath))

	if err != nil {
		return err
	} else if !exists {
		log.Printf("Kustomization for overlay does not exist, creating.")
		if err := os.MkdirAll(clusterPath, 0755); err != nil {
			return err
		}
		if err := utils.WriteKustomization(clusterPath, []string{nsPath}, []string{}); err != nil{
			return err
		}
	}

	if err := utils.AddKustomizeResource(clusterPath, nsPath); err != nil {
		return err
	}

	return nil
}

func addGroupToCluster(groupName, env, cluster string, users []string) error {
	appName := config.GetString("app-name")
	clusterPath := filepath.Join(repoDirectory, appName, overlaysPath, env, cluster)
	groupPath := filepath.Join(clusterPath, "groups", groupName + ".enc.yaml")
	var group models.Group
	var err error

	if exists, err := utils.PathExists(groupPath); err != nil {
		return err
	} else if exists {
		log.Printf("group %s in cluster %s already exists, appending existing users list", groupName, cluster)
		group, err = models.GroupFromYAMLPath(groupPath)
		if err != nil {
			return err
		}
		for _, user := range users {
			if !group.Contains(user) {
				group.Users = append(group.Users, user)
			}
		}
	} else {
		dirExists, err := utils.PathExists(filepath.Dir(groupPath))
		if err != nil {
			return err
		}
		if !dirExists {
			log.Printf("group directory for cluster not found, creating a new one... %s", filepath.Dir(groupPath))
			if err := os.MkdirAll(filepath.Dir(groupPath), 0755); err != nil {
				return fmt.Errorf("failed to create group directory: %w", err)
			}
		}
		group = models.NewGroup(groupName)
		group.Users = users
	}

	groupOut := models.ToYAML(group)

	err = ioutil.WriteFile(groupPath, groupOut, 0644)
	if err != nil {
		return fmt.Errorf("failed to write group: %w", err)
	}

	return nil
}

func addUsersToGroupInCluster(user, group, env, cluster string, pgpKeys []string) error {
	fmt.Println("adding namespace to cluster")
	return nil
}
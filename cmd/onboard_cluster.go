package cmd

import (
	"github.com/operate-first/opfcli/models"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	onboardClusterCfgFile string
	onboardCluster        = &cobra.Command{
		Use:   "cluster",
		Long: "Onboard a team/user to a cluster. This command creates " +
			"the manifests necessary to onboard a new team to a cluster. " +
			"If any of the manifests already exist, then they are either " +
			"ignored or updated to accomodate new changes.",
		Short: "Onboard a team/user to a cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			onboardClusterCfg, err := models.OnboardConfigFromYAMLPath(onboardClusterCfgFile)
			if err != nil {
				log.Fatalf("Encountered error marshalling config file: %v", err)
			}

			onboardingTemplate := onboardClusterCfg.OnboardingTemplate
			env, cluster, team := onboardingTemplate.Env, onboardingTemplate.Cluster, onboardingTemplate.TeamName

			for _, ns := range onboardingTemplate.Namespaces {
				if err = createNamespace(ns.Name, team, ns.DisplayName, true); err != nil {
					return err
				}
				if err = addNamespaceToCluster(ns.Name, env, cluster); err != nil {
					return err
				}
			}

			if err = createRoleBinding(team, "admin"); err != nil {
				return err
			}

			if err = createGroup(team, true); err != nil {
				return err
			}

			if err = addGroupToCluster(team, env, cluster, onboardingTemplate.Usernames); err != nil {
				return err
			}
			return nil
		},
	}
)

func init() {
	onboardCluster.PersistentFlags().StringVar(&onboardClusterCfgFile, "onboard-config", "", "cluster config file (required)")
	if err := onboardCluster.MarkPersistentFlagRequired("onboard-config"); err != nil {
		log.Debugf("Onboarding requires a config file, please use the -onboard-config flag: %v", err)
	}

	onboardCmd.AddCommand(onboardCluster)
}

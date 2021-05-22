package cmd

import (
	"github.com/spf13/cobra"
)

var createProjectCmd = &cobra.Command{
	Use:   "create-project projectName projectOwner",
	Short: "Onboard a new project into Operate First",
	Long: `Onboard a new project into Operate First.

- Register a new group
- Register a new namespace with appropriate role bindings for your group
`,
	Args:          cobra.ExactArgs(2),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		projectName := args[0]
		projectOwner := args[1]
		projectDescription := cmd.Flag("description").Value.String()

		if err = createNamespace(projectName, projectOwner, projectDescription, false); err != nil {
			return err
		}
		if err = createAdminRoleBinding(projectName, projectOwner); err != nil {
			return err
		}
		if err = createGroup(projectOwner, true); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(createProjectCmd)

	createProjectCmd.PersistentFlags().StringP(
		"description", "d", "", "Team description")
}

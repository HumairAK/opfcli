// Package cmd implements the commands supported by opfcli.
package cmd

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/operate-first/opfcli/utils"
)

const defaultAppName = "cluster-scope"
const namespacePath = "base/core/namespaces"
const groupPath = "base/user.openshift.io/groups"
const componentPath = "components"
const componentRelPath = "../../../../components"
const overlaysPath  = "overlays"
const appRootRelPath = "../../.."

var config = viper.New()

var cfgFile string
var appName string
var repoDirectory string

var rootCmd = &cobra.Command{
	Use:   "opfcli",
	Short: "A command line tool for Operate First GitOps",
	Long: `A command line tool for Operate First GitOps.

Use opfcli to interact with an Operate First style Kubernetes
configuration repository.`,
}

// Execute processes the opfcli command line.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error: %s", err)
	}
}

func init() {
	loglevel, err := strconv.Atoi(os.Getenv("OPF_LOGLEVEL"))
	if err != nil {
		loglevel = 1
	}

	switch {
	case loglevel >= 2:
		log.SetLevel(log.DebugLevel)
	case loglevel >= 1:
		log.SetLevel(log.InfoLevel)
	default:
		log.SetLevel(log.WarnLevel)
	}

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(
		&cfgFile, "config", "f", "", "configuration file")
	rootCmd.PersistentFlags().StringVarP(
		&appName, "app-name", "a", "cluster-scope", "application name")
	rootCmd.PersistentFlags().StringVarP(
		&repoDirectory, "repodir", "R", "", "path to opf repository")

	if err = config.BindPFlag("app-name", rootCmd.PersistentFlags().Lookup("app-name")); err != nil {
		panic(err)
	}
}

func initConfig() {
	config.SetEnvPrefix("opf")
	config.AutomaticEnv()
	config.SetConfigName(".opfcli")

	replacer := strings.NewReplacer("-", "_")
	config.SetEnvKeyReplacer(replacer)

	config.SetDefault("app-name", defaultAppName)

	if cfgFile != "" {
		config.SetConfigFile(cfgFile)
		if err := config.ReadInConfig(); err != nil {
			log.Debugf("failed to read configuration: %v", err)
		}
	} else {
		home, err := homedir.Dir()
		if err == nil {
			config.AddConfigPath(home)
		}

		if repoDirectory == "" {
			repoDirectory, err = utils.FindRepoDir()
			if err != nil {
				repoDirectory, err = filepath.Abs(".")
				if err != nil {
					log.Fatalf("failed to determine repository directory: %v", err)
				}
			}
		}
		log.Debugf("using %s as repository directory", repoDirectory)

		config.AddConfigPath(repoDirectory)
		if err = config.ReadInConfig(); err != nil {
			log.Debugf("failed to read configuration: %v", err)
		}
		if config.ConfigFileUsed() != "" {
			log.Printf("read configuration from %s", config.ConfigFileUsed())
		}
	}
}

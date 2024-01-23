package cmd

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/bednarradek/php-deployer/internal"
	"github.com/bednarradek/php-deployer/pkg/file_system"
	"github.com/spf13/cobra"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Run deploying process",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("deploy called")

		ctx := cmd.Context()

		t, err := cmd.Flags().GetString("type")
		if err != nil {
			log.Fatalf("Error while getting type flag: %s", err)
		}
		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			log.Fatalf("Error while getting config flag: %s", err)
		}

		configContent, err := file_system.NewSystemReader().Read(ctx, configPath)
		if err != nil {
			log.Fatalf("Error while reading config file: %s", err)
		}

		switch t {
		case "ftp":
			config := new(internal.FtpConfig)
			if err := json.Unmarshal(configContent, config); err != nil {
				log.Fatalf("Error while unmarshalling config file: %s", err)
			}
			deployer, err := internal.NewFtpDeployer(config)
			if err != nil {
				log.Fatalf("Error while creating deployer: %s", err)
			}
			defer func() {
				deployer.Close()
			}()
			if err := deployer.Deploy(ctx); err != nil {
				log.Fatalf("Error while deploying: %s", err)
			}
		default:
			log.Fatalf("Unknown type of deployer: %s", t)
		}
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)

	deployCmd.Flags().StringP("type", "t", "ftp", "Type of deployer")
	deployCmd.Flags().StringP("config", "c", "", "Path to config file")
}

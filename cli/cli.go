package cli

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/AfazTech/b9m/api"
	"github.com/AfazTech/b9m/config"
	"github.com/AfazTech/b9m/parser"
	"github.com/AfazTech/b9m/record"
	"github.com/AfazTech/b9m/servicemanager"
	"github.com/AfazTech/b9m/utils"
	"github.com/AfazTech/b9m/zone"
	"github.com/AfazTech/logger/v2"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "b9m",
	Short: "B9m Command Line Interface",
	Long:  "A CLI tool for managing BIND9 configurations and operations in b9m.",
}

var addDomainCmd = &cobra.Command{
	Use:   "add-domain [domain] [ns1] [ns2]",
	Short: "Add a new domain",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		domain, ns1, ns2 := args[0], args[1], args[2]
		if err := zone.AddDomain(domain, ns1, ns2); err != nil {
			logger.Fatal(err)
		}
		logger.Infof("Domain '%s' added successfully with nameservers '%s' and '%s'.", domain, ns1, ns2)
	},
}

var deleteDomainCmd = &cobra.Command{
	Use:   "delete-domain [domain]",
	Short: "Delete a domain",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		domain := args[0]
		if err := zone.DeleteDomain(domain); err != nil {
			logger.Fatal(err)
		}
		logger.Infof("Domain '%s' deleted successfully.", domain)
	},
}

var addRecordCmd = &cobra.Command{
	Use:   "add-record [domain] [name] [type] [value] [ttl]",
	Short: "Add a new DNS record",
	Args:  cobra.ExactArgs(5),
	Run: func(cmd *cobra.Command, args []string) {
		domain, name, rType, value, ttlStr := args[0], args[1], args[2], args[3], args[4]
		ttl, err := strconv.Atoi(ttlStr)
		if err != nil {
			logger.Fatalf("Invalid TTL value '%s': %v", ttlStr, err)
		}
		if err := record.AddRecord(domain, record.RecordType(rType), name, value, ttl); err != nil {
			logger.Fatal(err)
		}
		logger.Infof("Record added successfully: Domain: '%s', Name: '%s', Type: '%s', Value: '%s', TTL: %d.", domain, name, rType, value, ttl)
	},
}

var deleteRecordCmd = &cobra.Command{
	Use:   "delete-record [domain] [name] [type] [value]",
	Short: "Delete a DNS record",
	Args:  cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		domain, name, rType, value := args[0], args[1], args[2], args[3]
		if err := record.DeleteRecord(domain, name, record.RecordType(rType), value); err != nil {
			logger.Fatal(err)
		}
		logger.Infof("Record deleted successfully: Domain: '%s', Name: '%s', Type: '%s', Value: '%s'.", domain, name, rType, value)
	},
}

var getRecordsCmd = &cobra.Command{
	Use:   "get-records [domain]",
	Short: "Get all records of a domain",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		domain := args[0]
		records, err := record.GetAllRecords(domain)
		if err != nil {
			logger.Fatal(err)
		}
		logger.Infof("DNS records for domain '%s':", domain)
		for _, record := range records {
			logger.Infof("Record: Name: '%s', TTL: %d, Type: '%s', Value: '%s'.", record.Name, record.TTL, record.Type, record.Value)
		}
	},
}

var startAPICmd = &cobra.Command{
	Use:   "start-api [port] [apiKey]",
	Short: "Start the API server",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		port, apiKey := args[0], args[1]
		logger.Infof("Starting API server on port '%s' with provided API key.", port)
		api.StartServer(port, apiKey)
	},
}
var backupCmd = &cobra.Command{
	Use:   "backup [directory]",
	Short: "backup all config files",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		directory := args[0]
		logger.Infof("all files backuped to directory: '%s'.", directory)
		err := utils.Backup(directory)
		if err != nil {
			logger.Fatalf("failed to backup: %v", err)
		}
	},
}
var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload BIND9 configuration",
	Run: func(cmd *cobra.Command, args []string) {
		if err := servicemanager.ReloadBind(); err != nil {
			logger.Fatalf("Error reloading BIND9 configuration: %v", err)
		}
		logger.Info("BIND9 configuration reloaded successfully using 'rndc reload'.")
	},
}

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart BIND9 service",
	Run: func(cmd *cobra.Command, args []string) {
		if err := servicemanager.RestartBind(); err != nil {
			logger.Fatalf("Error restarting BIND9 service: %v", err)
		}
		logger.Info("BIND9 service restarted successfully.")
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop BIND9 service",
	Run: func(cmd *cobra.Command, args []string) {
		if err := servicemanager.StopBind(); err != nil {
			logger.Fatalf("Error stopping BIND9 service: %v", err)
		}
		logger.Info("BIND9 service stopped successfully.")
	},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start BIND9 service",
	Run: func(cmd *cobra.Command, args []string) {
		if err := servicemanager.StartBind(); err != nil {
			logger.Fatalf("Error starting BIND9 service: %v", err)
		}
		logger.Info("BIND9 service started successfully.")
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get the status of BIND9 service",
	Run: func(cmd *cobra.Command, args []string) {
		status, err := servicemanager.StatusBind()
		if err != nil {
			logger.Fatalf("Error fetching BIND9 service status: %v", err)
		}
		logger.Infof("BIND9 service status retrieved: '%s'.", status)
	},
}

var getDomainsCmd = &cobra.Command{
	Use:   "get-domains",
	Short: "Get all domains and their configuration files",
	Run: func(cmd *cobra.Command, args []string) {
		domains, err := parser.GetDomains()
		if err != nil {
			logger.Fatalf("Error fetching domains: %v", err)
		}
		logger.Info("List of domains and their associated configuration files:")
		for domain, file := range domains {
			logger.Infof("Domain: '%s', File: '%s'.", domain, file)
		}
	},
}

var getConfigCmd = &cobra.Command{
	Use:   "get-config",
	Short: "Get all bind9 configs",
	Run: func(cmd *cobra.Command, args []string) {
		confFile := config.GetConfigFile()
		configs, err := parser.ParseConfig(confFile)
		if err != nil {
			logger.Fatalf("failed to parsing configuration: %v", err)
		}

		prettyJSON, err := json.MarshalIndent(configs, "", "  ")
		if err != nil {
			logger.Fatalf("failed to format JSON: %v", err)
		}

		logger.Infof(string(prettyJSON))
	},
}

func init() {
	rootCmd.AddCommand(
		addDomainCmd,
		deleteDomainCmd,
		addRecordCmd,
		deleteRecordCmd,
		getRecordsCmd,
		startAPICmd,
		reloadCmd,
		restartCmd,
		stopCmd,
		startCmd,
		statusCmd,
		getDomainsCmd,
		getConfigCmd,
		backupCmd,
	)
}

func StartCLI() {
	if err := rootCmd.Execute(); err != nil {
		logger.Fatal(err)
		os.Exit(1)
	}
}

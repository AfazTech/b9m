package cli

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/imafaz/b9m/api"
	"github.com/imafaz/b9m/controller"
	"github.com/spf13/cobra"
)

var bindManager = controller.NewBindManager("/etc/bind/zones", "/etc/bind/named.conf.local")

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
		if err := bindManager.AddDomain(args[0], args[1], args[2]); err != nil {
			log.Fatalf("Error: %v", err)
		}
		fmt.Println("Domain added successfully.")
	},
}

var deleteDomainCmd = &cobra.Command{
	Use:   "delete-domain [domain]",
	Short: "Delete a domain",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := bindManager.DeleteDomain(args[0]); err != nil {
			log.Fatalf("Error: %v", err)
		}
		fmt.Println("Domain deleted successfully.")
	},
}

var addRecordCmd = &cobra.Command{
	Use:   "add-record [domain] [name] [type] [value] [ttl]",
	Short: "Add a new DNS record",
	Args:  cobra.ExactArgs(5),
	Run: func(cmd *cobra.Command, args []string) {
		ttl, err := strconv.Atoi(args[4])
		if err != nil {
			log.Fatalf("Invalid TTL value: %v", err)
		}
		if err := bindManager.AddRecord(args[0], controller.RecordType(args[2]), args[1], args[3], ttl); err != nil {
			log.Fatalf("Error: %v", err)
		}
		fmt.Println("Record added successfully.")
	},
}

var deleteRecordCmd = &cobra.Command{
	Use:   "delete-record [domain] [name]",
	Short: "Delete a DNS record",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if err := bindManager.DeleteRecord(args[0], args[1]); err != nil {
			log.Fatalf("Error: %v", err)
		}
		fmt.Println("Record deleted successfully.")
	},
}

var getRecordsCmd = &cobra.Command{
	Use:   "get-records [domain]",
	Short: "Get all records of a domain",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		records, err := bindManager.GetAllRecords(args[0])
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		for _, record := range records {
			fmt.Printf("%s %d IN %s %s\n", record.Name, record.TTL, record.Type, record.Value)
		}
	},
}

var startAPICmd = &cobra.Command{
	Use:   "start-api [port] [apiKey]",
	Short: "Start the API server",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Starting API server on port %s with API key %s\n", args[0], args[1])
		api.StartServer(args[0], args[1])
	},
}

var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload BIND9 configuration",
	Run: func(cmd *cobra.Command, args []string) {
		if err := bindManager.ReloadBind(); err != nil {
			log.Fatalf("Error reloading BIND: %v", err)
		}
		fmt.Println("BIND9 reloaded successfully.")
	},
}

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart BIND9 service",
	Run: func(cmd *cobra.Command, args []string) {
		if err := bindManager.RestartBind(); err != nil {
			log.Fatalf("Error restarting BIND: %v", err)
		}
		fmt.Println("BIND9 restarted successfully.")
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop BIND9 service",
	Run: func(cmd *cobra.Command, args []string) {
		if err := bindManager.StopBind(); err != nil {
			log.Fatalf("Error stopping BIND: %v", err)
		}
		fmt.Println("BIND9 stopped successfully.")
	},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start BIND9 service",
	Run: func(cmd *cobra.Command, args []string) {
		if err := bindManager.StartBind(); err != nil {
			log.Fatalf("Error starting BIND: %v", err)
		}
		fmt.Println("BIND9 started successfully.")
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get the status of BIND9 service",
	Run: func(cmd *cobra.Command, args []string) {
		status, err := bindManager.StatusBind()
		if err != nil {
			log.Fatalf("Error fetching status: %v", err)
		}
		fmt.Printf("BIND9 status: %s\n", status)
	},
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Get statistics of BIND9",
	Run: func(cmd *cobra.Command, args []string) {
		stats, err := bindManager.GetStats()
		if err != nil {
			log.Fatalf("Error fetching stats: %v", err)
		}
		fmt.Printf("Total Zones: %d\n", stats.TotalZones)
		fmt.Printf("Total Routers: %d\n", stats.TotalRouters)
	},
}

var getDomainsCmd = &cobra.Command{
	Use:   "get-domains",
	Short: "Get all domains and their configuration files",
	Run: func(cmd *cobra.Command, args []string) {
		domains, err := bindManager.GetDomains()
		if err != nil {
			log.Fatalf("Error fetching domains: %v", err)
		}
		for domain, file := range domains {
			fmt.Printf("Domain: %s, File: %s\n", domain, file)
		}
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
		statsCmd,
		getDomainsCmd,
	)
}

func StartCLI() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

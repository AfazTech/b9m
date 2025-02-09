package cli

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/imafaz/B9CA/api"
	"github.com/imafaz/B9CA/controller"
	"github.com/spf13/cobra"
)

var bindManager = controller.NewBindManager("/etc/bind/zones", "/etc/bind/named.conf.local")

var rootCmd = &cobra.Command{
	Use:   "b9ca",
	Short: "B9CA Command Line Interface",
	Long:  "A CLI tool for managing BIND9 configurations and operations in B9CA.",
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

func init() {
	rootCmd.AddCommand(addDomainCmd, deleteDomainCmd, addRecordCmd, deleteRecordCmd, getRecordsCmd, startAPICmd)
}

func StartCLI() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

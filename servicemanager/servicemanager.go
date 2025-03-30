package servicemanager

import (
	"fmt"
	"os/exec"
	"strings"
)

func detectBindServiceName() string {
	services := []string{"bind9", "named"}
	for _, service := range services {
		cmd := exec.Command("systemctl", "list-units", "--type=service", "--all")
		output, _ := cmd.CombinedOutput()
		if strings.Contains(string(output), service+".service") {
			return service
		}
	}
	return "named"
}

var bindServiceName = detectBindServiceName()

func ReloadBind() error {
	cmd := exec.Command("rndc", "reload")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to reload Bind: %w | output: %s", err, string(output))
	}
	return nil
}

func RestartBind() error {
	cmd := exec.Command("systemctl", "restart", bindServiceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restart Bind: %w | output: %s", err, string(output))
	}
	return nil
}

func StopBind() error {
	cmd := exec.Command("systemctl", "stop", bindServiceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop Bind: %w | output: %s", err, string(output))
	}
	return nil
}

func StartBind() error {
	cmd := exec.Command("systemctl", "start", bindServiceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start Bind: %w | output: %s", err, string(output))
	}
	return nil
}

func StatusBind() (string, error) {
	cmd := exec.Command("systemctl", "is-active", bindServiceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get Bind status: %w | output: %s", err, string(output))
	}
	return strings.TrimSpace(string(output)), nil
}

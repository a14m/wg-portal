package internal

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/samber/lo"
)

var interfaceRegex = regexp.MustCompile(`^interface:\s+(.+)$`)

type WireGuardConnection struct {
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

func GetStatus() (string, error) {
	output, err := showStatus()
	if err != nil {
		return "", err
	}
	status := lo.FilterMap(strings.Split(string(output), "\n"), func(line string, _ int) (string, bool) {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "interface") {
			return fmt.Sprintf("Connection: %s", strings.TrimPrefix(line, "interface:")), true
		}
		if strings.Contains(line, "latest handshake") {
			return fmt.Sprintf("Latest Handshake: %s", strings.TrimPrefix(line, "latest handshake:")), true
		}
		if strings.Contains(line, "transfer") {
			return fmt.Sprintf("Transfer: %s", strings.TrimPrefix(line, "transfer:")), true
		}
		return "", false
	})
	// This is a simple check on whether a connection started or not.
	// Instead of complex logic on looping on connections and figuring which connection might be missing info.
	// NOTE: This doesn't handle if 3x connections were started and none of them is still active.
	// NOTE: ToggleConnection stops all active connections and activate one.
	// to avoid issues with multiple VPNs configuring the same iptable that could happen with default wireguard configs
	if len(status)%3 != 0 {
		status = append(status, "Connection starting...")
	}
	return strings.Join(status, "\n"), nil
}

func GetConnections() ([]*WireGuardConnection, error) {
	activeConnection, err := getActiveConnections()
	if err != nil {
		return nil, err
	}
	allConnections, err := getAllConnections()
	if err != nil {
		return nil, err
	}

	connections := make([]*WireGuardConnection, 0, len(allConnections))
	for _, i := range allConnections {
		connections = append(connections, &WireGuardConnection{
			Name:   i,
			Active: slices.Contains(activeConnection, i),
		})
	}
	return connections, nil
}

func ToggleConnection(name string) ([]byte, error) {
	allConnections, err := GetConnections()
	if err != nil {
		return nil, err
	}
	activeConnections := lo.Filter(allConnections, func(i *WireGuardConnection, _ int) bool {
		return i.Active
	})
	connection, err := getConnection(name)
	if err != nil {
		return nil, err
	}
	output, err := stopActiveConnections(activeConnections)
	if err != nil {
		return nil, err
	}
	startOutput, err := startConnection(connection)
	if err != nil {
		return nil, err
	}
	output = append(output, startOutput...)
	return output, nil
}

func stopActiveConnections(activeConnections []*WireGuardConnection) ([]byte, error) {
	var output []byte
	for _, activeConnection := range activeConnections {
		log.Printf("Stopping connection %s.", activeConnection.Name)
		cmd := exec.Command("sudo", "wg-quick", "down", activeConnection.Name)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return nil, err
		}
		output = append(output, out...)
		log.Printf("Successfully stopped connection %s.", activeConnection.Name)
	}
	return output, nil
}

func startConnection(connection *WireGuardConnection) ([]byte, error) {
	if connection.Active {
		return nil, nil
	}
	log.Printf("Starting connection %s.", connection.Name)
	cmd := exec.Command("sudo", "wg-quick", "up", connection.Name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	log.Printf("Successfully started connection %s.", connection.Name)
	return output, nil
}

// Get the list of all wireguard connections using config files
func getAllConnections() ([]string, error) {
	files, err := filepath.Glob("/etc/wireguard/*.conf")
	if err != nil {
		return nil, err
	}
	files = lo.Map(files, func(f string, _ int) string {
		return strings.TrimSuffix(filepath.Base(f), filepath.Ext(f))
	})
	return files, nil
}

// Get the list of active wireguard connections using wg show command
func getActiveConnections() ([]string, error) {
	var activeConnections []string
	status, err := showStatus()
	if err != nil {
		return nil, err
	}
	for line := range strings.SplitSeq(string(status), "\n") {
		if matches := interfaceRegex.FindStringSubmatch(strings.TrimSpace(line)); len(matches) > 1 {
			activeConnections = append(activeConnections, matches[1])
		}
	}
	return activeConnections, nil
}

func getConnection(name string) (*WireGuardConnection, error) {
	allConnections, err := GetConnections()
	if err != nil {
		return nil, err
	}
	connection, ok := lo.Find(allConnections, func(dev *WireGuardConnection) bool {
		return dev.Name == name
	})
	if !ok {
		return nil, fmt.Errorf("failed to find connection: %s", name)
	}
	return connection, nil
}

func showStatus() ([]byte, error) {
	cmd := exec.Command("sudo", "wg", "show")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute wg show: %w", err)
	}
	return output, nil
}

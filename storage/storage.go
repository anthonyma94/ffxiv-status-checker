package storage

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/anthonyma94/ffxiv-status-checker/model"
)

// LoadLastServerState loads the last saved state of the server from a file.
func LoadLastServerState(filename string) (*model.Server, error) {
	file, err := os.Open(filename)
	if err != nil {
		// If the file does not exist, return nil.
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var server model.Server
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&server); err != nil {
		return nil, err
	}
	return &server, nil
}

// SaveServerState saves the current state of the server to a file.
func SaveServerState(filename string, server *model.Server) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(server); err != nil {
		return err
	}
	return nil
}

// FileNameForServer generates a safe filename for the server state.
// The file name is hidden (prefixed with a dot) so that it can be gitignored.
func FileNameForServer(serverName string) string {
	// Example: ".ffxiv-status-checker_state_Faerie.json"
	return fmt.Sprintf(".ffxiv-status-checker_state_%s.json", serverName)
}

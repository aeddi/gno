package gas

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gnolang/gno/contribs/gasstat/pkg/file"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Measurements represents the gas Measurements for different operations.
type Measurements struct {
	Calls     []CallMeasurement     `json:"calls" yaml:"calls"`
	Runs      []RunMeasurement      `json:"runs" yaml:"runs"`
	Transfers []TransferMeasurement `json:"transfers" yaml:"transfers"`
	AddPkgs   []AddPkgMeasurement   `json:"add_pkgs" yaml:"add_pkgs"`
}

// CallMeasurement represents the gas measurement for a function call.
type CallMeasurement struct {
	PkgPath string `json:"pkg_path" yaml:"pkg_path"`
	Func    string `json:"func" yaml:"func"`
	ExecResult
}

// RunMeasurement represents the gas measurement for a run operation.
type RunMeasurement struct {
	Files []File `json:"file_list" yaml:"file_list"`
	ExecResult
}

// TransferMeasurement represents the gas measurement for a transfer operation.
type TransferMeasurement struct {
	FromAddress crypto.Address `json:"from_address" yaml:"from_address"`
	ToAddress   crypto.Address `json:"to_address" yaml:"to_address"`
	Amount      std.Coins      `json:"amount" yaml:"amount"`
	ExecResult
}

// AddPkgMeasurement represents the gas measurement for adding a package.
type AddPkgMeasurement struct {
	PkgPath string    `json:"pkg_path" yaml:"pkg_path"`
	Files   []File    `json:"mem_files" yaml:"mem_files"`
	Deposit std.Coins `json:"deposit" yaml:"deposit"`
	ExecResult
}

// ExecResult represents the result of an execution.
type ExecResult struct {
	GasFee    int64         `json:"gas_fee" yaml:"gas_fee"`
	GasWanted int64         `json:"gas_wanted" yaml:"gas_wanted"`
	GasUsed   int64         `json:"gas_used" yaml:"gas_used"`
	Duration  time.Duration `json:"duration" yaml:"duration"`
}

// File represents a file with its path, size, and SHA256 checksum.
type File struct {
	Path      string `json:"path" yaml:"path"`
	Size      int    `json:"size" yaml:"size"`
	Sha256Sum string `json:"sha256_sum" yaml:"sha256_sum"`
}

// LoadFromCsv loads gas measurements from a CSV file.
func LoadFromCsv(filename string) (*Measurements, error) {
	// TODO: Implement CSV loading logic.
	return nil, fmt.Errorf("CSV loading not implemented")
}

// LoadFromJson loads gas measurements from a JSON file.
func LoadFromJson(filename string) (*Measurements, error) {
	// Read the file content.
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON content into an Measurements struct.
	measurements := new(Measurements)
	if err := json.Unmarshal(content, &measurements); err != nil {
		return nil, err
	}

	return measurements, nil
}

// LoadFromFile loads gas measurements from a file based on the file extension.
func LoadFromFile(filename string) (*Measurements, error) {
	// Determine the file format based on the file extension.
	switch file.FileFormatFromExt(filename) {
	case file.CSV:
		return LoadFromCsv(filename)
	case file.JSON:
		return LoadFromJson(filename)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", filename)
	}
}

// SaveToCsv saves the gas measurements to a CSV file.
func (u *Measurements) SaveToCsv(filename string) error {
	// TODO: Implement CSV saving logic.
	return fmt.Errorf("CSV saving not implemented")
}

// SaveToJson saves the gas measurements to a JSON file.
func (u *Measurements) SaveToJson(filename string) error {
	// Marshal the Measurements struct into JSON.
	content, err := json.MarshalIndent(u, "", "  ")
	if err != nil {
		return err
	}

	// Write the JSON content to the file.
	if err := os.WriteFile(filename, content, 0644); err != nil {
		return err
	}

	return nil
}

// SaveToFile saves the gas measurements to a file based on the file extension.
func (u *Measurements) SaveToFile(filename string) error {
	// Determine the file format based on the file extension.
	switch file.FileFormatFromExt(filename) {
	case file.CSV:
		return u.SaveToCsv(filename)
	case file.JSON:
		return u.SaveToJson(filename)
	default:
		return fmt.Errorf("unsupported file format: %s", filename)
	}
}

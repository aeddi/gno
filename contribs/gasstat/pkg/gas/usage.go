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

// Usage represents the gas usage for different operations.
type Usage struct {
	Calls   []CallUsage   `json:"calls" yaml:"calls"`
	Runs    []RunUsage    `json:"runs" yaml:"runs"`
	Sends   []SendUsage   `json:"sends" yaml:"sends"`
	AddPkgs []AddPkgUsage `json:"add_pkgs" yaml:"add_pkgs"`
}

// CallUsage represents the gas usage for a function call.
type CallUsage struct {
	PkgPath string `json:"pkg_path" yaml:"pkg_path"`
	Func    string `json:"func" yaml:"func"`
	ExecResult
}

// RunUsage represents the gas usage for a run operation.
type RunUsage struct {
	Files []File `json:"file_list" yaml:"file_list"`
	ExecResult
}

// SendUsage represents the gas usage for a send operation.
type SendUsage struct {
	FromAddress crypto.Address `json:"from_address" yaml:"from_address"`
	ToAddress   crypto.Address `json:"to_address" yaml:"to_address"`
	Amount      std.Coins      `json:"amount" yaml:"amount"`
	ExecResult
}

// AddPkgUsage represents the gas usage for adding a package.
type AddPkgUsage struct {
	PkgPath string    `json:"pkg_path" yaml:"pkg_path"`
	Files   []File    `json:"mem_files" yaml:"mem_files"`
	Deposit std.Coins `json:"deposit" yaml:"deposit"`
	ExecResult
}

// ExecResult represents the result of an execution.
type ExecResult struct {
	Duration time.Duration `json:"duration" yaml:"duration"`
	GasUsed  int64         `json:"gas_used" yaml:"gas_used"`
}

// File represents a file with its path, size, and SHA256 checksum.
type File struct {
	Path      string `json:"path" yaml:"path"`
	Size      int    `json:"size" yaml:"size"`
	Sha256Sum string `json:"sha256_sum" yaml:"sha256_sum"`
}

// LoadFromTxt loads gas usage from a TXT file.
func LoadFromTxt(filename string) (*Usage, error) {
	// TODO: Implement TXT loading logic.
	return nil, fmt.Errorf("TXT loading not implemented")
}

// LoadFromCsv loads gas usage from a CSV file.
func LoadFromCsv(filename string) (*Usage, error) {
	// TODO: Implement CSV loading logic.
	return nil, fmt.Errorf("CSV loading not implemented")
}

// LoadFromJson loads gas usage from a JSON file.
func LoadFromJson(filename string) (*Usage, error) {
	// Read the file content.
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON content into an Usage struct.
	usage := new(Usage)
	if err := json.Unmarshal(content, &usage); err != nil {
		return nil, err
	}

	return usage, nil
}

// LoadFromFile loads gas usage from a file based on the file extension.
func LoadFromFile(filename string) (*Usage, error) {
	// Determine the file format based on the file extension.
	switch file.FileFormatFromExt(filename) {
	case file.TXT:
		return LoadFromTxt(filename)
	case file.CSV:
		return LoadFromCsv(filename)
	case file.JSON:
		return LoadFromJson(filename)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", filename)
	}
}

// SaveToTxt saves the gas usage to a CSV file.
func (u *Usage) SaveToTxt(filename string) error {
	return fmt.Errorf("TXT saving not implemented")
}

// SaveToCsv saves the gas usage to a CSV file.
func (u *Usage) SaveToCsv(filename string) error {
	return fmt.Errorf("CSV saving not implemented")
}

// SaveToJson saves the gas usage to a JSON file.
func (u *Usage) SaveToJson(filename string) error {
	// Marshal the Usage struct into JSON.
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

// SaveToFile saves the gas usage to a file based on the file extension.
func (u *Usage) SaveToFile(filename string) error {
	// Determine the file format based on the file extension.
	switch file.FileFormatFromExt(filename) {
	case file.TXT:
		return u.SaveToTxt(filename)
	case file.CSV:
		return u.SaveToCsv(filename)
	case file.JSON:
		return u.SaveToJson(filename)
	default:
		return fmt.Errorf("unsupported file format: %s", filename)
	}
}

package distribution

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/gnolang/gno/contribs/gasstat/pkg/file"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Distribution represents a distribution of balances (address:amount).
type Distribution struct {
	Balances []gnoland.Balance `json:"balances"`
}

// CSV headers for the address and amount columns.
const (
	CSVAccountHeader = "Account Address"
	CSVBalanceHeader = "Balance (ugnot)"
)

// validateAmount checks if the amount is in ugnot and positive.
func validateAmount(amount std.Coins) error {
	// Check if the amount is not empty.
	if len(amount) == 0 {
		return fmt.Errorf("amount is empty")
	}

	// Check if there is more than one amount.
	if len(amount) > 1 {
		return fmt.Errorf("more than one amount")
	}

	// Check if the amount is not in ugnot.
	if amount[0].Denom != ugnot.Denom {
		return fmt.Errorf("amount is not in ugnot")
	}

	// Check if the amount is negative.
	if amount[0].Amount < 0 {
		return fmt.Errorf("amount is negative")
	}

	return nil
}

// LoadFromTxt loads account balances from a text file.
// The file should contain one account per line, with the tm2 address and balance
// separated by an equals sign and followed by ugnot.
// Example: g1p3ucd3ptpw902fluyjzhq3ffgq4ntddatev7s5=42027010477582ugnot
func LoadFromTxt(filename string) (*Distribution, error) {
	// Open the file for reading.
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var (
		scanner = bufio.NewScanner(file)
		distrib = new(Distribution)
		lineNum = 0
	)

	// Read the file line by line.
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Remove comments and trim spaces.
		line = strings.Split(line, "#")[0]
		line = strings.TrimSpace(line)

		// Skip empty lines.
		if line == "" {
			continue
		}

		// Parse the line into an account balance.
		var balance gnoland.Balance
		if err := balance.Parse(line); err != nil {
			return nil, fmt.Errorf("unable to parse line %d: %w", lineNum, err)
		}

		// Delete all coins that are not in ugnot.
		balance.Amount = slices.DeleteFunc(balance.Amount, func(c std.Coin) bool {
			return c.Denom != ugnot.Denom
		})

		// Validate the balance amount.
		if err := validateAmount(balance.Amount); err != nil {
			return nil, fmt.Errorf("invalid balance at line %d: %w", lineNum, err)
		}

		// Append the balance to the distribution.
		distrib.Balances = append(distrib.Balances, balance)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning txt file failed: %w", err)
	}

	return distrib, nil
}

// LoadFromCsv loads account balances from a CSV file.
// The file should contain one account per line, with the tm2 address and balance
// separated by a comma. The first line optionally contains the column names.
// Example: g1p3ucd3ptpw902fluyjzhq3ffgq4ntddatev7s5,42027010477582
func LoadFromCsv(filename string) (*Distribution, error) {
	// Open the file for reading.
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var (
		csvReader = csv.NewReader(file)
		distrib   = new(Distribution)
		lineNum   = 0
	)

	// parseRecord is a helper function to parse a CSV record into a balance.
	parseRecord := func(record []string) (gnoland.Balance, error) {
		var balance gnoland.Balance

		// Parse the address from the first field.
		balance.Address, err = crypto.AddressFromBech32(record[0])
		if err != nil {
			return balance, fmt.Errorf("invalid address %q: %w", record[0], err)
		}

		// Parse the amount from the second field.
		amount, err := strconv.ParseInt(record[1], 10, 64)
		if err != nil {
			return balance, fmt.Errorf("invalid amount %q: %w", record[1], err)
		}

		// Check if the amount is negative.
		if amount < 0 {
			return balance, fmt.Errorf("amount %q is negative", record[1])
		}

		// Set the denom to ugnot as it is the only supported denom.
		balance.Amount = std.Coins{{Denom: ugnot.Denom, Amount: amount}}

		return balance, nil
	}

	// Read the CSV records line by line.
	for {
		lineNum++
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("unable to read line %d: %w", lineNum, err)
		}

		// Check if the record has exactly two fields.
		if len(record) != 2 {
			return nil, fmt.Errorf("invalid record at line %d: expected 2 fields, got %d", lineNum, len(record))
		}

		// Parse the record into an account balance.
		balance, err := parseRecord(record)
		if err != nil {
			// Check if the first line is the optional header.
			if lineNum == 1 &&
				(strings.TrimSpace(record[0]) != CSVAccountHeader || strings.TrimSpace(record[1]) != CSVBalanceHeader) {
				return nil, fmt.Errorf("invalid header or record at line 1: %w", err)
			} else if lineNum > 1 {
				return nil, fmt.Errorf("invalid record at line %d: %w", lineNum, err)
			}
		}

		// Append the balance to the distribution.
		distrib.Balances = append(distrib.Balances, balance)
	}

	return distrib, nil
}

// LoadFromJson loads account balances from a JSON file.
func LoadFromJson(filename string) (*Distribution, error) {
	// Read the file content.
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON content into a Distribution struct.
	distrib := new(Distribution)
	if err := json.Unmarshal(content, &distrib); err != nil {
		return nil, err
	}

	// Filter out all coins that are not in ugnot and validate each amount.
	for i, balance := range distrib.Balances {
		// Delete all coins that are not in ugnot.
		balance.Amount = slices.DeleteFunc(balance.Amount, func(c std.Coin) bool {
			return c.Denom != ugnot.Denom
		})

		// Validate each balance amount.
		if err := validateAmount(balance.Amount); err != nil {
			return nil, fmt.Errorf("invalid balance %d: %w", i, err)
		}
	}

	return distrib, nil
}

// LoadFromFile loads account balances from a file based on the file extension.
func LoadFromFile(filename string) (*Distribution, error) {
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

// BalanceAmounts returns a slice of the balance amounts.
func (d *Distribution) BalanceAmounts() []float64 {
	// Allocate a slice of float64 for the balance amounts.
	balances := make([]float64, len(d.Balances))

	// Validate each balance amount and append it to the slice.
	for i := range d.Balances {
		if err := validateAmount(d.Balances[i].Amount); err != nil {
			panic(fmt.Sprintf("invalid balance %d: %v", i, err))
		}
		balances[i] = float64(d.Balances[i].Amount[0].Amount)
	}

	return balances
}

// SaveToTxt saves the distribution to a text file.
func (d *Distribution) SaveToTxt(filename string) error {
	// Allocate a slice of strings for the balance lines.
	lines := make([]string, len(d.Balances))

	// Format each balance as <address>=<amount>ugnot.
	for i := range d.Balances {
		if err := validateAmount(d.Balances[i].Amount); err != nil {
			return fmt.Errorf("invalid balance %d: %w", i, err)
		}
		lines[i] = fmt.Sprintf("%s=%dugnot", d.Balances[i].Address, d.Balances[i].Amount[0].Amount)
	}

	// Write the balance lines to the file.
	content := []byte(strings.Join(lines, "\n"))
	if err := os.WriteFile(filename, content, 0644); err != nil {
		return err
	}

	return nil
}

// SaveToCsv saves the distribution to a CSV file.
func (d *Distribution) SaveToCsv(filename string) error {
	// Open the file for writing.
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("unable to create file %s: %w", filename, err)
	}
	defer file.Close()

	// Create a CSV writer.
	csvWriter := csv.NewWriter(file)

	// Write the CSV header.
	if err := csvWriter.Write([]string{CSVAccountHeader, CSVBalanceHeader}); err != nil {
		return fmt.Errorf("unable to write CSV header: %w", err)
	}

	// Validate and write each balance as a CSV record.
	for i, balance := range d.Balances {
		if err := validateAmount(balance.Amount); err != nil {
			return fmt.Errorf("invalid balance %d: %w", i, err)
		}

		// Format the balance as a CSV record.
		record := []string{
			balance.Address.String(),
			fmt.Sprintf("%d", balance.Amount[0].Amount),
		}

		// Write the record to the CSV file.
		if err := csvWriter.Write(record); err != nil {
			return fmt.Errorf("unable to write record %d: %v: %w", i, record, err)
		}
	}

	// Flush the CSV writer to write any buffered data to the file.
	csvWriter.Flush()

	return nil
}

// SaveToJson saves the distribution to a JSON file.
func (d *Distribution) SaveToJson(filename string) error {
	// Validate each balance amount.
	for i, balance := range d.Balances {
		if err := validateAmount(balance.Amount); err != nil {
			return fmt.Errorf("invalid balance %d: %w", i, err)
		}
	}

	// Marshal the Distribution struct into JSON.
	content, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}

	// Write the JSON content to the file.
	if err := os.WriteFile(filename, content, 0644); err != nil {
		return err
	}

	return nil
}

// SaveToFile saves the distribution to a file based on the file extension.
func (d *Distribution) SaveToFile(filename string) error {
	// Determine the file format based on the file extension.
	switch file.FileFormatFromExt(filename) {
	case file.TXT:
		return d.SaveToTxt(filename)
	case file.CSV:
		return d.SaveToCsv(filename)
	case file.JSON:
		return d.SaveToJson(filename)
	default:
		return fmt.Errorf("unsupported file format: %s", filename)
	}
}

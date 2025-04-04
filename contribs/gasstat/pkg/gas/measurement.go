package gas

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/gnolang/gno/contribs/gasstat/pkg/file"
	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"gopkg.in/yaml.v3"
)

const chainID = "dev"

// Config represents the configuration for gas measurement.
type Config struct {
	Accounts []string `json:"accounts" yaml:"accounts"`

	Calls   []CallCfg   `json:"calls" yaml:"calls"`
	Runs    []RunCfg    `json:"runs" yaml:"runs"`
	Sends   []SendCfg   `json:"sends" yaml:"sends"`
	AddPkgs []AddPkgCfg `json:"add_pkgs" yaml:"add_pkgs"`
}

// CallCfg represents the configuration for a function call.
type CallCfg struct {
	BaseCfg
	Send    std.Coins `json:"send" yaml:"send"`
	PkgPath string    `json:"pkg_path" yaml:"pkg_path"`
	Func    string    `json:"func" yaml:"func"`
	Args    []string  `json:"args" yaml:"args"`
}

// RunCfg represents the configuration for a run operation.
type RunCfg struct {
	BaseCfg
	Send     std.Coins        `json:"send" yaml:"send"`
	MemFiles []*gnovm.MemFile `json:"mem_files" yaml:"mem_files"`
}

// SendCfg represents the configuration for a send operation.
type SendCfg struct {
	BaseCfg
	ToAddress crypto.Address `json:"to_address" yaml:"to_address"`
	Amount    std.Coins      `json:"amount" yaml:"amount"`
}

// AddPkgCfg represents the configuration for adding a package.
type AddPkgCfg struct {
	BaseCfg
	PkgPath  string           `json:"pkg_path" yaml:"pkg_path"`
	MemFiles []*gnovm.MemFile `json:"mem_files" yaml:"mem_files"`
	Deposit  std.Coins        `json:"deposit" yaml:"deposit"`
}

// BaseCfg represents the base configuration for all operations.
type BaseCfg struct {
	AccountIndex int `json:"account_index" yaml:"account_index"`
	gnoclient.BaseTxCfg
}

// exec executes an operation based on the provided configuration.
func exec(clients []gnoclient.Client, clientCfg any) (*ExecResult, error) {
	// Sanity check for the account index.
	if clientCfg.(BaseCfg).AccountIndex < 0 || clientCfg.(BaseCfg).AccountIndex >= len(clients) {
		return nil, fmt.Errorf("invalid account index: %d", clientCfg.(BaseCfg).AccountIndex)
	}

	// Get the client for the specified account index.
	client := clients[clientCfg.(BaseCfg).AccountIndex]
	infos, err := client.Signer.Info()
	if err != nil {
		return nil, fmt.Errorf("unable to get signer infos: %w", err)
	}

	var (
		res   *ctypes.ResultBroadcastTxCommit
		start = time.Now()
	)

	// Run the appropriate operation based on the configuration type.
	switch cfg := clientCfg.(type) {
	case CallCfg:
		res, err = client.Call(cfg.BaseTxCfg, vm.MsgCall{
			Caller:  infos.GetAddress(),
			Send:    cfg.Send,
			PkgPath: cfg.PkgPath,
			Func:    cfg.Func,
			Args:    cfg.Args,
		})

	case RunCfg:
		msgRun := vm.NewMsgRun(infos.GetAddress(), cfg.Send, cfg.MemFiles)
		res, err = client.Run(cfg.BaseTxCfg, msgRun)

	case SendCfg:
		res, err = client.Send(cfg.BaseTxCfg, bank.MsgSend{
			FromAddress: infos.GetAddress(),
			ToAddress:   cfg.ToAddress,
			Amount:      cfg.Amount,
		})

	case AddPkgCfg:
		msgAddPkg := vm.NewMsgAddPackage(infos.GetAddress(), cfg.PkgPath, cfg.MemFiles)
		msgAddPkg.Deposit = cfg.Deposit
		res, err = client.AddPackage(cfg.BaseTxCfg, msgAddPkg)
	}

	// Check for errors in the response.
	if err != nil {
		return nil, err
	}
	if res.CheckTx.Error != nil {
		return nil, fmt.Errorf("check error: %w", err)
	}
	if res.DeliverTx.Error != nil {
		return nil, fmt.Errorf("delivery error: %w", err)
	}

	return &ExecResult{
		Duration: time.Since(start),
		GasUsed:  res.DeliverTx.GasUsed,
	}, nil
}

// memFilesToFiles converts a slice of MemFile to a slice of File.
func memFilesToFiles(pkgPath string, memFiles []*gnovm.MemFile) []File {
	files := make([]File, len(memFiles))

	for i, memFile := range memFiles {
		files[i] = File{
			Path:      path.Join(pkgPath, memFile.Name),
			Size:      len(memFile.Body),
			Sha256Sum: fmt.Sprintf("%x", crypto.Sha256([]byte(memFile.Body))),
		}
	}

	return files
}

// configFromFile loads the configuration from a file.
func configFromFile(filename string) (*Config, error) {
	// Read the file content.
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config := new(Config)

	// Determine the file format based on the file extension.
	switch file.FileFormatFromExt(filename) {
	case file.YAML:
		// Unmarshal the YAML content into an Config struct.
		if err := yaml.Unmarshal(content, &config); err != nil {
			return nil, err
		}
	case file.JSON:
		// Unmarshal the JSON content into an Config struct.
		if err := json.Unmarshal(content, &config); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported file format: %s", filename)
	}

	return config, nil
}

// Measure executes the gas measurement based on the provided config file and remote address.
func Measure(remote string, configFile string) (*Usage, error) {
	// Load the configuration from the file.
	config, err := configFromFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("unable to load config file %q: %w", configFile, err)
	}

	// Initialize the RPC client.
	rpcClient, err := rpcclient.NewHTTPClient(remote)
	if err != nil {
		return nil, fmt.Errorf("unable to create http client for %q: %w", remote, err)
	}

	// Initialize one gnoclient for each account.
	clients := make([]gnoclient.Client, len(config.Accounts))
	for i, account := range config.Accounts {
		signer, err := gnoclient.SignerFromBip39(account, chainID, "", 0, 0)
		if err != nil {
			return nil, fmt.Errorf("unable to create signer from mnemonic %q: %w", account, err)
		}
		clients[i] = gnoclient.Client{
			Signer:    signer,
			RPCClient: rpcClient,
		}
	}

	// Allocate the usage struct slices.
	usage := &Usage{
		Calls:   make([]CallUsage, len(config.Calls)),
		Runs:    make([]RunUsage, len(config.Runs)),
		Sends:   make([]SendUsage, len(config.Sends)),
		AddPkgs: make([]AddPkgUsage, len(config.AddPkgs)),
	}

	// Execute all calls.
	for i, callCfg := range config.Calls {
		res, err := exec(clients, callCfg)
		if err != nil {
			return nil, fmt.Errorf("unable to call: %w", err)
		}

		usage.Calls[i] = CallUsage{
			PkgPath:    callCfg.PkgPath,
			Func:       callCfg.Func,
			ExecResult: *res,
		}
	}

	// Execute all runs.
	for i, runCfg := range config.Runs {
		res, err := exec(clients, runCfg)
		if err != nil {
			return nil, fmt.Errorf("unable to run: %w", err)
		}

		usage.Runs[i] = RunUsage{
			Files:      memFilesToFiles("", runCfg.MemFiles),
			ExecResult: *res,
		}
	}

	// Execute all sends.
	for i, sendCfg := range config.Sends {
		res, err := exec(clients, sendCfg)
		if err != nil {
			return nil, fmt.Errorf("unable to send: %w", err)
		}

		keyInfo, err := clients[sendCfg.BaseCfg.AccountIndex].Signer.Info()
		if err != nil {
			return nil, fmt.Errorf("unable to get signer info: %w", err)
		}

		usage.Sends[i] = SendUsage{
			FromAddress: keyInfo.GetAddress(),
			ToAddress:   sendCfg.ToAddress,
			Amount:      sendCfg.Amount,
			ExecResult:  *res,
		}
	}

	// Execute all addPkgs.
	for i, addPkgCfg := range config.AddPkgs {
		res, err := exec(clients, addPkgCfg)
		if err != nil {
			return nil, fmt.Errorf("unable to add package: %w", err)
		}

		usage.AddPkgs[i] = AddPkgUsage{
			PkgPath:    addPkgCfg.PkgPath,
			Files:      memFilesToFiles(addPkgCfg.PkgPath, addPkgCfg.MemFiles),
			Deposit:    addPkgCfg.Deposit,
			ExecResult: *res,
		}
	}

	return usage, nil
}

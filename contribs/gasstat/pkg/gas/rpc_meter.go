package gas

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/gnolang/gno/contribs/gasstat/pkg/file"
	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"gopkg.in/yaml.v3"
)

// RPCMeterConfig represents the configuration for the RPC gas meter.
type RPCMeterConfig struct {
	ChainID string `json:"chain_id" yaml:"chain_id"`

	Accounts []string `json:"accounts" yaml:"accounts"`

	Calls     []CallCfg     `json:"calls" yaml:"calls"`
	Runs      []RunCfg      `json:"runs" yaml:"runs"`
	Transfers []TransferCfg `json:"transfers" yaml:"transfers"`
	AddPkgs   []AddPkgCfg   `json:"add_pkgs" yaml:"add_pkgs"`
}

// CallCfg represents the configuration for a function call.
type CallCfg struct {
	BaseCfg `json:",inline" yaml:",inline"`
	Send    std.Coins  `json:"send" yaml:"send"`
	PkgPath string     `json:"pkg_path" yaml:"pkg_path"`
	Func    string     `json:"func" yaml:"func"`
	Args    [][]string `json:"args" yaml:"args"`
}

// RunCfg represents the configuration for a run operation.
type RunCfg struct {
	BaseCfg  `json:",inline" yaml:",inline"`
	Send     std.Coins      `json:"send" yaml:"send"`
	MemFiles []*std.MemFile `json:"mem_files" yaml:"mem_files"`
}

// TransferCfg represents the configuration for a bank transfer operation.
type TransferCfg struct {
	BaseCfg        `json:",inline" yaml:",inline"`
	ToAccountIndex int         `json:"to_account_index" yaml:"to_account_index"`
	Amounts        []std.Coins `json:"amounts" yaml:"amounts"`
}

// AddPkgCfg represents the configuration for adding a package.
type AddPkgCfg struct {
	BaseCfg  `json:",inline" yaml:",inline"`
	PkgPath  string         `json:"pkg_path" yaml:"pkg_path"`
	MemFiles []*std.MemFile `json:"mem_files" yaml:"mem_files"`
	Deposit  std.Coins      `json:"deposit" yaml:"deposit"`
}

// BaseCfg represents the base configuration for all operations.
type BaseCfg struct {
	AccountIndex        int `json:"account_index" yaml:"account_index"`
	gnoclient.BaseTxCfg `json:",inline" yaml:",inline"`
}

// execRequest executes a request and returns the gas usage and duration.
func execRequest(
	clients []gnoclient.Client,
	accountIndex int,
	doReq func(gnoclient.Client, crypto.Address) (*ctypes.ResultBroadcastTxCommit, error),
) (*ExecResult, error) {
	// Sanity check for the account index.
	if accountIndex < 0 || accountIndex >= len(clients) {
		return nil, fmt.Errorf("invalid account index: %d", accountIndex)
	}

	// Get the client for the specified account index.
	client := clients[accountIndex]
	infos, err := client.Signer.Info()
	if err != nil {
		return nil, fmt.Errorf("unable to get signer infos: %w", err)
	}

	// Measure the time taken for the request.
	start := time.Now()

	// Execute the request.
	res, err := doReq(client, infos.GetAddress())
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
func memFilesToFiles(pkgPath string, memFiles []*std.MemFile) []File {
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
func configFromFile(filename string) (*RPCMeterConfig, error) {
	// Read the file content.
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config := new(RPCMeterConfig)

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

// feeToInt64 converts a ugnot fee string to an int64 value.
func feeToInt64(fee string) (int64, error) {
	feeCoins, err := std.ParseCoin(fee)
	if err != nil {
		return 0, fmt.Errorf("unable to parse fee (%s): %w", fee, err)
	}

	if feeCoins.Denom != ugnot.Denom {
		return 0, fmt.Errorf("wrong fee denom (%s): must be %s", fee, ugnot.Denom)
	}

	return feeCoins.Amount, nil
}

// measureCalls executes the gas measurement for all call operations.
func measureCalls(clients []gnoclient.Client, callCfgs []CallCfg) ([]CallMeasurement, error) {
	// Count the number of calls to allocate.
	callsCount := 0
	for _, callCfg := range callCfgs {
		callsCount += len(callCfg.Args)
	}
	calls := make([]CallMeasurement, callsCount)

	// Execute and measure all call operations.
	resCount := 0
	for i, callCfg := range callCfgs {
		for j, args := range callCfg.Args {
			// Send the RPC request for this call operation.
			res, err := execRequest(
				clients,
				callCfg.AccountIndex,
				func(client gnoclient.Client, addr crypto.Address) (*ctypes.ResultBroadcastTxCommit, error) {
					return client.Call(callCfg.BaseTxCfg, vm.MsgCall{
						Caller:  addr,
						Send:    callCfg.Send,
						PkgPath: callCfg.PkgPath,
						Func:    callCfg.Func,
						Args:    args,
					})
				},
			)
			if err != nil {
				return nil, fmt.Errorf("unable to exec call %d args %d: %w: %+v", i, j, err, callCfg)
			}

			// Convert the fee string to int64.
			fee, err := feeToInt64(callCfg.GasFee)
			if err != nil {
				return nil, fmt.Errorf("unable to convert fee on call %d: %w: %+v", i, err, callCfg)
			}
			res.GasFee = fee
			res.GasWanted = callCfg.GasWanted

			// Store the result in the calls slice.
			calls[resCount] = CallMeasurement{
				PkgPath:    callCfg.PkgPath,
				Func:       callCfg.Func,
				ExecResult: *res,
			}
			resCount++
		}
	}

	return calls, nil
}

// measureRuns executes the gas measurement for all run operations.
func measureRuns(clients []gnoclient.Client, runCfgs []RunCfg) ([]RunMeasurement, error) {
	runs := make([]RunMeasurement, len(runCfgs))

	// Execute and measure all run operations.
	for i, runCfg := range runCfgs {
		// Send the RPC request for this run operation.
		res, err := execRequest(
			clients,
			runCfg.AccountIndex,
			func(client gnoclient.Client, addr crypto.Address) (*ctypes.ResultBroadcastTxCommit, error) {
				msgRun := vm.NewMsgRun(addr, runCfg.Send, runCfg.MemFiles)
				return client.Run(runCfg.BaseTxCfg, msgRun)
			},
		)
		if err != nil {
			return nil, fmt.Errorf("unable to exec run %d: %w: %+v", i, err, runCfg)
		}

		// Convert the fee string to int64.
		fee, err := feeToInt64(runCfg.GasFee)
		if err != nil {
			return nil, fmt.Errorf("unable to convert fee on run %d: %w: %+v", i, err, runCfg)
		}
		res.GasFee = fee
		res.GasWanted = runCfg.GasWanted

		// Store the result in the runs slice.
		runs[i] = RunMeasurement{
			Files:      memFilesToFiles("", runCfg.MemFiles),
			ExecResult: *res,
		}
	}

	return runs, nil
}

// measureTransfers executes the gas measurement for all transfer operations.
func measureTransfers(clients []gnoclient.Client, transferCfgs []TransferCfg) ([]TransferMeasurement, error) {
	// Count the number of transfers to allocate.
	transfersCount := 0
	for _, transferCfg := range transferCfgs {
		transfersCount += len(transferCfg.Amounts)
	}
	transfers := make([]TransferMeasurement, transfersCount)

	// Execute and measure all transfer operations.
	resCount := 0
	for i, transferCfg := range transferCfgs {
		for j, amount := range transferCfg.Amounts {
			// Sanity check for the account index.
			if transferCfg.ToAccountIndex < 0 || transferCfg.ToAccountIndex >= len(clients) {
				return nil, fmt.Errorf("invalid account index: %d", transferCfg.ToAccountIndex)
			}

			// Get the signer info for the specified account index.
			toInfo, err := clients[transferCfg.ToAccountIndex].Signer.Info()
			if err != nil {
				return nil, fmt.Errorf("unable to get signer infos: %w", err)
			}

			// Send the RPC request for this transfer operation.
			res, err := execRequest(
				clients,
				transferCfg.AccountIndex,
				func(client gnoclient.Client, addr crypto.Address) (*ctypes.ResultBroadcastTxCommit, error) {
					return client.Send(transferCfg.BaseTxCfg, bank.MsgSend{
						FromAddress: addr,
						ToAddress:   toInfo.GetAddress(),
						Amount:      amount,
					})
				},
			)
			if err != nil {
				return nil, fmt.Errorf("unable to exec transfer %d amount %d: %w: %+v", i, j, err, transferCfg)
			}

			// Convert the fee string to int64.
			fee, err := feeToInt64(transferCfg.GasFee)
			if err != nil {
				return nil, fmt.Errorf("unable to convert fee on transfer %d: %w: %+v", i, err, transferCfg)
			}
			res.GasFee = fee
			res.GasWanted = transferCfg.GasWanted

			// Store the result in the transfers slice.
			fromInfo, _ := clients[transferCfg.BaseCfg.AccountIndex].Signer.Info()
			transfers[resCount] = TransferMeasurement{
				FromAddress: fromInfo.GetAddress(),
				ToAddress:   toInfo.GetAddress(),
				Amount:      amount,
				ExecResult:  *res,
			}
			resCount++
		}
	}

	return transfers, nil
}

// measureAddPkgs executes the gas measurement for all add package operations.
func measureAddPkgs(clients []gnoclient.Client, addPkgCfgs []AddPkgCfg) ([]AddPkgMeasurement, error) {
	addPkgs := make([]AddPkgMeasurement, len(addPkgCfgs))

	// Execute and measure all add package operations.
	for i, addPkgCfg := range addPkgCfgs {
		// Send the RPC request for this add package operation.
		res, err := execRequest(
			clients,
			addPkgCfg.AccountIndex,
			func(client gnoclient.Client, addr crypto.Address) (*ctypes.ResultBroadcastTxCommit, error) {
				msgAddPkg := vm.NewMsgAddPackage(addr, addPkgCfg.PkgPath, addPkgCfg.MemFiles)
				msgAddPkg.MaxDeposit = addPkgCfg.Deposit
				return client.AddPackage(addPkgCfg.BaseTxCfg, msgAddPkg)
			},
		)
		if err != nil {
			return nil, fmt.Errorf("unable to exec add_pkg %d: %w: %+v", i, err, addPkgCfg)
		}

		// Convert the fee string to int64.
		fee, err := feeToInt64(addPkgCfg.GasFee)
		if err != nil {
			return nil, fmt.Errorf("unable to convert fee on add_pkg %d: %w: %+v", i, err, addPkgCfg)
		}
		res.GasFee = fee
		res.GasWanted = addPkgCfg.GasWanted

		// Store the result in the addPkgs slice.
		addPkgs[i] = AddPkgMeasurement{
			PkgPath:    addPkgCfg.PkgPath,
			Files:      memFilesToFiles(addPkgCfg.PkgPath, addPkgCfg.MemFiles),
			Deposit:    addPkgCfg.Deposit,
			ExecResult: *res,
		}
	}

	return addPkgs, nil
}

// Measure executes the gas measurement based on the provided config file and remote address.
func Measure(remote string, configFile string) (*Measurements, error) {
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
		signer, err := gnoclient.SignerFromBip39(account, config.ChainID, "", 0, 0)
		if err != nil {
			return nil, fmt.Errorf("unable to create signer from mnemonic %q: %w", account, err)
		}
		clients[i] = gnoclient.Client{
			Signer:    signer,
			RPCClient: rpcClient,
		}
	}

	// Measure the call operations.
	calls, err := measureCalls(clients, config.Calls)
	if err != nil {
		return nil, fmt.Errorf("unable to measure calls: %w", err)
	}

	// Measure the run operations.
	runs, err := measureRuns(clients, config.Runs)
	if err != nil {
		return nil, fmt.Errorf("unable to measure runs: %w", err)
	}

	// Measure the transfer operations.
	transfers, err := measureTransfers(clients, config.Transfers)
	if err != nil {
		return nil, fmt.Errorf("unable to measure transfers: %w", err)
	}

	// Measure the add package operations.
	addPkgs, err := measureAddPkgs(clients, config.AddPkgs)
	if err != nil {
		return nil, fmt.Errorf("unable to measure add packages: %w", err)
	}

	return &Measurements{
		Calls:     calls,
		Runs:      runs,
		Transfers: transfers,
		AddPkgs:   addPkgs,
	}, nil
}

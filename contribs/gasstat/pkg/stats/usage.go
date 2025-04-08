package stats

import (
	"sort"

	"github.com/gnolang/gno/contribs/gasstat/pkg/gas"
)

// Usage represents the gas usage statistics for different operations.
type Usage struct {
	RealmCalls    Realms     `json:"realm_calls" yaml:"realm_calls"`
	BankTransfers AverageGas `json:"bank_transfers" yaml:"bank_transfers"`
}

// Realms represents the gas usage statistics for all Realms.
type Realms struct {
	AverageGas `json:",inline" yaml:",inline"`
	Realms     []Realm `json:"realms" yaml:"realms"`
}

// Realm represents the gas usage statistics for a specific Realm and its functions.
type Realm struct {
	PkgPath    string `json:"pkg_path" yaml:"pkg_path"`
	AverageGas `json:",inline" yaml:",inline"`
	Funcs      []Func `json:"funcs" yaml:"funcs"`
}

// Func represents the gas usage statistics for a specific function in a Realm.
type Func struct {
	Name       string `json:"name" yaml:"name"`
	AverageGas `json:",inline" yaml:",inline"`
}

// AverageGas represents the average gas values for a set of transactions.
type AverageGas struct {
	Fee    int64 `json:"average_gas_fee" yaml:"average_gas_fee"`
	Wanted int64 `json:"average_gas_wanted" yaml:"average_gas_wanted"`
	Used   int64 `json:"average_gas_used" yaml:"average_gas_used"`
	Count  int64 `json:"count" yaml:"count"`
}

// gasSumCalculator is a helper struct to calculate the sum of gas values.
type gasSumCalculator struct {
	gasFeeSum    int64
	gasWantedSum int64
	gasUsedSum   int64
	count        int64
}

// add adds the gas values from a measurement to the calculator.
func (gsc *gasSumCalculator) add(measurement *gas.ExecResult) {
	gsc.gasFeeSum += measurement.GasFee
	gsc.gasWantedSum += measurement.GasWanted
	gsc.gasUsedSum += measurement.GasUsed
	gsc.count++
}

// addGasSum adds the gas sums from another calculator to this one.
func (gsc *gasSumCalculator) addGasSum(gasSum *gasSumCalculator) {
	gsc.gasFeeSum += gasSum.gasFeeSum
	gsc.gasWantedSum += gasSum.gasWantedSum
	gsc.gasUsedSum += gasSum.gasUsedSum
	gsc.count += gasSum.count
}

// average calculates the average gas values from the sums.
func (gsc *gasSumCalculator) average() AverageGas {
	if gsc.count == 0 {
		return AverageGas{}
	}
	return AverageGas{
		Fee:    gsc.gasFeeSum / gsc.count,
		Wanted: gsc.gasWantedSum / gsc.count,
		Used:   gsc.gasUsedSum / gsc.count,
		Count:  gsc.count,
	}
}

// computeRealmCalls calculates the average gas usage for all Realms and their functions.
func computeRealmCalls(measurements []*gas.Measurements) Realms {
	// Group measurements by package path and function name using a 2D map.
	realmsMap := map[string]map[string][]*gas.CallMeasurement{}

	for _, m := range measurements {
		for _, call := range m.Calls {
			if _, ok := realmsMap[call.PkgPath]; !ok {
				realmsMap[call.PkgPath] = make(map[string][]*gas.CallMeasurement)
			}
			realmsMap[call.PkgPath][call.Func] = append(realmsMap[call.PkgPath][call.Func], &call)
		}
	}

	var (
		realmIndex   = 0
		realms       = make([]Realm, len(realmsMap))
		realmsGasSum = &gasSumCalculator{}
	)

	// Iterate over each Realm.
	for pkgPath := range realmsMap {
		var (
			funcIndex   = 0
			funcs       = make([]Func, len(realmsMap[pkgPath]))
			realmGasSum = &gasSumCalculator{}
		)

		// Iterate over each Realm function.
		for funcName := range realmsMap[pkgPath] {
			funcGasSum := &gasSumCalculator{}

			// Iterate over each Realm function call.
			for _, call := range realmsMap[pkgPath][funcName] {
				// Sum the gas values for the function.
				funcGasSum.add(&call.ExecResult)
			}

			// Add the function to the functions list of the Realm.
			funcs[funcIndex] = Func{
				Name:       funcName,
				AverageGas: funcGasSum.average(),
			}

			// Sum the gas values for the Realm.
			realmGasSum.addGasSum(funcGasSum)
			funcIndex++
		}

		// Sort the functions by name.
		sort.Slice(funcs, func(i, j int) bool {
			return funcs[i].Name < funcs[j].Name
		})

		// Add the Realm to the Realms list.
		realms[realmIndex] = Realm{
			PkgPath:    pkgPath,
			Funcs:      funcs,
			AverageGas: realmGasSum.average(),
		}

		// Sum the gas values for all Realms.
		realmsGasSum.addGasSum(realmGasSum)
		realmIndex++
	}

	// Sort the Realms by package path.
	sort.Slice(realms, func(i, j int) bool {
		return realms[i].PkgPath < realms[j].PkgPath
	})

	return Realms{
		Realms:     realms,
		AverageGas: realmsGasSum.average(),
	}
}

// computeBankTransfer calculates the average gas usage for all bank transfers.
func computeBankTransfer(measurements []*gas.Measurements) AverageGas {
	var gasSum gasSumCalculator

	for _, m := range measurements {
		for _, t := range m.Transfers {
			gasSum.add(&t.ExecResult)
		}
	}

	return gasSum.average()
}

// NewUsage creates a new Usage instance from the given measurements.
func NewUsage(measurements []*gas.Measurements) *Usage {
	return &Usage{
		RealmCalls:    computeRealmCalls(measurements),
		BankTransfers: computeBankTransfer(measurements),
		// TODO: compute Run and AddPkg usage
	}
}

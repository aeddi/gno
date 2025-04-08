package stats

import (
	"slices"

	"github.com/gnolang/gno/contribs/gasstat/pkg/balances"
	"github.com/montanaflynn/stats"
)

// Distribution represents statistics about a list of balances, including the
// total balance, average balance, weight of the top and bottom 1000 accounts,
// variance, deviation, entropy, and balance percentiles.
type Distribution struct {
	AccountCount int `json:"account_count"`

	Total struct {
		All                     float64 `json:"all"`
		Top1000                 float64 `json:"top1000,omitzero"`
		Bottom1000              float64 `json:"bottom1000,omitzero"`
		AllWithoutTopBottom1000 float64 `json:"all_without_top_bottom_1000,omitzero"`
	} `json:"total"`

	Average struct {
		All                     float64 `json:"all"`
		Top1000                 float64 `json:"top1000,omitzero"`
		Bottom1000              float64 `json:"bottom1000,omitzero"`
		AllWithoutTopBottom1000 float64 `json:"all_without_top_bottom_1000,omitzero"`
	} `json:"average"`

	Weight struct {
		Top1000    float64 `json:"top1000,omitzero"`
		Bottom1000 float64 `json:"bottom1000,omitzero"`
	} `json:"weight,omitzero"`

	Variance  float64 `json:"variance"`
	Deviation float64 `json:"deviation"`
	Entropy   float64 `json:"entropy"`

	Percentiles map[uint8]float64 `json:"percentiles"`
}

// NewDistribution creates a new Distribution statistics instance from a list of balances.
func NewDistribution(distrib *balances.Balances) *Distribution {
	// Helper function to calculate stats and panic on error.
	mustStats := func(data stats.Float64Data, f func(stats.Float64Data) (float64, error)) float64 {
		sum, err := f(data)
		if err != nil {
			panic(err)
		}
		return sum
	}

	// Create a new statistics instance with the balance count.
	statistics := &Distribution{
		AccountCount: len(distrib.Balances),
	}

	// Get the balances and sort them in ascending order.
	balances := distrib.BalanceAmounts()
	slices.Sort(balances)

	// Calculate the total balances.
	statistics.Total.All = mustStats(balances, stats.Sum)

	// Calculate the average balances.
	statistics.Average.All = statistics.Total.All / float64(statistics.AccountCount)

	// If there are more than 2000 accounts, calculate the top and bottom 1000 average, total and weight.
	if len(balances) > 2000 {
		// Calculate the top and bottom 1000 total balances.
		statistics.Total.Top1000 = mustStats(balances[len(balances)-1000:], stats.Sum)
		statistics.Total.Bottom1000 = mustStats(balances[:1000], stats.Sum)
		statistics.Total.AllWithoutTopBottom1000 = statistics.Total.All - statistics.Total.Top1000 - statistics.Total.Bottom1000

		// Calculate the weight of the top and bottom 1000 accounts.
		statistics.Weight.Top1000 = statistics.Total.Top1000 / statistics.Total.All * 100.0
		statistics.Weight.Bottom1000 = statistics.Total.Bottom1000 / statistics.Total.All * 100.0

		// Calculate the top and bottom 1000 average balances.
		statistics.Average.Top1000 = statistics.Total.Top1000 / 1000.
		statistics.Average.Bottom1000 = statistics.Total.Bottom1000 / 1000.
		statistics.Average.AllWithoutTopBottom1000 = statistics.Total.AllWithoutTopBottom1000 / float64(statistics.AccountCount-2000)
	}

	// Calculate miscellaneous stats.
	statistics.Entropy = mustStats(balances, stats.Entropy)
	statistics.Variance = mustStats(balances, stats.Variance)
	statistics.Deviation = mustStats(balances, stats.StandardDeviation)

	// Calculate the balance percentiles.
	statistics.Percentiles = make(map[uint8]float64, 100)
	for percentile := uint8(1); percentile <= 100; percentile++ {
		balance, err := stats.Percentile(balances, float64(percentile))
		if err != nil {
			panic(err)
		}

		statistics.Percentiles[percentile] = balance
	}

	return statistics
}

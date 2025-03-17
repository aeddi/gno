package stats

import (
	"slices"

	"github.com/gnolang/gno/contribs/gasstat/pkg/distribution"
	"github.com/montanaflynn/stats"
)

// Statistics represents statistics about a distribution, including the
// total balance, average balance, weight of the top and bottom 1000 accounts,
// variance, deviation, entropy, and balance percentiles.
type Statistics struct {
	AccountCount int `json:"account_count"`

	Total struct {
		All                     float64 `json:"all"`
		Top1000                 float64 `json:"top1000"`
		Bottom1000              float64 `json:"bottom1000"`
		AllWithoutTopBottom1000 float64 `json:"all_without_top_bottom_1000"`
	} `json:"total"`

	Average struct {
		All                     float64 `json:"all"`
		Top1000                 float64 `json:"top1000"`
		Bottom1000              float64 `json:"bottom1000"`
		AllWithoutTopBottom1000 float64 `json:"all_without_top_bottom_1000"`
	} `json:"average"`

	Weight struct {
		Top1000    float64 `json:"top1000"`
		Bottom1000 float64 `json:"bottom1000"`
	} `json:"weight"`

	Variance  float64 `json:"variance"`
	Deviation float64 `json:"deviation"`
	Entropy   float64 `json:"entropy"`

	Percentiles map[uint8]float64 `json:"percentiles"`
}

// NewFromDistribution creates a new statistics instance from a Distribution.
func NewFromDistribution(distrib *distribution.Distribution) *Statistics {
	// Helper function to calculate stats and panic on error.
	mustStats := func(data stats.Float64Data, f func(stats.Float64Data) (float64, error)) float64 {
		sum, err := f(data)
		if err != nil {
			panic(err)
		}
		return sum
	}

	// Create a new statistics instance with the balance count.
	statistics := &Statistics{
		AccountCount: len(distrib.Balances),
	}

	// Get the balances and sort them in ascending order.
	balances := distrib.BalanceAmounts()
	slices.Sort(balances)

	// Calculate the total balances.
	statistics.Total.All = mustStats(balances, stats.Sum)
	statistics.Total.Top1000 = mustStats(balances[len(balances)-1000:], stats.Sum)
	statistics.Total.Bottom1000 = mustStats(balances[:1000], stats.Sum)
	statistics.Total.AllWithoutTopBottom1000 = statistics.Total.All - statistics.Total.Top1000 - statistics.Total.Bottom1000

	// Calculate the weight of the top and bottom 1000 accounts.
	statistics.Weight.Top1000 = statistics.Total.Top1000 / statistics.Total.All * 100.0
	statistics.Weight.Bottom1000 = statistics.Total.Bottom1000 / statistics.Total.All * 100.0

	// Calculate the average balances.
	statistics.Average.All = statistics.Total.All / float64(statistics.AccountCount)
	statistics.Average.Top1000 = statistics.Total.Top1000 / 1000.
	statistics.Average.Bottom1000 = statistics.Total.Bottom1000 / 1000.
	statistics.Average.AllWithoutTopBottom1000 = statistics.Total.AllWithoutTopBottom1000 / float64(statistics.AccountCount-2000)

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

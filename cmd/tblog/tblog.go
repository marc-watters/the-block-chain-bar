package main

import "github.com/davecgh/go-spew/spew"

type Trx struct {
	from  string
	to    string
	value uint64
}

func main() {
	// Extra touch pre-defining the Map length to avoid reallocation
	trxMap := make(map[string]Trx, 3)
	trxMap["trx1"] = Trx{"andrej", "babayaga", 10}

	getTrxAsSlice(trxMap)
}

func getTrxAsSlice(trxMap map[string]Trx) []Trx {
	// Defines the Slice capacity to match the Map elements count
	trxs := make([]Trx, 0, 3)
	spew.Dump(trxs[0])

	for _, trx := range trxMap {
		trxs = append(trxs, trx)
	}

	spew.Dump(trxs)

	return trxs
}

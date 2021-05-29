package randomx

import (
	"encoding/hex"
	"log"
	"runtime"
	"sync"

	"github.com/minernl/go-miningcore/go-randomx"
)

// Randomx algo
type Randomx struct {
	randomxVM      randomx.VM
	randomxCache   randomx.Cache
	randomxDataset randomx.Dataset
	RandomxReady   bool
	seedHash       string
}

//if !s.config.BypassShareValidation

// IsNewSeed compare saved seed
func (r *Randomx) IsNewSeed(seedHash string) bool {
	return seedHash != r.seedHash
}

// Init randomx dataset
func (r *Randomx) Init(seedHash string) error {
	var err error = nil
	if seedHash != r.seedHash {

		log.Println("randomx new seed init: ", seedHash)

		if r.randomxCache != nil {
			randomx.ReleaseCache(r.randomxCache)
		}
		if r.randomxDataset != nil {
			randomx.ReleaseDataset(r.randomxDataset)
		}
		if r.randomxVM != nil {
			randomx.DestroyVM(r.randomxVM)
		}
		r.randomxCache, err = randomx.AllocCache(randomx.FlagFullMEM, randomx.FlagJIT) // without lagePage to avoid panic
		if err != nil {
			return err
		}
		r.randomxDataset, err = randomx.AllocDataset(randomx.FlagFullMEM, randomx.FlagJIT)
		if err != nil {
			return err
		}
		r.seedHash = seedHash
		seed, err := hex.DecodeString(r.seedHash)
		if err != nil {
			return err
		}
		log.Println("randomx init dataset start")
		randomx.InitCache(r.randomxCache, seed)
		//log.Println("rxCache initialization finished")
		count := randomx.DatasetItemCount()
		log.Println("dataset count:", count)
		//randomx.InitDataset(r.randomxDataset, r.randomxCache, 0, count)
		var wg sync.WaitGroup
		var workerNum = uint32(runtime.NumCPU())
		for i := uint32(0); i < workerNum; i++ {
			wg.Add(1)
			a := (count * i) / workerNum
			b := (count * (i + 1)) / workerNum
			go func() {
				defer wg.Done()
				randomx.InitDataset(r.randomxDataset, r.randomxCache, a, b-a)
			}()
		}
		wg.Wait()
		log.Println("randomx init dataset finish")
		r.randomxVM, err = randomx.CreateVM(r.randomxCache, r.randomxDataset, randomx.FlagFullMEM, randomx.FlagJIT)
		if err != nil {
			return err
		}
		r.RandomxReady = true
	}

	return err
}

// CalcHash randomx
func (r *Randomx) CalcHash(convertedBlob []byte) []byte {
	hashBytes := randomx.CalculateHash(r.randomxVM, convertedBlob)
	return hashBytes
}

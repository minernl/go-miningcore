package stratum

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"log"
	"math/big"

	"github.com/minernl/go-miningcore/cnutil"
)

type BlockTemplate struct {
	diffInt64      int64
	height         int64
	difficulty     *big.Int
	reservedOffset int
	prevHash       string
	seedHash       string
	buffer         []byte
}

func (b *BlockTemplate) nextBlob(extraNonce uint32, instanceId []byte) string {
	extraBuff := new(bytes.Buffer)
	binary.Write(extraBuff, binary.BigEndian, extraNonce)

	blobBuff := make([]byte, len(b.buffer))
	copy(blobBuff, b.buffer)
	copy(blobBuff[b.reservedOffset+4:b.reservedOffset+7], instanceId)
	copy(blobBuff[b.reservedOffset:], extraBuff.Bytes())
	blob := cnutil.ConvertBlob(blobBuff)
	return hex.EncodeToString(blob)
}

func (s *StratumServer) fetchBlockTemplate() bool {
	r := s.rpc()
	reply, err := r.GetBlockTemplate(8, s.config.Address)
	if err != nil {
		log.Printf("Error while refreshing block template: %s", err)
		return false
	}
	t := s.currentBlockTemplate()

	if t != nil && t.prevHash == reply.PrevHash {
		// Fallback to height comparison
		if len(reply.PrevHash) == 0 && reply.Height > t.height {
			log.Printf("New block to mine on %s at height %v, diff: %v", r.Name, reply.Height, reply.Difficulty)
		} else {
			return false
		}
	} else {
		log.Printf("New block to mine on %s at height %v, diff: %v, prev_hash: %s", r.Name, reply.Height, reply.Difficulty, reply.PrevHash)
	}
	newTemplate := BlockTemplate{
		diffInt64:      reply.Difficulty,
		difficulty:     big.NewInt(reply.Difficulty),
		height:         reply.Height,
		prevHash:       reply.PrevHash,
		seedHash:       reply.SeedHash,
		reservedOffset: reply.ReservedOffset,
	}
	newTemplate.buffer, _ = hex.DecodeString(reply.Blob)
	s.blockTemplate.Store(&newTemplate)

	// randomx new seed
	if !s.config.BypassShareValidation {
		if s.rx.IsNewSeed(newTemplate.seedHash) {
			s.rx.RandomxReady = false
			s.rx.Init(newTemplate.seedHash)
		}
	}

	return true
}

package utxo

import (
	"math/big"
	"testing"
	"time"

	"crypto-tracker/internal/model"
)

func TestAnalyzeUTXOOutputs_DustAndChangeHeuristics(t *testing.T) {
	sender := "bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq"
	recipient := "34xp4vRoCGJym3xR7yCVPFHoCNxv4Twseo"      // P2SH (different script type, round amount)
	changeAddr := "bc1qxy2kgdygjrsqtzq2n0yrf2493p83kkfjhx0wlh" // P2WPKH (matches sender script type, odd amount)
	dustAddr := "1DustAttackSpamAddress111111111111"

	tx := model.Transaction{
		TxID:      "a1b2c3d4e5f6",
		Chain:     model.ChainBitcoin,
		Timestamp: time.Now(),
		Inputs: []model.TxInput{
			{
				Address:    sender,
				ValueRaw:   big.NewInt(150_000_000), // 1.5 BTC
				ValueFloat: 1.5,
				ScriptType: model.ScriptP2WPKH,
			},
		},
		Outputs: []model.TxOutput{
			{
				Index:      0,
				Address:    recipient,
				ValueRaw:   big.NewInt(100_000_000), // 1.0 BTC (round payment, P2SH)
				ValueFloat: 1.0,
				ScriptType: model.ScriptP2SH,
			},
			{
				Index:      1,
				Address:    changeAddr,
				ValueRaw:   big.NewInt(49_985_420), // 0.49985420 BTC (change, P2WPKH)
				ValueFloat: 0.49985420,
				ScriptType: model.ScriptP2WPKH,
			},
			{
				Index:      2,
				Address:    dustAddr,
				ValueRaw:   big.NewInt(546), // 546 sats dust
				ValueFloat: 0.00000546,
				ScriptType: model.ScriptP2PKH,
			},
		},
	}

	targets := AnalyzeUTXOOutputs(tx, sender, 1000)
	if len(targets) != 3 {
		t.Fatalf("expected 3 evaluated targets, got %d", len(targets))
	}

	// Target 0: Recipient should NOT be skipped
	if targets[0].IsSkipped || targets[0].ToAddress != recipient {
		t.Errorf("expected recipient %s to be followed, got skipped=%v", recipient, targets[0].IsSkipped)
	}

	// Target 1: Change address SHOULD be skipped with ReasonChangeSkipped
	if !targets[1].IsSkipped || targets[1].Reason != model.ReasonChangeSkipped {
		t.Errorf("expected change address %s to be skipped with ReasonChangeSkipped, got skipped=%v reason=%s",
			changeAddr, targets[1].IsSkipped, targets[1].Reason)
	}

	// Target 2: Dust output SHOULD be skipped with ReasonDustFiltered
	if !targets[2].IsSkipped || targets[2].Reason != model.ReasonDustFiltered {
		t.Errorf("expected dust address %s to be filtered with ReasonDustFiltered, got skipped=%v reason=%s",
			dustAddr, targets[2].IsSkipped, targets[2].Reason)
	}
}


package main

import (
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"github.com/etherria/bitcoin-tx-builder/bitcoin"
	"testing"
)

func TestCommitTxMessageHash(t *testing.T) {
	inscriptionDataList := make([]bitcoin.InscriptionData, 0)
	inscriptionDataList = append(inscriptionDataList, bitcoin.InscriptionData{
		ContentType: "text/plain;charset=utf-8",
		Body:        []byte(`{"p":"brc-20","op":"mint","tick":"xcvb","amt":"100"}`),
		RevealAddr:  "tb1pklh8lqax5l7m2ycypptv2emc4gata2dy28svnwcp9u32wlkenvsspcvhsr",
	})
	commitTxPrevOutputList := make([]*bitcoin.PrevOutput, 0)
	commitTxPrevOutputList = append(commitTxPrevOutputList, &bitcoin.PrevOutput{
		TxId:    "7298e2304d962fd7975ba1b8f4a7cb1af7c6bccd064d03ea43d2d6e8d0d49c8b",
		VOut:    0,
		Amount:  5460,
		Address: "2Mzo8Xo359ieu3AZ3LpyPRVRUtsZVcffUoy",
	})
	params := &BuildBrc20CommitTxRequest{
		CommitTxPrevOutputList: commitTxPrevOutputList,
		CommitFeeRate:          2,
		RevealFeeRate:          2,
		RevealOutValue:         546,
		InscriptionDataList:    inscriptionDataList,
		ChangeAddress:          "tb1pklh8lqax5l7m2ycypptv2emc4gata2dy28svnwcp9u32wlkenvsspcvhsr",
		PubKey:                 "024f0ce4a3f60d68468807b31e97e70f4880d382f29c7bcd1516048d71e6538459",
	}
	netParams := &chaincfg.TestNet3Params

	commitTxPrivateKeyListWif := []string{
		"cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
		"cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
		"cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
		"cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
	}
	serializedPubKey, err := hex.DecodeString(params.PubKey)
	if err != nil {
		t.Error(err)
	}
	parseResult, unsignedCommitTxHex, commitTxFee, err := bitcoin.BuildBrc20CommitTx(netParams, params.InscriptionDataList, params.CommitTxPrevOutputList, params.RevealOutValue, params.MinChangeValue, params.CommitFeeRate, params.RevealFeeRate, params.ChangeAddress, serializedPubKey, commitTxPrivateKeyListWif)
	if err != nil {
		t.Error(err)
	}
	var tx *wire.MsgTx
	if tx, err = bitcoin.NewTxFromHex(unsignedCommitTxHex); err != nil {
		t.Error(err)
	}
	tool := &bitcoin.InscriptionBuilder{
		Network: netParams,
	}
	commitTxPrevOutputFetcher, _, _, err := tool.ParseCommitTxPrevOutput(params.CommitTxPrevOutputList)
	pk, err := btcec.ParsePubKey(serializedPubKey)
	if err != nil {
		t.Error(err)
	}
	pubKeyBytes := pk.SerializeCompressed()
	messageHashMap, err := bitcoin.GetMessageHash(tx, pubKeyBytes, commitTxPrevOutputFetcher)
	if err != nil {
		t.Error(err)
	}
	res := &BuildBrc20CommitTxResponse{
		ParseResult:    parseResult,
		TxHex:          unsignedCommitTxHex,
		MessageHashMap: messageHashMap,
		CommitTxFee:    commitTxFee,
	}

	t.Log(res)
}

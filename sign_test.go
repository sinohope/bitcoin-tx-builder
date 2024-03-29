package main

import (
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/etherria/bitcoin-tx-builder/bitcoin"
	"testing"
)

func TestCommitTx(t *testing.T) {
	inscriptionDataList := make([]bitcoin.InscriptionData, 0)
	inscriptionDataList = append(inscriptionDataList, bitcoin.InscriptionData{
		ContentType: "text/plain;charset=utf-8",
		Body:        []byte(`{"p":"brc-20","op":"transfer","tick":"mpct","amt":"20"}`),
		RevealAddr:  "mwTnMZzAvUzVCQKGkdcAHYW1sMnJCumbJL",
	})
	commitTxPrevOutputList := make([]*bitcoin.PrevOutput, 0)
	commitTxPrevOutputList = append(commitTxPrevOutputList, &bitcoin.PrevOutput{
		TxId:    "9e42c75fbf79b3e72f3c637992a559c8d36e3df7a5d1694c34d9c73cdc8d4e31",
		VOut:    0,
		Amount:  22000,
		Address: "mwTnMZzAvUzVCQKGkdcAHYW1sMnJCumbJL",
	})
	params := &BuildBrc20CommitTxRequest{
		CommitTxPrevOutputList: commitTxPrevOutputList,
		CommitFeeRate:          2,
		RevealFeeRate:          2,
		RevealOutValue:         546,
		InscriptionDataList:    inscriptionDataList,
		ChangeAddress:          "mwTnMZzAvUzVCQKGkdcAHYW1sMnJCumbJL",
		PubKey:                 "03791de14f1d886f995f89df9bf4eab6f30a3c804d33d5ea6a729c5c22939ee92b",
		MinChangeValue:         805,
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

	signatureMap := make(map[int]string)
	signatureMap[0] = "2992c0a99f056414b4b25cfb6802b8480ba476a199cd913346a8aa01824f463201fad3b74ca15e148349a25a7634ff75869ff1cc9bbc7c30050cae0bc93aa79401"
	params2 := &BuildCommitTxRawDataRequest{
		CommitTxPrevOutputList: commitTxPrevOutputList,
		TxHex:                  unsignedCommitTxHex,
		SignatureMap:           signatureMap,
		PubKey:                 "03791de14f1d886f995f89df9bf4eab6f30a3c804d33d5ea6a729c5c22939ee92b",
	}

	txHex, err := bitcoin.BuildCommitTxRawData(netParams, params2.TxHex, params2.CommitTxPrevOutputList, params2.SignatureMap, params2.PubKey)
	t.Log(txHex)

	params3 := &BuildBrc20RevealTxRequest{
		CommitTxHash: "9a424b5e2deecd1b65317afd848126e20a5369faf4f35516b5160dd94ae97a77",
	}
	commitTxHash, err := chainhash.NewHashFromStr(params3.CommitTxHash)
	if err != nil {
		t.Error(err)
	}

	var witnessList [][]byte
	revealTxsHex, witnessList, revealTxFees, err := bitcoin.BuildBrc20RevealTx(netParams, *commitTxHash, parseResult.CtxDataList, []string{inscriptionDataList[0].RevealAddr}, 2, parseResult.RevealOutValue)
	if err != nil {
		t.Error(err)
	}
	res3 := &BuildBrc20RevealTxResponse{
		RevealTxsHex: revealTxsHex[0],
		WitnessList:  witnessList[0],
		RevealTxFees: revealTxFees[0],
		messageHash:  hex.EncodeToString(witnessList[0]), //review交易只有一个input，所以只对应一个messageHash
	}
	t.Log(res3, commitTxHash.String())
	revealSignature := "3ceed8a8ea0b543c029a54359c8471143a338a488993e347e5a451e936312aab488b8c6756f1b56d6d11fbd40ea99ccd8075cdedce6caf6b50abaa61e0a01fd0"
	//8950280836c6f069faffd41a8ab992c2e775533554ff8221aeb9e6ab0753647f48d80ce3b2a1a9903124dc111a31d275eec0f31f386d0edc574da4ce8d7a7e8f01
	//cadadf851a56318bac89b2b8429a0f10f3890a892592cb723ab786a50ec571da60b2fdccf96809cdb14b20c1fe29561a3375b7a5a6f6fe974ea594de2bf394d300
	//4ba9f34361895962eb76ad92ac16189482e66725bf2a2e53cbe6768823f78fc6e5a71aefdafe96885b588710e45a21538c953b1c18c9b8857f86d530220f60fb
	//3ceed8a8ea0b543c029a54359c8471143a338a488993e347e5a451e936312aab488b8c6756f1b56d6d11fbd40ea99ccd8075cdedce6caf6b50abaa61e0a01fd0
	signedRevealTxsHex, err := bitcoin.SignBrc20RevealTx2(netParams, revealTxsHex, revealSignature, parseResult.CtxDataList)

	t.Log(signedRevealTxsHex[0])
}

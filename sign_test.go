package main

import (
	"encoding/hex"
	"fmt"
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
		Body:        []byte(`{"p":"brc-20","op":"transfer","tick":"mpct","amt":"10"}`),
		RevealAddr:  "mzkW8wgqUfgc6qPd2wypDkotRUjzL4VECh",
	})
	commitTxPrevOutputList := make([]*bitcoin.PrevOutput, 0)
	commitTxPrevOutputList = append(commitTxPrevOutputList, &bitcoin.PrevOutput{
		TxId:    "11cb908c1f427ea843a9d92110a168d8ea817e0e3371a0bc0c68e37cb07c92da",
		VOut:    1,
		Amount:  9994975,
		Address: "mzkW8wgqUfgc6qPd2wypDkotRUjzL4VECh",
	})
	params := &BuildBrc20CommitTxRequest{
		CommitTxPrevOutputList: commitTxPrevOutputList,
		CommitFeeRate:          3,
		RevealFeeRate:          3,
		RevealOutValue:         546,
		InscriptionDataList:    inscriptionDataList,
		ChangeAddress:          "mzkW8wgqUfgc6qPd2wypDkotRUjzL4VECh",
		PubKey:                 "0277752ea4bfa8898f9ec542e6fd4afad58b30ad1ca45e3d0c8a074a3a82999879",
		MinChangeValue:         1322,
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
		PubKey:                 "0277752ea4bfa8898f9ec542e6fd4afad58b30ad1ca45e3d0c8a074a3a82999879",
	}

	txHex, err := bitcoin.BuildRawData(netParams, params2.TxHex, params2.CommitTxPrevOutputList, params2.SignatureMap, params2.PubKey)
	t.Log(txHex)

	params3 := &BuildBrc20RevealTxRequest{
		CommitTxHash: "6b0446ed1e599b648c3365b65f7f0a7f65821bf0ed57fb1071ba21226db67625",
	}
	commitTxHash, err := chainhash.NewHashFromStr(params3.CommitTxHash)
	if err != nil {
		t.Error(err)
	}

	var witnessList [][]byte
	revealTxsHex, witnessList, revealTxFees, err := bitcoin.BuildBrc20RevealTx(netParams, *commitTxHash, parseResult.CtxDataList, []string{inscriptionDataList[0].RevealAddr}, 3, parseResult.RevealOutValue)
	fmt.Println(*commitTxHash, parseResult.CtxDataList[0], []string{inscriptionDataList[0].RevealAddr}, 3, parseResult.RevealOutValue)
	t.Log(*commitTxHash, parseResult.CtxDataList)
	if err != nil {
		t.Error(err)
	}
	res3 := &BuildBrc20RevealTxResponse{
		RevealTxsHex: revealTxsHex[0],
		WitnessList:  witnessList[0],
		RevealTxFees: revealTxFees[0],
		MessageHash:  hex.EncodeToString(witnessList[0]), //review交易只有一个input，所以只对应一个messageHash
	}
	t.Log(res3, commitTxHash.String())
	revealSignature := "3af27ffb82be26620e8dd01d9ecf434c012006f84aa6fd24c1da7e232435bb8b61508bb08c6da086e5e1d837b5091c684da52be1320cfc1b977835ffa837590d"
	//fb56542be551e5f00d060af75eaa18f36cb40418bc8ae11cfcda046841413548dd2a29df92ffb2f609fafbab2a1c678ba8d9a229dcb6abbf040a17d40b700047
	//8950280836c6f069faffd41a8ab992c2e775533554ff8221aeb9e6ab0753647f48d80ce3b2a1a9903124dc111a31d275eec0f31f386d0edc574da4ce8d7a7e8f01
	//cadadf851a56318bac89b2b8429a0f10f3890a892592cb723ab786a50ec571da60b2fdccf96809cdb14b20c1fe29561a3375b7a5a6f6fe974ea594de2bf394d300
	//4ba9f34361895962eb76ad92ac16189482e66725bf2a2e53cbe6768823f78fc6e5a71aefdafe96885b588710e45a21538c953b1c18c9b8857f86d530220f60fb
	//3ceed8a8ea0b543c029a54359c8471143a338a488993e347e5a451e936312aab488b8c6756f1b56d6d11fbd40ea99ccd8075cdedce6caf6b50abaa61e0a01fd0
	signedRevealTxsHex, err := bitcoin.SignBrc20RevealTx2(netParams, revealTxsHex, revealSignature, parseResult.CtxDataList)

	t.Log(signedRevealTxsHex[0])
}

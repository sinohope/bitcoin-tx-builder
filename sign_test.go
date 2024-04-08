package main

import (
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
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
		TxId:    "5da5b05e77ab6d69f19134e1acac4c9b9bc5f547efae7013011ce132922a776d",
		VOut:    1,
		Amount:  9989953,
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
	signatureMap[0] = "2b32c38cf73a7ce1f313501bcc2315589d6cec1824613a99eb1d02c7fdafcc8127a0e3faf1b04e56dc1d39213be834d26223faa22065f0168b1d29fea00904d901"
	params2 := &BuildCommitTxRawDataRequest{
		CommitTxPrevOutputList: commitTxPrevOutputList,
		TxHex:                  unsignedCommitTxHex,
		SignatureMap:           signatureMap,
		PubKey:                 "0277752ea4bfa8898f9ec542e6fd4afad58b30ad1ca45e3d0c8a074a3a82999879",
	}

	txHex, err := bitcoin.BuildRawData(netParams, params2.TxHex, params2.CommitTxPrevOutputList, params2.SignatureMap, params2.PubKey)
	t.Log(txHex)

	params3 := &BuildBrc20RevealTxRequest{
		CommitTxHash: "1131c0300dc1ad44a0f67fe4372a16fc5bc2c1a979f61c75325872be3cfa5d79",
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
	revealSignature := "620d9c4bb034791045a49f52b4a5d08fd4ded147a2e0d324d4c0c8f3f4ccd5a65a7a3b4d639e15bb375a09eeea0617c3324b1aa519f9e6e20ba7f1c9fd0fe5d6"
	//fb56542be551e5f00d060af75eaa18f36cb40418bc8ae11cfcda046841413548dd2a29df92ffb2f609fafbab2a1c678ba8d9a229dcb6abbf040a17d40b700047
	//8950280836c6f069faffd41a8ab992c2e775533554ff8221aeb9e6ab0753647f48d80ce3b2a1a9903124dc111a31d275eec0f31f386d0edc574da4ce8d7a7e8f01
	//620d9c4bb034791045a49f52b4a5d08fd4ded147a2e0d324d4c0c8f3f4ccd5a65a7a3b4d639e15bb375a09eeea0617c3324b1aa519f9e6e20ba7f1c9fd0fe5d6
	//cadadf851a56318bac89b2b8429a0f10f3890a892592cb723ab786a50ec571da60b2fdccf96809cdb14b20c1fe29561a3375b7a5a6f6fe974ea594de2bf394d300
	//4ba9f34361895962eb76ad92ac16189482e66725bf2a2e53cbe6768823f78fc6e5a71aefdafe96885b588710e45a21538c953b1c18c9b8857f86d530220f60fb
	//3ceed8a8ea0b543c029a54359c8471143a338a488993e347e5a451e936312aab488b8c6756f1b56d6d11fbd40ea99ccd8075cdedce6caf6b50abaa61e0a01fd0
	signedRevealTxsHex, err := bitcoin.SignBrc20RevealTx2(netParams, revealTxsHex, revealSignature, parseResult.CtxDataList)
	fmt.Println(revealTxsHex, revealSignature, parseResult.CtxDataList[0])
	t.Log(signedRevealTxsHex[0])
}

func TestPubKey(t *testing.T) {
	privateKeyWif, err := btcutil.DecodeWIF("5KWKSRnmzxCjUP1NKR4dNyyHhaZWSGRTbGzBnm1vwgwpoe2AVGQ")
	if err != nil {
		t.Log(err)
	}
	privateKey := privateKeyWif.PrivKey
	pubKey := privateKey.PubKey()
	fmt.Println(hex.EncodeToString(pubKey.SerializeCompressed()))
}

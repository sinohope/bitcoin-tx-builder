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
	"github.com/etherria/bitcoin-tx-builder/bitcoin/txscript"
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
func TestSignature(t *testing.T) {
	signature, err := txscript.BuildSignature("e6f76ca4a01c0455685ddd506c83a64d6ce10cca687ae8f70aa202fb2c404624185c225fc65a24b4fccdc92b19ff3c07599830f0c2171882829e91fc2ac5dc7301")
	if err != nil {
	}
	fmt.Println(hex.EncodeToString(signature.Serialize()))
}

func TestTxSize(t *testing.T) {
	//rawData := "0100000002ea1a2716d431ce9e18885160a6510def5f2cd90850d8efd4f265317a2cb00d72000000006a47304402202600cb55e74b3de69ac3e5c00ca74ffea532bca17e3a1e60baa563440023a33402205256ac12b5939dcd4efeab065c754a00512e5b8ad7f21ab0d832e3db6892a1ac012102cedbc02644ff7e2b6d7bda7fcaaecca42a2ad5196b2c7efa051c42a408da5fcefdffffffaae4a6a1178dd7aa64650a8d9b39a6a696d8972ebecb3ea1d86597912f815016010000006b483045022100ec4a3ac7c31cf7faa36e5c36ad016cc8ea7f0b5751e27dbb4501d705b329b763022075d5cf556d6771db94aaaccecc7e2c5e83e263c33ba223bb2fd278f072976a5f012102cedbc02644ff7e2b6d7bda7fcaaecca42a2ad5196b2c7efa051c42a408da5fcefdffffff0222020000000000002251208541c446f5d4623427173b1cd88b5868fd46e4516f9f9ab02eadc6de455d23eab7153c00000000001976a914f69223e5d65be4aa498c40839c3ccc133883b4da88ac00000000"
	rawData := "0100000002ea1a2716d431ce9e18885160a6510def5f2cd90850d8efd4f265317a2cb00d72000000006b4830450221008ebb2e4aa32837ad45e7d706e81f464de0ab1a7ab298f48d3c69256100ea41c602201bf9fdc2d772ff9c4e1d96549c14ae2170c513cad4d011e348f4f1e9371a50cb01210357bbb2d4a9cb8a2357633f201b9c518c2795ded682b7913c6beef3fe23bd6d2ffdffffffaae4a6a1178dd7aa64650a8d9b39a6a696d8972ebecb3ea1d86597912f815016010000006a4730440220158d5d6900a15461e101d178c1cb628d7d7fd733cb1b6aadce667b5e82f3313302204b333d9f5473dd87a3b8b038c3ee1a6f19239b599c76b6df4a75561a43ba850201210357bbb2d4a9cb8a2357633f201b9c518c2795ded682b7913c6beef3fe23bd6d2ffdffffff0222020000000000002251208541c446f5d4623427173b1cd88b5868fd46e4516f9f9ab02eadc6de455d23eaa1030000000000002251208541c446f5d4623427173b1cd88b5868fd46e4516f9f9ab02eadc6de455d23ea00000000"
	tx, _ := bitcoin.NewTxFromHex(rawData)
	size := bitcoin.GetTxVirtualSize(btcutil.NewTx(tx))
	t.Log(size)

}

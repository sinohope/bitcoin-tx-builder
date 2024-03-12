package bitcoin

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/assert"
)

func TestInscribe(t *testing.T) {
	network := &chaincfg.TestNet3Params

	commitTxPrevOutputList := make([]*PrevOutput, 0)
	commitTxPrevOutputList = append(commitTxPrevOutputList, &PrevOutput{
		TxId:       "453aa6dd39f31f06cd50b72a8683b8c0402ab36f889d96696317503a025a21b5",
		VOut:       0,
		Amount:     546,
		Address:    "2NF33rckfiQTiE5Guk5ufUdwms8PgmtnEdc",
		PrivateKey: "cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
	})
	commitTxPrevOutputList = append(commitTxPrevOutputList, &PrevOutput{
		TxId:       "22c8a4869f2aa9ee5994959c0978106130290cda53f6e933a8dda2dcb82508d4",
		VOut:       0,
		Amount:     546,
		Address:    "tb1qtsq9c4fje6qsmheql8gajwtrrdrs38kdzeersc",
		PrivateKey: "cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
	})
	commitTxPrevOutputList = append(commitTxPrevOutputList, &PrevOutput{
		TxId:       "3c6f205ec2995696d5bc852709d234a63aad82131b5b7615504e2e3e9ff88987",
		VOut:       0,
		Amount:     546,
		Address:    "mouQtmBWDS7JnT65Grj2tPzdSmGKJgRMhE",
		PrivateKey: "cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
	})
	commitTxPrevOutputList = append(commitTxPrevOutputList, &PrevOutput{
		TxId:       "aa09fa48dda0e2b7de1843c3db8d3f2d7f2cbe0f83331a125b06516a348abd26",
		VOut:       4,
		Amount:     1142196,
		Address:    "tb1pklh8lqax5l7m2ycypptv2emc4gata2dy28svnwcp9u32wlkenvsspcvhsr",
		PrivateKey: "cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
	})

	inscriptionDataList := make([]InscriptionData, 0)
	inscriptionDataList = append(inscriptionDataList, InscriptionData{
		ContentType: "text/plain;charset=utf-8",
		Body:        []byte(`{"p":"brc-20","op":"mint","tick":"xcvb","amt":"100"}`),
		RevealAddr:  "tb1pklh8lqax5l7m2ycypptv2emc4gata2dy28svnwcp9u32wlkenvsspcvhsr",
	})
	inscriptionDataList = append(inscriptionDataList, InscriptionData{
		ContentType: "text/plain;charset=utf-8",
		Body:        []byte(`{"p":"brc-20","op":"mint","tick":"xcvb","amt":"10"}`),
		RevealAddr:  "mouQtmBWDS7JnT65Grj2tPzdSmGKJgRMhE",
	})
	inscriptionDataList = append(inscriptionDataList, InscriptionData{
		ContentType: "text/plain;charset=utf-8",
		Body:        []byte(`{"p":"brc-20","op":"mint","tick":"xcvb","amt":"10000"}`),
		RevealAddr:  "tb1qtsq9c4fje6qsmheql8gajwtrrdrs38kdzeersc",
	})
	inscriptionDataList = append(inscriptionDataList, InscriptionData{
		ContentType: "text/plain;charset=utf-8",
		Body:        []byte(`{"p":"brc-20","op":"mint","tick":"xcvb","amt":"1"}`),
		RevealAddr:  "2NF33rckfiQTiE5Guk5ufUdwms8PgmtnEdc",
	})

	request := &InscriptionRequest{
		CommitTxPrevOutputList: commitTxPrevOutputList,
		CommitFeeRate:          2,
		RevealFeeRate:          2,
		RevealOutValue:         546,
		InscriptionDataList:    inscriptionDataList,
		ChangeAddress:          "tb1pklh8lqax5l7m2ycypptv2emc4gata2dy28svnwcp9u32wlkenvsspcvhsr",
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	log.Println(string(requestBytes))

	txs, err := Inscribe(network, request)
	if err != nil {
		t.Fatal(err)
	}
	txsBytes, err := json.MarshalIndent(txs, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(txsBytes))
}

func TestInscribe2(t *testing.T) {

	network := &chaincfg.TestNet3Params

	inscriptionDataList := make([]InscriptionData, 0)
	inscriptionDataList = append(inscriptionDataList, InscriptionData{
		ContentType: "text/plain;charset=utf-8",
		Body:        []byte(`{"p":"brc-20","op":"mint","tick":"xcvb","amt":"100"}`),
		RevealAddr:  "tb1pklh8lqax5l7m2ycypptv2emc4gata2dy28svnwcp9u32wlkenvsspcvhsr",
	})
	inscriptionDataList = append(inscriptionDataList, InscriptionData{
		ContentType: "text/plain;charset=utf-8",
		Body:        []byte(`{"p":"brc-20","op":"mint","tick":"xcvb","amt":"10"}`),
		RevealAddr:  "mouQtmBWDS7JnT65Grj2tPzdSmGKJgRMhE",
	})
	inscriptionDataList = append(inscriptionDataList, InscriptionData{
		ContentType: "text/plain;charset=utf-8",
		Body:        []byte(`{"p":"brc-20","op":"mint","tick":"xcvb","amt":"10000"}`),
		RevealAddr:  "tb1qtsq9c4fje6qsmheql8gajwtrrdrs38kdzeersc",
	})
	inscriptionDataList = append(inscriptionDataList, InscriptionData{
		ContentType: "text/plain;charset=utf-8",
		Body:        []byte(`{"p":"brc-20","op":"mint","tick":"xcvb","amt":"1"}`),
		RevealAddr:  "2NF33rckfiQTiE5Guk5ufUdwms8PgmtnEdc",
	})

	firstPrevOutPrivateKey := "cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22"
	privateKeyWif, err := btcutil.DecodeWIF(firstPrevOutPrivateKey)
	assert.Nil(t, err)
	pubKey := privateKeyWif.PrivKey.PubKey().SerializeUncompressed()

	commitFeeRate := int64(2)
	revealFeeRate := int64(2)
	revealOutValue := int64(546)
	changeAddress := "tb1pklh8lqax5l7m2ycypptv2emc4gata2dy28svnwcp9u32wlkenvsspcvhsr"
	minChangeValue := int64(0)

	var parseResult *Brc20InscriptionParseResult

	var commitTx *wire.MsgTx
	var commitTxHash chainhash.Hash
	var commitTxFee int64

	// build commitTx
	{
		commitTxPrevOutputList := make([]*PrevOutput, 0)
		commitTxPrevOutputList = append(commitTxPrevOutputList, &PrevOutput{
			TxId:    "453aa6dd39f31f06cd50b72a8683b8c0402ab36f889d96696317503a025a21b5",
			VOut:    0,
			Amount:  546,
			Address: "2NF33rckfiQTiE5Guk5ufUdwms8PgmtnEdc",
		})
		commitTxPrevOutputList = append(commitTxPrevOutputList, &PrevOutput{
			TxId:    "22c8a4869f2aa9ee5994959c0978106130290cda53f6e933a8dda2dcb82508d4",
			VOut:    0,
			Amount:  546,
			Address: "tb1qtsq9c4fje6qsmheql8gajwtrrdrs38kdzeersc",
		})
		commitTxPrevOutputList = append(commitTxPrevOutputList, &PrevOutput{
			TxId:    "3c6f205ec2995696d5bc852709d234a63aad82131b5b7615504e2e3e9ff88987",
			VOut:    0,
			Amount:  546,
			Address: "mouQtmBWDS7JnT65Grj2tPzdSmGKJgRMhE",
		})
		commitTxPrevOutputList = append(commitTxPrevOutputList, &PrevOutput{
			TxId:    "aa09fa48dda0e2b7de1843c3db8d3f2d7f2cbe0f83331a125b06516a348abd26",
			VOut:    4,
			Amount:  1142196,
			Address: "tb1pklh8lqax5l7m2ycypptv2emc4gata2dy28svnwcp9u32wlkenvsspcvhsr",
		})

		commitTxPrivateKeyListWif := []string{
			"cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
			"cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
			"cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
			"cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
		}

		var txPreparedHex string
		totalSenderAmount := btcutil.Amount(0)

		// 1.prepare commit tx
		parseResult, txPreparedHex, totalSenderAmount, err = PrepareBrc20CommitTx(network, inscriptionDataList, commitTxPrevOutputList, revealOutValue, minChangeValue, revealFeeRate, changeAddress, pubKey)
		assert.Nil(t, err)

		// 2.pre sign commit tx
		var txForEstimateHex string
		txForEstimateHex, err = SignBrc20CommitTx(network, txPreparedHex, commitTxPrevOutputList, commitTxPrivateKeyListWif)
		assert.Nil(t, err)

		// 3. check tx virtualsize and adjust output
		var txCheckedHex string
		txCheckedHex, commitTxFee, err = AdjustBrc20CommitTx(network, txForEstimateHex, commitTxPrevOutputList, totalSenderAmount, parseResult.TotalRevealPrevOutputValue, commitFeeRate, parseResult.MinChangeValue)
		assert.Nil(t, err)

		// 4. sign
		var commitTxHex string
		commitTxHex, err = SignBrc20CommitTx(network, txCheckedHex, commitTxPrevOutputList, commitTxPrivateKeyListWif)
		assert.Nil(t, err)

		commitTx, err = NewTxFromHex(commitTxHex)
		assert.Nil(t, err)
		commitTxHash = commitTx.TxHash()
	}

	var unsignedRevealTxsHex []string
	var signedRevealTxsHex []string
	var revealTxs []*wire.MsgTx
	var revealTxFees []int64
	// build reveal tx
	{
		revealAddrs := make([]string, len(inscriptionDataList))
		for i := 0; i < len(inscriptionDataList); i++ {
			revealAddrs[i] = inscriptionDataList[i].RevealAddr
		}

		// build
		var witnessList [][]byte
		unsignedRevealTxsHex, witnessList, revealTxFees, err = BuildBrc20RevealTx(network, commitTxHash, parseResult.CtxDataList, revealAddrs, revealFeeRate, parseResult.RevealOutValue)
		assert.Nil(t, err)

		// sign
		signedRevealTxsHex, err = SignBrc20RevealTx(network, unsignedRevealTxsHex, witnessList, parseResult.CtxDataList, firstPrevOutPrivateKey)
		assert.Nil(t, err)

		// check
		err = CheckBrc20RevealTx(signedRevealTxsHex)
		assert.Nil(t, err)

		revealTxs = make([]*wire.MsgTx, len(signedRevealTxsHex))
		for i, txHex := range signedRevealTxsHex {
			tx, err := NewTxFromHex(txHex)
			assert.Nil(t, err)
			revealTxs[i] = tx
		}
	}

	tool := &InscriptionBuilder{
		Network:  network,
		CommitTx: commitTx,
		RevealTx: revealTxs,
	}

	var commitTxStr string
	commitTxStr, err = tool.GetCommitTxHex()
	assert.Nil(t, err)

	var revealTxsStr []string
	revealTxsStr, err = tool.GetRevealTxHexList()
	assert.Nil(t, err)

	txs := &InscribeTxs{
		CommitTx:     commitTxStr,
		RevealTxs:    revealTxsStr,
		CommitTxFee:  commitTxFee,
		RevealTxFees: revealTxFees,
		CommitAddrs:  parseResult.CommitAddrs,
	}

	txsBytes, err := json.MarshalIndent(txs, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(txsBytes))
}

func TestInscribe3(t *testing.T) {

	network := &chaincfg.TestNet3Params

	inscriptionDataList := make([]InscriptionData, 0)
	inscriptionDataList = append(inscriptionDataList, InscriptionData{
		ContentType: "text/plain;charset=utf-8",
		Body:        []byte(`{"p":"brc-20","op":"mint","tick":"xcvb","amt":"100"}`),
		RevealAddr:  "tb1pklh8lqax5l7m2ycypptv2emc4gata2dy28svnwcp9u32wlkenvsspcvhsr",
	})
	inscriptionDataList = append(inscriptionDataList, InscriptionData{
		ContentType: "text/plain;charset=utf-8",
		Body:        []byte(`{"p":"brc-20","op":"mint","tick":"xcvb","amt":"10"}`),
		RevealAddr:  "mouQtmBWDS7JnT65Grj2tPzdSmGKJgRMhE",
	})
	inscriptionDataList = append(inscriptionDataList, InscriptionData{
		ContentType: "text/plain;charset=utf-8",
		Body:        []byte(`{"p":"brc-20","op":"mint","tick":"xcvb","amt":"10000"}`),
		RevealAddr:  "tb1qtsq9c4fje6qsmheql8gajwtrrdrs38kdzeersc",
	})
	inscriptionDataList = append(inscriptionDataList, InscriptionData{
		ContentType: "text/plain;charset=utf-8",
		Body:        []byte(`{"p":"brc-20","op":"mint","tick":"xcvb","amt":"1"}`),
		RevealAddr:  "2NF33rckfiQTiE5Guk5ufUdwms8PgmtnEdc",
	})

	firstPrevOutPrivateKey := "cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22"
	privateKeyWif, err := btcutil.DecodeWIF(firstPrevOutPrivateKey)
	assert.Nil(t, err)
	pubKey := privateKeyWif.PrivKey.PubKey().SerializeUncompressed()

	commitFeeRate := int64(2)
	revealFeeRate := int64(2)
	revealOutValue := int64(546)
	changeAddress := "tb1pklh8lqax5l7m2ycypptv2emc4gata2dy28svnwcp9u32wlkenvsspcvhsr"
	minChangeValue := int64(0)

	var parseResult *Brc20InscriptionParseResult

	var commitTx *wire.MsgTx
	var commitTxHash chainhash.Hash
	var commitTxFee int64

	// build commitTx
	{
		commitTxPrevOutputList := make([]*PrevOutput, 0)
		commitTxPrevOutputList = append(commitTxPrevOutputList, &PrevOutput{
			TxId:    "453aa6dd39f31f06cd50b72a8683b8c0402ab36f889d96696317503a025a21b5",
			VOut:    0,
			Amount:  546,
			Address: "2NF33rckfiQTiE5Guk5ufUdwms8PgmtnEdc",
		})
		commitTxPrevOutputList = append(commitTxPrevOutputList, &PrevOutput{
			TxId:    "22c8a4869f2aa9ee5994959c0978106130290cda53f6e933a8dda2dcb82508d4",
			VOut:    0,
			Amount:  546,
			Address: "tb1qtsq9c4fje6qsmheql8gajwtrrdrs38kdzeersc",
		})
		commitTxPrevOutputList = append(commitTxPrevOutputList, &PrevOutput{
			TxId:    "3c6f205ec2995696d5bc852709d234a63aad82131b5b7615504e2e3e9ff88987",
			VOut:    0,
			Amount:  546,
			Address: "mouQtmBWDS7JnT65Grj2tPzdSmGKJgRMhE",
		})
		commitTxPrevOutputList = append(commitTxPrevOutputList, &PrevOutput{
			TxId:    "aa09fa48dda0e2b7de1843c3db8d3f2d7f2cbe0f83331a125b06516a348abd26",
			VOut:    4,
			Amount:  1142196,
			Address: "tb1pklh8lqax5l7m2ycypptv2emc4gata2dy28svnwcp9u32wlkenvsspcvhsr",
		})

		commitTxPrivateKeyListWif := []string{
			"cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
			"cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
			"cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
			"cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
		}

		var unsignedCommitTxHex string

		// 1.prepare commit tx
		parseResult, unsignedCommitTxHex, commitTxFee, err = BuildBrc20CommitTx(network, inscriptionDataList, commitTxPrevOutputList, revealOutValue, minChangeValue, commitFeeRate, revealFeeRate, changeAddress, pubKey, commitTxPrivateKeyListWif)
		assert.Nil(t, err)

		// 2. sign
		var commitTxHex string
		commitTxHex, err = SignBrc20CommitTx(network, unsignedCommitTxHex, commitTxPrevOutputList, commitTxPrivateKeyListWif)
		assert.Nil(t, err)

		commitTx, err = NewTxFromHex(commitTxHex)
		assert.Nil(t, err)
		commitTxHash = commitTx.TxHash()
	}

	var unsignedRevealTxsHex []string
	var signedRevealTxsHex []string
	var revealTxs []*wire.MsgTx
	var revealTxFees []int64
	// build reveal tx
	{
		revealAddrs := make([]string, len(inscriptionDataList))
		for i := 0; i < len(inscriptionDataList); i++ {
			revealAddrs[i] = inscriptionDataList[i].RevealAddr
		}

		// build
		var witnessList [][]byte
		unsignedRevealTxsHex, witnessList, revealTxFees, err = BuildBrc20RevealTx(network, commitTxHash, parseResult.CtxDataList, revealAddrs, revealFeeRate, parseResult.RevealOutValue)
		assert.Nil(t, err)

		// sign
		signedRevealTxsHex, err = SignBrc20RevealTx(network, unsignedRevealTxsHex, witnessList, parseResult.CtxDataList, firstPrevOutPrivateKey)
		assert.Nil(t, err)

		// check
		err = CheckBrc20RevealTx(signedRevealTxsHex)
		assert.Nil(t, err)

		revealTxs = make([]*wire.MsgTx, len(signedRevealTxsHex))
		for i, txHex := range signedRevealTxsHex {
			tx, err := NewTxFromHex(txHex)
			assert.Nil(t, err)
			revealTxs[i] = tx
		}
	}

	tool := &InscriptionBuilder{
		Network:  network,
		CommitTx: commitTx,
		RevealTx: revealTxs,
	}

	var commitTxStr string
	commitTxStr, err = tool.GetCommitTxHex()
	assert.Nil(t, err)

	var revealTxsStr []string
	revealTxsStr, err = tool.GetRevealTxHexList()
	assert.Nil(t, err)

	txs := &InscribeTxs{
		CommitTx:     commitTxStr,
		RevealTxs:    revealTxsStr,
		CommitTxFee:  commitTxFee,
		RevealTxFees: revealTxFees,
		CommitAddrs:  parseResult.CommitAddrs,
	}

	txsBytes, err := json.MarshalIndent(txs, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(txsBytes))
}

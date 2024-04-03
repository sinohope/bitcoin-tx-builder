package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/etherria/bitcoin-tx-builder/bitcoin"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"net/http"
)

type ResultData struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func successRes(ctx echo.Context, data interface{}) error {
	d, _ := json.Marshal(data)
	log.Infof("response:%s", string(d))
	return ctx.JSON(http.StatusOK, &ResultData{Code: 200, Data: data})
}
func errorRes(ctx echo.Context, msg string) error {
	log.Error(msg)
	return ctx.JSON(http.StatusInternalServerError, &ResultData{Code: http.StatusInternalServerError, Msg: msg})
}

func buildBrc20CommitTx(ctx echo.Context) error {
	network := ctx.Param("network")
	netParams := getNetwork(network)
	params := &BuildBrc20CommitTxRequest{}
	err := ctx.Bind(params)
	if err != nil {
		return errorRes(ctx, err.Error())
	}
	d, _ := json.Marshal(params)
	log.Infof("buildBrc20CommitTx request:%s", string(d))
	commitTxPrivateKeyListWif := []string{
		"cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
		"cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
		"cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
		"cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
	}
	serializedPubKey, err := hex.DecodeString(params.PubKey)
	if err != nil {
		return errorRes(ctx, err.Error())
	}
	parseResult, unsignedCommitTxHex, commitTxFee, err := bitcoin.BuildBrc20CommitTx(netParams, params.InscriptionDataList, params.CommitTxPrevOutputList, params.RevealOutValue, params.MinChangeValue, params.CommitFeeRate, params.RevealFeeRate, params.ChangeAddress, serializedPubKey, commitTxPrivateKeyListWif)

	if err != nil {
		return errorRes(ctx, err.Error())
	}
	var tx *wire.MsgTx
	if tx, err = bitcoin.NewTxFromHex(unsignedCommitTxHex); err != nil {
		return errorRes(ctx, err.Error())
	}
	tool := &bitcoin.InscriptionBuilder{
		Network: netParams,
	}
	commitTxPrevOutputFetcher, _, _, err := tool.ParseCommitTxPrevOutput(params.CommitTxPrevOutputList)
	pk, err := btcec.ParsePubKey(serializedPubKey)
	if err != nil {
		return errorRes(ctx, err.Error())
	}
	pubKeyBytes := pk.SerializeCompressed()
	messageHashMap, err := bitcoin.GetMessageHash(tx, pubKeyBytes, commitTxPrevOutputFetcher)
	if err != nil {
		return errorRes(ctx, err.Error())
	}

	return successRes(ctx, &BuildBrc20CommitTxResponse{
		ParseResult:    parseResult,
		TxHex:          unsignedCommitTxHex,
		MessageHashMap: messageHashMap,
		CommitTxFee:    commitTxFee,
	})
}

func buildCommitTxRawData(ctx echo.Context) error {
	network := ctx.Param("network")
	netParams := getNetwork(network)
	params := &BuildCommitTxRawDataRequest{}
	err := ctx.Bind(params)
	if err != nil {
		return errorRes(ctx, err.Error())
	}
	d, _ := json.Marshal(params)
	log.Infof("buildCommitTxRawData request:%s", string(d))
	txHex, err := bitcoin.BuildRawData(netParams, params.TxHex, params.CommitTxPrevOutputList, params.SignatureMap, params.PubKey)
	if err != nil {
		return errorRes(ctx, err.Error())
	}
	return successRes(ctx, &SignBrc20CommitTxResponse{
		TxHex: txHex,
	})
}

func buildBrc20RevealTx(ctx echo.Context) error {
	network := ctx.Param("network")
	netParams := getNetwork(network)
	params := &BuildBrc20RevealTxRequest{}
	err := ctx.Bind(params)
	if err != nil {
		return errorRes(ctx, err.Error())
	}
	d, _ := json.Marshal(params)
	log.Infof("buildBrc20RevealTx request:%s", string(d))

	var witnessList [][]byte
	commitTxHash, err := chainhash.NewHashFromStr(params.CommitTxHash)
	if err != nil {
		return errorRes(ctx, err.Error())
	}
	revealTxsHex, witnessList, revealTxFees, err := bitcoin.BuildBrc20RevealTx(netParams, *commitTxHash, params.CtxDataList, params.RevealAddrs, params.RevealFeeRate, params.RevealOutValue)
	if err != nil {
		return errorRes(ctx, err.Error())
	}
	fmt.Println(*commitTxHash, params.CtxDataList[0], params.RevealAddrs, params.RevealFeeRate, params.RevealOutValue)
	return successRes(ctx, &BuildBrc20RevealTxResponse{
		RevealTxsHex: revealTxsHex[0],
		WitnessList:  witnessList[0],
		RevealTxFees: revealTxFees[0],
		MessageHash:  hex.EncodeToString(witnessList[0]), //review交易只有一个input，所以只对应一个messageHash
	})
}

func buildReviewTxRawData(ctx echo.Context) error {
	network := ctx.Param("network")
	netParams := getNetwork(network)
	params := &BuildRevealTxRawDataRequest{}
	err := ctx.Bind(params)
	if err != nil {
		return errorRes(ctx, err.Error())
	}
	d, _ := json.Marshal(params)
	log.Infof("buildReviewTxRawData request:%s", string(d))
	signedRevealTxsHex, err := bitcoin.SignBrc20RevealTx2(netParams, params.RevealTxsHex, params.Signature, params.CtxDataList)
	fmt.Println("SignBrc20RevealTx2 params", params.RevealTxsHex, params.Signature, params.CtxDataList[0])
	if err != nil {
		return errorRes(ctx, err.Error())
	}

	return successRes(ctx, &BuildRevealTxRawDataResponse{
		RevealTxHex: signedRevealTxsHex[0],
	})
}

func buildNormalTx(ctx echo.Context) error {
	network := ctx.Param("network")
	netParams := getNetwork(network)
	params := &BuildUnsignedTxRequest{}
	err := ctx.Bind(params)
	if err != nil {
		return errorRes(ctx, err.Error())
	}
	d, _ := json.Marshal(params)
	log.Infof("buildNormalTx request:%s", string(d))

	txBuild := bitcoin.NewTxBuild(params.Version, netParams)
	for i := 0; i < len(params.Inputs); i++ {
		txBuild.AddInput2(params.Inputs[i].TxId, params.Inputs[i].VOut, "", params.Inputs[i].Address, params.Inputs[i].Amount)
	}

	for i := 0; i < len(params.Outputs); i++ {
		txBuild.AddOutput(params.Outputs[i].Address, params.Outputs[i].Amount)
	}

	tx, _, err := txBuild.Build(false)

	if err != nil {
		return errorRes(ctx, err.Error())
	}

	txHex, err := bitcoin.GetTxHex(tx)
	if err != nil {
		return errorRes(ctx, err.Error())
	}

	serializedPubKey, err := hex.DecodeString(params.PubKey)
	pk, err := btcec.ParsePubKey(serializedPubKey)
	if err != nil {
		return errorRes(ctx, err.Error())
	}
	pubKeyBytes := pk.SerializeCompressed()
	tool := &bitcoin.InscriptionBuilder{
		Network: netParams,
	}
	prevOutputFetcher, _, _, err := tool.ParseCommitTxPrevOutput(params.Inputs)
	messageHashMap, err := bitcoin.GetMessageHash(tx, pubKeyBytes, prevOutputFetcher)
	if err != nil {
		return errorRes(ctx, err.Error())
	}

	return successRes(ctx, &BuildUnsignedTxResponse{
		UnsignedTx:  txHex,
		MessageHash: messageHashMap,
	})
}

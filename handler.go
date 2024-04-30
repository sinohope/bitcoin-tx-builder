package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
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
func errorResByCode(ctx echo.Context, msg string, code int) error {
	log.Error(msg)
	return ctx.JSON(http.StatusOK, &ResultData{Code: code, Msg: msg})
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
	var commitTxPrivateKeyListWif = make([]string, len(params.CommitTxPrevOutputList))
	for i, _ := range params.CommitTxPrevOutputList {
		commitTxPrivateKeyListWif[i] = "cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22"
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
	if err != nil {
		return errorRes(ctx, err.Error())
	}
	pubKeyBytes, err := hex.DecodeString(params.PubKey)

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
	params.Version = 1
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

	pubKeyBytes, err := hex.DecodeString(params.PubKey)
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

func buildNormalTx2(ctx echo.Context) error {
	network := ctx.Param("network")
	netParams := getNetwork(network)
	params := &BuildUnsignedTxRequest{}
	err := ctx.Bind(params)
	if err != nil {
		return errorRes(ctx, err.Error())
	}
	d, _ := json.Marshal(params)
	log.Infof("buildNormalTx request:%s", string(d))
	params.Version = 1
	txBuild := bitcoin.NewTxBuild(params.Version, netParams)
	inputAmount := int64(0)
	for i := 0; i < len(params.Inputs); i++ {
		inputAmount += params.Inputs[i].Amount
		txBuild.AddInput2(params.Inputs[i].TxId, params.Inputs[i].VOut, "", params.Inputs[i].Address, params.Inputs[i].Amount)
	}
	outputAmount := int64(0)
	for i := 0; i < len(params.Outputs); i++ {
		outputAmount += params.Outputs[i].Amount
		txBuild.AddOutput(params.Outputs[i].Address, params.Outputs[i].Amount)
	}
	//先假设有找零，构造找零output
	txBuild.AddOutput(params.Inputs[0].Address, 0)
	tx, _, err := txBuild.Build(false)
	if err != nil {
		return errorRes(ctx, err.Error())
	}
	tool := &bitcoin.InscriptionBuilder{
		Network: netParams,
	}
	prevOutputFetcher, _, _, err := tool.ParseCommitTxPrevOutput(params.Inputs)
	//假签名，计算手续费
	if err = bitcoin.Sign(tx, getPriKeys(params.Inputs), prevOutputFetcher); err != nil {
		return errorRes(ctx, err.Error())
	}
	var changeAmount int64
	minChangeValue := int64(546)
	if tx, changeAmount, err = CompleteTx(tx, btcutil.Amount(inputAmount), outputAmount, params.FeeRate, minChangeValue); err != nil {
		maxVoutAmount := outputAmount + changeAmount
		if maxVoutAmount < 0 {
			maxVoutAmount = 0
		}
		return errorResByCode(ctx, fmt.Sprintf("Limit the capacity of transactions with btc and fee, your current maximum amount of coins is %d, please modify the transfer amount and try again", maxVoutAmount), 1001) //insufficient balance
	}
	if changeAmount >= minChangeValue {
		outputAmount += changeAmount
		params.Outputs = append(params.Outputs, RawOutput{params.Inputs[0].Address, changeAmount})
	}
	fee := inputAmount - outputAmount
	txHex, err := bitcoin.GetTxHex(tx)
	if err != nil {
		return errorRes(ctx, err.Error())
	}
	pubKeyBytes, err := hex.DecodeString(params.PubKey)

	messageHashMap, err := bitcoin.GetMessageHash(tx, pubKeyBytes, prevOutputFetcher)
	if err != nil {
		return errorRes(ctx, err.Error())
	}
	return successRes(ctx, &BuildUnsignedTxResponse{
		Fee:         fee,
		UnsignedTx:  txHex,
		MessageHash: messageHashMap,
		Outputs:     params.Outputs,
		Inputs:      params.Inputs,
	})
}

func CompleteTx(tx *wire.MsgTx, totalSenderAmount btcutil.Amount, outputAmount, commitFeeRate int64, minChangeValue int64) (*wire.MsgTx, int64, error) {
	size := btcutil.Amount(bitcoin.GetTxVirtualSize(btcutil.NewTx(tx)))
	log.Infof("tx size: %d", size)
	fee := size * btcutil.Amount(commitFeeRate)
	changeAmount := totalSenderAmount - btcutil.Amount(outputAmount) - fee
	if int64(changeAmount) >= minChangeValue {
		tx.TxOut[len(tx.TxOut)-1].Value = int64(changeAmount)
	} else {
		tx.TxOut = tx.TxOut[:len(tx.TxOut)-1]
		if changeAmount < 0 {
			//计算无找零fee
			sizeWithoutChange := btcutil.Amount(bitcoin.GetTxVirtualSize(btcutil.NewTx(tx)))
			feeWithoutChange := sizeWithoutChange * btcutil.Amount(commitFeeRate)
			log.Infof("tx sizeWithoutChange: %d", sizeWithoutChange)
			if totalSenderAmount-btcutil.Amount(outputAmount)-feeWithoutChange < 0 {
				changeAmount = totalSenderAmount - btcutil.Amount(outputAmount) - feeWithoutChange //此时的changeAmount是负值，用于计算最大可转金额
				return nil, int64(changeAmount), errors.New("insufficient balance")
			}
		}
	}
	return tx, int64(changeAmount), nil
}

func getPriKeys(inputs []*bitcoin.PrevOutput) []*btcec.PrivateKey {
	var commitTxPrivateKeyListWif = make([]string, len(inputs))
	for i, _ := range inputs {
		commitTxPrivateKeyListWif[i] = "cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22"
	}
	var commitTxPrivateKeyList []*btcec.PrivateKey
	for _, prvkey := range commitTxPrivateKeyListWif {
		privateKeyWif, _ := btcutil.DecodeWIF(prvkey)
		commitTxPrivateKeyList = append(commitTxPrivateKeyList, privateKeyWif.PrivKey)
	}
	return commitTxPrivateKeyList
}
func pubKey2Addr(ctx echo.Context) error {
	network := ctx.Param("network")
	netParams := getNetwork(network)
	params := &PubKey2AddrRequest{}
	err := ctx.Bind(params)
	if err != nil {
		return errorRes(ctx, err.Error())
	}
	publicKey, err := hex.DecodeString(params.PubKey)
	if err != nil {
		return errorRes(ctx, err.Error())
	}
	addr, err := bitcoin.PubKeyToAddr(publicKey, params.AddrType, netParams)
	if err != nil {
		return errorRes(ctx, err.Error())
	}
	return successRes(ctx, &PubKey2AddrResponse{
		Addr: addr,
	})
}

func health(ctx echo.Context) error {
	return successRes(ctx, "ok")
}

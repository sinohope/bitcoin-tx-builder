package main

import (
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/wire"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"net/http"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/etherria/bitcoin-tx-builder/bitcoin"
	"github.com/labstack/echo/v4"
)

type RequestData struct {
	ID     uint64      `json:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

type ResponseData struct {
	ID     uint64      `json:"id"`
	Error  string      `json:"error"`
	Result interface{} `json:"result"`
}

type PubKey2AddrRequest struct {
	PubKey   string `json:"pubKey"`
	AddrType string `json:"addrType"`
}

type PubKey2AddrResponse struct {
	Addr string `json:"addr"`
}

type BuildUnsignedTxRequest struct {
	Version int32       `json:"version"`
	Inputs  []RawInput  `json:"inputs"`
	Outputs []RawOutput `json:"outputs"`
}

type RawInput struct {
	TxId string `json:"txId"`
	VOut uint32 `json:"vout"`
	//RedeemScript string `json:"redeemScript"`
	Address string `json:"address"`
	Amount  int64  `json:"amount"`
}

type RawOutput struct {
	Address string `json:"address"`
	Amount  int64  `json:"amount"`
}

type BuildUnsignedTxResponse struct {
	UnsignedTx string `json:"unsignedTx"`
}

type PrepareBrc20CommitTxRequest struct {
	CommitTxPrevOutputList []*bitcoin.PrevOutput     `json:"commitTxPrevOutputList"`
	RevealFeeRate          int64                     `json:"revealFeeRate"`
	InscriptionDataList    []bitcoin.InscriptionData `json:"inscriptionDataList"`
	RevealOutValue         int64                     `json:"revealOutValue"`
	ChangeAddress          string                    `json:"changeAddress"`
	MinChangeValue         int64                     `json:"minChangeValue"`
	PubKey                 []byte                    `json:"pubKey"`
}

type PrepareBrc20CommitTxResponse struct {
	ParseResult       *bitcoin.Brc20InscriptionParseResult `json:"parseResult"`
	TxHex             string                               `json:"txHex"`
	TotalSenderAmount int64                                `json:"totalSenderAmount"`
}

type SignBrc20CommitTxRequest struct {
	TxHex                     string                `json:"txHex"`
	CommitTxPrevOutputList    []*bitcoin.PrevOutput `json:"commitTxPrevOutputList"`
	CommitTxPrivateKeyListWif []string              `json:"commitTxPrivateKeyListWif"`
}

type SignBrc20CommitTxResponse struct {
	TxHex string `json:"txHex"`
}

type AdjustBrc20CommitTxRequest struct {
	TxHex                      string                `json:"txHex"`
	CommitTxPrevOutputList     []*bitcoin.PrevOutput `json:"commitTxPrevOutputList"`
	TotalSenderAmount          int64                 `json:"totalSenderAmount"`
	TotalRevealPrevOutputValue int64                 `json:"totalRevealPrevOutputValue"`
	CommitFeeRate              int64                 `json:"commitFeeRate"`
	MinChangeValue             int64                 `json:"minChangeValue"`
}

type AdjustBrc20CommitTxResponse struct {
	TxHex       string `json:"txHex"`
	CommitTxFee int64  `json:"commitTxFee"`
}

type BuildBrc20CommitTxRequest struct {
	CommitTxPrevOutputList []*bitcoin.PrevOutput     `json:"commitTxPrevOutputList"`
	CommitFeeRate          int64                     `json:"commitFeeRate"`
	RevealFeeRate          int64                     `json:"revealFeeRate"`
	InscriptionDataList    []bitcoin.InscriptionData `json:"inscriptionDataList"`
	RevealOutValue         int64                     `json:"revealOutValue"`
	ChangeAddress          string                    `json:"changeAddress"`
	MinChangeValue         int64                     `json:"minChangeValue"`
	PubKey                 string                    `json:"pubKey"`
}

type BuildBrc20CommitTxResponse struct {
	ParseResult    *bitcoin.Brc20InscriptionParseResult `json:"parseResult"`
	TxHex          string                               `json:"txHex"`
	CommitTxFee    int64                                `json:"commitTxFee"`
	MessageHashMap map[int]string                       `json:"messageHashMap"`
}

type BuildBrc20RevealTxRequest struct {
	CommitTxHash   chainhash.Hash          `json:"commitTxHash"`
	CtxDataList    []*bitcoin.Brc20CtxData `json:"ctxDataList"`
	RevealAddrs    []string                `json:"revealAddrs"`
	RevealFeeRate  int64                   `json:"revealFeeRate"`
	RevealOutValue int64                   `json:"revealOutValue"`
}

type BuildBrc20RevealTxResponse struct {
	RevealTxsHex []string `json:"revealTxsHex"`
	WitnessList  [][]byte `json:"witnessList"`
	RevealTxFees []int64  `json:"revealTxFees"`
}

type SignBrc20RevealTxRequest struct {
	RevealTxsHex  []string                `json:"revealTxsHex"`
	WitnessList   [][]byte                `json:"witnessList"`
	CtxDataList   []*bitcoin.Brc20CtxData `json:"CtxDataList"`
	PrivateKeyWif string                  `json:"privateKeyWif"`
}

type SignBrc20RevealTxResponse struct {
	RevealTxsHex []string `json:"revealTxsHex"`
}

type BuildCommitTxRawDataRequest struct {
	CommitTxPrevOutputList []*bitcoin.PrevOutput `json:"commitTxPrevOutputList"`
	TxHex                  string                `json:"txHex"`
	SignatureMap           map[int]string        `json:"signatureMap"`
	PubKey                 string                `json:"pubKey"`
}

type BuildCommitTxRawDataResponse struct {
	RawData string `json:"rawData"`
}

type CheckBrc20RevealTxRequest struct {
	RevealTxsHex []string `json:"revealTxsHex"`
}

type CheckBrc20RevealTxResponse struct {
}

func main() {
	e := echo.New()

	e.POST("/:network", func(ctx echo.Context) error {

		network := ctx.Param("network")
		var netParams *chaincfg.Params

		if network == "mainnet" {
			netParams = &chaincfg.MainNetParams
		} else if network == "regtest" {
			netParams = &chaincfg.RegressionNetParams
		} else if network == "testnet3" {
			netParams = &chaincfg.TestNet3Params
		} else if network == "simnet" {
			netParams = &chaincfg.SimNetParams
		} else {
			return ctx.String(http.StatusNotFound, network)
		}

		req := new(RequestData)
		if err := ctx.Bind(req); err != nil {
			return err
		}

		rsp := &ResponseData{
			ID:     req.ID,
			Error:  "",
			Result: nil,
		}

		if req.Method == "pubKey2Addr" {

			params := req.Params.(*PubKey2AddrRequest)

			publicKey, err := hex.DecodeString(params.PubKey)
			if err != nil {
				rsp.Error = err.Error()
				return ctx.JSON(http.StatusOK, rsp)
			}

			addr, err := bitcoin.PubKeyToAddr(publicKey, params.AddrType, netParams)
			if err != nil {
				rsp.Error = err.Error()
				return ctx.JSON(http.StatusOK, rsp)
			}

			rsp.Result = &PubKey2AddrResponse{
				Addr: addr,
			}

			return ctx.JSON(http.StatusOK, rsp)

		} else if req.Method == "buildUnsignedTx" {
			params := req.Params.(*BuildUnsignedTxRequest)

			txBuild := bitcoin.NewTxBuild(params.Version, netParams)
			for i := 0; i < len(params.Inputs); i++ {
				txBuild.AddInput2(params.Inputs[i].TxId, params.Inputs[i].VOut, "", params.Inputs[i].Address, params.Inputs[i].Amount)
			}

			for i := 0; i < len(params.Outputs); i++ {
				txBuild.AddOutput(params.Outputs[i].Address, params.Outputs[i].Amount)
			}

			tx, _, err := txBuild.Build(false)

			if err != nil {
				rsp.Error = err.Error()
				return ctx.JSON(http.StatusOK, rsp)
			}

			txHex, err := bitcoin.GetTxHex(tx)
			if err != nil {
				rsp.Error = err.Error()
				return ctx.JSON(http.StatusOK, rsp)
			}

			rsp.Result = &BuildUnsignedTxResponse{
				UnsignedTx: txHex,
			}

			return ctx.JSON(http.StatusOK, rsp)

		} else if req.Method == "prepareBrc20CommitTx" {
			params := req.Params.(*PrepareBrc20CommitTxRequest)

			parseResult, txPreparedHex, totalSenderAmount, err := bitcoin.PrepareBrc20CommitTx(netParams, params.InscriptionDataList, params.CommitTxPrevOutputList, params.RevealOutValue, params.MinChangeValue, params.RevealFeeRate, params.ChangeAddress, params.PubKey)

			if err != nil {
				rsp.Error = err.Error()
				return ctx.JSON(http.StatusOK, rsp)
			}

			rsp.Result = &PrepareBrc20CommitTxResponse{
				ParseResult:       parseResult,
				TxHex:             txPreparedHex,
				TotalSenderAmount: int64(totalSenderAmount),
			}

			return ctx.JSON(http.StatusOK, rsp)

		} else if req.Method == "signBrc20CommitTx" {

			params := req.Params.(*SignBrc20CommitTxRequest)

			txForEstimateHex, err := bitcoin.SignBrc20CommitTx(netParams, params.TxHex, params.CommitTxPrevOutputList, params.CommitTxPrivateKeyListWif)

			if err != nil {
				rsp.Error = err.Error()
				return ctx.JSON(http.StatusOK, rsp)
			}

			rsp.Result = &SignBrc20CommitTxResponse{
				TxHex: txForEstimateHex,
			}

			return ctx.JSON(http.StatusOK, rsp)

		} else if req.Method == "adjustBrc20CommitTx" {

			params := req.Params.(*AdjustBrc20CommitTxRequest)

			totalSenderAmount := btcutil.Amount(params.TotalSenderAmount)

			txCheckedHex, commitTxFee, err := bitcoin.AdjustBrc20CommitTx(netParams, params.TxHex, params.CommitTxPrevOutputList, totalSenderAmount, params.TotalRevealPrevOutputValue, params.CommitFeeRate, params.MinChangeValue)

			if err != nil {
				rsp.Error = err.Error()
				return ctx.JSON(http.StatusOK, rsp)
			}

			rsp.Result = &AdjustBrc20CommitTxResponse{
				TxHex:       txCheckedHex,
				CommitTxFee: commitTxFee,
			}

			return ctx.JSON(http.StatusOK, rsp)
		} else if req.Method == "buildBrc20CommitTx" {
			params := req.Params.(*BuildBrc20CommitTxRequest)

			commitTxPrivateKeyListWif := []string{
				"cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
				"cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
				"cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
				"cPnvkvUYyHcSSS26iD1dkrJdV7k1RoUqJLhn3CYxpo398PdLVE22",
			}
			serializedPubKey, err := hexutil.Decode(params.PubKey)
			if err != nil {
				return err
			}
			parseResult, unsignedCommitTxHex, commitTxFee, err := bitcoin.BuildBrc20CommitTx(netParams, params.InscriptionDataList, params.CommitTxPrevOutputList, params.RevealOutValue, params.MinChangeValue, params.CommitFeeRate, params.RevealFeeRate, params.ChangeAddress, serializedPubKey, commitTxPrivateKeyListWif)

			if err != nil {
				rsp.Error = err.Error()
				return ctx.JSON(http.StatusOK, rsp)
			}
			var tx *wire.MsgTx
			if tx, err = bitcoin.NewTxFromHex(unsignedCommitTxHex); err != nil {
				rsp.Error = err.Error()
				return ctx.JSON(http.StatusOK, rsp)
			}
			tool := &bitcoin.InscriptionBuilder{
				Network: netParams,
			}
			commitTxPrevOutputFetcher, _, _, err := tool.ParseCommitTxPrevOutput(params.CommitTxPrevOutputList)
			pk, err := btcec.ParsePubKey(serializedPubKey)
			if err != nil {
				return err
			}
			pubKeyBytes := pk.SerializeCompressed()
			messageHashMap, err := bitcoin.GetMessageHash(tx, pubKeyBytes, commitTxPrevOutputFetcher)
			if err != nil {
				return err
			}
			rsp.Result = &BuildBrc20CommitTxResponse{
				ParseResult:    parseResult,
				TxHex:          unsignedCommitTxHex,
				MessageHashMap: messageHashMap,
				CommitTxFee:    commitTxFee,
			}

			return ctx.JSON(http.StatusOK, rsp)

		} else if req.Method == "buildBrc20RevealTx" {
			params := req.Params.(*BuildBrc20RevealTxRequest)

			var witnessList [][]byte
			revealTxsHex, witnessList, revealTxFees, err := bitcoin.BuildBrc20RevealTx(netParams, params.CommitTxHash, params.CtxDataList, params.RevealAddrs, params.RevealFeeRate, params.RevealOutValue)
			if err != nil {
				rsp.Error = err.Error()
				return ctx.JSON(http.StatusOK, rsp)
			}

			rsp.Result = &BuildBrc20RevealTxResponse{
				RevealTxsHex: revealTxsHex,
				WitnessList:  witnessList,
				RevealTxFees: revealTxFees,
			}

			return ctx.JSON(http.StatusOK, rsp)
		} else if req.Method == "signBrc20RevealTx" {
			//params := req.Params.(*SignBrc20RevealTxRequest)
			//
			//signedRevealTxsHex, err := bitcoin.SignBrc20RevealTx(netParams, params.RevealTxsHex, params.CtxDataList, params.PrivateKeyWif)
			//
			//if err != nil {
			//	rsp.Error = err.Error()
			//	return ctx.JSON(http.StatusOK, rsp)
			//}
			//
			//rsp.Result = &BuildBrc20RevealTxResponse{
			//	RevealTxsHex: signedRevealTxsHex,
			//}
			//
			//return ctx.JSON(http.StatusOK, rsp)
		} else if req.Method == "buildCommitTxRawData" {
			params := req.Params.(*BuildCommitTxRawDataRequest)

			txHex, err := bitcoin.BuildCommitTxRawData(netParams, params.TxHex, params.CommitTxPrevOutputList, params.SignatureMap, params.PubKey)

			if err != nil {
				rsp.Error = err.Error()
				return ctx.JSON(http.StatusOK, rsp)
			}

			rsp.Result = &SignBrc20CommitTxResponse{
				TxHex: txHex,
			}

			return ctx.JSON(http.StatusOK, rsp)
		} else if req.Method == "checkBrc20RevealTx" {
			params := req.Params.(*CheckBrc20RevealTxRequest)

			err := bitcoin.CheckBrc20RevealTx(params.RevealTxsHex)

			if err != nil {
				rsp.Error = err.Error()
				return ctx.JSON(http.StatusOK, rsp)
			}

			rsp.Result = &CheckBrc20RevealTxResponse{}

			return ctx.JSON(http.StatusOK, rsp)
		}

		return ctx.String(http.StatusNotFound, req.Method)
	})
	/*
		e.GET("/build_raw_tx", func(c echo.Context) error {

			txBuild := bitcoin.NewRawTxBuild(1, 0, "testnet3")
			txBuild.AddInput2("c44a7f98434e5e875a573339f77d36022c79c525771fa88c72fa53f3a55eeaf7", 1, "mouQtmBWDS7JnT65Grj2tPzdSmGKJgRMhE", 1488430)
			txBuild.AddOutput("mouQtmBWDS7JnT65Grj2tPzdSmGKJgRMhE", 1488200)
			//data, _ = json.Marshal(txBuild)

			return c.String(http.StatusOK, "Hello, World!")
		})

		e.GET("/brc20_commit", func(c echo.Context) error {

			txBuild := bitcoin.NewRawTxBuild(1, 0, "testnet3")
			txBuild.AddInput2("c44a7f98434e5e875a573339f77d36022c79c525771fa88c72fa53f3a55eeaf7", 1, "mouQtmBWDS7JnT65Grj2tPzdSmGKJgRMhE", 1488430)
			txBuild.AddOutput("mouQtmBWDS7JnT65Grj2tPzdSmGKJgRMhE", 1488200)
			//data, _ = json.Marshal(txBuild)

			return c.String(http.StatusOK, "Hello, World!")
		})

		e.GET("/brc20_reveal", func(c echo.Context) error {

			txBuild := bitcoin.NewRawTxBuild(1, 0, "testnet3")
			txBuild.AddInput2("c44a7f98434e5e875a573339f77d36022c79c525771fa88c72fa53f3a55eeaf7", 1, "mouQtmBWDS7JnT65Grj2tPzdSmGKJgRMhE", 1488430)
			txBuild.AddOutput("mouQtmBWDS7JnT65Grj2tPzdSmGKJgRMhE", 1488200)
			//data, _ = json.Marshal(txBuild)

			return c.String(http.StatusOK, "Hello, World!")
		})
	*/
	s := http.Server{
		Addr:    ":8081",
		Handler: e,
		//ReadTimeout: 30 * time.Second, // customize http.Server timeouts
	}
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		//log.Fatal(err)
		e.Logger.Fatal(err)
	}
}

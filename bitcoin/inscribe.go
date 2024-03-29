package bitcoin

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/etherria/bitcoin-tx-builder/bitcoin/txscript"
)

type InscriptionData struct {
	ContentType string `json:"contentType"`
	Body        []byte `json:"body"`
	RevealAddr  string `json:"revealAddr"`
}

type PrevOutput struct {
	TxId       string `json:"txId"`
	VOut       uint32 `json:"vOut"`
	Amount     int64  `json:"amount"`
	Address    string `json:"address"`
	PrivateKey string `json:"privateKey"`
}

type InscriptionRequest struct {
	CommitTxPrevOutputList []*PrevOutput     `json:"commitTxPrevOutputList"`
	CommitFeeRate          int64             `json:"commitFeeRate"`
	RevealFeeRate          int64             `json:"revealFeeRate"`
	InscriptionDataList    []InscriptionData `json:"inscriptionDataList"`
	RevealOutValue         int64             `json:"revealOutValue"`
	ChangeAddress          string            `json:"changeAddress"`
	MinChangeValue         int64             `json:"minChangeValue"`
}

type InscriptionTxCtxData struct {
	PrivateKey              *btcec.PrivateKey
	InscriptionScript       []byte
	CommitTxAddress         string
	CommitTxAddressPkScript []byte
	ControlBlockWitness     []byte
	RevealTxPrevOutput      *wire.TxOut
}

type InscriptionBuilder struct {
	Network                   *chaincfg.Params
	CommitTxPrevOutputFetcher *txscript.MultiPrevOutFetcher
	CommitTxPrivateKeyList    []*btcec.PrivateKey
	RevealTxPrevOutputFetcher *txscript.MultiPrevOutFetcher
	CommitTxPrevOutputList    []*PrevOutput
	RevealTx                  []*wire.MsgTx
	CommitTx                  *wire.MsgTx
	MustCommitTxFee           int64
	MustRevealTxFees          []int64
	CommitAddrs               []string
	CommitTxFee               int64
	RevealTxFees              []int64
}

type InscribeTxs struct {
	CommitTx     string   `json:"commitTx"`
	RevealTxs    []string `json:"revealTxs"`
	CommitTxFee  int64    `json:"commitTxFee"`
	RevealTxFees []int64  `json:"revealTxFees"`
	CommitAddrs  []string `json:"commitAddrs"`
}

type Brc20InscriptionParseResult struct {
	CtxDataList                []*Brc20CtxData `json:"ctxDataList"`
	RevealOutValue             int64           `json:"revealOutValue"`
	TotalRevealPrevOutputValue int64           `json:"totalRevealPrevOutputValue"`
	MinChangeValue             int64           `json:"minChangeValue"`
	CommitAddrs                []string        `json:"commitAddrs"`
}

type Brc20CtxData struct {
	CommitTxAddress     string `json:"commitTxAddress"`
	CommitTxOutPkScript []byte `json:"commitTxOutPkScript"`
	CommitTxOutValue    int64  `json:"commitTxOutValue"`

	InscriptionScript   []byte `json:"inscriptionScript"`
	ControlBlockWitness []byte `json:"controlBlockWitness"`

	RevealTxOutPkScript []byte `json:"revealTxOutPkScript"`
	RevealTxOutValue    int64  `json:"revealTxOutValue"`
}

const (
	DefaultTxVersion      = 2
	DefaultSequenceNum    = 0xfffffffd
	DefaultRevealOutValue = int64(546)
	DefaultMinChangeValue = int64(546)

	MaxStandardTxWeight = 4000000 / 10
	WitnessScaleFactor  = 4
)

func NewInscriptionTool(network *chaincfg.Params, request *InscriptionRequest) (*InscriptionBuilder, error) {
	var commitTxPrivateKeyList []*btcec.PrivateKey
	for _, prevOutput := range request.CommitTxPrevOutputList {
		privateKeyWif, err := btcutil.DecodeWIF(prevOutput.PrivateKey)
		if err != nil {
			return nil, err
		}
		commitTxPrivateKeyList = append(commitTxPrivateKeyList, privateKeyWif.PrivKey)
	}
	tool := &InscriptionBuilder{
		Network:                   network,
		CommitTxPrevOutputFetcher: txscript.NewMultiPrevOutFetcher(nil),
		CommitTxPrivateKeyList:    commitTxPrivateKeyList,
		RevealTxPrevOutputFetcher: txscript.NewMultiPrevOutFetcher(nil),
		CommitTxPrevOutputList:    request.CommitTxPrevOutputList,
	}
	return tool, tool.initTool(network, request)
}

func (builder *InscriptionBuilder) initTool(network *chaincfg.Params, request *InscriptionRequest) error {
	destinations := make([]string, len(request.InscriptionDataList))
	revealOutValue := DefaultRevealOutValue
	if request.RevealOutValue > 0 {
		revealOutValue = request.RevealOutValue
	}
	minChangeValue := DefaultMinChangeValue
	if request.MinChangeValue > 0 {
		minChangeValue = request.MinChangeValue
	}

	privateKeyWif, err := btcutil.DecodeWIF(request.CommitTxPrevOutputList[0].PrivateKey)
	if err != nil {
		return err
	}
	privateKey := privateKeyWif.PrivKey
	pubKey := privateKey.PubKey()

	inscriptionTxCtxDataList := make([]*InscriptionTxCtxData, len(request.InscriptionDataList))
	for i := 0; i < len(request.InscriptionDataList); i++ {

		inscriptionTxCtxData, err := newInscriptionTxCtxData(network, &request.InscriptionDataList[i], privateKey, pubKey)
		if err != nil {
			return err
		}
		inscriptionTxCtxDataList[i] = inscriptionTxCtxData
		destinations[i] = request.InscriptionDataList[i].RevealAddr
	}
	revealTxs, totalRevealPrevOutputValue, err := builder.BuildEmptyRevealTx(destinations, inscriptionTxCtxDataList, revealOutValue, request.RevealFeeRate)
	if err != nil {
		return err
	}

	var txStage1 *wire.MsgTx
	var commitTxPrevOutputFetcher *txscript.MultiPrevOutFetcher
	totalSenderAmount := btcutil.Amount(0)

	if commitTxPrevOutputFetcher, txStage1, totalSenderAmount, err = builder.ParseCommitTxPrevOutput(request.CommitTxPrevOutputList); err != nil {
		return err
	}

	if err = builder.FillCommitTxOutput(txStage1, inscriptionTxCtxDataList, request.ChangeAddress); err != nil {
		return err
	}

	txForEstimate := wire.NewMsgTx(txStage1.Version)
	txForEstimate.LockTime = txStage1.LockTime
	txForEstimate.TxIn = txStage1.TxIn
	txForEstimate.TxOut = txStage1.TxOut
	if err = Sign(txForEstimate, builder.CommitTxPrivateKeyList, commitTxPrevOutputFetcher); err != nil {
		return err
	}

	var tx *wire.MsgTx
	if tx, err = builder.CompleteCommitTx(txForEstimate, totalSenderAmount, totalRevealPrevOutputValue, request.CommitFeeRate, minChangeValue); err != nil {
		return err
	}

	if err = Sign(tx, builder.CommitTxPrivateKeyList, commitTxPrevOutputFetcher); err != nil {
		return err
	}

	builder.CommitTxFee = CalculateCommitTxFee(tx, commitTxPrevOutputFetcher)
	builder.CommitTx = tx

	var revealTxPrevOutputFetcher *txscript.MultiPrevOutFetcher
	var witnessList [][]byte
	if revealTxPrevOutputFetcher, witnessList, err = builder.FillRevealTx(revealTxs, builder.CommitTx.TxHash(), inscriptionTxCtxDataList); err != nil {
		return err
	}

	if err = builder.SignRevealTx(revealTxs, witnessList, inscriptionTxCtxDataList); err != nil {
		return err
	}

	if err = CheckRevealTx(revealTxs); err != nil {
		return err
	}

	builder.RevealTxFees = CalculateRevealTxFee(revealTxs, revealTxPrevOutputFetcher)
	builder.RevealTx = revealTxs

	return nil
}

func (builder *InscriptionBuilder) PreProcess(network *chaincfg.Params, inscriptionDataList []InscriptionData, argRevealOutValue int64, argMinChangeValue int64, revealFeeRate int64, pubKey []byte) (*Brc20InscriptionParseResult, error) {
	destinations := make([]string, len(inscriptionDataList))
	revealOutValue := DefaultRevealOutValue
	if argRevealOutValue > 0 {
		revealOutValue = argRevealOutValue
	}
	minChangeValue := DefaultMinChangeValue
	if argMinChangeValue > 0 {
		minChangeValue = argMinChangeValue
	}

	pk, err := btcec.ParsePubKey(pubKey)
	if err != nil {
		return nil, err
	}

	inscriptionTxCtxDataList := make([]*InscriptionTxCtxData, len(inscriptionDataList))
	for i := 0; i < len(inscriptionDataList); i++ {

		inscriptionTxCtxData, err := newInscriptionTxCtxData(network, &inscriptionDataList[i], nil, pk)
		if err != nil {
			return nil, err
		}
		inscriptionTxCtxDataList[i] = inscriptionTxCtxData
		destinations[i] = inscriptionDataList[i].RevealAddr
	}

	revealTxs, totalRevealPrevOutputValue, err := builder.BuildEmptyRevealTx(destinations, inscriptionTxCtxDataList, revealOutValue, revealFeeRate)
	if err != nil {
		return nil, err
	}

	ctxDataList := make([]*Brc20CtxData, len(inscriptionTxCtxDataList))
	for i := 0; i < len(inscriptionTxCtxDataList); i++ {

		data := &Brc20CtxData{
			CommitTxAddress:     inscriptionTxCtxDataList[i].CommitTxAddress,
			CommitTxOutPkScript: inscriptionTxCtxDataList[i].CommitTxAddressPkScript,
			CommitTxOutValue:    inscriptionTxCtxDataList[i].RevealTxPrevOutput.Value,

			InscriptionScript:   inscriptionTxCtxDataList[i].InscriptionScript,
			ControlBlockWitness: inscriptionTxCtxDataList[i].ControlBlockWitness,

			RevealTxOutPkScript: revealTxs[i].TxOut[0].PkScript,
			RevealTxOutValue:    revealTxs[i].TxOut[0].Value,
		}

		ctxDataList[i] = data
	}

	result := &Brc20InscriptionParseResult{
		CtxDataList:                ctxDataList,
		RevealOutValue:             revealOutValue,
		TotalRevealPrevOutputValue: totalRevealPrevOutputValue,
		MinChangeValue:             minChangeValue,
		CommitAddrs:                builder.CommitAddrs,
	}

	return result, nil
}

func newInscriptionTxCtxData(network *chaincfg.Params, inscriptionData *InscriptionData, privateKey *btcec.PrivateKey, pubKey *btcec.PublicKey) (*InscriptionTxCtxData, error) {

	inscriptionBuilder := txscript.NewScriptBuilder().
		AddData(schnorr.SerializePubKey(pubKey)).
		AddOp(txscript.OP_CHECKSIG).
		AddOp(txscript.OP_FALSE).
		AddOp(txscript.OP_IF).
		AddData([]byte("ord")).
		AddOp(txscript.OP_DATA_1).
		AddOp(txscript.OP_DATA_1).
		AddData([]byte(inscriptionData.ContentType)).
		AddOp(txscript.OP_0)
	maxChunkSize := 520
	// use taproot to skip txscript.MaxScriptSize 10000
	bodySize := len(inscriptionData.Body)
	for i := 0; i < bodySize; i += maxChunkSize {
		end := i + maxChunkSize
		if end > bodySize {
			end = bodySize
		}

		inscriptionBuilder.AddFullData(inscriptionData.Body[i:end])
	}
	inscriptionScript, err := inscriptionBuilder.Script()
	if err != nil {
		return nil, err
	}
	inscriptionScript = append(inscriptionScript, txscript.OP_ENDIF)

	proof := &txscript.TapscriptProof{
		TapLeaf:  txscript.NewBaseTapLeaf(schnorr.SerializePubKey(pubKey)),
		RootNode: txscript.NewBaseTapLeaf(inscriptionScript),
	}

	controlBlock := proof.ToControlBlock(pubKey)
	controlBlockWitness, err := controlBlock.ToBytes()
	if err != nil {
		return nil, err
	}

	tapHash := proof.RootNode.TapHash()
	commitTxAddress, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(txscript.ComputeTaprootOutputKey(pubKey, tapHash[:])), network)
	if err != nil {
		return nil, err
	}
	commitTxAddressPkScript, err := txscript.PayToAddrScript(commitTxAddress)
	if err != nil {
		return nil, err
	}

	return &InscriptionTxCtxData{
		PrivateKey:              privateKey,
		InscriptionScript:       inscriptionScript,
		CommitTxAddress:         commitTxAddress.EncodeAddress(),
		CommitTxAddressPkScript: commitTxAddressPkScript,
		ControlBlockWitness:     controlBlockWitness,
	}, nil
}

func (builder *InscriptionBuilder) BuildEmptyRevealTx(destination []string, inscriptionTxCtxDataList []*InscriptionTxCtxData, revealOutValue, revealFeeRate int64) ([]*wire.MsgTx, int64, error) {
	addTxInTxOutIntoRevealTx := func(tx *wire.MsgTx, index int) error {
		in := wire.NewTxIn(&wire.OutPoint{Index: uint32(index)}, nil, nil)
		in.Sequence = DefaultSequenceNum
		tx.AddTxIn(in)
		scriptPubKey, err := AddrToPkScript(destination[index], builder.Network)
		if err != nil {
			return err
		}
		out := wire.NewTxOut(revealOutValue, scriptPubKey)
		tx.AddTxOut(out)
		return nil
	}

	totalPrevOutputValue := int64(0)
	total := len(inscriptionTxCtxDataList)
	revealTx := make([]*wire.MsgTx, total)
	mustRevealTxFees := make([]int64, total)
	commitAddrs := make([]string, total)
	for i := 0; i < total; i++ {
		tx := wire.NewMsgTx(DefaultTxVersion)
		err := addTxInTxOutIntoRevealTx(tx, i)
		if err != nil {
			return revealTx, 0, err
		}
		prevOutputValue := revealOutValue + int64(tx.SerializeSize())*revealFeeRate
		emptySignature := make([]byte, 64)
		emptyControlBlockWitness := make([]byte, 33)
		fee := (int64(wire.TxWitness{emptySignature, inscriptionTxCtxDataList[i].InscriptionScript, emptyControlBlockWitness}.SerializeSize()+2+3) / 4) * revealFeeRate
		prevOutputValue += fee
		inscriptionTxCtxDataList[i].RevealTxPrevOutput = &wire.TxOut{
			PkScript: inscriptionTxCtxDataList[i].CommitTxAddressPkScript,
			Value:    prevOutputValue,
		}
		totalPrevOutputValue += prevOutputValue
		revealTx[i] = tx
		mustRevealTxFees[i] = int64(tx.SerializeSize())*revealFeeRate + fee
		commitAddrs[i] = inscriptionTxCtxDataList[i].CommitTxAddress
	}
	builder.MustRevealTxFees = mustRevealTxFees
	builder.CommitAddrs = commitAddrs

	return revealTx, totalPrevOutputValue, nil
}

func (builder *InscriptionBuilder) ParseCommitTxPrevOutput(commitTxPrevOutputList []*PrevOutput) (*txscript.MultiPrevOutFetcher, *wire.MsgTx, btcutil.Amount, error) {
	tx := wire.NewMsgTx(DefaultTxVersion)
	commitTxPrevOutputFetcher := txscript.NewMultiPrevOutFetcher(nil)
	totalSenderAmount := btcutil.Amount(0)

	for _, prevOutput := range commitTxPrevOutputList {
		txHash, err := chainhash.NewHashFromStr(prevOutput.TxId)
		if err != nil {
			return nil, nil, totalSenderAmount, err
		}
		outPoint := wire.NewOutPoint(txHash, prevOutput.VOut)
		pkScript, err := AddrToPkScript(prevOutput.Address, builder.Network)
		if err != nil {
			return nil, nil, totalSenderAmount, err
		}
		txOut := wire.NewTxOut(prevOutput.Amount, pkScript)
		commitTxPrevOutputFetcher.AddPrevOut(*outPoint, txOut)

		in := wire.NewTxIn(outPoint, nil, nil)
		in.Sequence = DefaultSequenceNum
		tx.AddTxIn(in)

		totalSenderAmount += btcutil.Amount(prevOutput.Amount)
	}

	return commitTxPrevOutputFetcher, tx, totalSenderAmount, nil
}

func (builder *InscriptionBuilder) FillCommitTxOutput(tx *wire.MsgTx, inscriptionTxCtxDataList []*InscriptionTxCtxData, changeAddress string) error {
	changePkScript, err := AddrToPkScript(changeAddress, builder.Network)
	if err != nil {
		return err
	}

	for i := range inscriptionTxCtxDataList {
		tx.AddTxOut(inscriptionTxCtxDataList[i].RevealTxPrevOutput)
	}

	tx.AddTxOut(wire.NewTxOut(0, changePkScript))

	return nil
}

func (builder *InscriptionBuilder) CompleteCommitTx(txForEstimate *wire.MsgTx, totalSenderAmount btcutil.Amount, totalRevealPrevOutputValue, commitFeeRate int64, minChangeValue int64) (*wire.MsgTx, error) {

	tx := wire.NewMsgTx(DefaultTxVersion)
	tx.TxIn = txForEstimate.TxIn
	tx.TxOut = txForEstimate.TxOut

	fee := btcutil.Amount(GetTxVirtualSize(btcutil.NewTx(txForEstimate))) * btcutil.Amount(commitFeeRate)
	changeAmount := totalSenderAmount - btcutil.Amount(totalRevealPrevOutputValue) - fee
	if int64(changeAmount) >= minChangeValue {
		tx.TxOut[len(tx.TxOut)-1].Value = int64(changeAmount)
	} else {
		tx.TxOut = tx.TxOut[:len(tx.TxOut)-1]
		if changeAmount < 0 {
			txForEstimate.TxOut = txForEstimate.TxOut[:len(txForEstimate.TxOut)-1]
			feeWithoutChange := btcutil.Amount(GetTxVirtualSize(btcutil.NewTx(txForEstimate))) * btcutil.Amount(commitFeeRate)
			if totalSenderAmount-btcutil.Amount(totalRevealPrevOutputValue)-feeWithoutChange < 0 {
				builder.MustCommitTxFee = int64(fee)
				return nil, errors.New("insufficient balance")
			}
		}
	}
	return tx, nil
}

func (builder *InscriptionBuilder) FillRevealTx(revealTxs []*wire.MsgTx, commitTxhash chainhash.Hash, inscriptionTxCtxDataList []*InscriptionTxCtxData) (*txscript.MultiPrevOutFetcher, [][]byte, error) {
	revealTxPrevOutputFetcher := txscript.NewMultiPrevOutFetcher(nil)
	witnessList := make([][]byte, len(inscriptionTxCtxDataList))

	for i := range inscriptionTxCtxDataList {
		revealTxPrevOutputFetcher.AddPrevOut(wire.OutPoint{
			Hash:  commitTxhash,
			Index: uint32(i),
		}, inscriptionTxCtxDataList[i].RevealTxPrevOutput)
		revealTxs[i].TxIn[0].PreviousOutPoint.Hash = commitTxhash
	}
	for i := range inscriptionTxCtxDataList {
		revealTx := revealTxs[i]
		witnessArray, err := txscript.CalcTapscriptSignaturehash(txscript.NewTxSigHashes(revealTx, revealTxPrevOutputFetcher),
			txscript.SigHashDefault, revealTx, 0, revealTxPrevOutputFetcher, txscript.NewBaseTapLeaf(inscriptionTxCtxDataList[i].InscriptionScript))
		if err != nil {
			return revealTxPrevOutputFetcher, witnessList, err
		}
		witnessList[i] = witnessArray
	}

	return revealTxPrevOutputFetcher, witnessList, nil
}

func (builder *InscriptionBuilder) SignRevealTx(revealTxs []*wire.MsgTx, witnessList [][]byte, inscriptionTxCtxDataList []*InscriptionTxCtxData) error {
	for _, tx := range revealTxs {
		tx.TxIn[0].Witness = wire.TxWitness{}
	}

	for i := range inscriptionTxCtxDataList {
		witnessArray := witnessList[i]
		signature, err := schnorr.Sign(inscriptionTxCtxDataList[i].PrivateKey, witnessArray)
		if err != nil {
			return err
		}
		witness := wire.TxWitness{signature.Serialize(), inscriptionTxCtxDataList[i].InscriptionScript, inscriptionTxCtxDataList[i].ControlBlockWitness}
		revealTxs[i].TxIn[0].Witness = witness
	}

	return nil
}
func (builder *InscriptionBuilder) SignRevealTx2(revealTxs []*wire.MsgTx, signature string, inscriptionTxCtxDataList []*InscriptionTxCtxData) error {
	for _, tx := range revealTxs {
		tx.TxIn[0].Witness = wire.TxWitness{}
	}

	for i := range inscriptionTxCtxDataList {
		signature, err := txscript.BuildSignature(signature)
		if err != nil {
			return err
		}
		witness := wire.TxWitness{signature.Serialize(), inscriptionTxCtxDataList[i].InscriptionScript, inscriptionTxCtxDataList[i].ControlBlockWitness}
		revealTxs[i].TxIn[0].Witness = witness
	}

	return nil
}

func CheckRevealTx(revealTxs []*wire.MsgTx) error {
	// check tx max tx wight
	for i, tx := range revealTxs {
		revealWeight := GetTransactionWeight(btcutil.NewTx(tx))
		if revealWeight > MaxStandardTxWeight {
			return errors.New(fmt.Sprintf("reveal(index %d) transaction weight greater than %d (MAX_STANDARD_TX_WEIGHT): %d", i, MaxStandardTxWeight, revealWeight))
		}
	}

	return nil
}

func Sign(tx *wire.MsgTx, privateKeys []*btcec.PrivateKey, prevOutFetcher *txscript.MultiPrevOutFetcher) error {
	for i, in := range tx.TxIn {
		prevOut := prevOutFetcher.FetchPrevOutput(in.PreviousOutPoint)
		txSigHashes := txscript.NewTxSigHashes(tx, prevOutFetcher)
		privKey := privateKeys[i]
		if txscript.IsPayToTaproot(prevOut.PkScript) {
			witness, err := txscript.TaprootWitnessSignature(tx, txSigHashes, i, prevOut.Value, prevOut.PkScript, txscript.SigHashDefault, privKey)
			if err != nil {
				return err
			}
			in.Witness = witness
		} else if txscript.IsPayToPubKeyHash(prevOut.PkScript) {
			sigScript, err := txscript.SignatureScript(tx, i, prevOut.PkScript, txscript.SigHashAll, privKey, true)
			if err != nil {
				return err
			}
			in.SignatureScript = sigScript
		} else {
			pubKeyBytes := privKey.PubKey().SerializeCompressed()
			script, err := PayToPubKeyHashScript(btcutil.Hash160(pubKeyBytes))
			if err != nil {
				return err
			}
			amount := prevOut.Value
			witness, err := txscript.WitnessSignature(tx, txSigHashes, i, amount, script, txscript.SigHashAll, privKey, true)
			if err != nil {
				return err
			}
			in.Witness = witness

			if txscript.IsPayToScriptHash(prevOut.PkScript) {
				redeemScript, err := PayToWitnessPubKeyHashScript(btcutil.Hash160(pubKeyBytes))
				if err != nil {
					return err
				}
				in.SignatureScript = append([]byte{byte(len(redeemScript))}, redeemScript...)
			}
		}
	}

	return nil
}

func GetTxHex(tx *wire.MsgTx) (string, error) {
	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf.Bytes()), nil
}

func (builder *InscriptionBuilder) GetCommitTxHex() (string, error) {
	return GetTxHex(builder.CommitTx)
}

func (builder *InscriptionBuilder) GetRevealTxHexList() ([]string, error) {
	txHexList := make([]string, len(builder.RevealTx))
	for i := range builder.RevealTx {
		txHex, err := GetTxHex(builder.RevealTx[i])
		if err != nil {
			return nil, err
		}
		txHexList[i] = txHex
	}
	return txHexList, nil
}

func CalculateCommitTxFee(commitTx *wire.MsgTx, commitTxPrevOutputFetcher *txscript.MultiPrevOutFetcher) int64 {
	commitTxFee := int64(0)
	for _, in := range commitTx.TxIn {
		commitTxFee += commitTxPrevOutputFetcher.FetchPrevOutput(in.PreviousOutPoint).Value
	}
	for _, out := range commitTx.TxOut {
		commitTxFee -= out.Value
	}

	return commitTxFee
}

func CalculateRevealTxFee(revealTxs []*wire.MsgTx, revealTxPrevOutputFetcher *txscript.MultiPrevOutFetcher) []int64 {
	revealTxFees := make([]int64, 0)
	for _, tx := range revealTxs {
		revealTxFee := int64(0)
		for i, in := range tx.TxIn {
			revealTxFee += revealTxPrevOutputFetcher.FetchPrevOutput(in.PreviousOutPoint).Value
			revealTxFee -= tx.TxOut[i].Value
			revealTxFees = append(revealTxFees, revealTxFee)
		}
	}
	return revealTxFees
}

func Inscribe(network *chaincfg.Params, request *InscriptionRequest) (*InscribeTxs, error) {
	tool, err := NewInscriptionTool(network, request)
	if err != nil && err.Error() == "insufficient balance" {
		return &InscribeTxs{
			CommitTx:     "",
			RevealTxs:    []string{},
			CommitTxFee:  tool.MustCommitTxFee,
			RevealTxFees: tool.MustRevealTxFees,
			CommitAddrs:  tool.CommitAddrs,
		}, nil
	}

	if err != nil {
		return nil, err
	}

	commitTx, err := tool.GetCommitTxHex()
	if err != nil {
		return nil, err
	}
	revealTxs, err := tool.GetRevealTxHexList()
	if err != nil {
		return nil, err
	}

	return &InscribeTxs{
		CommitTx:     commitTx,
		RevealTxs:    revealTxs,
		CommitTxFee:  tool.CommitTxFee,
		RevealTxFees: tool.RevealTxFees,
		CommitAddrs:  tool.CommitAddrs,
	}, nil
}

// GetTransactionWeight computes the value of the weight metric for a given
// transaction. Currently the weight metric is simply the sum of the
// transactions's serialized size without any witness data scaled
// proportionally by the WitnessScaleFactor, and the transaction's serialized
// size including any witness data.
func GetTransactionWeight(tx *btcutil.Tx) int64 {
	msgTx := tx.MsgTx()

	baseSize := msgTx.SerializeSizeStripped()
	totalSize := msgTx.SerializeSize()

	// (baseSize * 3) + totalSize
	return int64((baseSize * (WitnessScaleFactor - 1)) + totalSize)
}

// GetTxVirtualSize computes the virtual size of a given transaction. A
// transaction's virtual size is based off its weight, creating a discount for
// any witness data it contains, proportional to the current
// blockchain.WitnessScaleFactor value.
func GetTxVirtualSize(tx *btcutil.Tx) int64 {
	// vSize := (weight(tx) + 3) / 4
	//       := (((baseSize * 3) + totalSize) + 3) / 4
	// We add 3 here as a way to compute the ceiling of the prior arithmetic
	// to 4. The division by 4 creates a discount for wit witness data.
	return (GetTransactionWeight(tx) + (WitnessScaleFactor - 1)) / WitnessScaleFactor
}

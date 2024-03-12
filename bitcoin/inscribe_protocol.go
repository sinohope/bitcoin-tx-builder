package bitcoin

import (
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

func PrepareBrc20CommitTx(network *chaincfg.Params, inscriptionDataList []InscriptionData, commitTxPrevOutputList []*PrevOutput,
	revealOutValue int64, minChangeValue int64, revealFeeRate int64, changeAddress string, pubKey []byte) (*Brc20InscriptionParseResult, string, btcutil.Amount, error) {
	tool := &InscriptionBuilder{
		Network: network,
	}

	var err error
	var parseResult *Brc20InscriptionParseResult
	var tx *wire.MsgTx
	var txHex string
	totalSenderAmount := btcutil.Amount(0)

	// brc20 parse
	if parseResult, err = tool.PreProcess(network, inscriptionDataList, revealOutValue, minChangeValue, revealFeeRate, pubKey); err != nil {
		return parseResult, txHex, totalSenderAmount, err
	}

	if _, tx, totalSenderAmount, err = tool.ParseCommitTxPrevOutput(commitTxPrevOutputList); err != nil {
		return parseResult, txHex, totalSenderAmount, err
	}

	inscriptionTxCtxDataList := make([]*InscriptionTxCtxData, len(inscriptionDataList))
	for i := 0; i < len(inscriptionDataList); i++ {
		inscriptionTxCtxData := &InscriptionTxCtxData{
			InscriptionScript:       parseResult.CtxDataList[i].InscriptionScript,
			CommitTxAddress:         parseResult.CtxDataList[i].CommitTxAddress,
			CommitTxAddressPkScript: parseResult.CtxDataList[i].CommitTxOutPkScript,
			ControlBlockWitness:     parseResult.CtxDataList[i].ControlBlockWitness,
			RevealTxPrevOutput: &wire.TxOut{
				PkScript: parseResult.CtxDataList[i].CommitTxOutPkScript,
				Value:    parseResult.CtxDataList[i].CommitTxOutValue,
			},
		}
		inscriptionTxCtxDataList[i] = inscriptionTxCtxData
	}

	if err = tool.FillCommitTxOutput(tx, inscriptionTxCtxDataList, changeAddress); err != nil {
		return parseResult, txHex, totalSenderAmount, err
	}

	if txHex, err = GetTxHex(tx); err != nil {
		return parseResult, txHex, totalSenderAmount, err
	}

	return parseResult, txHex, totalSenderAmount, nil
}

func SignBrc20CommitTx(network *chaincfg.Params, txHex string, commitTxPrevOutputList []*PrevOutput, commitTxPrivateKeyListWif []string) (string, error) {
	tool := &InscriptionBuilder{
		Network: network,
	}

	var txSignedHex string

	var commitTxPrivateKeyList []*btcec.PrivateKey
	for _, prvkey := range commitTxPrivateKeyListWif {
		privateKeyWif, err := btcutil.DecodeWIF(prvkey)
		if err != nil {
			return txSignedHex, err
		}
		commitTxPrivateKeyList = append(commitTxPrivateKeyList, privateKeyWif.PrivKey)
	}

	commitTxPrevOutputFetcher, _, _, err := tool.ParseCommitTxPrevOutput(commitTxPrevOutputList)
	if err != nil {
		return txSignedHex, err
	}

	var tx *wire.MsgTx
	if tx, err = NewTxFromHex(txHex); err != nil {
		return txSignedHex, err
	}

	if err = Sign(tx, commitTxPrivateKeyList, commitTxPrevOutputFetcher); err != nil {
		return txSignedHex, err
	}

	if txSignedHex, err = GetTxHex(tx); err != nil {
		return txSignedHex, err
	}

	return txSignedHex, nil
}

func AdjustBrc20CommitTx(network *chaincfg.Params, txForEstimateHex string, commitTxPrevOutputList []*PrevOutput,
	totalSenderAmount btcutil.Amount, totalRevealPrevOutputValue int64, commitFeeRate int64, minChangeValue int64) (string, int64, error) {
	tool := &InscriptionBuilder{
		Network: network,
	}

	var commitTx *wire.MsgTx
	var commitTxHex string

	commitTxPrevOutputFetcher, _, _, err := tool.ParseCommitTxPrevOutput(commitTxPrevOutputList)
	if err != nil {
		return commitTxHex, 0, err
	}

	var txForEstimate *wire.MsgTx
	if txForEstimate, err = NewTxFromHex(txForEstimateHex); err != nil {
		return commitTxHex, 0, err
	}

	if commitTx, err = tool.CompleteCommitTx(txForEstimate, totalSenderAmount, totalRevealPrevOutputValue, commitFeeRate, minChangeValue); err != nil {
		return commitTxHex, 0, err
	}

	if commitTxHex, err = GetTxHex(commitTx); err != nil {
		return commitTxHex, 0, err
	}

	commitTxFee := CalculateCommitTxFee(commitTx, commitTxPrevOutputFetcher)

	return commitTxHex, commitTxFee, err
}

func BuildBrc20CommitTx(network *chaincfg.Params, inscriptionDataList []InscriptionData, commitTxPrevOutputList []*PrevOutput,
	revealOutValue int64, minChangeValue int64, commitFeeRate int64, revealFeeRate int64, changeAddress string,
	pubKey []byte, commitTxPrivateKeyListWif []string) (*Brc20InscriptionParseResult, string, int64, error) {

	// 1.prepare commit tx
	parseResult, txPreparedHex, totalSenderAmount, err := PrepareBrc20CommitTx(network, inscriptionDataList, commitTxPrevOutputList, revealOutValue, minChangeValue, revealFeeRate, changeAddress, pubKey)
	if err != nil {
		return nil, "", 0, err
	}

	// 2.pre sign commit tx
	var txForEstimateHex string
	txForEstimateHex, err = SignBrc20CommitTx(network, txPreparedHex, commitTxPrevOutputList, commitTxPrivateKeyListWif)
	if err != nil {
		return nil, "", 0, err
	}

	// 3. check tx virtualsize and adjust output
	var txCheckedHex string
	var commitTxFee int64
	txCheckedHex, commitTxFee, err = AdjustBrc20CommitTx(network, txForEstimateHex, commitTxPrevOutputList, totalSenderAmount, parseResult.TotalRevealPrevOutputValue, commitFeeRate, parseResult.MinChangeValue)
	if err != nil {
		return nil, "", 0, err
	}

	return parseResult, txCheckedHex, commitTxFee, nil
}

func BuildBrc20RevealTx(network *chaincfg.Params, commitTxHash chainhash.Hash, ctxDataList []*Brc20CtxData, revealAddrs []string,
	revealFeeRate int64, revealOutValue int64) ([]string, [][]byte, []int64, error) {
	tool := &InscriptionBuilder{
		Network: network,
	}

	inscriptionTxCtxDataList := make([]*InscriptionTxCtxData, len(ctxDataList))
	for i := 0; i < len(ctxDataList); i++ {

		inscriptionTxCtxData := &InscriptionTxCtxData{
			InscriptionScript:       ctxDataList[i].InscriptionScript,
			CommitTxAddress:         ctxDataList[i].CommitTxAddress,
			CommitTxAddressPkScript: ctxDataList[i].CommitTxOutPkScript,
			ControlBlockWitness:     ctxDataList[i].ControlBlockWitness,
			RevealTxPrevOutput: &wire.TxOut{
				PkScript: ctxDataList[i].CommitTxOutPkScript,
				Value:    ctxDataList[i].CommitTxOutValue,
			},
		}
		inscriptionTxCtxDataList[i] = inscriptionTxCtxData
	}

	var witnessList [][]byte
	var revealTxFees []int64
	reavealTxsHex := make([]string, 0)

	revealTxs, _, err := tool.BuildEmptyRevealTx(revealAddrs, inscriptionTxCtxDataList, revealOutValue, revealFeeRate)
	if err != nil {
		return reavealTxsHex, witnessList, revealTxFees, err
	}

	var revealTxPrevOutputFetcher *txscript.MultiPrevOutFetcher
	if revealTxPrevOutputFetcher, witnessList, err = tool.FillRevealTx(revealTxs, commitTxHash, inscriptionTxCtxDataList); err != nil {
		return reavealTxsHex, witnessList, revealTxFees, err
	}

	revealTxFees = CalculateRevealTxFee(revealTxs, revealTxPrevOutputFetcher)

	for _, tx := range revealTxs {
		txHex, err := GetTxHex(tx)
		if err != nil {
			return reavealTxsHex, witnessList, revealTxFees, err
		}
		reavealTxsHex = append(reavealTxsHex, txHex)
	}

	return reavealTxsHex, witnessList, revealTxFees, nil
}

func SignBrc20RevealTx(network *chaincfg.Params, revealTxsHex []string, witnessList [][]byte, ctxDataList []*Brc20CtxData, firstPrevOutPrivateKey string) ([]string, error) {
	tool := &InscriptionBuilder{
		Network: network,
	}

	var signedRevealTxsHex []string

	// must use the first privatekey, previous is the pubkey of the first privatekey
	privateKeyWif, err := btcutil.DecodeWIF(firstPrevOutPrivateKey)
	if err != nil {
		return signedRevealTxsHex, err
	}

	revealTxs := make([]*wire.MsgTx, len(revealTxsHex))
	for i, txHex := range revealTxsHex {
		tx, err := NewTxFromHex(txHex)
		if err != nil {
			return signedRevealTxsHex, err
		}

		revealTxs[i] = tx
	}

	inscriptionTxCtxDataList := make([]*InscriptionTxCtxData, len(ctxDataList))
	for i := 0; i < len(ctxDataList); i++ {

		inscriptionTxCtxData := &InscriptionTxCtxData{
			PrivateKey:              privateKeyWif.PrivKey,
			InscriptionScript:       ctxDataList[i].InscriptionScript,
			CommitTxAddress:         ctxDataList[i].CommitTxAddress,
			CommitTxAddressPkScript: ctxDataList[i].CommitTxOutPkScript,
			ControlBlockWitness:     ctxDataList[i].ControlBlockWitness,
			RevealTxPrevOutput: &wire.TxOut{
				PkScript: ctxDataList[i].CommitTxOutPkScript,
				Value:    ctxDataList[i].CommitTxOutValue,
			},
		}
		inscriptionTxCtxDataList[i] = inscriptionTxCtxData
	}

	if err = tool.SignRevealTx(revealTxs, witnessList, inscriptionTxCtxDataList); err != nil {
		return signedRevealTxsHex, err
	}

	for _, tx := range revealTxs {
		txHex, err := GetTxHex(tx)
		if err != nil {
			return signedRevealTxsHex, err
		}
		signedRevealTxsHex = append(signedRevealTxsHex, txHex)
	}

	return signedRevealTxsHex, nil
}

func CheckBrc20RevealTx(revealTxsHex []string) error {
	revealTxs := make([]*wire.MsgTx, len(revealTxsHex))
	for i, txHex := range revealTxsHex {
		tx, err := NewTxFromHex(txHex)
		if err != nil {
			return err
		}

		revealTxs[i] = tx
	}

	return CheckRevealTx(revealTxs)
}

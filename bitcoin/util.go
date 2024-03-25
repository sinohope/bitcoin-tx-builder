package bitcoin

import (
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/wire"
	"github.com/etherria/bitcoin-tx-builder/bitcoin/txscript"
)

func getMessageHash(tx *wire.MsgTx, privateKeys []*btcec.PrivateKey, prevOutFetcher *txscript.MultiPrevOutFetcher) (map[int][]byte, error) {
	var messageHashes = make(map[int][]byte)
	for i, in := range tx.TxIn {
		prevOut := prevOutFetcher.FetchPrevOutput(in.PreviousOutPoint)
		txSigHashes := txscript.NewTxSigHashes(tx, prevOutFetcher)
		privKey := privateKeys[i]
		if txscript.IsPayToTaproot(prevOut.PkScript) {
			//witness, err := txscript.TaprootWitnessSignature(tx, txSigHashes, i, prevOut.Value, prevOut.PkScript, txscript.SigHashDefault, privKey)
			sigHash, err := txscript.CalcTaprootSignatureHashRaw(
				txSigHashes, txscript.SigHashDefault, tx, i,
				txscript.NewCannedPrevOutputFetcher(prevOut.PkScript, prevOut.Value),
			)
			if err != nil {
				return messageHashes, err
			}
			messageHashes[i] = sigHash
		} else if txscript.IsPayToPubKeyHash(prevOut.PkScript) {
			//sigScript, err := txscript.SignatureScript(tx, i, prevOut.PkScript, txscript.SigHashAll, privKey, true)
			hash, err := txscript.CalcSignatureHash(prevOut.PkScript, txscript.SigHashAll, tx, i)
			if err != nil {
				return messageHashes, err
			}
			messageHashes[i] = hash
		} else {
			pubKeyBytes := privKey.PubKey().SerializeCompressed()
			script, err := PayToPubKeyHashScript(btcutil.Hash160(pubKeyBytes))
			if err != nil {
				return messageHashes, err
			}
			amount := prevOut.Value
			//witness, err := txscript.WitnessSignature(tx, txSigHashes, i, amount, script, txscript.SigHashAll, privKey, true)
			hash, err := txscript.CalcWitnessSignatureHashRaw(script, txSigHashes, txscript.SigHashAll, tx,
				i, amount)
			if err != nil {
				return messageHashes, err
			}
			messageHashes[i] = hash
		}
	}
	return messageHashes, nil
}

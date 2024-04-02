package bitcoin

import (
	"encoding/hex"
	"errors"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/etherria/bitcoin-tx-builder/bitcoin/txscript"
)

func GetMessageHash(tx *wire.MsgTx, pubKeyBytes []byte, prevOutFetcher *txscript.MultiPrevOutFetcher) (map[int]string, error) {
	var messageHashes = make(map[int]string)
	for i, in := range tx.TxIn {
		prevOut := prevOutFetcher.FetchPrevOutput(in.PreviousOutPoint)
		txSigHashes := txscript.NewTxSigHashes(tx, prevOutFetcher)
		if txscript.IsPayToTaproot(prevOut.PkScript) {
			//witness, err := txscript.TaprootWitnessSignature(tx, txSigHashes, i, prevOut.Value, prevOut.PkScript, txscript.SigHashDefault, privKey)
			sigHash, err := txscript.CalcTaprootSignatureHashRaw(
				txSigHashes, txscript.SigHashDefault, tx, i,
				txscript.NewCannedPrevOutputFetcher(prevOut.PkScript, prevOut.Value),
			)
			if err != nil {
				return messageHashes, err
			}
			messageHashes[i] = hexutil.Encode(sigHash)
		} else if txscript.IsPayToPubKeyHash(prevOut.PkScript) {
			//sigScript, err := txscript.SignatureScript(tx, i, prevOut.PkScript, txscript.SigHashAll, privKey, true)
			hash, err := txscript.CalcSignatureHash(prevOut.PkScript, txscript.SigHashAll, tx, i)
			if err != nil {
				return messageHashes, err
			}
			messageHashes[i] = hexutil.Encode(hash)
		} else {
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
			messageHashes[i] = hexutil.Encode(hash)
		}
	}
	return messageHashes, nil
}

func BuildRawData(network *chaincfg.Params, txHex string, commitTxPrevOutputList []*PrevOutput, signatureMap map[int]string, pubKey string) (string, error) {
	tool := &InscriptionBuilder{
		Network: network,
	}

	var txSignedHex string

	commitTxPrevOutputFetcher, _, _, err := tool.ParseCommitTxPrevOutput(commitTxPrevOutputList)
	if err != nil {
		return txSignedHex, err
	}

	var tx *wire.MsgTx
	if tx, err = NewTxFromHex(txHex); err != nil {
		return txSignedHex, err
	}

	if err = SignBySignature(tx, commitTxPrevOutputFetcher, signatureMap, pubKey); err != nil {
		return txSignedHex, err
	}

	if txSignedHex, err = GetTxHex(tx); err != nil {
		return txSignedHex, err
	}

	return txSignedHex, nil
}

func SignBySignature(tx *wire.MsgTx, prevOutFetcher *txscript.MultiPrevOutFetcher, signatureMap map[int]string, pubKey string) error {
	for i, in := range tx.TxIn {
		prevOut := prevOutFetcher.FetchPrevOutput(in.PreviousOutPoint)
		txSigHashes := txscript.NewTxSigHashes(tx, prevOutFetcher)
		if txscript.IsPayToTaproot(prevOut.PkScript) {
			//witness, err := txscript.TaprootWitnessSignature(tx, txSigHashes, i, prevOut.Value, prevOut.PkScript, txscript.SigHashDefault, privKey)
			//if err != nil {
			//	return err
			//}
			//in.Witness = witness
			return errors.New("not supper taproot address")
		} else if txscript.IsPayToPubKeyHash(prevOut.PkScript) {
			sigScript, err := txscript.SignatureScript2(tx, i, prevOut.PkScript, txscript.SigHashAll, signatureMap[i], pubKey, true)
			if err != nil {
				return err
			}
			in.SignatureScript = sigScript
		} else {
			serialized, err := hex.DecodeString(pubKey)
			if err != nil {
				return err
			}
			pk, err := btcec.ParsePubKey(serialized)
			if err != nil {
				return err
			}
			pubKeyBytes := pk.SerializeCompressed()
			script, err := PayToPubKeyHashScript(btcutil.Hash160(pubKeyBytes))
			if err != nil {
				return err
			}
			amount := prevOut.Value
			witness, err := txscript.WitnessSignature2(tx, txSigHashes, i, amount, script, txscript.SigHashAll, signatureMap[i], pubKeyBytes, true)
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

func ParsePubKey(pubKeyStr string) (*btcec.PublicKey, error) {
	serializedPubKey, err := hex.DecodeString(pubKeyStr)
	pk, err := btcec.ParsePubKey(serializedPubKey)
	if err != nil {
		return nil, err
	}
	return pk, nil
}

package controllers

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/grupokindynos/common/blockbook"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/grupokindynos/common/plutus"
	"github.com/grupokindynos/plutus/models"
	"github.com/tyler-smith/go-bip39"
	"os"
	"strconv"
	"strings"
)

type Params struct {
	Coin string
	Body []byte
	Txid string
}

var ethAccount = "0x4dc011f9792d18cd67f5afa4f1678e9c6c4d8e0e"

const addrGap = 20

type AddrInfo struct {
	LastUsed int
	AddrInfo []models.AddrInfo
}

type Controller struct {
	Address map[string]AddrInfo
}

func (c *Controller) GetBalance(params Params) (interface{}, error) {
	coinConfig, err := coinfactory.GetCoin(params.Coin)
	if err != nil {
		return nil, err
	}
	if !coinConfig.Token && coinConfig.Tag != "ETH" {
		blockBookWrap, err := blockbook.NewBlockBookWrapper(coinConfig.BlockExplorer)
		if err != nil {
			return nil, err
		}
		acc, err := getAccFromMnemonic(coinConfig)
		if err != nil {
			return nil, err
		}
		pub, err := acc.Neuter()
		if err != nil {
			return nil, err
		}
		info, err := blockBookWrap.GetXpub(pub.String())
		if err != nil {
			return nil, err
		}
		confirmed, err := strconv.ParseFloat(info.Balance, 64)
		if err != nil {
			return nil, err
		}
		unconfirmed, err := strconv.ParseFloat(info.UnconfirmedBalance, 64)
		response := plutus.Balance{
			Confirmed:   confirmed / 1e8,
			Unconfirmed: unconfirmed / 1e8,
		}
		return response, nil
	} else {
		ethConfig, err := coinfactory.GetCoin("ETH")
		if err != nil {
			return nil, err
		}
		blockBookWrap, err := blockbook.NewBlockBookWrapper(ethConfig.BlockExplorer)
		if err != nil {
			return nil, err
		}
		info, err := blockBookWrap.GetEthAddress(ethAccount)
		if err != nil {
			return nil, err
		}
		if coinConfig.Token {
			var tokenInfo *blockbook.EthTokens
			for _, token := range info.Tokens {
				if coinConfig.Contract == token.Contract {
					tokenInfo = &token
				}
			}
			if tokenInfo == nil {
				response := plutus.Balance{
					Confirmed: 0,
				}
				return response, nil
			}
			balance, err := strconv.ParseFloat(tokenInfo.Balance, 64)
			if err != nil {
				return nil, err
			}
			response := plutus.Balance{
				Confirmed: balance,
			}
			return response, nil
		} else {
			balance, err := strconv.ParseFloat(info.Balance, 64)
			if err != nil {
				return nil, err
			}
			response := plutus.Balance{
				Confirmed: balance / 1e18,
			}
			return response, nil
		}
	}
}

func (c *Controller) GetAddress(params Params) (interface{}, error) {
	coinConfig, err := coinfactory.GetCoin(params.Coin)
	if err != nil {
		return nil, err
	}
	if coinConfig.Token || coinConfig.Tag == "ETH" {
		return ethAccount, nil
	}
	acc, err := getAccFromMnemonic(coinConfig)
	if err != nil {
		return nil, err
	}
	// Create a new xpub and derive the address from the hdwallet
	directExtended, err := acc.Child(0)
	if err != nil {
		return nil, err
	}
	addrExtPub, err := directExtended.Child(uint32(c.Address[coinConfig.Tag].LastUsed + 1))
	if err != nil {
		return nil, err
	}
	addr, err := addrExtPub.Address(coinConfig.NetParams)
	if err != nil {
		return nil, err
	}
	newAddrInfo := AddrInfo{
		LastUsed: c.Address[coinConfig.Tag].LastUsed + 1,
		AddrInfo: c.Address[coinConfig.Tag].AddrInfo,
	}
	newAddrInfo.AddrInfo = append(newAddrInfo.AddrInfo, models.AddrInfo{
		Addr: addr.String(), Path: c.Address[coinConfig.Tag].LastUsed + 1,
	})
	c.Address[coinConfig.Tag] = newAddrInfo
	return addr.String(), nil
}

func (c *Controller) SendToAddress(params Params) (interface{}, error) {
	var SendToAddressData plutus.SendAddressBodyReq
	err := json.Unmarshal(params.Body, &SendToAddressData)
	if err != nil {
		return nil, err
	}
	value, err := btcutil.NewAmount(SendToAddressData.Amount)
	if err != nil {
		return nil, err
	}
	coinConfig, err := coinfactory.GetCoin(SendToAddressData.Coin)
	if err != nil {
		return nil, err
	}
	acc, err := getAccFromMnemonic(coinConfig)
	if err != nil {
		return nil, err
	}
	blockBookWrap, err := blockbook.NewBlockBookWrapper(coinConfig.BlockExplorer)
	if err != nil {
		return nil, err
	}
	utxos, err := blockBookWrap.GetUtxo(acc.String(), false)
	if err != nil {
		return nil, err
	}
	if len(utxos) == 0 {
		return nil, errors.New("no balance available")
	}
	var Tx wire.MsgTx
	var txVersion int32
	if coinConfig.Tag == "POLIS" || coinConfig.Tag == "DASH" {
		txVersion = 2
	} else {
		txVersion = 1
	}
	var availableAmount btcutil.Amount
	var changeAddrPubKeyHash string
	// Add the inputs without signatures
	for i, utxo := range utxos {
		if i == 0 {
			changeAddrPubKeyHash = utxo.Address
		}
		intValue, err := strconv.ParseInt(utxo.Value, 10, 64)
		if err != nil {
			return nil, err
		}
		utxoAmount := btcutil.Amount(intValue)
		availableAmount += utxoAmount
		txidHash, err := chainhash.NewHashFromStr(utxo.Txid)
		if err != nil {
			return nil, err
		}
		prevOut := wire.NewOutPoint(txidHash, uint32(utxo.Vout))
		in := wire.NewTxIn(prevOut, nil, nil)
		Tx.AddTxIn(in)
	}
	// Retrieve information for outputs
	payAddr, err := btcutil.DecodeAddress(SendToAddressData.Address, coinConfig.NetParams)
	changeAddr, err := btcutil.DecodeAddress(changeAddrPubKeyHash, coinConfig.NetParams)
	pkScriptPay, err := txscript.PayToAddrScript(payAddr)
	pkScriptChange, err := txscript.PayToAddrScript(changeAddr)
	txOut := &wire.TxOut{
		Value:    int64(value.ToUnit(btcutil.AmountSatoshi)),
		PkScript: pkScriptPay,
	}
	fee, err := blockBookWrap.GetFee("6")
	if err != nil {
		return nil, err
	}
	var feeRate int64
	if fee.Result == "-1" {
		feeRate = 2000
	} else {
		feeParse, err := strconv.ParseFloat(fee.Result, 64)
		if err != nil {
			return nil, err
		}
		feeRate = int64(feeParse * 1e8)
	}
	txSize := (len(Tx.TxIn) * 180) + (len(Tx.TxOut) * 34)
	payingFee := btcutil.Amount((feeRate / 1024) * int64(txSize))
	if availableAmount-payingFee-value > 0 {
		txOutChange := &wire.TxOut{
			Value:    int64(((availableAmount - value) - payingFee).ToUnit(btcutil.AmountSatoshi)),
			PkScript: pkScriptChange,
		}
		Tx.AddTxOut(txOutChange)
	}
	Tx.AddTxOut(txOut)
	Tx.Version = txVersion
	// Create the signatures
	for i, utxo := range utxos {
		utxoPrevOutHash, err := chainhash.NewHashFromStr(utxo.Txid)
		if err != nil {
			return nil, err
		}
		path := strings.Split(utxo.Path, "/")
		pathParse, err := strconv.ParseInt(path[5], 10, 64)
		if err != nil {
			return nil, err
		}
		privKey, err := getPrivKeyFromPath(coinConfig, uint32(pathParse))
		if err != nil {
			return nil, err
		}
		addr, err := btcutil.DecodeAddress(utxo.Address, coinConfig.NetParams)
		if err != nil {
			return nil, err
		}
		subscript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, err
		}
		sigScript, err := txscript.SignatureScript(&Tx, i, subscript, txscript.SigHashSingle, privKey, true)
		if err != nil {
			return nil, err
		}
		for _, in := range Tx.TxIn {
			if in.PreviousOutPoint.Hash.IsEqual(utxoPrevOutHash) {
				in.SignatureScript = sigScript
			}
		}
	}
	buf := bytes.NewBuffer([]byte{})
	err = Tx.BtcEncode(buf, 0, wire.BaseEncoding)
	if err != nil {
		return nil, err
	}
	rawTx := hex.EncodeToString(buf.Bytes())
	return blockBookWrap.SendTx(rawTx)
}

func (c *Controller) ValidateAddress(params Params) (interface{}, error) {
	var ValidateAddressData models.AddressValidationBodyReq
	err := json.Unmarshal(params.Body, &ValidateAddressData)
	if err != nil {
		return nil, err
	}
	coinConfig, err := coinfactory.GetCoin(ValidateAddressData.Coin)
	if err != nil {
		return nil, err
	}
	var isMine bool
	for _, addr := range c.Address[coinConfig.Tag].AddrInfo {
		if addr.Addr == ValidateAddressData.Address {
			isMine = true
		}
	}
	return isMine, nil
}

func (c *Controller) ValidateRawTx(params Params) (interface{}, error) {
	var ValidateTxData plutus.ValidateRawTxReq
	err := json.Unmarshal(params.Body, &ValidateTxData)
	if err != nil {
		return nil, err
	}
	value := btcutil.Amount(ValidateTxData.Amount)
	coinConfig, err := coinfactory.GetCoin(ValidateTxData.Coin)
	if err != nil {
		return nil, err
	}
	rawTxBytes, err := hex.DecodeString(ValidateTxData.RawTx)
	if err != nil {
		return nil, err
	}
	tx, err := btcutil.NewTxFromBytes(rawTxBytes)
	if err != nil {
		return nil, err
	}
	var isValue, isAddress bool
	for _, out := range tx.MsgTx().TxOut {
		outAmount := btcutil.Amount(out.Value)
		if outAmount == value {
			isValue = true
		}
		for _, addr := range c.Address[coinConfig.Tag].AddrInfo {
			Addr, err := btcutil.DecodeAddress(addr.Addr, coinConfig.NetParams)
			if err != nil {
				return nil, err
			}
			scriptAddr, err := txscript.PayToAddrScript(Addr)
			if err != nil {
				return nil, err
			}
			if bytes.Equal(scriptAddr, out.PkScript) {
				isAddress = true
			}
		}
	}
	if isValue && isAddress {
		return true, nil
	} else {
		return false, nil
	}
}

func (c *Controller) getAddrs(coinConfig *coins.Coin) error {
	acc, err := getAccFromMnemonic(coinConfig)
	if err != nil {
		return err
	}
	blockBookWrap, err := blockbook.NewBlockBookWrapper(coinConfig.BlockExplorer)
	if err != nil {
		return err
	}
	info, err := blockBookWrap.GetXpub(acc.String())
	if err != nil {
		return err
	}
	var addrInfoSlice []models.AddrInfo
	for i := info.UsedTokens; i < info.UsedTokens+addrGap; i++ {
		addr, err := getPubKeyHashFromPath(acc, coinConfig, uint32(i))
		if err != nil {
			return err
		}
		addrInfo := models.AddrInfo{Addr: addr, Path: i}
		addrInfoSlice = append(addrInfoSlice, addrInfo)
	}
	c.Address[coinConfig.Tag] = AddrInfo{
		LastUsed: info.UsedTokens,
		AddrInfo: addrInfoSlice,
	}
	return nil
}

func getAccFromMnemonic(coinConfig *coins.Coin) (*hdkeychain.ExtendedKey, error) {
	if coinConfig.Mnemonic == "" {
		return nil, errors.New("the coin is not available")
	}
	seed := bip39.NewSeed(coinConfig.Mnemonic, os.Getenv("MNEMONIC_PASSWORD"))
	mKey, err := hdkeychain.NewMaster(seed, coinConfig.NetParams)
	if err != nil {
		return nil, err
	}
	purposeChild, err := mKey.Child(hdkeychain.HardenedKeyStart + 44)
	if err != nil {
		return nil, err
	}
	coinType, err := purposeChild.Child(hdkeychain.HardenedKeyStart + coinConfig.NetParams.HDCoinType)
	if err != nil {
		return nil, err
	}
	accChild, err := coinType.Child(hdkeychain.HardenedKeyStart + 0)
	if err != nil {
		return nil, err
	}
	return accChild, nil
}

func getPubKeyHashFromPath(acc *hdkeychain.ExtendedKey, coinConfig *coins.Coin, path uint32) (string, error) {
	directExtended, err := acc.Child(0)
	if err != nil {
		return "", err
	}
	addrExtPub, err := directExtended.Child(path)
	if err != nil {
		return "", err
	}
	addr, err := addrExtPub.Address(coinConfig.NetParams)
	if err != nil {
		return "", err
	}
	return addr.String(), nil
}

func getPrivKeyFromPath(coinConfig *coins.Coin, path uint32) (*btcec.PrivateKey, error) {
	if coinConfig.Mnemonic == "" {
		return nil, errors.New("the coin is not available")
	}
	seed := bip39.NewSeed(coinConfig.Mnemonic, os.Getenv("MNEMONIC_PASSWORD"))
	mKey, err := hdkeychain.NewMaster(seed, coinConfig.NetParams)
	if err != nil {
		return nil, err
	}
	purposeChild, err := mKey.Child(hdkeychain.HardenedKeyStart + 44)
	if err != nil {
		return nil, err
	}
	coinType, err := purposeChild.Child(hdkeychain.HardenedKeyStart + coinConfig.NetParams.HDCoinType)
	if err != nil {
		return nil, err
	}
	accChild, err := coinType.Child(hdkeychain.HardenedKeyStart + 0)
	if err != nil {
		return nil, err
	}
	directChild, err := accChild.Child(0)
	if err != nil {
		return nil, err
	}
	privKeyChild, err := directChild.Child(path)
	if err != nil {
		return nil, err
	}
	return privKeyChild.ECPrivKey()
}

func NewPlutusController() *Controller {
	ctrl := &Controller{
		Address: make(map[string]AddrInfo),
	}
	// Here we handle only active coins
	for _, coin := range coinfactory.Coins {
		coinConf, err := coinfactory.GetCoin(coin.Tag)
		if err != nil {
			panic(err)
		}
		if coin.Tag == "POLIS" {
			err = chaincfg.Register(coinConf.NetParams)
			if err != nil {
				panic(err)
			}
		}
		if !coin.Token && coin.Tag != "ETH" {
			err := ctrl.getAddrs(coinConf)
			if err != nil {
				panic(err)
			}
		}
	}
	return ctrl
}

package controllers

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eabz/btcutil"
	"github.com/eabz/btcutil/chaincfg"
	"github.com/eabz/btcutil/hdkeychain"
	"github.com/eabz/btcutil/txscript"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/martian/log"
	"github.com/grupokindynos/common/blockbook"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/grupokindynos/common/plutus"
	"github.com/grupokindynos/plutus/models"
	"github.com/martinboehm/btcd/btcec"
	"github.com/martinboehm/btcd/chaincfg/chainhash"
	"github.com/martinboehm/btcd/wire"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/tyler-smith/go-bip39"
	"math/big"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Params struct {
	Coin string
	Body []byte
	Txid string
}

//var ethAccount = "0x363cf89578DcC1F820C636161C4dD7435e111108"
var ethAccount = "0xC97316DBa300DcFe93A261B9481058C01e07C970"

var myClient = &http.Client{Timeout: 10 * time.Second}

const addrGap = 20

type AddrInfo struct {
	LastUsed int
	AddrInfo []models.AddrInfo
}

type Controller struct {
	Address map[string]AddrInfo
}

type GasStation struct {
	Fast        float64 `json:"fast"`
	Fastest     float64 `json:"fastest"`
	SafeLow     float64 `json:"safeLow"`
	Average     float64 `json:"average"`
	SafeLowWait float64 `json:"safeLowWait"`
	AvgWait     float64 `json:"avgWait"`
	FastWait    float64 `json:"fastWait"`
	FastestWait float64 `json:"fastestWait"`
}

//**for testing only
type NestedElement struct {
	Address string  `json:"address"`
	Coin    string  `json:"coin"`
	Amount  float64 `json:"amount"`
}
type TestJ struct {
	Coin          string `json:"coin"`
	NestedElement `json:"body"`
	Txid          string `json:"txid"`
}

//**end of testing block

func (c *Controller) GetBalance(params Params) (interface{}, error) {
	coinConfig, err := coinfactory.GetCoin(params.Coin)
	if err != nil {
		return nil, err
	}
	if !coinConfig.Info.Token && coinConfig.Info.Tag != "ETH" {
		blockBookWrap := blockbook.NewBlockBookWrapper(coinConfig.Info.Blockbook)
		acc, err := getAccFromMnemonic(coinConfig, false)
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
		blockBookWrap := blockbook.NewBlockBookWrapper(ethConfig.Info.Blockbook)
		info, err := blockBookWrap.GetEthAddress(ethAccount)
		if err != nil {
			return nil, err
		}
		if coinConfig.Info.Token {
			var tokenInfo *blockbook.EthTokens
			for _, token := range info.Tokens {
				if coinConfig.Info.Contract == token.Contract {
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
	if coinConfig.Info.Token || coinConfig.Info.Tag == "ETH" {
		return ethAccount, nil
	}
	acc, err := getAccFromMnemonic(coinConfig, false)
	if err != nil {
		return nil, err
	}
	// Create a new xpub and derive the address from the hdwallet
	directExtended, err := acc.Child(0)
	if err != nil {
		return nil, err
	}
	addrExtPub, err := directExtended.Child(uint32(c.Address[coinConfig.Info.Tag].LastUsed + 1))
	if err != nil {
		return nil, err
	}
	addr, err := addrExtPub.Address(coinConfig.NetParams)
	if err != nil {
		return nil, err
	}
	newAddrInfo := AddrInfo{
		LastUsed: c.Address[coinConfig.Info.Tag].LastUsed + 1,
		AddrInfo: c.Address[coinConfig.Info.Tag].AddrInfo,
	}
	newAddrInfo.AddrInfo = append(newAddrInfo.AddrInfo, models.AddrInfo{
		Addr: addr.String(), Path: c.Address[coinConfig.Info.Tag].LastUsed + 1,
	})
	c.Address[coinConfig.Info.Tag] = newAddrInfo
	return addr.String(), nil
}

func (c *Controller) SendToAddress(params Params) (interface{}, error) {
	var SendToAddressData plutus.SendAddressBodyReq
	err := json.Unmarshal(params.Body, &SendToAddressData)
	if err != nil {
		return nil, err
	}
	//**only for testing block
	var bobody TestJ
	_ = json.Unmarshal(params.Body, &bobody)
	SendToAddressData.Amount = bobody.Amount
	SendToAddressData.Address = bobody.Address
	//**end of testing block
	coinConfig, err := coinfactory.GetCoin(SendToAddressData.Coin)
	if err != nil {
		return "", err
	}
	var txid string
	if coinConfig.Info.Token || coinConfig.Info.Tag == "ETH" {
		txid, err = c.sendToAddressEth(SendToAddressData, coinConfig)
		if err != nil {
			return nil, err
		}
	} else {
		txid, err = c.sendToAddress(SendToAddressData, coinConfig)
		if err != nil {
			return nil, err
		}
	}
	return txid, nil
}

func (c *Controller) sendToAddress(SendToAddressData plutus.SendAddressBodyReq, coinConfig *coins.Coin) (string, error) {
	value, err := btcutil.NewAmount(SendToAddressData.Amount)
	if err != nil {
		return "", err
	}
	acc, err := getAccFromMnemonic(coinConfig, true)
	if err != nil {
		return "", err
	}
	accPub, err := acc.Neuter()
	if err != nil {
		return "", err
	}
	blockBookWrap := blockbook.NewBlockBookWrapper(coinConfig.Info.Blockbook)
	utxos, err := blockBookWrap.GetUtxo(accPub.String(), false)
	if err != nil {
		return "", err
	}
	if len(utxos) == 0 {
		return "", errors.New("no balance available")
	}
	var Tx wire.MsgTx
	var txVersion int32
	if coinConfig.Info.Tag == "POLIS" || coinConfig.Info.Tag == "DASH" || coinConfig.Info.Tag == "GRS" {
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
			return "", err
		}
		utxoAmount := btcutil.Amount(intValue)
		availableAmount += utxoAmount
		txidHash, err := chainhash.NewHashFromStr(utxo.Txid)
		if err != nil {
			return "", err
		}
		prevOut := wire.NewOutPoint(txidHash, uint32(utxo.Vout))
		in := wire.NewTxIn(prevOut, nil, nil)
		Tx.AddTxIn(in)
	}
	// Retrieve information for outputs
	payAddr, err := btcutil.DecodeAddress(SendToAddressData.Address, coinConfig.NetParams)
	if err != nil {
		return "", err
	}
	changeAddr, err := btcutil.DecodeAddress(changeAddrPubKeyHash, coinConfig.NetParams)
	if err != nil {
		return "", err
	}
	pkScriptPay, err := txscript.PayToAddrScript(payAddr)
	if err != nil {
		return "", err
	}
	pkScriptChange, err := txscript.PayToAddrScript(changeAddr)
	if err != nil {
		return "", err
	}
	txOut := &wire.TxOut{
		Value:    int64(value.ToUnit(btcutil.AmountSatoshi)),
		PkScript: pkScriptPay,
	}
	var fee blockbook.Fee
	if SendToAddressData.Coin == "BTC" {
		fee, err = blockBookWrap.GetFee("4")
		if err != nil {
			return "", err
		}
	} else {
		fee, err = blockBookWrap.GetFee("2")
		if err != nil {
			return "", err
		}
	}
	var feeRate int64
	if fee.Result == "-1" || fee.Result == "0" {
		feeRate = 4000
	} else {
		feeParse, err := strconv.ParseFloat(fee.Result, 64)
		if err != nil {
			return "", err
		}
		feeRate = int64(feeParse * 1e8)
	}
	txSize := (len(Tx.TxIn) * 180) + (len(Tx.TxOut) * 34) + 124
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
		path := strings.Split(utxo.Path, "/")
		pathParse, err := strconv.ParseInt(path[5], 10, 64)
		if err != nil {
			return "", err
		}
		privKey, err := getPrivKeyFromPath(acc, uint32(pathParse))
		if err != nil {
			return "", err
		}
		addr, err := btcutil.DecodeAddress(utxo.Address, coinConfig.NetParams)
		if err != nil {
			return "", err
		}
		subscript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return "", err
		}
		var sigHash txscript.SigHashHasher
		if coinConfig.Info.Tag == "GRS" {
			sigHash = txscript.Sha256
		} else {
			sigHash = txscript.Sha256d
		}
		sigScript, err := txscript.SignatureScript(&Tx, i, subscript, txscript.SigHashAll, privKey, true, sigHash)
		if err != nil {
			return "", err
		}
		Tx.TxIn[i].SignatureScript = sigScript
	}
	buf := bytes.NewBuffer([]byte{})
	err = Tx.BtcEncode(buf, 0, wire.BaseEncoding)
	if err != nil {
		return "", err
	}
	rawTx := hex.EncodeToString(buf.Bytes())
	return blockBookWrap.SendTx(rawTx)
}

func (c *Controller) sendToAddressEth(SendToAddressData plutus.SendAddressBodyReq, coinConfig *coins.Coin) (string, error) {
	//**generate a valid tx amount
	value := big.NewInt(int64(SendToAddressData.Amount * 1000000000000000000))

	//**get the senders address, public key and private key?
	//acc, err := getAccFromMnemonic(coinConfig, true)
	//if err != nil {
	//	return "", err
	//}
	mnemonic := "roof stable huge chuckle where else sniff apology museum maze parade delay"
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return "", err
	}
	path := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := wallet.Derive(path, true)
	if err != nil {
		return "", err
	}
	ethAccount := account.Address.Hex()

	blockBookWrap := blockbook.NewBlockBookWrapper(coinConfig.Info.Blockbook)

	//** get the balance, check if its > 0
	//if len(utxos) == 0 {
	//	return "", errors.New("no balance available")
	//}
	info, err := blockBookWrap.GetEthAddress(ethAccount)
	if err != nil {
		return "", err
	}
	balance, err := strconv.ParseFloat(info.Balance, 64)
	if err != nil {
		return "", err
	}
	balance = balance / 1e18
	if balance == 0 {
		return "", errors.New("no balance available")
	}
	fmt.Println(balance)
	nonce, err := strconv.ParseUint(info.Nonce, 0, 64)
	if err != nil {
		return "", err
	}
	fmt.Println(nonce)

	//** Retrieve information for outputs: out adrdress
	//payAddr, err := btcutil.DecodeAddress(SendToAddressData.Address, coinConfig.NetParams)
	//if err != nil {
	//	return "", err
	//}
	toAddress := common.HexToAddress(SendToAddressData.Address)

	//**calculate fee/gas cost, add the amount
	gasLimit := uint64(21000)
	gasStation := GasStation{}
	_ = getJson("https://ethgasstation.info/json/ethgasAPI.json", &gasStation)
	fmt.Println(gasStation)
	gasPrice := big.NewInt(int64(1000000000 * (gasStation.Average / 10)))
	fmt.Println(gasPrice)
	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	// **sign and send
	signedTx, err := wallet.SignTx(account, tx, nil)
	if err != nil {
		return "", err
	}
	ts := types.Transactions{signedTx}
	rawTxBytes := ts.GetRlp(0)
	rawTxHex := hex.EncodeToString(rawTxBytes)

	return blockBookWrap.SendTx("0x" + rawTxHex)
}

func getJson(url string, target interface{}) error {
	r, err := myClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
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
	if coinConfig.Info.Token || coinConfig.Info.Tag == "ETH" {
		return reflect.DeepEqual(ValidateAddressData.Address, ethAccount), nil
	}
	var isMine bool
	for _, addr := range c.Address[coinConfig.Info.Tag].AddrInfo {
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
		for _, addr := range c.Address[coinConfig.Info.Tag].AddrInfo {
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
	acc, err := getAccFromMnemonic(coinConfig, false)
	if err != nil {
		return err
	}
	blockBookWrap := blockbook.NewBlockBookWrapper(coinConfig.Info.Blockbook)
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
	c.Address[coinConfig.Info.Tag] = AddrInfo{
		LastUsed: info.UsedTokens,
		AddrInfo: addrInfoSlice,
	}
	return nil
}

func getAccFromMnemonic(coinConfig *coins.Coin, priv bool) (*hdkeychain.ExtendedKey, error) {
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
	if priv {
		return accChild, nil
	} else {
		return accChild.Neuter()
	}
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

func getPrivKeyFromPath(acc *hdkeychain.ExtendedKey, path uint32) (*btcec.PrivateKey, error) {
	directExtended, err := acc.Child(0)
	if err != nil {
		return nil, err
	}
	accPath, err := directExtended.Child(path)
	if err != nil {
		return nil, err
	}
	return accPath.ECPrivKey()
}

func NewPlutusController() *Controller {
	ctrl := &Controller{
		Address: make(map[string]AddrInfo),
	}
	// Here we handle only active coins
	var i uint32
	for _, coin := range coinfactory.Coins {
		i += 1
		coinConf, err := coinfactory.GetCoin(coin.Info.Tag)
		if err != nil {
			panic(err)
		}
		if !coin.Info.Token && coin.Info.Tag != "ETH" {
			coin.NetParams.Net = wire.BitcoinNet(i)
			coin.NetParams.AddressMagicLen = 1
			registered := chaincfg.IsRegistered(coin.NetParams)
			if !registered {
				err := chaincfg.Register(coin.NetParams)
				if err != nil {
					panic(err)
				}
			}
			err := ctrl.getAddrs(coinConf)
			if err != nil {
				log.Infof("Error: %v, Coin: %v", err.Error(), coin.Info.Name)
			}
		}
	}
	return ctrl
}

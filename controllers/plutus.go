package controllers

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"math/big"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/eabz/btcutil"
	"github.com/eabz/btcutil/chaincfg"
	"github.com/eabz/btcutil/hdkeychain"
	"github.com/eabz/btcutil/txscript"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
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
	"golang.org/x/crypto/sha3"
)

type Params struct {
	Coin string
	Body []byte
	Txid string
}

var ethWallet *hdwallet.Wallet

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
		acc, err := getEthAccFromMnemonic(ethConfig, false)
		if err != nil {
			return nil, err
		}
		blockBookWrap := blockbook.NewBlockBookWrapper(ethConfig.Info.Blockbook)
		info, err := blockBookWrap.GetEthAddress(acc.Address.Hex())
		if err != nil {
			return nil, err
		}
		if coinConfig.Info.Tag != "ETH" {
			tokenInfo := ercDetails(info, coinConfig.Info.Contract)
			if tokenInfo == nil {
				response := plutus.Balance{
					Confirmed: 0,
				}
				return response, nil
			}
			balance, err := strconv.ParseFloat(tokenInfo.Balance, 64)
			balance = balance / (math.Pow(10, float64(tokenInfo.Decimals)))
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

func ercDetails(info blockbook.EthAddr, contract string) *blockbook.EthTokens {
	var tokenInfo *blockbook.EthTokens
	for _, token := range info.Tokens {
		if common.HexToAddress(contract) == common.HexToAddress(token.Contract) {
			tokenInfo = &token
			break
		}
	}
	return tokenInfo
}

func (c *Controller) GetAddress(params Params) (interface{}, error) {
	coinConfig, err := coinfactory.GetCoin(params.Coin)
	if err != nil {
		return nil, err
	}
	if coinConfig.Info.Token || coinConfig.Info.Tag == "ETH" {
		var acc accounts.Account
		if coinConfig.Mnemonic == "" {
			ethConfig, err := coinfactory.GetCoin("ETH")
			if err != nil {
				return nil, err
			}
			acc, err = getEthAccFromMnemonic(ethConfig, false)
			if err != nil {
				return nil, err
			}
		} else {
			acc, err = getEthAccFromMnemonic(coinConfig, false)
			if err != nil {
				return nil, err
			}
		}
		return acc.Address.Hex(), nil
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
	// To prevent address collision we need to de-register all networks and register just the network using
	chaincfg.ResetParams()
	chaincfg.Register(coinConfig.NetParams)
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
	if fee.Result == "-1" || fee.Result == "0" || fee.Result == "" {
		feeRate = 4000
	} else {
		feeParse, err := strconv.ParseFloat(fee.Result, 64)
		if err != nil {
			return "", err
		}
		feeRate = int64(feeParse * 1e8)
	}
	txSize := (len(Tx.TxIn) * 180) + (len(Tx.TxOut) * 34) + 124
	feeSats := float64(feeRate) / 1024.0 * float64(txSize)
	payingFee := btcutil.Amount(int64(feeSats))
	if availableAmount-payingFee-value > 0 {
		txOutChange := &wire.TxOut{
			Value:    int64(((availableAmount - value) - payingFee).ToUnit(btcutil.AmountSatoshi)),
			PkScript: pkScriptChange,
		}

		Tx.AddTxOut(txOutChange)
	}
	Tx.AddTxOut(txOut)
	Tx.Version = txVersion
	// To prevent address collision we need to de-register all networks and register just the network using
	chaincfg.ResetParams()
	chaincfg.Register(coinConfig.NetParams)

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
	// using the ethereum account to hl the tokens
	ethConfig, err := coinfactory.GetCoin("ETH")
	if err != nil {
		return "", err
	}
	//**get the account that holds the private keys and addresses
	account, err := getEthAccFromMnemonic(ethConfig, true)
	if err != nil {
		return "", err
	}
	ethAccount := account.Address.Hex()

	blockBookWrap := blockbook.NewBlockBookWrapper(ethConfig.Info.Blockbook)

	//** get the balance, check if its > 0 or less than the amount
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
		return "", errors.New("no eth available")
	}
	decimals := 18
	if coinConfig.Info.Token && coinConfig.Info.Tag != "ETH" {
		// check balance of token
		tokenInfo := ercDetails(info, coinConfig.Info.Contract)
		if tokenInfo == nil {
			return "", errors.New("no token balance available")
		}
		tokenBalance, err := strconv.ParseFloat(tokenInfo.Balance, 64)
		if err != nil {
			return "", err
		}
		tokenBalance = tokenBalance / (math.Pow(10, float64(tokenInfo.Decimals)))
		if tokenBalance == 0 || tokenBalance < SendToAddressData.Amount {
			return "", errors.New("not enough token available")
		}
		decimals = tokenInfo.Decimals
	} else {
		if balance < SendToAddressData.Amount {
			return "", errors.New("not enough balance")
		}
	}
	// get the nonce
	nonce, err := strconv.ParseUint(info.Nonce, 0, 64)
	if err != nil {
		return "", errors.New("nonce failed")
	}

	//** Retrieve information for outputs: out address
	toAddress := common.HexToAddress(SendToAddressData.Address)
	//**calculate fee/gas cost
	gasLimit := uint64(21000)
	if coinConfig.Info.Tag != "ETH" {
		gasLimit = uint64(200000)
	}
	gasStation := GasStation{}
	err = getJSON("https://ethgasstation.info/json/ethgasAPI.json", &gasStation)
	if err != nil {
		return "", errors.New("could not retrieve the gas price")
	}
	gasPrice := big.NewInt(int64(1000000000 * (gasStation.Average / 10))) //(10^9*(gweiValue/10))
	var data []byte
	var tx *types.Transaction

	if coinConfig.Info.Token && coinConfig.Info.Tag != "ETH" {
		// the additional data for the token transaction
		tokenAddress := common.HexToAddress(coinConfig.Info.Contract)

		transferFnSignature := []byte("transfer(address,uint256)")
		hash := sha3.NewLegacyKeccak256()
		hash.Write(transferFnSignature)
		methodID := hash.Sum(nil)[:4]

		paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)

		val := new(big.Float)
		pot := new(big.Float)
		val.SetFloat64(SendToAddressData.Amount)
		pot.SetFloat64(math.Pow10(decimals))
		val.Mul(val, pot)
		amount, _ := val.Int(nil)
		paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)

		data = append(data, methodID...)
		data = append(data, paddedAddress...)
		data = append(data, paddedAmount...)
		value := big.NewInt(0) // in wei (0 eth)
		tx = types.NewTransaction(nonce, tokenAddress, value, gasLimit, gasPrice, data)
	} else {
		value := big.NewInt(int64(SendToAddressData.Amount * 1000000000000000000)) // the amount in wei
		tx = types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)
	}
	// **sign and send
	signedTx, err := signEthTx(ethConfig, account, tx, nil)
	if err != nil {
		return "", errors.New("failed to sign transaction")
	}
	ts := types.Transactions{signedTx}
	rawTxBytes := ts.GetRlp(0)
	rawTxHex := hex.EncodeToString(rawTxBytes)
	return blockBookWrap.SendTx("0x" + rawTxHex)
	//return "", nil
}

func getJSON(url string, target interface{}) error {
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
		coinConfig, err = coinfactory.GetCoin("ETH")
		if err != nil {
			return nil, err
		}
		acc, err := getEthAccFromMnemonic(coinConfig, false)
		if err != nil {
			return nil, err
		}
		return reflect.DeepEqual(ValidateAddressData.Address, acc.Address.Hex()), nil
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
	coinConfig, err := coinfactory.GetCoin(ValidateTxData.Coin)
	if err != nil {
		return nil, err
	}

	var isValue, isAddress bool

	//ethereum-like coins (and ERC20)
	if coinConfig.Info.Token || coinConfig.Info.Tag == "ETH" {
		value := ValidateTxData.Amount
		var tx *types.Transaction
		if ValidateTxData.RawTx[0:2] == "0x" {
			ValidateTxData.RawTx = ValidateTxData.RawTx[2:]
		}
		rawtx, err := hex.DecodeString(ValidateTxData.RawTx)
		if err != nil {
			return nil, err
		}
		err = rlp.DecodeBytes(rawtx, &tx)
		if err != nil {
			return nil, err
		}
		//compare amount from the tx and the input body
		var txBodyAmount int64
		var txAddr common.Address
		if coinConfig.Info.Token && coinConfig.Info.Tag != "ETH" {
			address, amount := DecodeERC20Data([]byte(hex.EncodeToString(tx.Data())))
			txAddr = common.HexToAddress(string(address))
			txBodyAmount = amount.Int64()
		} else {
			txBodyAmount = tx.Value().Int64()
			txAddr = *tx.To()
		}
		if txBodyAmount == value {
			isValue = true
		}
		bodyAddr := common.HexToAddress(ValidateTxData.Address)
		//compare the address from the tx and the input body
		if bytes.Equal(bodyAddr.Bytes(), txAddr.Bytes()) {
			isAddress = true
		}

	} else {
		//bitcoin-like coins
		value := btcutil.Amount(ValidateTxData.Amount)

		rawTxBytes, err := hex.DecodeString(ValidateTxData.RawTx)
		if err != nil {
			return nil, err
		}
		tx, err := btcutil.NewTxFromBytes(rawTxBytes)
		if err != nil {
			return nil, err
		}
		// To prevent address collision we need to de-register all networks and register just the network using
		chaincfg.ResetParams()
		chaincfg.Register(coinConfig.NetParams)
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
	chaincfg.ResetParams()
	chaincfg.Register(coinConfig.NetParams)
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
	}
	return accChild.Neuter()
}

func getEthAccFromMnemonic(coinConfig *coins.Coin, saveWallet bool) (accounts.Account, error) {
	if coinConfig.Mnemonic == "" {
		return accounts.Account{}, errors.New("the coin is not available")
	}
	wallet, err := hdwallet.NewFromMnemonic(coinConfig.Mnemonic)
	if err != nil {
		return accounts.Account{}, err
	}
	// standard for eth wallets like Metamask
	path := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := wallet.Derive(path, true)
	if err != nil {
		return accounts.Account{}, err
	}
	if saveWallet {
		ethWallet = wallet
	}
	return account, nil
}

func signEthTx(coinConfig *coins.Coin, account accounts.Account, tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	if coinConfig.Mnemonic == "" {
		return nil, errors.New("the coin is not available")
	}
	signedTx, err := ethWallet.SignTx(account, tx, chainID)
	if err != nil {
		return nil, err
	}
	return signedTx, nil
}

func DecodeERC20Data(b []byte) ([]byte, *big.Int) {
	to := b[32:72]
	tokens := b[74:136]
	hexed, _ := hex.DecodeString(string(tokens))
	amount := big.NewInt(0)
	amount.SetBytes(hexed)
	return to, amount
}

func decimalToToken(decimalAmount float64, decimals int) *big.Int {
	val := new(big.Float)
	pot := new(big.Float)
	val.SetFloat64(decimalAmount)
	pot.SetFloat64(math.Pow10(decimals))
	val.Mul(val, pot)
	amount, _ := val.Int(nil)
	return amount
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
	for _, coin := range coinfactory.Coins {
		if !coin.Info.Token && coin.Info.Tag != "ETH" {
			err := ctrl.getAddrs(coin)
			if err != nil {
				log.Infof("Error: %v, Coin: %v", err.Error(), coin.Info.Name)
			}
		}
	}
	return ctrl
}

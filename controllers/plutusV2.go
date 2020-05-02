package controllers

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/eabz/btcutil"
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
	"github.com/martinboehm/btcd/chaincfg/chainhash"
	"github.com/martinboehm/btcd/wire"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"golang.org/x/crypto/sha3"
	"math"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type ParamsV2 struct {
	Coin    string
	Body    []byte
	Txid    string
	Service string
}

type ControllerV2 struct {
	Address map[string]AddrInfo
}

var ethWalletV2 *hdwallet.Wallet

const coinV2 = "ETHV2"

func (c *ControllerV2) GetBalanceV2(params ParamsV2) (interface{}, error) {
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
		if params.Service == "tyche" || params.Service == "ladon" {
			ethConfig.Mnemonic = os.Getenv("MNEMONIC_" + coinV2)
		}
		acc, err := getEthAccFromMnemonicV2(ethConfig, false)
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

func (c *ControllerV2) GetAddressV2(params ParamsV2) (interface{}, error) {
	coinConfig, err := coinfactory.GetCoin(params.Coin)
	if err != nil {
		return nil, err
	}
	if coinConfig.Info.Token || coinConfig.Info.Tag == "ETH" {
		ethConfig, err := coinfactory.GetCoin("ETH")
		if err != nil {
			return nil, err
		}
		if params.Service == "tyche" || params.Service == "ladon" {
			ethConfig.Mnemonic = os.Getenv("MNEMONIC_" + coinV2)
		}
		var acc accounts.Account
		acc, err = getEthAccFromMnemonicV2(ethConfig, false)
		if err != nil {
			return nil, err
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

func (c *ControllerV2) SendToAddressV2(params ParamsV2) (interface{}, error) {
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
		txid, err = c.sendToAddressEthV2(SendToAddressData, coinConfig, params.Service)
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

func (c *ControllerV2) sendToAddress(SendToAddressData plutus.SendAddressBodyReq, coinConfig *coins.Coin) (string, error) {
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

func (c *ControllerV2) sendToAddressEthV2(SendToAddressData plutus.SendAddressBodyReq, coinConfig *coins.Coin, service string) (string, error) {
	// using the ethereum account to hl the tokens
	ethConfig, err := coinfactory.GetCoin("ETH")
	if err != nil {
		return "", err
	}
	//**get the account that holds the private keys and addresses
	if service == "tyche" || service == "ladon" {
		ethConfig.Mnemonic = os.Getenv("MNEMONIC_" + coinV2)
	}
	account, err := getEthAccFromMnemonicV2(ethConfig, true)
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
	var gasStation GasStation
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
	signedTx, err := signEthTxV2(ethConfig, account, tx, nil)
	if err != nil {
		return "", errors.New("failed to sign transaction")
	}
	ts := types.Transactions{signedTx}
	rawTxBytes := ts.GetRlp(0)
	rawTxHex := hex.EncodeToString(rawTxBytes)
	return blockBookWrap.SendTx("0x" + rawTxHex)
	//return "", nil
}

func (c *ControllerV2) ValidateAddressV2(params ParamsV2) (interface{}, error) {
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
		if params.Service == "tyche" || params.Service == "ladon" {
			coinConfig.Mnemonic = os.Getenv("MNEMONIC_" + coinV2)
		}
		acc, err := getEthAccFromMnemonicV2(coinConfig, false)
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

func (c *ControllerV2) ValidateRawTxV2(params ParamsV2) (interface{}, error) {
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
		if (params.Service == "ladon" || params.Service == "tyche") && coinConfig.Info.Tag != "ETH" {
			//remove the 1e8
			valueNoSatoshi := float64(value) / 1e8
			value = decimalToToken(valueNoSatoshi, coinConfig.Info.Decimals).Int64()
		} else if (params.Service == "ladon" || params.Service == "tyche") && coinConfig.Info.Tag == "ETH" {
			value = value * 1e10
		}
		var tx *types.Transaction
		if ValidateTxData.RawTx[0:2] == "0x" {
			ValidateTxData.RawTx = ValidateTxData.RawTx[2:]
		}
		//if ValidateTxData.RawTx[]
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

func getEthAccFromMnemonicV2(coinConfig *coins.Coin, saveWallet bool) (accounts.Account, error) {
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
		ethWalletV2 = wallet
	}
	return account, nil
}

func signEthTxV2(coinConfig *coins.Coin, account accounts.Account, tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	if coinConfig.Mnemonic == "" {
		return nil, errors.New("the coin is not available")
	}
	signedTx, err := ethWalletV2.SignTx(account, tx, chainID)
	if err != nil {
		return nil, err
	}
	return signedTx, nil
}

func (c *ControllerV2) getAddrs(coinConfig *coins.Coin) error {
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

func NewPlutusControllerV2() *ControllerV2 {
	ctrl := &ControllerV2{
		Address: make(map[string]AddrInfo),
	}
	for _, coin := range coinfactory.Coins {
		coinConf, err := coinfactory.GetCoin(coin.Info.Tag)
		if err != nil {
			panic(err)
		}
		if !coin.Info.Token && coin.Info.Tag != "ETH" {
			err = ctrl.getAddrs(coinConf)
			if err != nil {
				log.Infof("Error: %v, Coin: %v", err.Error(), coin.Info.Name)
			}
		}
	}
	return ctrl
}

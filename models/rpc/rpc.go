package rpc

type GetBlockchainInfo struct {
	Chain                string  `json:"chain"`
	Blocks               int     `json:"blocks"`
	Headers              int     `json:"headers"`
	Bestblockhash        string  `json:"bestblockhash"`
	Difficulty           float64 `json:"difficulty"`
	Mediantime           int     `json:"mediantime"`
	Verificationprogress float64 `json:"verificationprogress"`
	Chainwork            string  `json:"chainwork"`
	Pruned               bool    `json:"pruned"`
	Softforks            []struct {
		ID      string `json:"id"`
		Version int    `json:"version"`
		Reject  struct {
			Status bool `json:"status"`
		} `json:"reject"`
	} `json:"softforks"`
	Bip9Softforks struct {
		Csv struct {
			Status    string `json:"status"`
			StartTime int    `json:"startTime"`
			Timeout   int    `json:"timeout"`
			Since     int    `json:"since"`
		} `json:"csv"`
		Dip0001 struct {
			Status    string `json:"status"`
			StartTime int    `json:"startTime"`
			Timeout   int    `json:"timeout"`
			Since     int    `json:"since"`
		} `json:"dip0001"`
		Bip147 struct {
			Status    string `json:"status"`
			StartTime int    `json:"startTime"`
			Timeout   int    `json:"timeout"`
			Since     int    `json:"since"`
		} `json:"bip147"`
	} `json:"bip9_softforks"`
}

type GetWalletInfo struct {
	Walletversion      int     `json:"walletversion"`
	Balance            float64 `json:"balance"`
	UnconfirmedBalance float64 `json:"unconfirmed_balance"`
	ImmatureBalance    float64 `json:"immature_balance"`
	Txcount            int     `json:"txcount"`
	Keypoololdest      int     `json:"keypoololdest"`
	Keypoolsize        int     `json:"keypoolsize"`
	KeysLeft           int     `json:"keys_left"`
	Paytxfee           float64 `json:"paytxfee"`
}

type GetNetworkInfo struct {
	Version         int    `json:"version"`
	Subversion      string `json:"subversion"`
	Protocolversion int    `json:"protocolversion"`
	Localservices   string `json:"localservices"`
	Localrelay      bool   `json:"localrelay"`
	Timeoffset      int    `json:"timeoffset"`
	Networkactive   bool   `json:"networkactive"`
	Connections     int    `json:"connections"`
	Networks        []struct {
		Name                      string `json:"name"`
		Limited                   bool   `json:"limited"`
		Reachable                 bool   `json:"reachable"`
		Proxy                     string `json:"proxy"`
		ProxyRandomizeCredentials bool   `json:"proxy_randomize_credentials"`
	} `json:"networks"`
	Relayfee       float64 `json:"relayfee"`
	Incrementalfee float64 `json:"incrementalfee"`
	Localaddresses []struct {
		Address string `json:"address"`
		Port    int    `json:"port"`
		Score   int    `json:"score"`
	} `json:"localaddresses"`
	Warnings string `json:"warnings"`
}

type ValidateAddress struct {
	Address      string `json:"address"`
	ScriptPubKey string `json:"scriptPubKey"`
	Ismine       bool   `json:"ismine"`
	Solvable     bool   `json:"solvable"`
	Desc         string `json:"desc"`
	Iswatchonly  bool   `json:"iswatchonly"`
	Isscript     bool   `json:"isscript"`
	Iswitness    bool   `json:"iswitness"`
	Script       string `json:"script"`
	Hex          string `json:"hex"`
	Pubkey       string `json:"pubkey"`
	Embedded     struct {
		Isscript       bool   `json:"isscript"`
		Iswitness      bool   `json:"iswitness"`
		WitnessVersion int    `json:"witness_version"`
		WitnessProgram string `json:"witness_program"`
		Pubkey         string `json:"pubkey"`
		Address        string `json:"address"`
		ScriptPubKey   string `json:"scriptPubKey"`
	} `json:"embedded"`
	Label               string `json:"label"`
	Ischange            bool   `json:"ischange"`
	Timestamp           int    `json:"timestamp"`
	Hdkeypath           string `json:"hdkeypath"`
	Hdseedid            string `json:"hdseedid"`
	Hdmasterfingerprint string `json:"hdmasterfingerprint"`
	Labels              []struct {
		Name    string `json:"name"`
		Purpose string `json:"purpose"`
	} `json:"labels"`
}

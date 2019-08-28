package responses

type Balance struct {
	Confirmed   float64 `json:"confirmed"`
	Unconfirmed float64 `json:"unconfirmed"`
}

type Info struct {
	Blocks      int    `json:"blocks"`
	Headers     int    `json:"headers"`
	Chain       string `json:"chain"`
	Protocol    int    `json:"protocol"`
	Version     int    `json:"version"`
	Subversion  string `json:"subversion"`
	Connections int    `json:"connections"`
}

type Status struct {
	NodeBlocks      int  `json:"node_blocks"`
	NodeHeaders     int  `json:"node_headers"`
	ExternalBlocks  int  `json:"external_blocks"`
	ExternalHeaders int  `json:"external_headers"`
	Synced          bool `json:"synced"`
}

package zkchainconfig

import "github.com/ledgerwatch/erigon/params/networkname"

const XlayerTestnetChainId = 195
const XlayerMainnetChainId = 196
const HermezMainnetChainId = 1101
const HermezBaliChainId = 2440
const HermezCardonaChainId = 2442
const HermezEtrogChainId = 10010
const HermezLocalDevnetChainId = 999999
const HermezEsTestChainId = 123

var chainIds = []uint64{
	XlayerTestnetChainId,     // xlayer-testnet
	XlayerMainnetChainId,     // xlayer-mainet
	HermezMainnetChainId,     // mainnet
	HermezBaliChainId,        // cardona internal
	HermezCardonaChainId,     // cardona
	HermezEtrogChainId,       //etrog testnet
	HermezLocalDevnetChainId, // local devnet
	HermezEsTestChainId,      // estestnet
}

var chainIdToName = map[uint64]string{
	XlayerTestnetChainId:     networkname.XLayerTestnetChainName,
	XlayerMainnetChainId:     networkname.XLayerMainnetChainName,
	HermezMainnetChainId:     networkname.HermezMainnetChainName,
	HermezBaliChainId:        networkname.HermezBaliChainName,
	HermezCardonaChainId:     networkname.HermezCardonaChainName,
	HermezEtrogChainId:       networkname.HermezEtrogChainName,
	HermezLocalDevnetChainId: networkname.HermezLocalDevnetChainName,
	HermezEsTestChainId:      networkname.HermezESTestChainName,
}

func IsZk(chainId uint64) bool {
	for _, validId := range chainIds {
		if chainId == validId {
			return true
		}
	}
	return false
}

func IsXLayerTestnetChain(chainId uint64) bool {
	return chainId == XlayerTestnetChainId
}

func GetChainName(chainId uint64) string {
	return chainIdToName[chainId]
}

func IsTestnet(chainId uint64) bool {
	return chainId == 1442
}

func IsDevnet(chainId uint64) bool {
	return chainId == 1440
}

func CheckForkOrder() error {
	return nil
}

func SetDynamicChainDetails(chainId uint64, chainName string) {
	chainIdToName[chainId] = chainName
	chainIds = append(chainIds, chainId)
}

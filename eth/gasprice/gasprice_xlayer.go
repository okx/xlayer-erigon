package gasprice

import "math/big"

const (
	weiToEth     = 1e18
	minUSDTPrice = 1e-18
)

func OKBToOKBWei(eth *big.Float) *big.Int {
	// Convert eth to wei
	wte := big.NewFloat(0).SetFloat64(weiToEth)
	val := eth.Mul(eth, wte)
	wei := new(big.Int)
	val.Int(wei)
	return wei
}

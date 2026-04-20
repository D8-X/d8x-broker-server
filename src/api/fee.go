package api

// getBrokerFeeTbps returns the broker fee. traderAddr can be an empty string
// and chainId can be -1
func (a *App) getBrokerFeeTbps(traderAddr string, chainId int) uint16 {
	return a.BrokerFeeTbps
}

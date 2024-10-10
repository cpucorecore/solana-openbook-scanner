package main

func main() {
	startSlot := uint64(294577906)

	rpcEndpoint := "https://api.mainnet-beta.solana.com"
	worker := NewWorker(rpcEndpoint)
	worker.getBlock(startSlot)
}

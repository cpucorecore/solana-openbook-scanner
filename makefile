test:
	go build
	./solana-openbook-scanner >scanner.log 2>&1

clean:
	rm -rf solana-openbook-scanner scanner.log blocks.json	

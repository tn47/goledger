build: clean
	go build -o ledger

dev: build
	./ledger examples/first.ldg

clean:
	rm -rf ledger

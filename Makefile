build: clean
	go build -o ledger

dev: build
	./ledger -f examples/first.ldg balance

clean:
	rm -rf ledger

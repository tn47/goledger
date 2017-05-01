build: clean
	go build -o ledger

dev: build
	./ledger -f examples/first.ldg -log debug

clean:
	rm -rf ledger

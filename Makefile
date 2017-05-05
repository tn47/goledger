SUBDIRS := api dblentry reports testdata

build: clean
	go build -o ledger

dev: build
	./ledger -f examples/first.ldg balance

test: clean build
	@for dir in $(SUBDIRS); do \
		echo $$dir "..."; \
		$(MAKE) -C $$dir test; \
	done
	go test -race -test.run=. -test.bench=. -test.benchmem=true


clean:
	rm -rf ledger

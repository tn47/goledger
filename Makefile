SUBDIRS := api dblentry reports testdata

build: clean
	go build

install: clean
	go install

dev: build
	./goledger -f examples/first.ldg balance

test: clean build
	@for dir in $(SUBDIRS); do \
		echo $$dir "..."; \
		$(MAKE) -C $$dir test; \
	done
	go test -v -race -test.run=. -test.bench=. -test.benchmem=true


clean:
	rm -rf ledger goledger

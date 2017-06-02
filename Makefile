SUBDIRS := api dblentry testdata reports

build: clean
	go build

install: clean
	go install

test: clean build
	@for dir in $(SUBDIRS); do \
		echo $$dir "..."; \
		$(MAKE) -C $$dir test; \
	done
	go test -v -race -test.run=. -test.bench=. -test.benchmem=true

coverage:
	@for dir in $(SUBDIRS); do \
		echo $$dir "..."; \
		$(MAKE) -C $$dir coverage; \
	done
	go test -coverprofile=coverage.out
	go tool cover -html=coverage.out
	rm -rf coverage.out

clean:
	rm -rf ledger goledger


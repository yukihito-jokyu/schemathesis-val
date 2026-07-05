.PHONY: run test-schemathesis test-schemathesis-auth test-schemathesis-stateful

# Use .venv/bin/schemathesis if it exists, otherwise fallback to system schemathesis
SCHEMATHESIS := $(shell if [ -f .venv/bin/schemathesis ]; then echo .venv/bin/schemathesis; else echo schemathesis; fi)

run:
	go run ./cmd/api

test-schemathesis:
	$(SCHEMATHESIS) run openapi.yaml --url http://localhost:8080

test-schemathesis-auth:
	$(SCHEMATHESIS) run openapi.yaml --url http://localhost:8080 -H "Authorization: Bearer test-token"

test-schemathesis-stateful:
	$(SCHEMATHESIS) run openapi.yaml --url http://localhost:8080 --stateful=links

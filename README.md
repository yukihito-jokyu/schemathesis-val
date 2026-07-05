# Schemathesis Go Verification API

This API is designed to verify the capability of Schemathesis in detecting schema deviations, boundary value validations, authentication verification, stateful transitions, and intentional bugs.

## Start API

```bash
go run ./cmd/api
```

Alternatively:
```bash
make run
```

## Run Schemathesis

```bash
schemathesis run openapi.yaml --url http://localhost:8080
```

Alternatively:
```bash
make test-schemathesis
```

## Run with auth header

```bash
schemathesis run openapi.yaml \
  --url http://localhost:8080 \
  -H "Authorization: Bearer test-token"
```

Alternatively:
```bash
make test-schemathesis-auth
```

## Run stateful testing

```bash
schemathesis run openapi.yaml \
  --url http://localhost:8080 \
  --stateful=links
```

Alternatively:
```bash
make test-schemathesis-stateful
```

## Expected findings

Schemathesis should detect failures from the following intentional bugs:

* `GET /bugs/schema-mismatch`: Returns invalid types and missing fields.
* `GET /bugs/status-mismatch`: Returns a 418 teapot code which is not documented in the schema.
* `POST /bugs/panic-on-zero`: Panics when a zero value is sent, returning 500.
* `GET /bugs/invalid-email`: Returns a user with a malformed email format (`not-an-email`).
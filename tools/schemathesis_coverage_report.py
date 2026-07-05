#!/usr/bin/env python3
"""Generate a small OpenAPI operation coverage dashboard from a Schemathesis HAR."""

from __future__ import annotations

import argparse
import html
import json
import re
from collections import Counter
from dataclasses import dataclass, field
from pathlib import Path
from urllib.parse import urlparse

import yaml


HTTP_METHODS = {"get", "put", "post", "delete", "patch", "head", "options", "trace"}


@dataclass
class Operation:
    method: str
    path: str
    operation_id: str
    documented_statuses: set[str]
    checks: list[str]
    request_count: int = 0
    observed_statuses: Counter[int] = field(default_factory=Counter)
    undocumented_statuses: Counter[int] = field(default_factory=Counter)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--schema", default="openapi.yaml", help="OpenAPI YAML file")
    parser.add_argument("--har", required=True, help="Schemathesis HAR report")
    parser.add_argument("--output", required=True, help="HTML output path")
    return parser.parse_args()


def status_is_documented(status: int, documented: set[str]) -> bool:
    return str(status) in documented or "default" in documented


def path_to_regex(path_template: str) -> re.Pattern[str]:
    escaped = re.escape(path_template)
    pattern = re.sub(r"\\\{[^/]+\\\}", r"[^/]+", escaped)
    return re.compile(f"^{pattern}$")


def resolve_ref(schema: dict, value: object) -> object:
    if not isinstance(value, dict) or "$ref" not in value:
        return value
    ref = value["$ref"]
    if not isinstance(ref, str) or not ref.startswith("#/"):
        return value
    current: object = schema
    for part in ref[2:].split("/"):
        if not isinstance(current, dict):
            return value
        current = current.get(part)
    return current


def describe_schema_constraints(schema: dict, node: object, prefix: str = "") -> list[str]:
    node = resolve_ref(schema, node)
    if not isinstance(node, dict):
        return []

    constraints: list[str] = []
    label = prefix or "body"
    for key in (
        "type",
        "enum",
        "minimum",
        "maximum",
        "minLength",
        "maxLength",
        "pattern",
        "format",
        "minItems",
        "maxItems",
        "additionalProperties",
    ):
        if key in node:
            value = node[key]
            if key == "enum":
                value = "|".join(map(str, value))
            constraints.append(f"{label}: {key}={value}")

    required = node.get("required")
    if isinstance(required, list) and required:
        constraints.append(f"{label}: required={','.join(map(str, required))}")

    properties = node.get("properties")
    if isinstance(properties, dict):
        for name, child in properties.items():
            child_prefix = f"{prefix}.{name}" if prefix else name
            constraints.extend(describe_schema_constraints(schema, child, child_prefix))

    items = node.get("items")
    if items is not None:
        constraints.extend(describe_schema_constraints(schema, items, f"{label}[]"))

    return constraints


def operation_checks(schema: dict, path_item: dict, operation: dict) -> list[str]:
    checks: list[str] = []
    parameters = []
    for source in (path_item.get("parameters"), operation.get("parameters")):
        if isinstance(source, list):
            parameters.extend(source)

    for parameter in parameters:
        parameter = resolve_ref(schema, parameter)
        if not isinstance(parameter, dict):
            continue
        name = parameter.get("name", "?")
        location = parameter.get("in", "?")
        constraints = describe_schema_constraints(schema, parameter.get("schema", {}), f"{location}.{name}")
        checks.extend(constraints or [f"{location}.{name}: present"])

    request_body = resolve_ref(schema, operation.get("requestBody"))
    if isinstance(request_body, dict):
        content = request_body.get("content")
        if isinstance(content, dict):
            for media_type, media in content.items():
                if isinstance(media, dict) and "schema" in media:
                    checks.append(f"requestBody: {media_type}")
                    checks.extend(describe_schema_constraints(schema, media["schema"]))

    return sorted(dict.fromkeys(checks))


def load_operations(schema_path: Path) -> list[Operation]:
    schema = yaml.safe_load(schema_path.read_text()) or {}
    operations: list[Operation] = []
    for path, path_item in (schema.get("paths") or {}).items():
        if not isinstance(path_item, dict):
            continue
        for method, operation in path_item.items():
            if method not in HTTP_METHODS or not isinstance(operation, dict):
                continue
            responses = operation.get("responses") or {}
            operations.append(
                Operation(
                    method=method.upper(),
                    path=path,
                    operation_id=str(operation.get("operationId") or f"{method}_{path}"),
                    documented_statuses={str(status) for status in responses},
                    checks=operation_checks(schema, path_item, operation),
                )
            )
    return operations


def load_har_entries(har_path: Path) -> list[dict]:
    if not har_path.exists():
        return []
    data = json.loads(har_path.read_text())
    return data.get("log", {}).get("entries", [])


def apply_har(operations: list[Operation], entries: list[dict]) -> None:
    matchers = [(operation, path_to_regex(operation.path)) for operation in operations]
    for entry in entries:
        request = entry.get("request", {})
        response = entry.get("response", {})
        method = str(request.get("method", "")).upper()
        path = urlparse(str(request.get("url", ""))).path
        status = int(response.get("status", 0) or 0)
        for operation, matcher in matchers:
            if operation.method == method and matcher.match(path):
                operation.request_count += 1
                operation.observed_statuses[status] += 1
                if not status_is_documented(status, operation.documented_statuses):
                    operation.undocumented_statuses[status] += 1
                break


def badge(text: str, kind: str) -> str:
    return f'<span class="badge {kind}">{html.escape(text)}</span>'


def render_html(operations: list[Operation], output_path: Path, har_path: Path) -> None:
    total = len(operations)
    covered = sum(1 for operation in operations if operation.request_count)
    undocumented = sum(sum(operation.undocumented_statuses.values()) for operation in operations)
    total_requests = sum(operation.request_count for operation in operations)
    percent = round((covered / total) * 100, 1) if total else 0

    rows = []
    for operation in operations:
        covered_badge = badge("covered", "ok") if operation.request_count else badge("not covered", "warn")
        status_badges = " ".join(
            badge(f"{status} x{count}", "bad" if status in operation.undocumented_statuses else "ok")
            for status, count in sorted(operation.observed_statuses.items())
        ) or "-"
        documented = ", ".join(sorted(operation.documented_statuses))
        checks = "<br>".join(html.escape(check) for check in operation.checks[:12])
        if len(operation.checks) > 12:
            checks += f"<br>... and {len(operation.checks) - 12} more"
        rows.append(
            "<tr>"
            f"<td>{covered_badge}</td>"
            f"<td><strong>{html.escape(operation.method)}</strong></td>"
            f"<td><code>{html.escape(operation.path)}</code><br><small>{html.escape(operation.operation_id)}</small></td>"
            f"<td>{operation.request_count}</td>"
            f"<td>{status_badges}</td>"
            f"<td>{html.escape(documented)}</td>"
            f"<td class=\"constraints\">{checks or '-'}</td>"
            "</tr>"
        )

    generated_from = html.escape(str(har_path))
    document = f"""<!doctype html>
<html lang="ja">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Schemathesis Coverage</title>
  <style>
    body {{ font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; margin: 0; color: #1f2937; background: #f8fafc; }}
    header {{ padding: 28px 32px; background: #0f172a; color: white; }}
    h1 {{ margin: 0 0 8px; font-size: 28px; }}
    main {{ padding: 24px 32px 40px; }}
    .summary {{ display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 12px; margin-bottom: 20px; }}
    .card {{ background: white; border: 1px solid #e5e7eb; border-radius: 8px; padding: 16px; }}
    .number {{ font-size: 30px; font-weight: 700; margin-top: 6px; }}
    .bar {{ height: 14px; background: #e5e7eb; border-radius: 999px; overflow: hidden; margin: 12px 0 22px; }}
    .bar > div {{ height: 100%; width: {percent}%; background: #2563eb; }}
    table {{ width: 100%; border-collapse: collapse; background: white; border: 1px solid #e5e7eb; }}
    th, td {{ padding: 10px 12px; border-bottom: 1px solid #e5e7eb; text-align: left; vertical-align: top; }}
    th {{ background: #f1f5f9; font-size: 13px; }}
    code {{ font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; }}
    small {{ color: #64748b; }}
    .badge {{ display: inline-block; padding: 3px 8px; border-radius: 999px; font-size: 12px; font-weight: 600; margin: 1px 2px 1px 0; }}
    .ok {{ background: #dcfce7; color: #166534; }}
    .warn {{ background: #fef3c7; color: #92400e; }}
    .bad {{ background: #fee2e2; color: #991b1b; }}
    .constraints {{ font-size: 12px; line-height: 1.5; max-width: 420px; }}
  </style>
</head>
<body>
  <header>
    <h1>Schemathesis Coverage</h1>
    <div>OpenAPI operation coverage generated from <code>{generated_from}</code></div>
  </header>
  <main>
    <section class="summary">
      <div class="card">Operation coverage<div class="number">{percent}%</div></div>
      <div class="card">Covered operations<div class="number">{covered}/{total}</div></div>
      <div class="card">HTTP requests<div class="number">{total_requests}</div></div>
      <div class="card">Undocumented statuses<div class="number">{undocumented}</div></div>
    </section>
    <div class="bar"><div></div></div>
    <table>
      <thead>
        <tr>
          <th>Coverage</th>
          <th>Method</th>
          <th>Operation</th>
          <th>Requests</th>
          <th>Observed statuses</th>
          <th>Documented statuses</th>
          <th>Generated input constraints</th>
        </tr>
      </thead>
      <tbody>
        {''.join(rows)}
      </tbody>
    </table>
  </main>
</body>
</html>
"""
    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(document)


def main() -> None:
    args = parse_args()
    operations = load_operations(Path(args.schema))
    har_path = Path(args.har)
    apply_har(operations, load_har_entries(har_path))
    render_html(operations, Path(args.output), har_path)


if __name__ == "__main__":
    main()

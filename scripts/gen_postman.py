"""Generate docs/postman_collection.json and docs/postman_environment.json from docs/openapi.yaml.

Usage (from the repository root):  python scripts/gen_postman.py
Requires PyYAML (pip install pyyaml).
"""
import json
import re
import uuid
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parent.parent
SPEC = yaml.safe_load((ROOT / "docs" / "openapi.yaml").read_text(encoding="utf-8"))
METHODS = ("get", "post", "put", "patch", "delete")


def resolve(ref):
    node = SPEC
    for part in ref.lstrip("#/").split("/"):
        node = node[part]
    return node


def example(schema, depth=0):
    """Build a small example value from an OpenAPI schema."""
    if depth > 4 or schema is None:
        return None
    if "$ref" in schema:
        return example(resolve(schema["$ref"]), depth + 1)
    if "allOf" in schema:
        merged = {}
        for part in schema["allOf"]:
            value = example(part, depth + 1)
            if isinstance(value, dict):
                merged.update(value)
        for k, v in schema.get("properties", {}).items():
            merged[k] = example(v, depth + 1)
        return merged
    if "example" in schema:
        return schema["example"]
    if "enum" in schema:
        return schema["enum"][0]
    kind = schema.get("type")
    fmt = schema.get("format")
    if kind == "object" or "properties" in schema:
        return {k: example(v, depth + 1) for k, v in schema.get("properties", {}).items()}
    if kind == "array":
        return [example(schema.get("items"), depth + 1)]
    if kind == "string":
        return {"uuid": "{{id}}", "email": "user@example.com", "date": "2026-12-01", "date-time": "2026-12-01T09:00:00Z"}.get(fmt, "string")
    if kind in ("number", "integer"):
        return 1
    if kind == "boolean":
        return True
    return None


def build_item(path, method, op):
    full = path
    path_vars = re.findall(r"\{(\w+)\}", path)
    full = re.sub(r"\{(\w+)\}", lambda m: "{{%s}}" % ("id" if m.group(1) == "id" else m.group(1)), full)

    shared = SPEC["paths"][path].get("parameters", [])
    query = []
    for param in shared + op.get("parameters", []):
        if "$ref" in param:
            param = resolve(param["$ref"])
        if param.get("in") == "query":
            query.append({"key": param["name"], "value": "", "disabled": True})

    url = {
        "raw": "{{baseUrl}}/api/v1" + full + ("?" + "&".join(q["key"] + "=" for q in query) if query else ""),
        "host": ["{{baseUrl}}"],
        "path": ["api", "v1"] + [seg for seg in full.strip("/").split("/") if seg],
        "query": query,
    }
    request = {"method": method.upper(), "header": [], "url": url, "description": op.get("description", "")}

    secured = op.get("security", SPEC.get("security", [])) != []
    if not secured:
        request["auth"] = {"type": "noauth"}

    body = op.get("requestBody")
    if body:
        content = body.get("content", {})
        if "application/json" in content:
            request["header"].append({"key": "Content-Type", "value": "application/json"})
            request["body"] = {
                "mode": "raw",
                "raw": json.dumps(example(content["application/json"].get("schema")), indent=2),
                "options": {"raw": {"language": "json"}},
            }
        elif "multipart/form-data" in content:
            request["body"] = {"mode": "formdata", "formdata": [{"key": "image", "type": "file", "src": []}]}

    item = {"name": op.get("summary", f"{method.upper()} {path}"), "request": request}
    if path == "/auth/login":
        item["event"] = [{
            "listen": "test",
            "script": {"type": "text/javascript", "exec": [
                "const d = pm.response.json().data;",
                "if (d) { pm.collectionVariables.set('token', d.access_token); pm.collectionVariables.set('refreshToken', d.refresh_token); }",
            ]},
        }]
        request["body"] = {"mode": "raw", "raw": json.dumps({"email": "sales@leventa.com", "password": "sales123"}, indent=2),
                           "options": {"raw": {"language": "json"}}}
    return item


def main():
    folders = {}
    for path, ops in SPEC["paths"].items():
        for method in METHODS:
            if method not in ops:
                continue
            op = ops[method]
            tag = (op.get("tags") or ["Other"])[0]
            folders.setdefault(tag, []).append(build_item(path, method, op))

    order = [t["name"] for t in SPEC.get("tags", [])]
    items = [{"name": t, "item": folders[t]} for t in order if t in folders]
    items += [{"name": t, "item": v} for t, v in folders.items() if t not in order]

    collection = {
        "info": {
            "_postman_id": str(uuid.uuid5(uuid.NAMESPACE_URL, "pms-collection")),
            "name": SPEC["info"]["title"],
            "description": SPEC["info"].get("description", ""),
            "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
        },
        "auth": {"type": "bearer", "bearer": [{"key": "token", "value": "{{token}}", "type": "string"}]},
        "variable": [
            {"key": "baseUrl", "value": "http://localhost:8080"},
            {"key": "token", "value": ""},
            {"key": "refreshToken", "value": ""},
            {"key": "id", "value": ""},
            {"key": "imageId", "value": ""},
            {"key": "marketingId", "value": ""},
            {"key": "surveyId", "value": ""},
            {"key": "propertyID", "value": ""},
        ],
        "item": items,
    }
    (ROOT / "docs" / "postman_collection.json").write_text(json.dumps(collection, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")

    environment = {
        "id": str(uuid.uuid5(uuid.NAMESPACE_URL, "pms-env-local")),
        "name": "PMS Local",
        "values": [{"key": "baseUrl", "value": "http://localhost:8080", "enabled": True}],
        "_postman_variable_scope": "environment",
    }
    (ROOT / "docs" / "postman_environment.json").write_text(json.dumps(environment, indent=2) + "\n", encoding="utf-8")
    total = sum(len(v) for v in folders.values())
    print(f"postman_collection.json: {total} requests in {len(items)} folders")


if __name__ == "__main__":
    main()

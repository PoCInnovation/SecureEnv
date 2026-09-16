# HTTP API v1

Every `/v1` request needs `Authorization: Bearer <vault token>`. The token is passed to Vault, so its policies decide what the caller may do.

Responses are JSON with `Cache-Control: no-store`. Errors look like:

```json
{ "error": { "code": "project_not_found", "message": "project not found: ..." } }
```

## Endpoints

| Method | Path | Body | Success |
|---|---|---|---|
| `GET` | `/healthz` | | `200`, process is alive |
| `GET` | `/readyz` | | `200` when Vault is reachable and unsealed, `503` otherwise |
| `GET` | `/v1/projects` | | `200` `{"projects": ["app"]}` |
| `POST` | `/v1/projects` | `{"name": "app"}` | `201`, `Location` header |
| `GET` | `/v1/projects/{project}` | | `200` `{"name", "current_version", "created_at", "updated_at"}` |
| `PATCH` | `/v1/projects/{project}` | `{"name": "new-name"}` | `204`, history is copied |
| `DELETE` | `/v1/projects/{project}` | | `204`, all versions are destroyed |
| `GET` | `/v1/projects/{project}/variables` | | `200` `{"version": 3, "variables": {"KEY": "value"}}`, `ETag: "3"` |
| `PUT` | `/v1/projects/{project}/variables` | `{"variables": {...}}` | `200` `{"version": 4}` |
| `GET` | `/v1/projects/{project}/variables/{key}` | | `200` `{"key", "value"}` |
| `PUT` | `/v1/projects/{project}/variables/{key}` | `{"value": "..."}` | `200` `{"version": 4}` |
| `DELETE` | `/v1/projects/{project}/variables/{key}` | | `200` `{"version": 5}` |

Send `If-Match: "3"` on `PUT /variables` to replace the variables only if the project is still at version 3. Otherwise the request fails with `412 version_conflict`.

## Validation

- Project names match `^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`.
- Variable keys match `^[A-Za-z_][A-Za-z0-9_]*$` and must not start with `SECURE_ENV_`.
- Request bodies are limited to 1 MiB and unknown JSON fields are rejected.

## Error codes

| Code | Status |
|---|---|
| `invalid_request` | 400 (413 when the body is too large) |
| `invalid_project_name`, `invalid_variable_key`, `reserved_variable_key` | 400 |
| `unauthorized` | 401 |
| `forbidden` | 403 |
| `project_not_found`, `variable_not_found` | 404 |
| `project_exists` | 409 |
| `version_conflict` | 412 |
| `internal` | 500, details are only logged server side |

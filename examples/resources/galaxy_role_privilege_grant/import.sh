# Role privilege grant can be imported using a composite ID.
# Basic format: role_id/entity_id/entity_kind/privilege/grant_kind
terraform import galaxy_role_privilege_grant.example <role_id>/<entity_id>/<entity_kind>/<privilege>/<grant_kind>

# Schema-scoped format (required when a role holds multiple Schema grants for
# different schema_names under the same catalog):
# role_id/entity_id/entity_kind/privilege/grant_kind/schema_name
# terraform import galaxy_role_privilege_grant.example <role_id>/<entity_id>/<entity_kind>/<privilege>/<grant_kind>/<schema_name>

# Extended format (when multiple Table or Column grants share the same entity/privilege/grant_kind):
# role_id/entity_id/entity_kind/privilege/grant_kind/schema_name/table_name/column_name
# terraform import galaxy_role_privilege_grant.example <role_id>/<entity_id>/<entity_kind>/<privilege>/<grant_kind>/<schema_name>/<table_name>/<column_name>

# Location entity IDs may contain slashes (e.g. s3:// or gs:// paths) and do not require encoding.
# terraform import galaxy_role_privilege_grant.example r-123/s3://bucket/prefix/*/Location/CreateSql/Allow

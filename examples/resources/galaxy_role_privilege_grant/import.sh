# Role privilege grant can be imported using a composite ID.
# Basic format: role_id/entity_id/entity_kind/privilege/grant_kind
terraform import galaxy_role_privilege_grant.example <role_id>/<entity_id>/<entity_kind>/<privilege>/<grant_kind>

# Extended format (when multiple grants share the same entity/privilege/grant_kind):
# role_id/entity_id/entity_kind/privilege/grant_kind/schema_name/table_name/column_name
# terraform import galaxy_role_privilege_grant.example <role_id>/<entity_id>/<entity_kind>/<privilege>/<grant_kind>/<schema_name>/<table_name>/<column_name>

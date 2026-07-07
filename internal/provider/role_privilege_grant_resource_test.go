// Copyright Starburst Data, Inc. All rights reserved.
//
// The source code is the proprietary and confidential information of Starburst Data, Inc. and
// may be used only for reference purposes in connection with the Terraform Registry. All rights,
// title, interest and ownership of the code and any derivatives, updates, upgrades, enhancements
// and modifications thereof remain with Starburst Data, Inc. You are not permitted to distribute,
// disclose, sell, lease, transfer, assign, modify, create derivative works of, or sublicense the
// code, or use the code to create or develop any products or services.

package provider

import (
	"testing"
)

func TestParseRolePrivilegeGrantImportID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		want    parsedImportID
		wantErr bool
	}{
		{
			name: "valid 5-part ID - basic catalog-level grant",
			id:   "role1/catalog1/Catalog/SELECT/Allow",
			want: parsedImportID{
				RoleID:     "role1",
				EntityID:   "catalog1",
				EntityKind: "Catalog",
				Privilege:  "SELECT",
				GrantKind:  "Allow",
				SchemaName: "",
				TableName:  "",
				ColumnName: "",
			},
			wantErr: false,
		},
		{
			name: "valid 6-part ID - schema-scoped grant",
			id:   "role1/catalog1/Schema/SELECT/Allow/schema1",
			want: parsedImportID{
				RoleID:     "role1",
				EntityID:   "catalog1",
				EntityKind: "Schema",
				Privilege:  "SELECT",
				GrantKind:  "Allow",
				SchemaName: "schema1",
				TableName:  "",
				ColumnName: "",
			},
			wantErr: false,
		},
		{
			name: "valid 8-part ID - table-scoped grant",
			id:   "role1/catalog1/Table/SELECT/Allow/schema1/table1/col1",
			want: parsedImportID{
				RoleID:     "role1",
				EntityID:   "catalog1",
				EntityKind: "Table",
				Privilege:  "SELECT",
				GrantKind:  "Allow",
				SchemaName: "schema1",
				TableName:  "table1",
				ColumnName: "col1",
			},
			wantErr: false,
		},
		{
			name: "entity ID containing slashes",
			id:   "role1/s3://bucket/path/Catalog/SELECT/Allow",
			want: parsedImportID{
				RoleID:     "role1",
				EntityID:   "s3://bucket/path",
				EntityKind: "Catalog",
				Privilege:  "SELECT",
				GrantKind:  "Allow",
				SchemaName: "",
				TableName:  "",
				ColumnName: "",
			},
			wantErr: false,
		},
		{
			name: "entity ID containing slashes with multiple path segments",
			id:   "role1/s3://bucket/path/to/object/Location/READ/Deny",
			want: parsedImportID{
				RoleID:     "role1",
				EntityID:   "s3://bucket/path/to/object",
				EntityKind: "Location",
				Privilege:  "READ",
				GrantKind:  "Deny",
				SchemaName: "",
				TableName:  "",
				ColumnName: "",
			},
			wantErr: false,
		},
		{
			name:    "invalid scope shape - Schema with 3 scope parts",
			id:      "role1/catalog1/Schema/SELECT/Allow/schema1/table1/col1",
			wantErr: true,
		},
		{
			name:    "invalid scope shape - Table with 1 scope part",
			id:      "role1/catalog1/Table/SELECT/Allow/schema1",
			wantErr: true,
		},
		{
			name:    "invalid scope shape - Column with 1 scope part",
			id:      "role1/catalog1/Column/SELECT/Allow/schema1",
			wantErr: true,
		},
		{
			name:    "missing entity_id",
			id:      "role1//Catalog/SELECT/Allow",
			wantErr: true,
		},
		{
			name:    "missing grant_kind",
			id:      "role1/catalog1/Catalog/SELECT",
			wantErr: true,
		},
		{
			name:    "unknown entity_kind",
			id:      "role1/catalog1/Unknown/SELECT/Allow",
			wantErr: true,
		},
		{
			name:    "unknown grant_kind",
			id:      "role1/catalog1/Catalog/SELECT/Maybe",
			wantErr: true,
		},
		{
			name: "valid Deny grant",
			id:   "role1/catalog1/Catalog/SELECT/Deny",
			want: parsedImportID{
				RoleID:     "role1",
				EntityID:   "catalog1",
				EntityKind: "Catalog",
				Privilege:  "SELECT",
				GrantKind:  "Deny",
				SchemaName: "",
				TableName:  "",
				ColumnName: "",
			},
			wantErr: false,
		},
		{
			name: "Location entity with no scope parts",
			id:   "role1/s3://location/path/Location/READ/Allow",
			want: parsedImportID{
				RoleID:     "role1",
				EntityID:   "s3://location/path",
				EntityKind: "Location",
				Privilege:  "READ",
				GrantKind:  "Allow",
				SchemaName: "",
				TableName:  "",
				ColumnName: "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRolePrivilegeGrantImportID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRolePrivilegeGrantImportID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got != tt.want {
				t.Errorf("parseRolePrivilegeGrantImportID() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

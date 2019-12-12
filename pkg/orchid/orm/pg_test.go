package orm

import (
	"testing"
)

func TestPg_ColumnTypeParser(t *testing.T) {
	type args struct {
		jsonSchemaType string
		format         string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "Error",
			args:    args{jsonSchemaType: "", format: ""},
			want:    "",
			wantErr: true,
		},
		{
			name:    "integer-int32",
			args:    args{jsonSchemaType: "integer", format: "int32"},
			want:    PgTypeInt,
			wantErr: false,
		},
		{
			name:    "empty-int32",
			args:    args{jsonSchemaType: "", format: "int32"},
			want:    PgTypeInt,
			wantErr: false,
		},
		{
			name:    "integer-empty",
			args:    args{jsonSchemaType: "integer", format: ""},
			want:    PgTypeInt,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ColumnTypeParser(tt.args.jsonSchemaType, tt.args.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("ColumnTypeParser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ColumnTypeParser() = %v, want %v", got, tt.want)
			}
		})
	}
}

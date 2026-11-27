package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWrapDollarQuotedStatements(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		sql     string
		want    string
		wantErr string
	}{
		{
			name: "plain statements are unchanged",
			sql: strings.Join([]string{
				"CREATE TABLE a(",
				"  id bigserial PRIMARY KEY",
				");",
				"",
				"CREATE INDEX a_idx ON a USING btree(id);",
			}, "\n"),
			want: strings.Join([]string{
				"CREATE TABLE a(",
				"  id bigserial PRIMARY KEY",
				");",
				"",
				"CREATE INDEX a_idx ON a USING btree(id);",
			}, "\n"),
		},
		{
			name: "function with dollar-quoted body is wrapped",
			sql: strings.Join([]string{
				"CREATE OR REPLACE FUNCTION river_job_notify()",
				"  RETURNS TRIGGER",
				"  AS $$",
				"DECLARE",
				"  payload json;",
				"BEGIN",
				"  IF NEW.state = 'available' THEN",
				"    -- Notify will coalesce duplicate notifications within a transaction, so",
				"    payload = json_build_object('queue', NEW.queue);",
				"    PERFORM",
				"      pg_notify('river_insert', payload::text);",
				"  END IF;",
				"  RETURN NULL;",
				"END;",
				"$$",
				"LANGUAGE plpgsql;",
				"",
				"CREATE TRIGGER river_notify",
				"  AFTER INSERT ON river_job",
				"  FOR EACH ROW",
				"  EXECUTE PROCEDURE river_job_notify();",
			}, "\n"),
			want: strings.Join([]string{
				"-- +goose StatementBegin",
				"CREATE OR REPLACE FUNCTION river_job_notify()",
				"  RETURNS TRIGGER",
				"  AS $$",
				"DECLARE",
				"  payload json;",
				"BEGIN",
				"  IF NEW.state = 'available' THEN",
				"    -- Notify will coalesce duplicate notifications within a transaction, so",
				"    payload = json_build_object('queue', NEW.queue);",
				"    PERFORM",
				"      pg_notify('river_insert', payload::text);",
				"  END IF;",
				"  RETURN NULL;",
				"END;",
				"$$",
				"LANGUAGE plpgsql;",
				"-- +goose StatementEnd",
				"",
				"CREATE TRIGGER river_notify",
				"  AFTER INSERT ON river_job",
				"  FOR EACH ROW",
				"  EXECUTE PROCEDURE river_job_notify();",
			}, "\n"),
		},
		{
			name: "do block with tagged dollar quotes is wrapped",
			sql: strings.Join([]string{
				"DO",
				"$body$",
				"BEGIN",
				"    IF (SELECT to_regclass('river_migration') IS NOT NULL) THEN",
				"        ALTER TABLE river_migration",
				"            RENAME TO river_migration_old;",
				"    END IF;",
				"END;",
				"$body$",
				"LANGUAGE 'plpgsql';",
			}, "\n"),
			want: strings.Join([]string{
				"-- +goose StatementBegin",
				"DO",
				"$body$",
				"BEGIN",
				"    IF (SELECT to_regclass('river_migration') IS NOT NULL) THEN",
				"        ALTER TABLE river_migration",
				"            RENAME TO river_migration_old;",
				"    END IF;",
				"END;",
				"$body$",
				"LANGUAGE 'plpgsql';",
				"-- +goose StatementEnd",
			}, "\n"),
		},
		{
			name: "closing tag and semicolon on the same line",
			sql: strings.Join([]string{
				"CREATE FUNCTION f()",
				"RETURNS boolean",
				"LANGUAGE SQL",
				"IMMUTABLE",
				"AS $$",
				"    SELECT 1;",
				"$$;",
			}, "\n"),
			want: strings.Join([]string{
				"-- +goose StatementBegin",
				"CREATE FUNCTION f()",
				"RETURNS boolean",
				"LANGUAGE SQL",
				"IMMUTABLE",
				"AS $$",
				"    SELECT 1;",
				"$$;",
				"-- +goose StatementEnd",
			}, "\n"),
		},
		{
			name: "comments before the statement stay outside the markers",
			sql: strings.Join([]string{
				"--",
				"-- Create the function.",
				"--",
				"",
				"CREATE FUNCTION f()",
				"  AS $$ SELECT 1 $$ LANGUAGE plpgsql;",
			}, "\n"),
			want: strings.Join([]string{
				"--",
				"-- Create the function.",
				"--",
				"",
				"-- +goose StatementBegin",
				"CREATE FUNCTION f()",
				"  AS $$ SELECT 1 $$ LANGUAGE plpgsql;",
				"-- +goose StatementEnd",
			}, "\n"),
		},
		{
			name: "dollar quotes inside single-quoted strings are ignored",
			sql:  `INSERT INTO a VALUES ('it''s $$ time');`,
			want: `INSERT INTO a VALUES ('it''s $$ time');`,
		},
		{
			name: "placeholder parameters are not dollar quotes",
			sql:  "SELECT set_config('k', $1, false);",
			want: "SELECT set_config('k', $1, false);",
		},
		{
			name:    "unterminated dollar quote",
			sql:     "CREATE FUNCTION f() AS $$\nBEGIN\n  RETURN 1;",
			wantErr: `unterminated dollar-quoted string "$$"`,
		},
		{
			name:    "statement starting mid-line is rejected",
			sql:     "DROP FUNCTION f; CREATE FUNCTION g() AS $$ SELECT 1 $$ LANGUAGE plpgsql;",
			wantErr: `statement starting mid-line is not supported`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := wrapDollarQuotedStatements(tt.sql)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want it to contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("result mismatch:\ngot:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestCheckRiverOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		files   []string
		wantErr string
	}{
		{
			name: "ordered files",
			files: []string{
				"00001_river_001.sql",
				"00002_river_002.sql",
				"00003_river_004.sql",
				"00004_app_migration.sql",
			},
		},
		{
			name: "app migrations interleaved between river migrations",
			files: []string{
				"00001_river_001.sql",
				"00002_app_migration.sql",
				"00003_river_003.sql",
			},
		},
		{
			name: "regenerated file got a new sequence number",
			files: []string{
				"00001_river_001.sql",
				"00002_river_003.sql",
				"00003_river_004.sql",
				"00004_river_002.sql",
			},
			wantErr: "00003_river_004.sql (river version 4) comes before 00004_river_002.sql (river version 2)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			for _, name := range tt.files {
				path := filepath.Join(dir, name)
				if err := os.WriteFile(path, []byte("-- +goose Up\nSELECT 1;"), 0o644); err != nil {
					t.Fatalf("error writing %s: %v", path, err)
				}
			}

			err := checkRiverOrder(dir)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want it to contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

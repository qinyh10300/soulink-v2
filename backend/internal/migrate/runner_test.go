package migrate

import "testing"

func TestNormalizeSQLMigrationStripsUTF8BOM(t *testing.T) {
	input := append([]byte{0xEF, 0xBB, 0xBF}, []byte("CREATE TABLE demo(id BIGINT);")...)

	got := normalizeSQLMigration(input)
	want := "CREATE TABLE demo(id BIGINT);"
	if got != want {
		t.Fatalf("normalizeSQLMigration() = %q, want %q", got, want)
	}
}

func TestNormalizeSQLMigrationKeepsPlainSQL(t *testing.T) {
	want := "CREATE TABLE demo(id BIGINT);"

	got := normalizeSQLMigration([]byte(want))
	if got != want {
		t.Fatalf("normalizeSQLMigration() = %q, want %q", got, want)
	}
}

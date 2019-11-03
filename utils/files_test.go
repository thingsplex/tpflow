package utils

import "testing"

func TestBackupDirectory(t *testing.T) {
	err := BackupDirectory("../testdata/var", "../testdata/backups", "flow")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}
package main

import (
	"github.com/mholt/archiver/v3"
	"os"
	"strings"
	"testing"
)

func TestArchiverTest(t *testing.T) {
	files := globList(strings.Split(os.Getenv("source"), ","))
	err := archiver.Archive(files.Source, "./test.tar")
	if err != nil {
		t.Errorf("create archiver error:%s", err.Error())
	}
}

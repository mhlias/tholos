package tholos

import (
	"os"
	"testing"
)

const fixtures_dir = "../test-fixtures"
const test_dir = "../test-tmp"

func TestConfigLoad(t *testing.T) {

	tholos := &Tholos_config{}

	os.Setenv("HOME", fixtures_dir)

	tholos = tholos.Configure(false)

	if tholos.Levels != 3 || tholos.Tf_modules_dir != "tfmodules" || tholos.Project_config_file != "project.yaml" || tholos.Root_profile != "root" {
		t.Fatal("Failed to properly load tholos config and match the parsed values.")
	}

}

func TestConfigInput(t *testing.T) {

	tholos := &Tholos_config{}

	os.Setenv("HOME", test_dir)

	inputs := "3,tfmodules,project.yaml,root"

	tholos = tholos.Configure(true, inputs)

	tholos = tholos.Configure(false)

	if tholos.Levels != 3 || tholos.Tf_modules_dir != "tfmodules" || tholos.Project_config_file != "project.yaml" || tholos.Root_profile != "root" {
		t.Fatal("Failed to properly save inputs to tholos config and match the parsed values of the stored file.")
	}

}

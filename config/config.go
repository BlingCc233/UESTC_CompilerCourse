package config

import (
	"os"
)

// File paths
const (
	SOURCE_PATH = "input/test.pas"
	ERR_PATH    = "output/output.err"
	DYD_PATH    = "output/output.dyd"
	DYS_PATH    = "output/output.dys"
	VAR_PATH    = "output/output.var"
	PRO_PATH    = "output/output.pro"
)

// Init creates the output directory if it doesn't exist
func Init() error {
	if _, err := os.Stat("output"); os.IsNotExist(err) {
		err := os.Mkdir("output", 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

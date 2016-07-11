package aws_helper

import (
//"log"
//"os"

//"github.com/craigmj/gototp"
)

type mfa_conf struct {
	device_name string
	serial_uuid string
	mfa_token   int16
}

func (m *mfa_conf) read_token() {

}

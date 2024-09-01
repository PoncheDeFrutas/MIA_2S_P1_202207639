package commands

import (
	"backend/global"
)

func ParserLogout(tokens []string) (string, error) {
	if err := global.LogUserOut(); err != nil {
		return "", err
	}
	return "", nil
}

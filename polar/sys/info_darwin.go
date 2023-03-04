package sys

import (
	"fmt"
	"polar/polar/utils"
	"strings"
)

func GetOs() string {
	prodName, err := utils.Invoke("sw_vers", "-productName")
	if err != nil {
		return "Unknown"
	}
	prodVer, err := utils.Invoke("sw_vers", "-productVersion")
	if err != nil {
		return strings.TrimSpace(string(prodName))
	}
	return fmt.Sprintf("%s %s", strings.TrimSpace(string(prodName)), strings.TrimSpace(string(prodVer)))
}

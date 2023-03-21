package sys

import (
	"fmt"
	"igloo/igloo/utils"
	"strings"
)

func GetOs() string {
	prodName, err := utils.Invoke("sw_vers", "-productName")
	if err != nil {
		return "Unknown"
	}
	prodVer, err := utils.Invoke("sw_vers", "-productVersion")
	if err != nil {
		return strings.TrimSpace(prodName)
	}
	return fmt.Sprintf("%s %s", strings.TrimSpace(prodName), strings.TrimSpace(prodVer))
}

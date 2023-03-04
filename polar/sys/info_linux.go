package sys

import (
	"github.com/dekobon/distro-detect/linux"
)

func GetOs() string {
	return linux.DiscoverDistro().Name
}

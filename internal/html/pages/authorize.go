package pages

import (
	"github.com/varsotech/prochat-server/internal/html/components"
)

type AuthorizePage struct {
	HeadInner components.HeadInner
	Name      string
	ClientID  string
}

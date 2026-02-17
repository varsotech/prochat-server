package pages

import (
	"github.com/varsotech/prochat-server/internal/homeserver/html/components"
)

type AuthorizePage struct {
	HeadInner   components.HeadInner
	Name        string
	ClientID    string
	QueryString string
}

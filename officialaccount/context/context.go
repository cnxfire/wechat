package context

import (
	"github.com/cnxfire/wechat/credential"
	"github.com/cnxfire/wechat/officialaccount/config"
)

// Context struct
type Context struct {
	*config.Config
	credential.AccessTokenHandle
}

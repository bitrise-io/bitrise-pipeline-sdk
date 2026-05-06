package step

// Common first-party step constants.
const (
	idActivateSSHKey    = "activate-ssh-key"
	idCachePullStep     = "cache-pull"
	idCachePushStep     = "cache-push"
	idSlackStep         = "slack"
)

// ActivateSSHKey returns a builder for the activate-ssh-key step.
func ActivateSSHKey() *Builder { return From(idActivateSSHKey, "1") }

// CachePull returns a builder for the cache-pull step.
func CachePull() *Builder { return From(idCachePullStep, "1") }

// CachePush returns a builder for the cache-push step.
func CachePush() *Builder { return From(idCachePushStep, "1") }

// Slack returns a builder for the slack message step.
func Slack() *Builder { return From(idSlackStep, "1") }

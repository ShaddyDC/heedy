package assets

func (c *Configuration) GetRequestBodyByteLimit() int64 {
	if c.RequestBodyByteLimit != nil {
		return *c.RequestBodyByteLimit
	}
	return 0
}

func (c *Configuration) GetHost() string {
	if c.Host != nil {
		return *c.Host
	}
	return ""
}

func (c *Configuration) GetPort() uint16 {
	if c.Port != nil {
		return *c.Port
	}
	return 0
}

func (c *Configuration) GetNewUserScopes() []string {
	if c.NewUserScopes != nil {
		return *c.NewUserScopes
	}
	return []string{}
}

func (c *Configuration) GetNewGroupScopes() []string {
	if c.NewGroupScopes != nil {
		return *c.NewGroupScopes
	}
	return []string{}
}

func (c *Configuration) GetNewConnectionScopes() []string {
	if c.NewConnectionScopes != nil {
		return *c.NewConnectionScopes
	}
	return []string{}
}

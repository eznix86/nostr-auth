package authorization

type Policy struct {
	authorizer *Authorizer
}

func CompilePolicy(cfg FileConfig) (*Policy, error) {
	authorizer, err := Compile(cfg)
	if err != nil {
		return nil, err
	}

	return &Policy{authorizer: authorizer}, nil
}

func LoadPolicyFile(path string) (*Policy, error) {
	cfg, err := LoadFileConfig(path)
	if err != nil {
		return nil, err
	}

	return CompilePolicy(cfg)
}

func (p *Policy) Allowed(host, pubkey, nip05 string) bool {
	if p == nil || p.authorizer == nil {
		return false
	}

	return p.authorizer.Allowed(host, pubkey, nip05)
}

func (p *Policy) Groups(host, pubkey, nip05 string) []string {
	if p == nil || p.authorizer == nil {
		return nil
	}

	return p.authorizer.Groups(host, pubkey, nip05)
}

package authorization

type Policy struct {
	authorizer *Authorizer
}

func LoadPolicyFile(path string) (*Policy, error) {
	authorizer, err := LoadFile(path)
	if err != nil {
		return nil, err
	}

	return &Policy{authorizer: authorizer}, nil
}

func (p *Policy) Allowed(host, pubkey, nip05 string) bool {
	if p == nil || p.authorizer == nil {
		return false
	}

	return p.authorizer.Allowed(host, pubkey, nip05)
}

package workspace

// Environments: named sets of variables, one active per project. The active
// one wins over the project's own variables, so a project can hold what every
// environment shares and each environment what only it knows.

// ActiveEnvironment returns the environment in use, or nil when none is.
func (p Project) ActiveEnvironment() *Environment {
	if p.ActiveEnv == "" {
		return nil
	}
	for i := range p.Environments {
		if p.Environments[i].ID == p.ActiveEnv {
			return &p.Environments[i]
		}
	}
	return nil
}

// Variables are the values a request is resolved against: the project's, with
// the active environment's on top.
func (p Project) Variables() []Pair {
	active := p.ActiveEnvironment()
	if active == nil {
		return p.Vars
	}

	out := make([]Pair, 0, len(p.Vars)+len(active.Vars))
	out = append(out, p.Vars...)
	for _, pair := range active.Vars {
		replaced := false
		for i := range out {
			if out[i].Key == pair.Key {
				out[i].Value = pair.Value
				replaced = true
				break
			}
		}
		if !replaced {
			out = append(out, pair)
		}
	}
	return out
}

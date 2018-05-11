package registry

import "github.com/netreduce/netreduce/nred"

type Registry struct{}

func (r *Registry) Definition(key string) (nred.Definition, bool) { return nred.Definition{}, false }

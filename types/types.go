package types

type Config struct {
	Frontends map[string]Frontend `toml:"frontends"`
	Backends map[string]Backend `toml:"backends"`
}

type Backend struct {
	Hosts []string
	Path string
}

type Frontend struct {
	Bind string
	Backends []string
	Strategy string
}

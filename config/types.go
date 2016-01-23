package config

type Config struct {
	Name, Directory, Command string
	UID, GID                 uint32
	Env                      []string
}

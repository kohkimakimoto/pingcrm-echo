package app

import (
	"embed"
	"path/filepath"

	"github.com/kohkimakimoto/inertia-echo"
)

type Config struct {
	Addr       string
	IsDev      bool
	IsDemo     bool
	Secret     []byte
	DataDir    string
	PublicFs   embed.FS
	ViewsFs    embed.FS
	DatabaseFs embed.FS
}

func NewConfig() *Config {
	return &Config{
		Addr:    ":8080",
		IsDev:   false,
		Secret:  nil,
		DataDir: "",
	}
}

func (c *Config) GetViteManifest() inertia.ViteManifest {
	b, err := c.PublicFs.ReadFile("public/dist/manifest.json")
	if err != nil {
		panic(err)
	}
	manifest, err := inertia.ParseViteManifest(b)
	if err != nil {
		panic(err)
	}
	return manifest
}

func (c *Config) DatabasePath() string {
	return filepath.Join(c.DataDir, "db.sqlite3")
}

func (c *Config) SchemaDDL() string {
	b, err := c.DatabaseFs.ReadFile("database/schema.sql")
	if err != nil {
		panic(err)
	}
	return string(b)
}

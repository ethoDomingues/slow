package slow

import (
	"html/template"
	"time"
)

func NewConfig() *Config { return &Config{} }

type Config struct {
	Env            string // environmnt
	SecretKey      string // for sign session
	Servername     string // for build url routes and route match
	ListeningInTLS bool   // UrlFor return a URl with schema in "https:"

	TemplateFolder string // for render Templates Html. Default "templates/"
	TemplateFuncs  template.FuncMap

	StaticFolder  string // for serve static files
	StaticUrlPath string // url uf request static file
	EnableStatic  bool   // enable static endpoint for serving static files

	Silent  bool   // don't print logs
	LogFile string // save log info in file

	SessionExpires          time.Duration
	SessionPermanentExpires time.Duration
}

func (c *Config) checkConfig() {
	if c.Env == "" {
		c.Env = "development"
	}
	if c.StaticFolder == "" {
		c.StaticFolder = "assets"
	}
	if c.TemplateFolder == "" {
		c.TemplateFolder = "templates/"
	}
	if c.TemplateFuncs == nil {
		c.TemplateFuncs = make(template.FuncMap)
	}
	if c.StaticUrlPath == "" {
		c.StaticUrlPath = "/assets"
	}
	if c.SessionExpires == 0 {
		c.SessionExpires = time.Minute * 30
	}
	if c.SessionPermanentExpires == 0 {
		c.SessionPermanentExpires = time.Hour * 744
	}
}

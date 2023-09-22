package slow

import (
	"html/template"
	"reflect"
	"time"
)

func NewConfig() *Config {
	return &Config{}
}

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

func (c *Config) getField(name string) reflect.Value {
	f := reflect.ValueOf(c).Elem().FieldByName(name)
	if !f.IsValid() {
		return reflect.Value{}
	}
	if f.Kind() == reflect.Pointer {
		return f.Elem()
	}
	return f
}

func (c *Config) GetField(name string) (any, bool) {
	f := c.getField(name)

	if f.IsValid() {
		return f.Interface(), true
	}
	return nil, false
}

func (c *Config) Set(name string, value any) bool {
	f := c.getField(name)

	if !f.IsValid() {
		return false
	}
	f.Set(reflect.ValueOf(value))
	return true
}

func (c *Config) Fields() []string {
	fields := []string{}
	rv := reflect.TypeOf(c).Elem()
	for i := 0; i < rv.NumField(); i++ {
		f := rv.Field(i)
		fields = append(fields, f.Name)
	}
	return fields
}

func (c *Config) Update(cfg *Config) {
	for _, v := range c.Fields() {
		f := cfg.getField(v)
		if f.IsValid() && f.CanInterface() {
			c.Set(v, f.Interface())
		}
	}
}

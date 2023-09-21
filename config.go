package slow

import (
	"reflect"
	"time"
)

func NewConfig() *Config {
	return &Config{
		StaticFolder:            "assets",
		StaticUrlPath:           "/assets",
		TemplateFolder:          "templates/",
		EnableStatic:            true,
		SessionExpires:          time.Minute * 30,
		SessionPermanentExpires: time.Hour * 744,
	}
}

type Config struct {
	Env                     string // environmnt
	LogFile                 string // save log info in file
	SecretKey               string // for sign session
	Servername              string // for build url routes and route match
	StaticFolder            string // for serve static files
	TemplateFolder          string // for render Templates Html. Default "templates/"
	StaticUrlPath           string // url uf request static file
	Silent                  bool   // don't print logs
	EnableStatic            bool   // enable static endpoint for serving static files
	ListeningInTLS          bool   // UrlFor return a URl with schema in "https:"
	SessionExpires          time.Duration
	SessionPermanentExpires time.Duration
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

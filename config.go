package slow

import "reflect"

func NewConfig() *Config {
	return &Config{
		StaticFolder:   "./assets",
		TemplateFolder: "./templates",
		StaticUrlPath:  "/assets",
		EnableStatic:   true,
	}
}

type Config struct {
	Env            string // environmnt
	LogFile        string // save log info in file
	SecretKey      string // for sign session
	Servername     string // for build url routes and route match
	StaticFolder   string // for serve static files
	StaticUrlPath  string // url uf request static file
	TemplateFolder string // for render template (html) files
	Silent         bool   // don't print logs
	EnableStatic   bool   // enable static endpoint for serving static files
	ListeningInTLS bool   // UrlFor return a URl with schema in "https:"
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
		c.Set(v, cfg.getField(v).Interface())
	}
}

// config.go

package config

type Config struct {
	Port                  string `yaml:"port"`
	VersionPrefixedRoutes bool   `yaml:"version_prefixed_routes"`
	EnableDebugProfiling  bool   `yaml:"enable_debug_profiling"`
	Installpath           string `yaml:"installpath"`
	Keypath               string `yaml:"keypath"`
	Basepath              string `yaml:"basepath"`
	ProxyUrl              string `yaml:"proxyurl"`
	RatingMapUrl          string `yaml:"ratingmapurl"`
	RatingType            string `yaml:"ratingtype"`

	Authentication struct {
		Enable bool `yaml:"enable"`
	}

	Log struct {
		ConsoleLevel   string `yaml:"consolelevel"`
		UseFile        bool   `yaml:"usefile"`
		FileLevel      string `yaml:"filelevel"`
		FilePath       string `yaml:"filepath"`
		FileMaxSize    int    `yaml:"filemaxsize"`
		FileMaxBackups int    `yaml:"filemaxbackup"`
		FileMaxAge     int    `yaml:"filemaxage"`
	} `yaml:"log"`

	Cors struct {
		AllowedOrigins string `yaml:"allowed_origins"`
		AllowedMethods string `yaml:"allowed_methods"`
		AllowedHeaders string `yaml:"allowed_headers"`
	} `yaml:"cors"`

	Caching struct {
		TTL      int `yaml:"default_ttl"`
		ErrorTTL int `yaml:"error_ttl"`
	} `yaml:"caching"`

	Assistant Assistant `yaml:"assistant" json:"assistant"`
}

func (c *Config) Init() {
	c.Port = "8080"
	c.VersionPrefixedRoutes = true
	c.Installpath = "."
	c.Keypath = "."
	c.Basepath = ""
	c.ProxyUrl = ""
	c.RatingMapUrl = "https://lab-gcp.ifeelsmart.net/tools/mirror/rating-system.json"
	c.RatingType = "mpaa"

	c.Cors.AllowedOrigins = "*"
	c.Cors.AllowedMethods = "POST,GET,DELETE"
	c.Cors.AllowedHeaders = "X-Requested-With,Content-Type"

	c.Log.ConsoleLevel = "panic"
	c.Log.UseFile = true
	c.Log.FileLevel = "debug"
	c.Log.FilePath = "."
	c.Log.FileMaxSize = 50
	c.Log.FileMaxBackups = 3
	c.Log.FileMaxAge = 28

	c.Caching.TTL = 300
	c.Caching.ErrorTTL = 2

	var ass Assistant
	ass.Init()
	c.Assistant = ass
}

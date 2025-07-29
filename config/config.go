package config

type Config struct {
	AllowedExtensions map[string]bool
	MaxFilesPerTask   int
	MaxActiveTasks    int
	Port              string
}

func LoadConfig() Config {
	return Config{
		AllowedExtensions: map[string]bool{
			".pdf":  true,
			".jpeg": true,
			".jpg":  true,
		},
		MaxFilesPerTask: 3,
		MaxActiveTasks:  3,
		Port:            "8080",
	}
}

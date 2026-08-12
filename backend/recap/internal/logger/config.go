package logger

type Config struct {
	Level       string `env:"LEVEL" envDefault:"info"`
	Development bool   `env:"DEVELOPMENT" envDefault:"false"`
}

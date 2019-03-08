package config

type Config struct {
	Token string `default:"fake_token"`
	Redis string `default:"127.0.0.1:6379/1"`
}

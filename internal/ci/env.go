package ci

import (
	"flag"

	"github.com/joho/godotenv"
)

func LoadEnvFileFromFlag(args []string) error {
	fs := flag.NewFlagSet("ci", flag.ContinueOnError)
	envFile := fs.String("env", "", "Path to .env file")
	fs.Parse(args[1:])
	if *envFile == "" {
		return nil
	}
	return godotenv.Load(*envFile)
}

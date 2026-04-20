package helpers

import "os"

func GetMainDiscordServerID() string {
	return os.Getenv("DISCORD_SERVER_ID")
}

func GetAPIKey() string {
	return os.Getenv("ZAU_API_KEY")
}

func GetAPIURL() string {
	return os.Getenv("ZAU_API_URL")
}

func GetBotToken() string {
	return os.Getenv("BOT_TOKEN")
}

func GetRedisURI() string {
	return os.Getenv("REDIS_URI")
}

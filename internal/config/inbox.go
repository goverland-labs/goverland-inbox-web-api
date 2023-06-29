package config

type Inbox struct {
	StorageAddress string `env:"INBOX_API_STORAGE_ADDRESS" envDefault:"localhost:11055"`
}

package database

type Config struct {
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"DBName"`
	SSLMode  string `json:"SSLMode"`
}

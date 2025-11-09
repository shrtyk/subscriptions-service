package main

import (
	"fmt"

	"github.com/shrtyk/subscriptions-service/internal/config"
	"github.com/shrtyk/subscriptions-service/internal/infra/postgres"
)

func main() {
	cfg := config.MustInitConfig()
	_ = postgres.MustCreateConnectionPool(&cfg.PostgresCfg)
	fmt.Println("hey hooman")
}

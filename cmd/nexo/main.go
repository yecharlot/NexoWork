package main

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/yecharlot/NexoWork/internal/runtime"
)

func main() {
	cfg := runtime.ConfigFromEnv()
	// Resolve web dir relative to executable or cwd
	if _, err := os.Stat(cfg.WebDir); err != nil {
		candidates := []string{
			"web",
			filepath.Join("..", "web"),
			filepath.Join(filepath.Dir(os.Args[0]), "web"),
			filepath.Join(filepath.Dir(os.Args[0]), "..", "web"),
		}
		for _, c := range candidates {
			if st, err := os.Stat(c); err == nil && st.IsDir() {
				cfg.WebDir = c
				break
			}
		}
	}
	log.Printf("NEXO starting · PrismaTec-Core · HTTP %s · AIP %s · data %s · web %s",
		cfg.HTTPAddr, cfg.AIPAddr, cfg.DataDir, cfg.WebDir)

	rt, err := runtime.New(cfg)
	if err != nil {
		log.Fatalf("runtime: %v", err)
	}
	if err := rt.Start(); err != nil {
		log.Fatalf("listen: %v", err)
	}
	log.Printf("NEXO listo · abre http://127.0.0.1%s  · admin token vía NEXO_ADMIN_TOKEN", cfg.HTTPAddr)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	log.Printf("apagando…")
	_ = rt.Stop()
}

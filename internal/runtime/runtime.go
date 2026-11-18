package runtime

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yecharlot/NexoWork/internal/httpapi"
	"github.com/yecharlot/NexoWork/internal/store"
	"github.com/yecharlot/PrismaTec-Core/api/aip"
	"github.com/yecharlot/PrismaTec-Core/core"
)

type Config struct {
	DataDir    string
	NodeName   string
	HTTPAddr   string
	AIPAddr    string
	AdminToken string
	WebDir     string
}

func ConfigFromEnv() Config {
	home, _ := os.UserHomeDir()
	dataDir := os.Getenv("NEXO_DATA_DIR")
	if dataDir == "" {
		dataDir = filepath.Join(home, ".local", "share", "nexo")
	}
	name := strings.TrimSpace(os.Getenv("NEXO_NODE_NAME"))
	if name == "" {
		name = "nexo-node"
	}
	httpAddr := strings.TrimSpace(os.Getenv("NEXO_HTTP_ADDR"))
	if httpAddr == "" {
		httpAddr = ":8080"
	}
	aipAddr := strings.TrimSpace(os.Getenv("NEXO_AIP_ADDR"))
	if aipAddr == "" {
		aipAddr = ":8081"
	}
	webDir := strings.TrimSpace(os.Getenv("NEXO_WEB_DIR"))
	if webDir == "" {
		webDir = "web"
	}
	return Config{
		DataDir:    dataDir,
		NodeName:   name,
		HTTPAddr:   httpAddr,
		AIPAddr:    aipAddr,
		AdminToken: strings.TrimSpace(os.Getenv("NEXO_ADMIN_TOKEN")),
		WebDir:     webDir,
	}
}

type Runtime struct {
	node       *core.Node
	cancel     context.CancelFunc
	httpServer *http.Server
	aipServer  *http.Server
}

func New(cfg Config) (*Runtime, error) {
	if cfg.AdminToken == "" {
		cfg.AdminToken = "nexo-admin-change-me"
	}
	node, err := core.NewNode(core.Config{
		DataDir: cfg.DataDir,
		Name:    cfg.NodeName,
	})
	if err != nil {
		return nil, fmt.Errorf("create PrismaTec node: %w", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	if err := node.Start(ctx); err != nil {
		cancel()
		return nil, fmt.Errorf("start PrismaTec node: %w", err)
	}

	st := store.New(node.Organisms(), string(node.ID()))
	api := httpapi.New(st, cfg.AdminToken, cfg.WebDir)

	rt := &Runtime{node: node, cancel: cancel}
	rt.httpServer = &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           api.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	aipSrv := &aip.Server{Node: node, Addr: cfg.AIPAddr}
	rt.aipServer = &http.Server{
		Addr:              cfg.AIPAddr,
		Handler:           aipSrv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	return rt, nil
}

func (r *Runtime) Start() error {
	errCh := make(chan error, 2)
	go func() {
		errCh <- r.httpServer.ListenAndServe()
	}()
	go func() {
		errCh <- r.aipServer.ListenAndServe()
	}()
	// Wait briefly for bind errors
	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			return err
		}
	case <-time.After(200 * time.Millisecond):
	}
	return nil
}

func (r *Runtime) Stop() error {
	r.cancel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = r.httpServer.Shutdown(ctx)
	_ = r.aipServer.Shutdown(ctx)
	return r.node.Stop()
}

func (r *Runtime) Node() *core.Node { return r.node }

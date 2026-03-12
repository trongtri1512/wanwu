package redis

import (
	"context"
	"fmt"
)

const (
	_dbApp = 8
)

var (
	_redisApp *client
)

func InitApp(ctx context.Context, cfg Config) error {
	if _redisApp != nil {
		return fmt.Errorf("redis app client already init")
	}
	c, err := newClient(ctx, cfg, _dbApp)
	if err != nil {
		return err
	}
	_redisApp = c
	return nil
}

func StopApp() {
	if _redisApp != nil {
		_redisApp.Stop()
		_redisApp = nil
	}
}

func App() *client {
	return _redisApp
}

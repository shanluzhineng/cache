package starter

import (
	"context"
	"fmt"
	"time"

	"github.com/abmpio/abmp/pkg/log"
	"github.com/abmpio/app"
	"github.com/abmpio/app/web"
	"github.com/abmpio/configurationx"
	"go.uber.org/zap"

	redisx "github.com/abmpio/redisx"
	"github.com/go-redis/redis/v8"
)

func init() {
	web.ConfigureService(serviceConfigurator)
}

func serviceConfigurator(wa web.WebApplication) {
	redisClient := initRedis()

	redisxOption := redisx.NewRedisOptions(redisClient)
	redisxOption.KeyPrefix = configurationx.GetInstance().Redis.GetDefaultOptions().KeyPrefix
	app.Context.RegistInstance(redisxOption)
	redisOption := redisx.NewRedisOptions(redisClient)
	app.Context.RegistInstance(redisOption)

	newRedisService := redisx.NewRedisService(redisOption)
	//注册IRedisService接口
	app.Context.RegistInstanceAs(newRedisService, new(redisx.IRedisService))
	newRedisxService := redisx.NewRedisService(redisxOption)
	app.Context.RegistInstanceAs(newRedisxService, new(redisx.IRedisService))
}

func initRedis() *redis.Client {
	client := createRedisClient()

	for {
		err := redisHealthCheck(client)
		if err == nil {
			break
		}
		log.Logger.Error(err.Error() + ",sleep 5 seconds...")
		time.Sleep(5 * time.Second)
	}
	app.Context.RegistInstance(client)
	return client
}

func createRedisClient() *redis.Client {
	defaultRedisOptions := configurationx.GetInstance().Redis.GetDefaultOptions()
	if defaultRedisOptions == nil {
		err := fmt.Errorf("没有配置好redis")
		log.Logger.Error(err.Error())
		panic(err)
	}

	client := redis.NewClient(&redis.Options{
		Network:  defaultRedisOptions.Network,
		Addr:     defaultRedisOptions.Addr,
		Password: defaultRedisOptions.Password,
		DB:       defaultRedisOptions.DB,
	})
	return client
}

func redisHealthCheck(client *redis.Client) error {
	pong, err := client.Ping(context.Background()).Result()
	if err != nil {
		err := fmt.Errorf(fmt.Sprintf("redis connect ping failed, err:%s", err.Error()))
		return err
	}
	log.Logger.Info("redis connect ping response:", zap.String("pong", pong))
	return nil
}

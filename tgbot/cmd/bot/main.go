package main

import (
	"bot/internal/config"
	"bot/internal/server"
	"bot/internal/storage"
	"bot/internal/tgbot"
	"bot/pkg/logger"
	_ "bot/pkg/stepik_parser"
	"context"
	_ "fmt"
	_ "time"
)

func main(){
	// timer:= time.Now()
	cfg:= config.NewConfig()
	logger , err:= logger.InitLogger(true, "", cfg.App.LogLevel)
	if err!= nil{
		panic(err)
	}


	db := storage.NewStorage(cfg.Storage.Type, cfg.Storage.ConnectionString,logger)


	botCfg:= tgbot.NewConfig(cfg.Bot.Token, cfg.Bot.Admins,cfg.Bot.Timeout)
	bot,err:= tgbot.NewBot(botCfg,logger,db)
	if err!= nil{
		panic(err)
	}
	ctx := context.Background()
	logger.Info("start bot")
	s:= server.NewServer(cfg.Server.Host,cfg.Server.Port,db,cfg.Server.JwtToken,cfg.Server.TokenTTL,logger)
	go bot.Start(ctx)
	s.Start()

// 	courseCh := make(chan parser.Course)
// 	errCh:= make(chan error)
// 	go parser.Parser("https://stepik.org/course/228315/promo",errCh,courseCh)
// 	var course parser.Course
// 	select{
// 	case course = <- courseCh:
// 		fmt.Println("good")
// 	case err := <- errCh:
// 		panic(err)
// 	}
// 	fmt.Println(course.Title)
// 	fmt.Println(course.Lessons)
// 	fmt.Println(course.Duration)
// 	fmt.Println(course.Description)
// 	fmt.Println(course.Level)
// 	for i:=0;i<len(course.Knowledges);i++{
// 		fmt.Println(i+1,":",course.Knowledges[i])
// 	}
// 	fmt.Println(course.Students)
// 	fmt.Println(time.Since(timer))
}
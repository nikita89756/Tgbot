package tgbot

import (
	"bot/internal/errors"
	"bot/internal/models"
	"bot/internal/storage"
	parser "bot/pkg/stepik_parser"
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	tb "gopkg.in/telebot.v4"
)
var Id int

type Bot struct {
	bot *tb.Bot
	logger *zap.Logger
	db storage.Storage
	cfg *Config
}

func NewBot(cfg *Config,logger *zap.Logger,db storage.Storage) (*Bot,error) {
	setting:=tb.Settings{
		Token: cfg.Token,
		Poller: &tb.LongPoller{
			Timeout: cfg.Timeout,
		},
	}
	bot,err:=tb.NewBot(setting)
	if err!= nil{
		return nil,err
	}

	b:= &Bot{bot:bot,logger: logger,cfg: cfg,db: db}
	b.prepare()

	return b,nil
}

func (b *Bot) prepare() {
	menu:=&tb.ReplyMarkup{ResizeKeyboard: true}
	btnSendAchievment:=menu.Text("Отправить достижение")
	btnGetProfile := menu.Text("Профиль")
	btnTasks := menu.Text("Задания")
	btnBuyThings:= menu.Text("Приобрести вещи")
	btnGetCourseRating:= menu.Text("Узнать рейтинг курса")
	menu.Reply(
		menu.Row(btnSendAchievment,btnGetProfile),
		menu.Row(btnTasks,btnBuyThings),
		menu.Row(btnGetCourseRating),
	)

	b.bot.Handle("/start", func(c tb.Context)error {
		id := int(c.Sender().ID)
		err := b.db.InsertNewUser(id)
		if err != nil {
			if err == errors.ErrUserAlreadyExists{ 
				b.logger.Info("Пользователь запустил бота",zap.Int("user_id",id))
				return c.Send("hello",menu)
			}
		b.logger.Error("Database error", zap.Error(err))
		return c.Send("Произошла ошибка")
		}
		b.logger.Info("Пользователь запустил бота",zap.Int("user_id",id))
		return c.Send("hello",menu)
	})

	b.bot.Handle(&btnSendAchievment, func(c tb.Context) error {
		return b.createCourseHandler(c)
	})

	b.bot.Handle(&btnGetProfile, func(c tb.Context) error {
		return b.getUserProfileHandler(c)
	})

	b.bot.Handle(&btnBuyThings, func(c tb.Context) error {
		return b.buyThingsHandler(c)
	})

	b.bot.Handle(&btnGetCourseRating, func(c tb.Context) error {
		return b.getCourseRatingHandler(c)
	})

	b.bot.Handle(&btnTasks, func(c tb.Context) error {
		return b.getTasksHandler(c,menu)
	})

}

func (b *Bot)getTasksHandler(c tb.Context,menu *tb.ReplyMarkup) error {
	menuTasks:=&tb.ReplyMarkup{ResizeKeyboard: true}
	btnSendTasks:=menuTasks.Text("Добавить задания")
	btnGetTasks := menuTasks.Text("Получить задания")
	btnBack := menuTasks.Text("Назад")
	menuTasks.Reply(
		menuTasks.Row(btnSendTasks,btnGetTasks),
		menuTasks.Row(btnBack),
	)

	c.Send("Выберите действие",menuTasks)

	b.bot.Handle(&btnSendTasks, func(c tb.Context) error {
		c.Send("Отправьте задание, которое вам необходимо выполнить")
		b.bot.Handle(tb.OnText, func(c tb.Context) error {
			task:=c.Text()
			b.db.InsertTask(int(c.Sender().ID),task)
			return c.Respond(&tb.CallbackResponse{Text: "Задание успешно добавлено"})
		})
		return nil
	})

	b.bot.Handle(&btnGetTasks, func(c tb.Context) error {
		tasks, err := b.db.GetTasks(int(c.Sender().ID))
		if err != nil {
			b.logger.Error("Database error", zap.Error(err))
			return c.Respond(&tb.CallbackResponse{Text: "Произошла ошибка"})
		}
		buttons:=make([]tb.Row,len(tasks))
		menuUsersTasks := &tb.ReplyMarkup{ResizeKeyboard: true}
		for i,task:=range tasks{
			buttons[i]=menuUsersTasks.Row(menuUsersTasks.Data(task.Task,fmt.Sprintf("task_%d",task.ID)))
		}
		menuUsersTasks.Inline(buttons...)
		return c.Send("Выберите выполненное задание",menuUsersTasks)
	})

	b.bot.Handle(tb.OnCallback, func(c tb.Context) error {
		var id int
		data := strings.TrimSpace(c.Callback().Data)
		_,err:=fmt.Sscanf(data,"task_%d",&id)
		if err!=nil{
			return c.Respond(&tb.CallbackResponse{Text: "Произошла ошибка"})
		}
		err = b.db.UpdateTaskStatus(id)
		if err != nil {
			b.logger.Error("Database error", zap.Error(err))
			return c.Respond(&tb.CallbackResponse{Text: "Произошла ошибка"})
		}
		return c.Respond(&tb.CallbackResponse{Text: "Задание выполнено"})

	})

	b.bot.Handle(&btnBack, func(c tb.Context) error {
		return c.Send("Выберите действие",menu)
	})
	return nil
}
func (b *Bot)getCourseRatingHandler(c tb.Context) error {
		c.Send("Пожалуйста отправьте ссылку на курс")
		b.bot.Handle(tb.OnText, func(c tb.Context) error {
			link:=c.Text()
			_,title,topics,grades,err:=b.db.GetCourseByLink(link)
			if err!=nil{
				
				if err == sql.ErrNoRows{
					return c.Respond(&tb.CallbackResponse{Text: "Курс не найден"})
				}

				return c.Respond(&tb.CallbackResponse{Text: "Произошла ошибка"})
			}
			answer:=fmt.Sprintf("Курс: %s\n\n",title)
			for i:=0;i<len(topics);i++{
				answer = answer + fmt.Sprintf("\n%d. %s \n**Рейтинг: %2f**\n",i+1,topics[i],grades[i])
			}
			return c.Send(answer)
			
		})
		return nil
}

func (b *Bot) buyThingsHandler(c tb.Context) error {

	items , err :=b.db.GetItems()
	if err!=nil {
		b.logger.Error("Database error", zap.Error(err))
		return c.Respond(&tb.CallbackResponse{Text: "Произошла ошибка"})
	}

	if len(items)==0{
		return c.Respond(&tb.CallbackResponse{Text: "Вещи отсутствуют"})
	}
	
	buttons:=make([]tb.Row,len(items))
	menu:=&tb.ReplyMarkup{ResizeKeyboard: true}
	for i:=0;i<len(items);i++{
		title:= fmt.Sprintf("Название: %s\nСтоимость %d, Множитель:%f",items[i].Name,items[i].Price,items[i].Multiplier)

		data:=fmt.Sprintf("item_%d_%d",items[i].ID,items[i].Price)
		btn:=menu.Data(title,data)
		buttons[i] = menu.Row(btn)
	}
	menu.Inline(buttons...)
	
	b.bot.Handle(tb.OnCallback,func(c tb.Context)error{
		var item models.Item
		data:=strings.TrimSpace(c.Callback().Data)
		_,err:=fmt.Sscanf(data,"item_%d_%d",&item.ID,&item.Price)
		if err!=nil{
			b.logger.Error("Invalid callback data",zap.String("data",data))
		}
		err = b.db.InsertUserItems(int(c.Sender().ID),item.ID,item.Price)
		if err != nil{
			if err == errors.ErrNotEnothCoins{
				return c.Respond(&tb.CallbackResponse{Text:"Недостаточно монет"})
			}else if err == errors.ErrUserItemAlreadyExists{
				return c.Respond(&tb.CallbackResponse{Text:"У вас есть уже этот предмет"})
			}else{
				b.logger.Error("Database Error",zap.Error(err))
				return c.Respond(&tb.CallbackResponse{Text:"Произзошла ошибка"})
			}
		}
		return c.Respond(&tb.CallbackResponse{Text:"Покупка совершена"})
	})
	return c.Send("Купить бусты:",menu)
}


func (b *Bot)getUserProfileHandler(c tb.Context) error {
	id := int(c.Sender().ID)
	coins,exp, err := b.db.GetUser(id)
	if err != nil {
		b.logger.Error("Database error", zap.Error(err))
		return c.Respond(&tb.CallbackResponse{Text: "Произошла ошибка"})
	}
	items,err:= b.db.GetUserItems(int(c.Sender().ID))
	answer := fmt.Sprintf("Ваша статистика:\n\nДеньги: %d\nОпыт: %d\n\nВаши предметы:\n", coins, exp)

	for i:=0;i<len(items);i++{
		itemInfo := fmt.Sprintf("\n%d. %s | Множитель:%2f",i+1,items[i].Name,items[i].Multiplier)
		answer = answer + itemInfo
	}
	return c.Send(answer)
}
func (b *Bot) createCourseHandler(c tb.Context)error{
	yesNoMenu:=&tb.ReplyMarkup{ResizeKeyboard: true}
			btnYes:=yesNoMenu.Data("Да","yes")
			btnNo:=yesNoMenu.Data("Нет","no")
			yesNoMenu.Inline(
				yesNoMenu.Row(btnYes,btnNo),
			)
			b.bot.Send(c.Sender(), "Пожалуйста, отправьте ссылку на курс")
			b.bot.Handle(tb.OnText, func(c tb.Context) error {
				var course parser.Course
				coursCh := make(chan parser.Course)
				errCh := make(chan error,2)
				
				go parser.Parser(c.Text(),errCh,coursCh)
				select {
				case course = <-coursCh:
					b.logger.Info("курс успешно найден")
				case err := <-errCh:
					b.logger.Error("Ошибка parser", zap.Error(err))
					return c.Send("Ошибка при чтении ссылки")
				}
				err :=c.Send("Пожалуйста, подождите ...")
				if err != nil {
					b.logger.Error("Ошибка отправки сообщения", zap.Error(err))
					return err
				}
				Id,err = b.db.GetCourseID(course.Title) 
				b.logger.Info("ошибочка",zap.Error(err))
				if err != nil {
					if err==sql.ErrNoRows{
						// create an connection with ml that give me course rating
						b.logger.Info("курс успешно создан в базе данных",zap.Error(err))
						Id,err = b.db.InsertCourse(course.Title,course.Url,course.Knowledges,6,100)
						b.logger.Info("id",zap.Int("id",Id))
						if err != nil {
							b.logger.Error("Database error", zap.Error(err))
							return c.Send("Произошла ошибка")
						}
					}else{
					b.logger.Error("Database error", zap.Error(err))
					return c.Send("Произошла ошибка")
					}
				}
				return c.Send(fmt.Sprintf("Название курса: %s\n",course.Title),yesNoMenu)
	})	
	b.bot.Handle(&btnNo, func(c tb.Context) error {
		callback := &tb.Callback{ID: c.Callback().ID, Message: c.Message()}
		b.bot.Respond(callback, &tb.CallbackResponse{Text: "Отправьте ссылку на курс заново"})
		b.bot.Send(c.Sender(), "Отправьте ссылку на курс заново")
		b.bot.Handle(tb.OnText, func(c tb.Context) error {
				var course parser.Course
				coursCh := make(chan parser.Course)
				errCh := make(chan error,2)
				go parser.Parser(c.Text(),errCh,coursCh)
				select {
				case course = <-coursCh:
					b.logger.Info("курс успешно найден")
				case err := <-errCh:
					b.logger.Error("Ошибка parser", zap.Error(err))
					return c.Send("Ошибка при чтении ссылки")
				}
				Id,err := b.db.GetCourseID(course.Title) 
				if err != nil {
					if err==sql.ErrNoRows{
						// create an connection with ml that give me course rating
						Id,err = b.db.InsertCourse(course.Title,course.Url,course.Knowledges,6,100)
						b.logger.Info("id",zap.Int("id",Id))
						if err != nil {
							b.logger.Error("Database error", zap.Error(err))
							return c.Send("Произошла ошибка")
						}
					}else{
					b.logger.Error("Database error", zap.Error(err))
					return c.Send("Произошла ошибка")
					}
				}
				return c.Send(fmt.Sprintf("Название курса: %s\n",course.Title),yesNoMenu)
		})
		return nil
	})

	b.bot.Handle(&btnYes, func(c tb.Context) error {
		callback := &tb.Callback{ID: c.Callback().ID, Message: c.Message()}
		b.bot.Respond(callback, &tb.CallbackResponse{Text: "Отправьте сертификат,подтверждающий прохождение курса"})
		c.Send("Отправьте сертификат,подтверждающий прохождение курса")
		b.bot.Handle(tb.OnPhoto, func(c tb.Context) error {
			photo := c.Message().Photo
			file,err:=b.bot.FileByID(photo.FileID)
			if err != nil {
				b.logger.Error("Ошибка при получении файла", zap.Error(err))
				return err
			}
			return b.handleFile(c, file)
			
		})
		return nil
	})
	return nil
	
}

func (b *Bot) Start(_ context.Context) error {
	b.bot.Start()
	return nil
}


func (b *Bot)downloadFile(url string, filePath string) error {
	b.logger.Info("Downloading file", zap.String("url", url), zap.String("path", filePath))
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func (b *Bot)handleFile(c tb.Context,file tb.File) error {

	questionMenu:=&tb.ReplyMarkup{ResizeKeyboard: true}
			btnQuestion:=questionMenu.Data("Пройти опрос по курсу","yes")
			questionMenu.Inline(
				questionMenu.Row(btnQuestion),
			)

	filepath:=fmt.Sprintf("downloads/%s",file.FileID+".jpg")
	fileUrl :="https://api.telegram.org/file/bot"+b.cfg.Token+"/"+file.FilePath
	err:= b.downloadFile(fileUrl, filepath)
	if err != nil {
		b.logger.Error("Ошибка при скачивании файла", zap.Error(err))
		return c.Send("Не удалось сохранить файл.")
	}
	c.Send("Идет верификация прохождения курса...")
	// errChan := make(chan error)
	result := make(chan string)
	go backgroundCheck(result)
	// for err := range errChan{
	// 	if err!=nil{
	// 		return c.Send("Произошла ошибка")
	// 	}
	// }
	if <-result == "ok"{
		//ПЕРЕДЕЛАТЬ В ОДИН МЕТОД С ТРАНЗАКЦИЯМИ
		if err != nil {
			if err == errors.ErrUserAlreadyExists{ 
				return c.Send("Вы уже прошли курс")
			}
				b.logger.Error("Database error", zap.Error(err))
				return c.Send("Произошла ошибка")
		}

		items,err:=b.db.GetUserItems(int(c.Sender().ID))
		if err != nil {
			b.logger.Error("Database error", zap.Error(err))
			return c.Send("Произошла ошибка")
		}
		var multiplier float64
		for i:=0;i<len(items);i++{
			multiplier+=items[i].Multiplier
		}
		if multiplier==0{
			multiplier=1
		}

		_,_,rating,coins,err:=b.db.GetCourseByID(Id)
		if err != nil {
			b.logger.Error("Database error", zap.Error(err))
			return c.Send("Произошла ошибка")
		}

		coins = int(float64(coins) * multiplier)
		exp:=int(float64(rating) * multiplier)
		err=b.db.UpdateUser(int(c.Sender().ID),exp,coins)
		if err != nil {
			b.logger.Error("Database error", zap.Error(err))
			return c.Send("Произошла ошибка")
		}
		err = b.db.InsertUserCourses(int(c.Sender().ID),Id)
		if err != nil {
			if err == errors.ErrUserCourseAlreadyExists{
				return c.Send("Вы уже прошли курс")
			}
			b.logger.Error("Database error", zap.Error(err))
			return c.Send("Произошла ошибка")
		}
		c.Send(fmt.Sprintf("Курс успешно прошел проверку, вам начислено %d монет и %d опыта ",coins,exp),questionMenu)
	}
	b.bot.Handle(&btnQuestion, func(c tb.Context) error {
		b.logger.Info("Пользователь зашел в опрос")
		err := b.rateHandler(c)
		if err != nil {
			return err
		}
		
		return nil
})
return nil
}

func (b *Bot) rateHandler(c tb.Context) error {
		topics,ids,_,err := b.db.GetTopics(Id)
		b.logger.Info("Пользователь получает топики", zap.Int("count1",len(ids)))
		b.logger.Info("Пользователь получает топики",zap.Int("count",len(topics)))
		if err != nil {
			b.logger.Error("Database error", zap.Error(err))
			return c.Send("Произошла ошибка")
		}
		for i:=0;i<len(ids);i++{
		gradeMenu := &tb.ReplyMarkup{}
		var gradeButtons []tb.Btn

		for j := 1; j <= 10; j++ {
			
			btn := gradeMenu.Data(fmt.Sprintf("%d ⭐", j),fmt.Sprintf("rate_%d_%d", ids[i], j))
			gradeButtons = append(gradeButtons, btn)
		}

		// Размещаем кнопки в 2 ряда
		gradeMenu.Inline(
			gradeMenu.Row(gradeButtons[:5]...),
			gradeMenu.Row(gradeButtons[5:]...),
		)
			b.logger.Info("Пользователь получает сообщение",zap.String("topic",topics[i]))
			c.Send(fmt.Sprintf("Тема: %s\n",topics[i]),gradeMenu)
		}
		b.bot.Handle(tb.OnCallback, func(c tb.Context) error {
    // Извлекаем данные: ожидаем формат "rate_<topic_id>_<rating>"
    var topicID, rating int
		data:=strings.TrimSpace(c.Callback().Data)
		b.logger.Info("Пользователь оценивает тему",zap.String("data",data))
    _, err := fmt.Sscanf(data, "rate_%d_%d", &topicID, &rating)
    if err != nil {
				b.logger.Error("Invalid callback data", zap.String("data", data))
        return c.Respond(&tb.CallbackResponse{Text: "Неверный формат данных!"})
    }

    // Сохраняем рейтинг в базе данных
    err = b.db.InsertGrade(int(c.Sender().ID), topicID, rating)
    if err != nil {
			if err == errors.ErrGradeAlreadyExists{
				return c.Respond(&tb.CallbackResponse{Text: "Вы уже оценивали эту тему"})
			}
        b.logger.Error("Database error", zap.Error(err))
        return c.Respond(&tb.CallbackResponse{Text: "Ошибка сохранения рейтинга"})
    }

    // Отвечаем на callback, можно отредактировать сообщение или отправить уведомление
    return c.Respond(&tb.CallbackResponse{Text: "Спасибо за оценку!"})
})

		return nil
}

// функция затычка для проверки фото 
func backgroundCheck(result chan string){
	time.Sleep(1*time.Second)
	result<- "ok"
}
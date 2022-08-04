package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"os/exec"
	"strings"
)

var (
	configPath string
)

type Config struct {
	Token  string `json:"TokenTG"`
	Debug  bool   `json:"DebugMode"`
	DB     string `json:"DBpatch"`
	Sctipt string `json:"SctiptPatch"`
}

func init() {
	// принимаем на входе флаг -c
	flag.StringVar(&configPath, "c", "/etc/rotateBot/config", "config")
	flag.Parse()
}

func main() {
	//инициализация чтения конфига
	var config Config
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal("невозможно прочитать конфиг")
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatal("ошибка в конфиге")
	}

	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		log.Panic("token: ", err)
	}

	bot.Debug = config.Debug

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			rtArg, err := scanDB(update.Message.Text, config.DB)
			if err != nil {
				continue
			}

			cmd := exec.Command(config.Sctipt, rtArg)
			result, err := cmd.Output()
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, string(result))
			msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)
		}
	}
}

func scanDB(g string, DB string) (string, error) {
	var result error = nil
	fileDB, err := os.Open(DB)
	if err != nil {
		log.Fatal(err)
	}
	defer fileDB.Close()

	var rtNumber string
	scanner := bufio.NewScanner(fileDB)
	for scanner.Scan() {
		var port []string
		port = strings.Split(scanner.Text(), "=")
		if port[0] == g {
			rtNumber = port[1]
			break
		}
	}
	if rtNumber == "" {
		result = errors.New("нет совпадений")
	}
	return rtNumber, result
}

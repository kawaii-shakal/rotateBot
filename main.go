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
	"regexp"
	"strings"
)

var (
	configPath string
)

type Config struct {
	Token   string  `json:"TokenTG"`
	Debug   bool    `json:"DebugMode"`
	DB      string  `json:"DBpatch"`
	Script  string  `json:"ScriptPatch"`
	UserACL []int64 `json:"UserAcl"`
}

func init() {
	// принимаем на входе флаг -c
	flag.StringVar(&configPath, "c", "/etc/rotateBot/config", "config")
	flag.Parse()
}

func readConfig(path string, config *Config) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal("невозможно прочитать конфиг")
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatal("ошибка в конфиге")
	}
}

func main() {

	//инициализация чтения конфига
	var config Config
	readConfig(configPath, &config)

	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		log.Panic("token: ", err)
	}

	//логирование в файл
	f, err := os.OpenFile("/var/log/rotateBot.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	bot.Debug = config.Debug

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			log.Println(update.Message)
			if false == checkACL(update.Message.From.ID, config.UserACL) {
				log.Println("не авторизован")
				continue
			}

			var msg tgbotapi.MessageConfig
			var result []byte

			var re = regexp.MustCompile(`(?m)^[0-9]{4}-[0-9]{4}$`)
			switch {
			case strings.Contains(update.Message.Text, "-") && re.MatchString(update.Message.Text):
				//todo: цикл который будет в потоках запускать скрипт
				fmt.Println("placeholder ", update.Message.Text)
			case len(update.Message.Text) == 4:
				rt, err := scanDB(update.Message.Text, config.DB)
				if err != nil {
					//todo: запись переменной ошибки в отправляемое сообщение
					fmt.Println("placeholder ")
				}
				result = reboot(config.Script, rt)
			case update.Message.Text == "all":
				//todo: парисинг всех роутеров из базы и заебашивание цикла
			}

			msg = tgbotapi.NewMessage(update.Message.Chat.ID, string(result))
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
		}
	}
}

func reboot(script string, rtArg string) []byte {
	cmd := exec.Command(script, rtArg)
	result, err := cmd.Output()
	if err != nil {
		log.Println(err.Error())
		return result
	}
	return result
}

func checkACL(i int64, users []int64) bool {
	result := false
	for _, a := range users {
		if a == i {
			//log.Println("совпадение с id: ", i)
			result = true
		}
	}
	return result
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

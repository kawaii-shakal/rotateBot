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
	"strconv"
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

func readConfig(config *Config) {
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
	readConfig(&config)

	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		log.Panic("token: ", err)

	}

	//логирование в файл
	f, err := os.OpenFile("/var/log/rotateBot.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
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

			var (
				msg        tgbotapi.MessageConfig
				result     string
				multiPort  = regexp.MustCompile(`(?m)^[0-9]{4}-[0-9]{4}$`)
				singlePort = regexp.MustCompile(`(?m)^[0-9]{4}$`)
				portDB     = createDB(config.DB)
			)

			switch {
			case multiPort.MatchString(update.Message.Text):
				firstPort, secondPort, err := getMultiPort(update.Message.Text, portDB)
				if err != nil {
					result = fmt.Sprintf("%s", err)
					err = nil
					break
				}

				if firstPort < secondPort {
					go func() {
						for i := firstPort; i <= secondPort; i++ {
							reboot(config.Script, i)
						}
					}()
				} else {
					go func() {
						for i := secondPort; i <= firstPort; i++ {
							reboot(config.Script, i)
						}
					}()
				}
				result = "send in rotation"

			case singlePort.MatchString(update.Message.Text):
				if rt, ok := portDB[update.Message.Text]; ok {
					result = reboot(config.Script, rt)
					break
				} else {
					result = "такого порта не найдено"
					break
				}

			case update.Message.Text == "all":
				//todo: парисинг всех роутеров из базы и заебашивание цикла
				log.Println("placeholder")
				break
			}

			msg = tgbotapi.NewMessage(update.Message.Chat.ID, result)
			msg.ReplyToMessageID = update.Message.MessageID
			_, err := bot.Send(msg)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func getMultiPort(x string, DB map[string]int) (int, int, error) {
	var (
		first  int
		second int
		err1   error = nil
		err2   error = nil
		err    error = nil
	)
	port := strings.Split(x, "-")

	if _, ok := DB[port[0]]; ok {
		first = DB[port[0]]
	} else {
		err1 = errors.New("нет порта: " + port[0])
	}
	if _, ok := DB[port[1]]; ok {
		second = DB[port[1]]
	} else {
		err2 = errors.New("нет порта: " + port[1])
	}

	switch {
	case (err1 != nil) && (err2 != nil):
		err = errors.New("портов " + port[0] + " и " + port[1] + " не существует")
		break
	case (err1 != nil) && (err2 == nil):
		err = err1
		break
	case (err1 == nil) && (err2 != nil):
		err = err2
		break
	}

	return first, second, err
}

func reboot(script string, rtArg int) string {
	cmd := exec.Command(script, strconv.Itoa(rtArg))
	result, err := cmd.Output()
	if err != nil {
		log.Println(err.Error())
		return string(result)
	}
	log.Println(result)
	return string(result)
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

func createDB(DB string) map[string]int {

	portsDB := map[string]int{}

	fileDB, err := os.Open(DB)
	if err != nil {
		log.Println(err)
	}
	defer fileDB.Close()

	scanner := bufio.NewScanner(fileDB)
	for scanner.Scan() {
		port := strings.Split(scanner.Text(), "=")
		portsDB[port[0]], _ = strconv.Atoi(port[1])
	}
	return portsDB
}

func _(f string, DB string) (string, error) {
	var (
		result   error = nil
		rtNumber string
	)

	fileDB, err := os.Open(DB)
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer fileDB.Close()

	scanner := bufio.NewScanner(fileDB)
	for scanner.Scan() {
		var port []string
		port = strings.Split(scanner.Text(), "=")
		if port[0] == f {
			rtNumber = port[1]
			break
		}
	}
	if rtNumber == "" {
		result = errors.New("нет совпадений")
	}
	return rtNumber, result
}

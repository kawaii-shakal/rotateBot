package main

import (
	"bufio"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	bot, err := tgbotapi.NewBotAPI("5337635051:AAEl0yMI14DOL-qbMBTxPe58pDIKe7HTW7U")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message
			//log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			//log.Println(update.Message.Text)
			//result := reboot(scanDB(update.Message.Text))
			cmd := exec.Command("/root/rebootLTE.sh", scanDB(update.Message.Text))
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

func scanDB(g string) string {
	fileDB, err := os.Open("/root/DB")
	if err != nil {
		log.Fatal(err)
	}
	defer fileDB.Close()

	var port []string
	scanner := bufio.NewScanner(fileDB)
	for scanner.Scan() {
		port = strings.Split(scanner.Text(), "=")
		if port[0] == g {
			break
		}
	}
	return port[1]
}

//func reboot(proxy string) string {
//	key, err := os.ReadFile("/root/.ssh/id_rsa")
//	//key, err := os.ReadFile("C:\\Users\\dima\\.ssh\\id_rsa")
//	if err != nil {
//		log.Fatalf("unable to read private key: %v", err)
//	}
//	signer, err := ssh.ParsePrivateKey(key)
//	if err != nil {
//		log.Fatalf("unable to parse private key: %v", err)
//	}
//
//	config := &ssh.ClientConfig{
//		User: "root",
//		Auth: []ssh.AuthMethod{
//			// Use the PublicKeys method for remote authentication.
//			ssh.PublicKeys(signer),
//		},
//		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
//	}
//
//	client, err := ssh.Dial("tcp", "188.232.207.52:8888", config)
//	if err != nil {
//		log.Fatalf("unable to connect: %v", err)
//	}
//	defer client.Close()
//
//	session, err := client.NewSession()
//	if err != nil {
//		log.Fatal("Failed to create session: ", err)
//	}
//	defer session.Close()
//
//	var b bytes.Buffer
//	session.Stdout = &b
//	command := "~/rebootLTE.sh " + proxy
//	if err := session.Run(command); err != nil {
//		log.Fatal("Failed to run: " + err.Error())
//	}
//	return b.String()
//}

// Merci de ne pas voler le code mais de le forked. Bonne utilisation !

/*
	author: AslakOffi
	company: HexalLuna | https://github.com/HexalLuna
*/

package main

import (
	"flag"
	"log"
	"net"

	"github.com/marcusolsson/tui-go"
)

var buf [512]byte

func startClientUI(serverIp string, port string) {
	conn := openConnection(serverIp, port)
	ui, messageArea := initUI(conn)

	go uiReceiveMessagesRoutine(conn, ui, messageArea)
	if err := ui.Run(); err != nil {
		log.Fatal(err)
	}
	closeConnection(conn)
}

func openConnection(serverIp string, port string) (conn net.Conn) {
	log.Println("Connexion en cours...")
	addr := serverIp + ":" + port
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatal("Connexion: " + addr + " a échoué.")
	}
	log.Println("Connexion établie.")
	return
}

func initUI(conn net.Conn) (tui.UI, *tui.Box) {
	messageArea := tui.NewVBox()
	messageAreaScroll := tui.NewScrollArea(messageArea)
	messageAreaScroll.SetAutoscrollToBottom(true)
	messageAreaBox := tui.NewVBox(messageAreaScroll)
	messageAreaBox.SetBorder(true)

	input := tui.NewEntry()
	input.SetFocused(true)
	input.SetSizePolicy(tui.Expanding, tui.Maximum)

	inputBox := tui.NewHBox(input)
	inputBox.SetBorder(true)
	inputBox.SetSizePolicy(tui.Expanding, tui.Maximum)

	chat := tui.NewVBox(messageAreaBox, inputBox)
	chat.SetSizePolicy(tui.Expanding, tui.Expanding)

	input.OnSubmit(func(e *tui.Entry) {
		message := e.Text()
		err := sendMessage(conn, message)
		if err != nil {
			message = "Vous n’êtes pas connecté au serveur, vous ne pouvez pas envoyer de messages."
			messageArea.Append(tui.NewHBox(tui.NewLabel(message), tui.NewSpacer()))
		}
		input.SetText("")
	})

	root := tui.NewHBox(chat)

	ui, err := tui.New(root)
	if err != nil {
		log.Fatal(err)
	}

	ui.SetKeybinding("Esc", func() {
		ui.Quit()
		err = sendMessage(conn, "!q\n")
	})

	return ui, messageArea
}

func uiReceiveMessagesRoutine(conn net.Conn, ui tui.UI, messageArea *tui.Box) {
	for {
		message, err := receiveMessage(conn)
		if err != nil {
			message = "Vous vous êtes déconnecté du serveur, vous ne recevez plus de messages."
			ui.Update(func() {
				messageArea.Append(tui.NewHBox(tui.NewLabel(message), tui.NewSpacer()))
			})
			return
		}
		if message == "!q\n" {
			break
		}
		ui.Update(func() {
			messageArea.Append(tui.NewHBox(tui.NewLabel(message), tui.NewSpacer()))
		})
	}
}

func sendMessage(conn net.Conn, message string) error {
	_, err := conn.Write([]byte(message))
	return err
}

func receiveMessage(conn net.Conn) (message string, err error) {
	n, err := conn.Read(buf[0:])
	if err != nil {
		return
	}
	message = string(buf[0:n])
	return
}

func closeConnection(conn net.Conn) {
	log.Println("Déconnexion en cours...")
	err := conn.Close()
	if err != nil {
		log.Println("La déconnexion a échoué")
	}
	log.Println("Déconnexion réussie")
}

func main() {
	var serverIp string
	var port string

	flag.StringVar(&serverIp, "ip", "127.0.0.1", "Adresse IP du serveur de chat")
	flag.StringVar(&port, "port", "8001", "Port du serveur de chat")
	flag.Parse()

	startClientUI(serverIp, port)
}

// Merci de ne pas voler le code mais de le forked. Bonne utilisation !

/*
	author: AslakOffi
	company: HexalLuna | https://github.com/HexalLuna
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

var clients = make(map[string]net.Conn)
var lock = sync.RWMutex{}
var buf [512]byte

func startServer(port string) error {
	ln, err := listen(port)
	if err != nil {
		return fmt.Errorf("\nuimpossible de démarrer le serveur: %v\n", err)
	}
	err = acceptConnections(ln)
	return err
}

func listen(port string) (net.Listener, error) {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, fmt.Errorf("\nimpossible de trouver le port: %s en raison d'une erreur: %v", port, err)
	}
	log.Println("Ecoute sur adresse: " + ln.Addr().String())
	return ln, nil
}

func acceptConnections(ln net.Listener) error {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Échec de l'acceptation à la demande de connexion.")
			continue
		}
		log.Println("Client connecté avec l'adresse: " + conn.RemoteAddr().String())
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	log.Println("Traitement client...")

	clientName, err := getClientName(conn)
	if err != nil {
		log.Println("Erreur de connexion: " + conn.RemoteAddr().String() + ": " + err.Error())
		return
	}
	if clientName == "!q\n" {
		sendMessage(conn, "!q\n")
		log.Println("Fermeture de la connexion...")
		err := conn.Close()
		if err != nil {
			log.Println("La fermeture de la connexion a échoué.")
		}
		log.Println("Connexion avec l'adresse: " + conn.RemoteAddr().String() + " fermée.")
		return
	}

	err = echoMessages(conn, clientName)
	if err != nil {
		log.Println("Erreur de connexion: + " + conn.RemoteAddr().String() + ": " + err.Error())
		return
	}

	closeConnection(conn, clientName)
}

func getClientName(conn net.Conn) (string, error) {
	clientName := ""
	sendMessage(conn, "Vous êtes connecté au serveur, choisissez un nom d'utilisateur. Appuyez sur Echap pour quitter.")

	for {
		receivedName, err := receiveMessage(conn)
		if err != err {
			return "", err
		}
		if receivedName == "!q\n" {
			return "!q\n", nil
		}
		receivedName = strings.TrimRight(receivedName, "\n")
		lock.RLock()
		_, in := clients[receivedName]
		lock.RUnlock()
		if !in {
			lock.Lock()
			clients[receivedName] = conn
			lock.Unlock()
			clientName = receivedName
			break
		}
		sendMessage(conn, "Le nom est déjà pris, veuillez en choisir un autre.")
	}

	sendMessage(conn, "Bienvenue dans la room, "+clientName)
	lock.RLock()
	for name, conn := range clients {
		if name != clientName {
			go sendMessage(conn, clientName+" a joint la room.")
		}
	}
	lock.RUnlock()

	return clientName, nil
}

func echoMessages(conn net.Conn, clientName string) error {
	pre := clientName + ": "
	for {
		message, err := receiveMessage(conn)
		if err != nil {
			lock.Lock()
			delete(clients, clientName)
			lock.Unlock()

			broadcastMessage(clientName + " a quitté la room.")
			return err
		}
		if message == "!q\n" {
			sendMessage(conn, "!q\n")
			return nil
		} else {
			broadcastMessage(pre + message)
		}
	}
}

func closeConnection(conn net.Conn, clientName string) {
	log.Println("Déconnexion...")
	err := conn.Close()
	if err != nil {
		log.Println("La déconnexion a échoué")
	}
	log.Println("Connexion avec l'adresse: " + conn.RemoteAddr().String() + " fermée.")

	lock.Lock()
	delete(clients, clientName)
	lock.Unlock()

	broadcastMessage(clientName + " a quitté la room.")
}

func broadcastMessage(message string) {
	lock.RLock()
	for _, conn := range clients {
		go sendMessage(conn, message)
	}
	lock.RUnlock()
}

func sendMessage(conn net.Conn, message string) {
	_, err := conn.Write([]byte(message))
	if err != nil {
		log.Println("Impossible d'écrire sur cette connexion.")
	}
}

func receiveMessage(conn net.Conn) (message string, err error) {
	n, err := conn.Read(buf[0:])
	if err != nil {
		return
	}
	message = string(buf[0:n])
	return
}

func main() {
	var port string

	flag.StringVar(&port, "port", "8001", "port sur lequel le serveur verra les nouvelles connexions")
	flag.Parse()

	err := startServer(port)
	if err != nil {
		log.Fatal(err)
	}
}

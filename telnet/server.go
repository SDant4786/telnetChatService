package telnet

import (
	"bufio"
	"chatservice/config"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

// Holds the help menu options for printing
var helpMenu = map[string]string{
	"/quit":           "quit chat\n",
	"/listchannels":   "list all channels\n",
	"/listusers":      "list all active users\n",
	"/create":         "create a new channels\n",
	"/join":           "join a channels\n",
	"/leave":          "leave a channels\n",
	"/ignoreuser":     "ignore messsages from a user\n",
	"/unignoreuser":   "receive messages from ignored user\n",
	"/pm":             "send private message to user\n",
	"/sendchannel":    "send message into channel\n",
	"/listmychannels": "list channels you're subscribed to\n",
	"/help":           "display help menu\n",
}

var Channels = map[string][]*User{} // Map to store channel names and the users in the channel
var Users = map[string]*User{}      // Map of all users. (map instead of slice for simpler lookups and deletes)

// Thread safe counter for stats
type Counter struct {
	mu sync.Mutex
	C  int
}

// Counter of messages sent
var MessagesSent = Counter{
	sync.Mutex{},
	0,
}

// Inits the telnet server
func InitTelnetServer(cfg config.Config, shutdown <-chan os.Signal, wg *sync.WaitGroup) {

	listener, err := net.Listen("tcp", cfg.TelNetIp+":"+cfg.TelNetPort)
	if err != nil {
		log.Fatalf("could not create telnet server %v", err)
	}
	defer listener.Close()
	log.Printf("created telnet server")
	wg.Done()
	for {
		select {
		case <-shutdown:
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				log.Fatalf("could not accept new connection %v", err)
			}

			go CreateUser(conn)
		}
	}
}

// Called when a user connects to the server to create an account
func CreateUser(conn net.Conn) {
	log.Printf("creating new user. conn: %v", conn.RemoteAddr())

	for {
		//Get user name
		username, err := ReadInput(conn, "Enter username: ")
		if err != nil {
			log.Fatalf("error reading input. err: %s", err)
		}

		//If user name alrady exists, get a new one
		if _, ok := Users[username]; ok {
			conn.Write([]byte("user already exists, please pick another user name\n"))
		} else {
			//Create new user
			user := &User{
				username:    username,
				conn:        conn,
				messageChan: make(chan string),
				channels:    []string{},
				ignored:     []*User{},
				closeChan:   make(chan bool),
			}
			Users[username] = user

			log.Printf("new user created. conn: %v, username: %s", user.conn.RemoteAddr(), user.username)

			//Start go routines for user
			go user.ReadFromCLI()
			go user.ReceiveMessage()
			err := PrintHelpMenu(user.conn)
			if err != nil {
				log.Fatalf("unable to print help menu. err:%s", err)
			}

			_, err = user.conn.Write([]byte("Welcome to the chat serivce\n"))
			if err != nil {
				log.Fatalf("unable to write welcome message. err:%s", err)
			}
			break
		}
	}
}

// Used to prompt user for input as well as just reading what they submit through the cli
func ReadInput(conn net.Conn, msg string) (string, error) {
	conn.Write([]byte(msg))
	s, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		log.Printf("readinput: could not read input from stdin: %v from client %v", err, conn.RemoteAddr().String())
		return "", err
	}
	s = strings.Trim(s, "\r\n")
	return s, nil
}

// Print the help menu to the user
func PrintHelpMenu(conn net.Conn) error {
	_, err := conn.Write([]byte("/**************Help Menu***************/\n"))
	if err != nil {
		return err
	}
	for k, v := range helpMenu {
		_, err = conn.Write([]byte(k + " | " + v))
		if err != nil {
			return err
		}
	}
	_, err = conn.Write([]byte("/**************************************/\n"))

	if err != nil {
		return err
	}
	return nil
}

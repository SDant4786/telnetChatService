package telnet

import (
	"errors"
	"log"
	"net"
	"strings"
	"time"
)

type User struct {
	username    string
	conn        net.Conn
	messageChan chan string
	channels    []string
	ignored     []*User
	closeChan   chan bool
}

const timeFormat = "02/01/2006 15:04:05" // Used to format the timestamp consistently

// Reads input from CLI, checks if its a command, if not, sends message to chat room
func (u *User) ReadFromCLI() {
	for {
		select {
		case <-u.closeChan:
			log.Printf("messege sending channel closed for user: %s", u.username)
			return
		default:
			msg, err := ReadInput(u.conn, "")
			if err != nil {
				_, err = u.conn.Write([]byte("unable to read input from cli. error: " + err.Error()))
				if err != nil {
					log.Printf("error writing to connection %v. error %s", u.conn.RemoteAddr(), err)
				}
				close(u.closeChan)
			}

			if len(msg) == 0 {
				continue
			}
			//Check for command input
			if msg[0] == '/' {
				err = u.commandHandler(msg)
				if err != nil {
					_, err = u.conn.Write([]byte("invalid command. error: " + err.Error() + "\r\n"))
					if err != nil {
						log.Printf("error writing to connection %v. error %s", u.conn.RemoteAddr(), err)
					}
				}
			} else { //Send to all users
				msgToSend := time.Now().Format(timeFormat) + "|" + u.username + "|" + msg

				for _, user := range Users {
					user.messageChan <- msgToSend
				}
				log.Printf("message sent to all: %s", msgToSend)

				MessagesSent.mu.Lock()
				MessagesSent.C++
				MessagesSent.mu.Unlock()
			}
		}
	}
}

// Reads from a users message channel, if the message is sent from an unblocked user, display it.
func (u *User) ReceiveMessage() {
	for {
		select {
		case <-u.closeChan:
			log.Printf("receive channel closed for user: %s", u.username)
			return
		case msg := <-u.messageChan:
			//Check for ignored user
			split := strings.Split(msg, "|")
			user := split[1]

			ignored := false
			for _, ignoredUser := range u.ignored {
				if user == ignoredUser.username {
					ignored = true
					break
				}
			}
			if !ignored {
				_, err := u.conn.Write([]byte(msg + "\n"))
				if err != nil {
					log.Printf("error writing to connection %v. error %s", u.conn.RemoteAddr(), err)
				}
			}
		}
	}
}

// Switch statement for handling all command inputs
func (u *User) commandHandler(msg string) error {
	switch msg {
	case "/quit":
		err := u.quit()
		if err != nil {
			return err
		}
	case "/listchannels":
		err := u.listChannels()
		if err != nil {
			return err
		}
	case "/listusers":
		err := u.listUsers()
		if err != nil {
			return err
		}
	case "/create":
		err := u.createChannel()
		if err != nil {
			return err
		}
	case "/join":
		err := u.joinChannel()
		if err != nil {
			return err
		}
	case "/leave":
		err := u.leaveChannel()
		if err != nil {
			return err
		}
	case "/ignoreuser":
		err := u.ignoreUser()
		if err != nil {
			return err
		}
	case "/unignoreuser":
		err := u.unIgnoreUser()
		if err != nil {
			return err
		}
	case "/pm":
		err := u.sendPM()
		if err != nil {
			return err
		}
	case "/sendchannel":
		err := u.sendIntoChannel()
		if err != nil {
			return err
		}
	case "/listmychannels":
		err := u.listMyChannels()
		if err != nil {
			return err
		}
	case "/help":
		PrintHelpMenu(u.conn)
	default:
		return errors.New("unknown command")
	}
	return nil
}

// Disconnects user from server and closes their go routines
func (u *User) quit() error {
	//Delete from channels
	for _, uc := range u.channels {
		if users, ok := Channels[uc]; ok {
			for i, user := range users {
				if user.username == u.username {
					Channels[uc] = append(Channels[uc][:i], Channels[uc][i+1:]...)
				}
			}
		}
	}
	//Delete from user list
	delete(Users, u.username)
	//Signal go routines to stop
	close(u.closeChan)

	_, err := u.conn.Write([]byte("You have quit the chat. Escape character is '^]', then enter 'close' to exit telnet"))

	if err != nil {
		return err
	}
	log.Printf("user: %s quit", u.username)
	return nil
}

// Displays available channels
func (u *User) listChannels() error {
	_, err := u.conn.Write([]byte("/**************Channels****************/\n"))
	if err != nil {
		return err
	}
	for ch := range Channels {
		_, err = u.conn.Write([]byte(ch + "\n"))
		if err != nil {
			return err
		}
	}
	_, err = u.conn.Write([]byte("/**************************************/\n"))

	if err != nil {
		return err
	}
	return nil
}

// Displays available users for pms
func (u *User) listUsers() error {
	_, err := u.conn.Write([]byte("/****************Users*****************/\n"))
	if err != nil {
		return err
	}
	for user := range Users {
		_, err = u.conn.Write([]byte(user + "\n"))
		if err != nil {
			return err
		}
	}
	_, err = u.conn.Write([]byte("/**************************************/\n"))

	if err != nil {
		return err
	}
	return nil
}

// Create a new channel
func (u *User) createChannel() error {
	channelName, err := ReadInput(u.conn, "Enter new channel name: ")
	if err != nil {
		return err
	}
	if _, ok := Channels[channelName]; !ok {
		Channels[channelName] = []*User{}
	} else {
		return errors.New("channel already exists")
	}
	_, err = u.conn.Write([]byte("Channel: " + channelName + " created \n"))
	if err != nil {
		return err
	}
	return nil
}

// Join a channel
func (u *User) joinChannel() error {
	channelName, err := ReadInput(u.conn, "Enter channel name to join: ")
	if err != nil {
		return err
	}
	if _, ok := Channels[channelName]; ok {
		for _, c := range u.channels {
			if c == channelName {
				_, err = u.conn.Write([]byte("Already in channel: " + channelName + " \n"))
				if err != nil {
					return err
				}
				return nil
			}
		}
		Channels[channelName] = append(Channels[channelName], u)
		u.channels = append(u.channels, channelName)
	} else {
		return errors.New("channel does not exist")
	}
	_, err = u.conn.Write([]byte("Joined channel: " + channelName + " \n"))
	if err != nil {
		return err
	}
	return nil
}

// Leave a channel
func (u *User) leaveChannel() error {
	channelName, err := ReadInput(u.conn, "Enter channel to leave: ")
	if err != nil {
		return err
	}
	if userList, ok := Channels[channelName]; ok {
		for i, user := range userList {
			if user.username == u.username {
				Channels[channelName] = append(Channels[channelName][:i], Channels[channelName][i+1:]...)
				break
			}
		}
		for i, c := range u.channels {
			if c == channelName {
				u.channels = append(u.channels[:i], u.channels[i+1:]...)
				break
			}
		}
		_, err := u.conn.Write([]byte("Left channel: " + channelName + " \n"))
		if err != nil {
			return err
		}
	} else {
		return errors.New("channel does not exist")
	}
	return nil
}

// Add user to ignore list
func (u *User) ignoreUser() error {
	userName, err := ReadInput(u.conn, "Enter user to ignore: ")
	if err != nil {
		return err
	}
	if user, ok := Users[userName]; ok {
		u.ignored = append(u.ignored, user)
	} else {
		return errors.New("user does not exist")
	}
	_, err = u.conn.Write([]byte("Ignored user: " + userName + " \n"))
	if err != nil {
		return err
	}
	return nil
}

// Remove user from ignore list
func (u *User) unIgnoreUser() error {
	userName, err := ReadInput(u.conn, "Enter user to unignore: ")
	if err != nil {
		return err
	}
	found := false
	for i, ignoredUser := range u.ignored {
		if ignoredUser.username == userName {
			u.ignored = append(u.ignored[:i], u.ignored[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return errors.New("user does not exist")
	}
	_, err = u.conn.Write([]byte("Unignored user: " + userName + " \n"))
	if err != nil {
		return err
	}
	return nil
}

// Send a private message
func (u *User) sendPM() error {
	user, err := ReadInput(u.conn, "Enter user to send pm to: ")
	if err != nil {
		return err
	}
	if user, ok := Users[user]; ok {
		msg, err := ReadInput(u.conn, "Enter message: ")
		if err != nil {
			return err
		}
		msgToSend := time.Now().Format(timeFormat) + "|" + u.username + "|" + msg
		user.messageChan <- msgToSend
		log.Printf("message sent to pm: %s", msgToSend)
		MessagesSent.mu.Lock()
		MessagesSent.C++
		MessagesSent.mu.Unlock()
		return nil
	} else {
		return errors.New("user does not exist")
	}
}

// Send into channel
func (u *User) sendIntoChannel() error {
	channel, err := ReadInput(u.conn, "Enter channel to send message to: ")
	if err != nil {
		return err
	}
	if userList, ok := Channels[channel]; ok {
		msg, err := ReadInput(u.conn, "Enter message: ")
		if err != nil {
			return err
		}
		msgToSend := time.Now().Format(timeFormat) + "|" + u.username + "|" + channel + "|" + msg
		for _, user := range userList {
			user.messageChan <- msgToSend
		}
		log.Printf("message sent to channel: %s", msgToSend)
		MessagesSent.mu.Lock()
		MessagesSent.C++
		MessagesSent.mu.Unlock()
		return nil
	} else {
		return errors.New("channel does not exist")
	}
}

// List channels a user is subscribed to
func (u *User) listMyChannels() error {
	_, err := u.conn.Write([]byte("/************My Channels***************/\n"))
	if err != nil {
		return err
	}
	for _, ch := range u.channels {
		_, err = u.conn.Write([]byte(ch + "\n"))
		if err != nil {
			return err
		}
	}
	_, err = u.conn.Write([]byte("/**************************************/\n"))

	if err != nil {
		return err
	}
	return nil
}

// Sends message from http to a channel
func HTTPSendChannelMessage(msg string, channel string, userList []*User) {
	msgToSend := time.Now().Format(timeFormat) + "|http|" + channel + "|" + msg
	for _, user := range userList {
		user.messageChan <- msgToSend
	}
	MessagesSent.mu.Lock()
	MessagesSent.C++
	MessagesSent.mu.Unlock()
	log.Printf("message sent to channel: %s", msgToSend)
}

// Sends message from http to a specific user
func HTTPSendUserMessage(msg string, user *User) {
	msgToSend := time.Now().Format(timeFormat) + "|http|" + msg
	user.messageChan <- msgToSend
	MessagesSent.mu.Lock()
	MessagesSent.C++
	MessagesSent.mu.Unlock()
	log.Printf("message sent to pm: %s", msgToSend)
}

// Sends message from http to a all users
func HTTPSendAllMessage(msg string) {
	msgToSend := time.Now().Format(timeFormat) + "|http|" + msg
	for _, user := range Users {
		user.messageChan <- msgToSend
	}
	MessagesSent.mu.Lock()
	MessagesSent.C++
	MessagesSent.mu.Unlock()
	log.Printf("message sent to all: %s", msgToSend)
}

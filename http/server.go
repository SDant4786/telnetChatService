package http

import (
	"chatservice/config"
	"chatservice/telnet"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type submitPost struct {
	Channel string
	Message string
	User    string
}

var logfile string

// Spin up handler and start server
func InitHttpServer(cfg config.Config) {
	logfile = cfg.LogFile
	//Spin up handlers and server
	http.HandleFunc("/submitMessage", submitMessage)
	http.HandleFunc("/getLogs", getLogs)
	http.HandleFunc("/stats", getStats)
	go http.ListenAndServe(cfg.HttpIp+":"+cfg.HttpPort, nil)
	log.Println("Created http server")
}

// Allows the http user to send in messages
func submitMessage(w http.ResponseWriter, r *http.Request) {
	var req submitPost
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println("json decoding error: ", err)
		return
	}

	if req.Channel != "" {
		if userList, ok := telnet.Channels[req.Channel]; ok {
			telnet.HTTPSendChannelMessage(req.Message, req.Channel, userList)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Channel does not exist"))
			return
		}
	} else if req.User != "" {
		if user, ok := telnet.Users[req.User]; ok {
			telnet.HTTPSendUserMessage(req.Message, user)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("User does not exist"))
			return
		}
	} else {
		telnet.HTTPSendAllMessage(req.Message)
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Message submitted successfully"))
}

// Allows the http user to get their messages
func getLogs(w http.ResponseWriter, r *http.Request) {
	contents, err := os.ReadFile(logfile)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("file reading error: ", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(contents))
}

// Returns stats for the chat service
func getStats(w http.ResponseWriter, r *http.Request) {
	retMap := map[string]int{
		"users":         len(telnet.Users),
		"channels":      len(telnet.Channels),
		"messages_sent": telnet.MessagesSent.C,
	}
	ret, err := json.Marshal(retMap)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("json marshelling error: ", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(ret)
}

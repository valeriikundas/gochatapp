package main

import (
	"encoding/json"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2/log"
	"gorm.io/gorm"
)

// userID to websocket connection
var websocketConnections = map[uint]*websocket.Conn{}

type BaseMessageSchema struct {
	Type string
}

type JoinChatRequestSchema struct {
	BaseMessageSchema

	ChatID uint
	UserID uint
}

type SendMessageRequestSchema struct {
	BaseMessageSchema

	ChatID  uint
	UserID  uint
	Message string
}

type BroadcastMessageSchema struct {
	BaseMessageSchema

	FromUserEmail string
	Message       string
}

func WebsocketHandler(c *websocket.Conn) {
	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		log.Fatal("error getting `db` from c.Locals()")
	}

	for {
		messageType, message, err := c.ReadMessage()
		if err != nil {
			log.Debugf("read error:", err)
			break
		}
		log.Debugf("recv: %d %+v", messageType, string(message))

		var v BaseMessageSchema
		err = json.Unmarshal(message, &v)
		if err != nil {
			log.Fatal("json unmarshall WebsocketRequestSchema error: %s\n", err)
		}

		switch messageType := v.Type; messageType {
		case "join_chat":
			handleJoinChat(c, message)

		case "send_message":
			// get users that are in chat and not userID
			handleSendMessage(db, message)

		default:
			log.Errorf("unhandled message type=%s v=%s\n", messageType, v)
		}
	}
}

func handleSendMessage(db *gorm.DB, message []byte) {
	var requestData SendMessageRequestSchema
	err := json.Unmarshal(message, &requestData)
	if err != nil {
		log.Fatalf("json unmarshall SendMessageRequestSchema error: %s\n", err)
	}
	log.Debugf("`send message` message=%+v\n", requestData)

	messageContent := string(requestData.Message)
	log.Debugf("messageContent=%s\n", messageContent)
	messageObj := Message{
		ChatID:  requestData.ChatID,
		FromID:  requestData.UserID,
		Content: messageContent,
	}
	tx := db.Create(&messageObj)
	if tx.Error != nil {
		log.Fatal(tx.Error)
	}

	userIDsToSendMessageTo, err := getChatUsersExcept(db, requestData.ChatID, requestData.UserID)
	if err != nil {
		log.Fatalf("getChatUsersExcept err=%s\n", err)
	}
	log.Debugf("will send message to userIDsToSendMessageTo=%+v\n", userIDsToSendMessageTo)

	for _, userID := range userIDsToSendMessageTo {
		sendMessageToUser(db, userID, messageContent)
	}
}

func sendMessageToUser(db *gorm.DB, userID uint, messageContent string) {
	memberConn := websocketConnections[userID]

	if memberConn == nil {
		log.Debugf("no connection for userID=%d\n", userID)
		return
	}

	twentySecondFromNow := time.Now().Add(time.Second * 10)
	err := memberConn.WriteControl(websocket.PingMessage, []byte("hello from the other side"), twentySecondFromNow)
	if err != nil {
		log.Fatal(err)
	}

	var fromUser User
	err = db.Find(&fromUser, userID).Error
	if err != nil {
		log.Fatalf("get user by id failed id=%d\n", userID)
	}

	broadcastMessageData := BroadcastMessageSchema{
		BaseMessageSchema: BaseMessageSchema{
			Type: "new_message",
		},
		FromUserEmail: fromUser.Email,
		Message:       messageContent,
	}
	b, err := json.Marshal(broadcastMessageData)
	if err != nil {
		log.Fatalf("json marshall err:%s\n", err)
	}

	err = memberConn.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		log.Fatalf("error WriteMessage %s\n", err)
	}
}

func handleJoinChat(c *websocket.Conn, message []byte) {
	var requestData JoinChatRequestSchema
	err := json.Unmarshal(message, &requestData)
	if err != nil {
		log.Fatalf("json unmarshall JoinChatRequestSchema error: %s\n", err)
	}
	log.Debugf("`join chat` message=%+v\n", requestData)

	userID := requestData.UserID
	_, exists := websocketConnections[userID]
	if exists {
		websocketConnections[userID].Close()
	}
	websocketConnections[userID] = c

	// TODO: broadcast to other users in chat, than a new user has joined
}

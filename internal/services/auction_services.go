package services

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type MessageKind int

const (
	PlaceBid MessageKind = iota
	//OK
	SuccessfullyPlacedBid
	//Errors
	FailedToPlaceBid
	InvalidJson
	//Info
	NewBidPlaced
	//Auction Finished
	AuctionFinished
)

type Message struct {
	Message string      `json:"message,omitempty"`
	Kind    MessageKind `json:"kind"`
	UserId  uuid.UUID   `json:"user_id,omitempty"`
	Amount  float64     `json:"amount,omitempty"`
}

type AuctionLobby struct {
	sync.Mutex
	Rooms map[uuid.UUID]*AuctionRoom
}

type AuctionRoom struct {
	Id          uuid.UUID
	Clients     map[uuid.UUID]*Client
	Register    chan *Client
	Unregister  chan *Client
	Broadcast   chan Message
	Context     context.Context
	BidsService BidsService
}

type Client struct {
	Conn   *websocket.Conn
	UserId uuid.UUID
	Send   chan Message
	Room   *AuctionRoom
}

func NewAuctionRoom(ctx context.Context, id uuid.UUID, BidsService BidsService) *AuctionRoom {
	return &AuctionRoom{
		Id:          id,
		Broadcast:   make(chan Message),
		Clients:     make(map[uuid.UUID]*Client),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		Context:     ctx,
		BidsService: BidsService,
	}
}

func NewClient(conn *websocket.Conn, userId uuid.UUID, room *AuctionRoom) *Client {
	return &Client{
		Conn:   conn,
		Room:   room,
		Send:   make(chan Message, 512),
		UserId: userId,
	}
}

// Auction Room Methods

func (r *AuctionRoom) registerClient(client *Client) {
	slog.Info("New User Connected!", "Client", client)
	r.Clients[client.UserId] = client
}

func (r *AuctionRoom) unregisterClient(client *Client) {
	slog.Info("User Disconnected!", "Client", client)
	delete(r.Clients, client.UserId)
}

func (r *AuctionRoom) broadcastMessage(m Message) {
	slog.Info("New message received", "RoomID", r.Id, "Message", m.Message, "Kind", m.Kind, "UserId", m.UserId)
	switch m.Kind {
	case PlaceBid:
		slog.Info("New Bid Placed", "Bid", m.Amount)
		bid, err := r.BidsService.PlaceBid(r.Context, r.Id, m.UserId, m.Amount)
		if err != nil {
			if errors.Is(err, ErrBidIsTooLow) {
				slog.Info("Bid is too low", "Bid", m.Amount)
				if client, ok := r.Clients[m.UserId]; ok {
					client.Send <- Message{
						Message: ErrBidIsTooLow.Error(),
						Kind:    FailedToPlaceBid,
						UserId:  m.UserId,
					}
				}
			}
			return
		}
		if client, ok := r.Clients[m.UserId]; ok {
			client.Send <- Message{
				Message: "Your bid was placed successfully!",
				Kind:    SuccessfullyPlacedBid,
				UserId:  m.UserId,
			}
		}
		for id, client := range r.Clients {
			newBidMessage := Message{
				Message: "A new bid was placed",
				Kind:    NewBidPlaced,
				Amount:  bid.BidAmount,
				UserId:  m.UserId,
			}
			if id == m.UserId {
				continue
			}
			client.Send <- newBidMessage
		}
	case InvalidJson:
		client, ok := r.Clients[m.UserId]
		if !ok {
			slog.Info("Client not found for invalid json message", "UserId", m.UserId)
			return
		}
		client.Send <- m
	}
}

func (room *AuctionRoom) Run() {
	slog.Info("Auction Room started for product", "Product Id:", room.Id)
	defer func() {
		close(room.Broadcast)
		close(room.Register)
		close(room.Unregister)
	}()

	for {
		select {
		case client := <-room.Register:
			room.registerClient(client)
		case client := <-room.Unregister:
			room.unregisterClient(client)
		case message := <-room.Broadcast:
			room.broadcastMessage(message)
		case <-room.Context.Done():
			slog.Info("Auction has ended", "Product Id:", room.Id)
			for _, client := range room.Clients {
				client.Send <- Message{
					Message: "Auction has ended",
					Kind:    AuctionFinished,
				}
			}
			return
		}
	}
}

// Client Constants
const (
	maxMessageSize = 512
	readDeadline   = 60 * time.Second
	pingPeriod     = (readDeadline * 9) / 10
	writeWait      = 10 * time.Second
)

// Client Methods
func (c *Client) ReadEventLoop() {
	defer func() {
		c.Room.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(readDeadline))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(readDeadline))
		return nil
	})

	for {
		var m Message
		m.UserId = c.UserId
		err := c.Conn.ReadJSON(&m)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				slog.Error("Unexpected close Error. Error in reading message from client", "error", err)
			}
			c.Room.Broadcast <- Message{
				Message: "this message could not be processed. invalid json format",
				Kind:    InvalidJson,
				UserId:  m.UserId,
			}
			continue
		}
		c.Room.Broadcast <- m
	}
}

func (c *Client) WriteEventLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteJSON(Message{
					Kind:    websocket.CloseMessage,
					Message: "closing websocket conn",
				})
				return
			}
			if message.Kind == AuctionFinished {
				close(c.Send)
				return
			}
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := c.Conn.WriteJSON(message)
			if err != nil {
				slog.Error("Error in writing message to client", "error", err)
				c.Room.Unregister <- c
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				slog.Error("Error in sending ping message to client", "error", err)
				return
			}
			slog.Info("Ping sent to client", "UserId", c.UserId)

		}
	}
}

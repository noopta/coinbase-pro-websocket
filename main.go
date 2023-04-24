package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/preichenberger/go-coinbasepro/v2"
	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

type bitcoinData struct {
	DataType    string     `json:"clip"`
	ProductID   string     `json:"product_id"`
	Transaction [][]string `json:"changes"`
	TimeStamp   string     `json:"time"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func sendSMS(orderType string, orderAmount float64) {
	client := twilio.NewRestClient()

	params := &openapi.CreateMessageParams{}
	params.SetTo(os.Getenv("TO_PHONE_NUMBER"))
	params.SetFrom(os.Getenv("TWILIO_PHONE_NUMBER"))
	params.SetBody("A" + orderType + " order worth " + strconv.FormatFloat(orderAmount, 'E', -1, 32) + " was made")

	_, err := client.Api.CreateMessage(params)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("message sent successfully")
	}
}

func reader(conn *websocket.Conn) {
	for {
		// read in a message
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		// print out that message for clarity
		log.Println(string(p))

		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Println(err)
			return
		}

	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home HTTP")
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	// upgrade this connection to a WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	log.Println("Client Connected")
	err = ws.WriteMessage(1, []byte("Hi Client!"))
	if err != nil {
		log.Println(err)
	}

	reader(ws)
}

func setupRoutes() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/ws", wsEndpoint)
}

func dialServer() {
	ws, _, err := websocket.DefaultDialer.Dial("wss://ws-feed.exchange.coinbase.com", nil)
	if err != nil {
		panic(err)
	}
	defer ws.Close()

	if err != nil {
		fmt.Println("error reading")
		fmt.Println(err)
		return
	}

	subscribe := coinbasepro.Message{
		Type: "subscribe",
		Channels: []coinbasepro.MessageChannel{
			// coinbasepro.MessageChannel{
			// 	Name: "heartbeat",
			// 	ProductIds: []string{
			// 		"BTC-USD",
			// 	},
			// },
			coinbasepro.MessageChannel{
				Name: "level2",
				ProductIds: []string{
					"BTC-USD",
				},
			},
		},
	}

	if err := ws.WriteJSON(subscribe); err != nil {
		fmt.Println("error writing")
		println(err.Error())
		return
	}

	for {
		// Read a message from websocket connection
		_, msg, err := ws.ReadMessage()
		if err != nil {
			return
		}

		transactionData := bitcoinData{}

		meanBitcoinPriceToday := 27498.84

		err = json.Unmarshal([]byte(string(msg)), &transactionData)

		if err != nil {
			fmt.Println(err)
		}

		// fmt.Println(transactionData.Transaction)
		if transactionData.Transaction != nil {
			i := 0

			for i = 0; i < len(transactionData.Transaction); i++ {

				splitStrings := transactionData.Transaction[i]
				orderAmount, err := strconv.ParseFloat(splitStrings[2], 64)

				if err != nil {
					fmt.Println(err)
				}

				currentBitcoinPrice, err := strconv.ParseFloat(splitStrings[1], 64)

				if (meanBitcoinPriceToday / currentBitcoinPrice) > 1.10 {
					fmt.Println("Bitcoin price is >= 10% market price")
					fmt.Println(currentBitcoinPrice)
				}

				if (meanBitcoinPriceToday / currentBitcoinPrice) < 0.90 {
					fmt.Println("Bitcoin price is <= 10% below market price")
					fmt.Println(currentBitcoinPrice)
				}

				if splitStrings != nil && splitStrings[0] == "buy" && (orderAmount >= 200) {
					// buy order
					// check how much investments we've already made
					// if we can buy $500 and our monthly budget isn't complete, then buy
					sendSMS(splitStrings[0], orderAmount)
					fmt.Println("Whale buy")
					fmt.Println(transactionData)
				}

				if splitStrings != nil && splitStrings[0] == "sell" && (orderAmount >= 100) {
					// buy order
					// check how much investments we've already made
					// if we can buy $500 and our monthly budget isn't complete, then buy
					sendSMS(splitStrings[0], orderAmount)
					fmt.Println("Whale sell")
					fmt.Println(transactionData)
				}
			}
		}

		// fmt.Println("*********")
		// fmt.Println(transactionData)
		// fmt.Println("parsed transaction")
		// fmt.Println(string(msg))
		// fmt.Println("*********")

		// writeToFile(string(msg))

		// uncomment below if you need send message to remote server
		//if err = ws.WriteMessage(websocket.TextMessage, msg); err != nil {
		//  return
		//}
	}
}

func writeToFile(msg string) {
	f, err := os.Create("data.txt")

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	_, err2 := f.WriteString(msg)

	if err2 != nil {
		log.Fatal(err2)
	}
}

func main() {
	fmt.Println("Hello World")
	// setupRoutes()
	// sendSMS()
	// sendEmail()
	dialServer()
	// log.Fatal(http.ListenAndServe(":8080", nil))
}

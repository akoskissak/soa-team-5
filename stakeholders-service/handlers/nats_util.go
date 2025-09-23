package handlers

import (
	"context"
	"encoding/json"
	"log"

	stakeproto "stakeholders-service/proto/stakeholders"

	"github.com/nats-io/nats.go"
)

func SubscribePurchaseCheckout(natsConn *nats.Conn, stakeholdersServer *StakeholdersServer) {
	_, err := natsConn.Subscribe("purchase_publish", func(msg *nats.Msg) {
		var event stakeproto.UpdateBalanceRequest
		err := json.Unmarshal(msg.Data, &event)
		if err != nil {
			log.Printf("Failed to parse event: %v", err)
			return
		}

		var resp *stakeproto.UpdateBalanceResponse
		if event.Command == "SUBTRACT" {
			resp, err = stakeholdersServer.SubtractBalance(context.Background(), &event)
			} else {
			resp, err = stakeholdersServer.AddBalance(context.Background(), &event)
			return
		}
		if err != nil {
			log.Printf("Failed to update balance for user %s: %v", event.UserId, err)
			respEvent := map[string]interface{}{
				"userId": event.UserId,
				"amount": event.Amount,
				"status": "FAILED",
			}
			respBytes, _ := json.Marshal(respEvent)
			natsConn.Publish("purchase_reply", respBytes)
			return
		}

		log.Printf("Balance updated for user %s ", event.UserId)

		respEvent := map[string]interface{}{
			"userId": resp.UserId,
			"amount": resp.Amount,
			"status": resp.Status,
		}
		respBytes, _ := json.Marshal(respEvent)
		_ = natsConn.Publish("purchase_reply", respBytes)
	})
	if err != nil {
		log.Fatalf("Failed to subscribe to purchase_publish: %v", err)
	}
}

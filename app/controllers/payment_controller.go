package controllers

import (
	"crypto/sha512"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/midtrans/midtrans-go/snap"
	"github.com/novaru/go-shop/app/consts"
	"github.com/novaru/go-shop/app/models"
	"net/http"
	"os"
	"strconv"
)

func (server *Server) Midtrans(w http.ResponseWriter, r *http.Request) {
	var paymentNotification models.MidtransNotification

	err := json.NewDecoder(r.Body).Decode(&paymentNotification)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		res := Result{Code: http.StatusBadRequest, Message: err.Error()}
		response, _ := json.Marshal(res)

		w.Write(response)
		return
	}
	defer r.Body.Close()

	err = validateSignatureKey(&paymentNotification)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)

		res := Result{Code: http.StatusForbidden, Message: err.Error()}
		response, _ := json.Marshal(res)

		w.Write(response)
		return
	}

	orderModel := models.Order{}
	order, err := orderModel.FindByID(server.DB, paymentNotification.OrderID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)

		res := Result{Code: http.StatusForbidden, Message: err.Error()}
		response, _ := json.Marshal(res)

		w.Write(response)
		return
	}

	if order.IsPaid() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)

		res := Result{Code: http.StatusForbidden, Message: "Already paid before."}
		response, _ := json.Marshal(res)

		w.Write(response)
		return
	}

	paymentModel := models.Payment{}

	amount, _ := strconv.ParseFloat(paymentNotification.GrossAmount, 32)
	jsonPayload, _ := json.Marshal(paymentNotification)
	payload := (*json.RawMessage)(&jsonPayload)

	_, err = paymentModel.CreatePayment(server.DB, &models.Payment{
		OrderID:           order.ID,
		Amount:            int(amount),
		TransactionID:     paymentNotification.TransactionID,
		TransactionStatus: paymentNotification.TransactionStatus,
		Payload:           payload,
		PaymentType:       paymentNotification.PaymentType,
	})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		res := Result{Code: http.StatusBadRequest, Message: "Could not process the payment."}
		response, _ := json.Marshal(res)
		w.Write(response)
		return
	}

	if isPaymentSuccess(&paymentNotification) {
		err = order.MarkAsPaid(server.DB)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)

			res := Result{Code: http.StatusBadRequest, Message: err.Error()}
			response, _ := json.Marshal(res)

			w.Write(response)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	res := Result{Code: http.StatusOK, Message: "Payment saved."}
	response, _ := json.Marshal(res)

	w.Write(response)
}

func isPaymentSuccess(payload *models.MidtransNotification) bool {
	paymentStatus := false

	if payload.PaymentType == string(snap.PaymentTypeCreditCard) {
		paymentStatus = (payload.TransactionStatus == consts.PaymentStatusCapture) && (payload.FraudStatus == consts.FraudStatusAccept)
	} else {
		paymentStatus = (payload.TransactionStatus == consts.PaymentStatusSettlement) && (payload.FraudStatus == consts.FraudStatusAccept)
	}

	return paymentStatus
}

// this will validate the signature key in the midtrans payload
func validateSignatureKey(payload *models.MidtransNotification) error {
	environment := os.Getenv("APP_ENV")
	if environment == "development" {
		return nil
	}

	signaturePayload := payload.OrderID + payload.StatusCode + payload.GrossAmount + os.Getenv("API_MIDTRANS_SERVER_KEY")
	sha512Value := sha512.New()
	sha512Value.Write([]byte(signaturePayload))

	signatureKey := fmt.Sprintf("%x", sha512Value.Sum(nil))

	if signatureKey != payload.SignatureKey {
		return errors.New("invalid signature key")
	}

	return nil

}

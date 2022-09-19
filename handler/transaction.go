package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"go_e-commerce-api/transaction"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

type transactionHandler struct {
	transactionService transaction.Service
}

func NewTransactionHandler(transactionService transaction.Service) *transactionHandler {
	return &transactionHandler{transactionService}
}

func (h *transactionHandler) GetBooksList(c *gin.Context) {
	transactions, err := h.transactionService.FindAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
	}

	var transactionsResponse []transaction.TransactionResponse

	for _, b := range transactions {
		transactionResponse := converToTransactionResponse(b)
		transactionsResponse = append(transactionsResponse, transactionResponse)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": transactionsResponse,
	})
}

func (h *transactionHandler) GetBookById(c *gin.Context) {
	idString := c.Param("id")
	id, _ := strconv.Atoi(idString)

	b, err := h.transactionService.FindByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
	}

	transactionResponse := converToTransactionResponse(b)

	c.JSON(http.StatusOK, gin.H{
		"data": transactionResponse,
	})
}

func (h *transactionHandler) GetBookByUser(c *gin.Context) {
	email_user := c.Param("email_buyer")

	allproductss, err := h.transactionService.FindByUser(email_user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
	}

	var allproductssResponse []transaction.TransactionResponse

	for _, b := range allproductss {
		allproductsResponse := converToTransactionResponse(b)
		allproductssResponse = append(allproductssResponse, allproductsResponse)
	}

	if allproductssResponse != nil {
		c.JSON(http.StatusOK, gin.H{
			"data": allproductssResponse,
		})
	}
}

func (h *transactionHandler) ChargeToken(c *gin.Context) {

	bodyz, _ := ioutil.ReadAll(c.Request.Body)
	var result interface{}
	json.Unmarshal([]byte(bodyz), &result)

	m := result.(map[string]interface{})

	customerMap := m["customer_details"]
	customerValue := customerMap.(map[string]interface{})

	billMap := customerValue["billing_address"]
	billValue := billMap.(map[string]interface{})

	shipMap := customerValue["shipping_address"]
	shipValue := shipMap.(map[string]interface{})

	itemMap := m["item_details"]
	itemValues := itemMap.([]interface{})
	itemValue := itemValues[0].(map[string]interface{})

	transactionMap := m["transaction_details"]
	transactionValue := transactionMap.(map[string]interface{})

	var s snap.Client
	s.New("SB-Mid-server-LRjpvuhR8PgIV0AVXQjyd6kk", midtrans.Sandbox)

	req := &snap.Request{
		CustomerDetail: &midtrans.CustomerDetails{
			BillAddr: &midtrans.CustomerAddress{
				Address:  billValue["address"].(string),
				City:     billValue["city"].(string),
				Postcode: billValue["postal_code"].(string),
			},
			Email: customerValue["email"].(string),
			FName: customerValue["first_name"].(string),
			LName: customerValue["last_name"].(string),
			Phone: customerValue["phone"].(string),
			ShipAddr: &midtrans.CustomerAddress{
				Address:  shipValue["address"].(string),
				City:     shipValue["city"].(string),
				Postcode: shipValue["postal_code"].(string),
			},
		},
		Items: &[]midtrans.ItemDetails{
			{
				ID:    itemValue["id"].(string),
				Name:  itemValue["name"].(string),
				Price: int64(itemValue["price"].(float64)),
				Qty:   int32(itemValue["quantity"].(float64)),
			},
		},
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  transactionValue["order_id"].(string),
			GrossAmt: int64(transactionValue["gross_amount"].(float64)),
		},
		UserId: m["user_id"].(string),
	}

	snapResp, _ := s.CreateTransaction(req)

	c.JSON(http.StatusOK, gin.H{
		"token": snapResp.Token,
	})
}
func (h *transactionHandler) PostBooksHandler(c *gin.Context) {
	var transactionRequest transaction.TransactionRequest

	err := c.ShouldBindJSON(&transactionRequest)

	if err != nil {

		for _, e := range err.(validator.ValidationErrors) {
			errMessage := fmt.Sprintf("Error on filled %s, condition: %s", e.Field(), e.ActualTag())
			c.JSON(http.StatusBadRequest, errMessage)

			return
		}
	}

	transaction, err := h.transactionService.Create(transactionRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": transaction,
	})
}

func (h *transactionHandler) UpdateBook(c *gin.Context) {
	var transactionRequest transaction.TransactionRequest

	err := c.ShouldBindJSON(&transactionRequest)

	if err != nil {

		for _, e := range err.(validator.ValidationErrors) {
			errMessage := fmt.Sprintf("Error on filled %s, condition: %s", e.Field(), e.ActualTag())
			c.JSON(http.StatusBadRequest, errMessage)

			return
		}
	}

	idString := c.Param("id")
	id, _ := strconv.Atoi(idString)
	transaction, err := h.transactionService.Update(id, transactionRequest)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": transaction,
	})
}

func (h *transactionHandler) DeleteBook(c *gin.Context) {
	idString := c.Param("id")
	id, _ := strconv.Atoi(idString)

	b, err := h.transactionService.Delete(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
	}

	transactionResponse := converToTransactionResponse(b)

	c.JSON(http.StatusOK, gin.H{
		"data":    transactionResponse,
		"Message": "Delete data success",
	})
}

func converToTransactionResponse(b transaction.Transaction) transaction.TransactionResponse {
	return transaction.TransactionResponse{
		Id:           b.Id,
		Name_product: b.Name_product,
		Image_url:    b.Image_url,
		Description:  b.Description,
		Price:        b.Price,
		Name_user:    b.Name_user,
		Email_user:   b.Email_user,
		Name_buyer:   b.Name_buyer,
		Email_buyer:  b.Email_buyer,
	}
}

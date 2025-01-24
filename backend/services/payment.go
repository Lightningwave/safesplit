package services

import (
	"context"
	"errors"
	"strconv"

	"github.com/braintree-go/braintree-go"
)

type PaymentService struct {
	gateway *braintree.Braintree
}

func NewPaymentService() (*PaymentService, error) {
	g := braintree.New(
		braintree.Sandbox,
		"bnbn68spnz45nftp",
		"pqks8j59jj2bbwnq",
		"65c8bab66c4a8012ca66cb6d4587f07e",
	)

	return &PaymentService{
		gateway: g,
	}, nil
}

type PaymentRequest struct {
    CardNumber     string  `json:"CardNumber"`
    CVV           string  `json:"CVV"`
    ExpiryMonth   int     `json:"ExpiryMonth"`
    ExpiryYear    int     `json:"ExpiryYear"`
    CardHolder    string  `json:"CardHolder"`
    BillingCycle  string  `json:"BillingCycle"`
    Amount        float64
    Currency      string
    Type          string
    // Billing details
    BillingName    string `json:"BillingName"`
    BillingEmail   string `json:"BillingEmail"`
    BillingAddress string `json:"BillingAddress"`
    CountryCode    string `json:"CountryCode"`
}

func (s *PaymentService) ProcessPayment(ctx context.Context, req *PaymentRequest) (*braintree.Transaction, error) {
	if req.Amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	if req.Type == "" {
		return nil, errors.New("transaction type is required")
	}

	txReq := &braintree.TransactionRequest{
		Type:   req.Type,
		Amount: braintree.NewDecimal(int64(req.Amount*100), 2),
		CreditCard: &braintree.CreditCard{
			Number:          req.CardNumber,
			CVV:             req.CVV,
			ExpirationMonth: strconv.Itoa(req.ExpiryMonth),
			ExpirationYear:  strconv.Itoa(req.ExpiryYear),
			CardholderName:  req.CardHolder,
		},
		Options: &braintree.TransactionOptions{
			SubmitForSettlement:   true,
			StoreInVaultOnSuccess: true,
		},
	}

	tx, err := s.gateway.Transaction().Create(ctx, txReq)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (s *PaymentService) GetPaymentStatus(ctx context.Context, paymentID string) (*braintree.Transaction, error) {
	if paymentID == "" {
		return nil, errors.New("payment ID is required")
	}

	tx, err := s.gateway.Transaction().Find(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

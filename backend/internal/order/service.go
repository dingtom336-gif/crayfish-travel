package order

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// Order represents a confirmed travel order.
type Order struct {
	ID                    uuid.UUID  `json:"id"`
	SessionID             uuid.UUID  `json:"session_id"`
	PaymentID             uuid.UUID  `json:"payment_id"`
	QuoteID               uuid.UUID  `json:"quote_id"`
	TraceID               string     `json:"trace_id"`
	OrderNo               string     `json:"order_no"`
	Status                string     `json:"status"`
	ContactName           string     `json:"contact_name"`
	ContactPhone          string     `json:"contact_phone"`
	TotalAmountCents      int64      `json:"total_amount_cents"`
	BasePriceCents        int64      `json:"base_price_cents"`
	RefundGuaranteeFeeCents int64    `json:"refund_guarantee_fee_cents"`
	Supplier              string     `json:"supplier"`
	PackageTitle          string     `json:"package_title"`
	Destination           string     `json:"destination"`
	StartDate             time.Time  `json:"start_date"`
	EndDate               time.Time  `json:"end_date"`
	Adults                int        `json:"adults"`
	Children              int        `json:"children"`
	SmsSent               bool       `json:"sms_sent"`
	SmsSentAt             *time.Time `json:"sms_sent_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// SmsNotifier defines the interface for sending SMS notifications.
type SmsNotifier interface {
	SendOrderConfirmation(phone, orderNo, destination, dates string) error
}

// MockSmsNotifier logs SMS to stdout instead of sending real messages.
type MockSmsNotifier struct{}

// NewMockSmsNotifier creates a mock SMS notifier.
func NewMockSmsNotifier() *MockSmsNotifier {
	return &MockSmsNotifier{}
}

// SendOrderConfirmation prints the notification to stdout.
func (m *MockSmsNotifier) SendOrderConfirmation(phone, orderNo, destination, dates string) error {
	fmt.Printf("[MockSMS] To: %s | Order: %s | Destination: %s | Dates: %s\n",
		phone, orderNo, destination, dates)
	return nil
}

// PIIDecryptor decrypts sealed PII blobs (nonce-prefixed AES-256-GCM).
type PIIDecryptor interface {
	OpenStringWithNonce(sealed []byte) (string, error)
}

// OrderService handles order business logic.
type OrderService struct {
	db        *sql.DB
	decryptor PIIDecryptor
}

// NewOrderService creates a new order service.
func NewOrderService(db *sql.DB, decryptor PIIDecryptor) *OrderService {
	return &OrderService{db: db, decryptor: decryptor}
}

// GenerateOrderNo creates a unique order number: "CO" + timestamp + 8 random chars.
func GenerateOrderNo() string {
	ts := time.Now().Format("20060102150405")
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	suffix := make([]byte, 8)
	for i := range suffix {
		suffix[i] = chars[rand.Intn(len(chars))]
	}
	return fmt.Sprintf("CO%s%s", ts, string(suffix))
}

// CreateOrder creates an order from payment and quote data.
func (s *OrderService) CreateOrder(ctx context.Context, sessionID, paymentID, quoteID uuid.UUID, traceID string) (*Order, error) {
	// Pull contact info from identity_records (PII is AES-256-GCM encrypted)
	var contactName, contactPhone string
	var nameEnc, phoneEnc []byte
	err := s.db.QueryRowContext(ctx, `
		SELECT name_encrypted, phone_encrypted FROM identity_records
		WHERE session_id = $1 ORDER BY created_at DESC LIMIT 1`,
		sessionID,
	).Scan(&nameEnc, &phoneEnc)
	if err == sql.ErrNoRows {
		contactName = "Guest"
		contactPhone = ""
	} else if err != nil {
		return nil, fmt.Errorf("fetch contact info: %w", err)
	} else {
		// Decrypt only if we have data
		if len(nameEnc) > 0 && s.decryptor != nil {
			if n, decErr := s.decryptor.OpenStringWithNonce(nameEnc); decErr == nil {
				contactName = n
			}
		}
		if len(phoneEnc) > 0 && s.decryptor != nil {
			if p, decErr := s.decryptor.OpenStringWithNonce(phoneEnc); decErr == nil {
				contactPhone = p
			}
		}
	}

	// Pull quote details (join sessions for start_date, end_date, adults, children
	// which are not stored on supplier_quotes)
	var packageTitle, destination, supplier string
	var totalAmountCents, basePriceCents, refundGuaranteeFeeCents int64
	var startDate, endDate time.Time
	var adults, children int
	err = s.db.QueryRowContext(ctx, `
		SELECT q.package_title, q.destination, q.supplier,
		       q.total_price_cents, q.base_price_cents, q.refund_guarantee_fee_cents,
		       s.start_date, s.end_date, s.adults, s.children
		FROM supplier_quotes q
		JOIN sessions s ON s.id = q.session_id
		WHERE q.id = $1 AND q.session_id = $2`,
		quoteID, sessionID,
	).Scan(&packageTitle, &destination, &supplier,
		&totalAmountCents, &basePriceCents, &refundGuaranteeFeeCents,
		&startDate, &endDate, &adults, &children)
	if err != nil {
		return nil, fmt.Errorf("fetch quote details: %w", err)
	}

	orderNo := GenerateOrderNo()

	// Insert order
	var order Order
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO orders (
			session_id, payment_id, quote_id, trace_id, order_no,
			contact_name, contact_phone, total_amount_cents, base_price_cents,
			refund_guarantee_fee_cents, supplier, package_title, destination,
			start_date, end_date, adults, children
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id, session_id, payment_id, quote_id, trace_id, order_no, status,
		          contact_name, contact_phone, total_amount_cents, base_price_cents,
		          refund_guarantee_fee_cents, supplier, package_title, destination,
		          start_date, end_date, adults, children, sms_sent, created_at, updated_at`,
		sessionID, paymentID, quoteID, traceID, orderNo,
		contactName, contactPhone, totalAmountCents, basePriceCents,
		refundGuaranteeFeeCents, supplier, packageTitle, destination,
		startDate, endDate, adults, children,
	).Scan(
		&order.ID, &order.SessionID, &order.PaymentID, &order.QuoteID,
		&order.TraceID, &order.OrderNo, &order.Status,
		&order.ContactName, &order.ContactPhone,
		&order.TotalAmountCents, &order.BasePriceCents,
		&order.RefundGuaranteeFeeCents, &order.Supplier, &order.PackageTitle,
		&order.Destination, &order.StartDate, &order.EndDate,
		&order.Adults, &order.Children, &order.SmsSent,
		&order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert order: %w", err)
	}

	return &order, nil
}

// GetByID returns a full order by ID.
func (s *OrderService) GetByID(ctx context.Context, orderID uuid.UUID) (*Order, error) {
	var order Order
	err := s.db.QueryRowContext(ctx, `
		SELECT id, session_id, payment_id, quote_id, trace_id, order_no, status,
		       contact_name, contact_phone, total_amount_cents, base_price_cents,
		       refund_guarantee_fee_cents, supplier, package_title, destination,
		       start_date, end_date, adults, children, sms_sent, sms_sent_at,
		       created_at, updated_at
		FROM orders WHERE id = $1`, orderID,
	).Scan(
		&order.ID, &order.SessionID, &order.PaymentID, &order.QuoteID,
		&order.TraceID, &order.OrderNo, &order.Status,
		&order.ContactName, &order.ContactPhone,
		&order.TotalAmountCents, &order.BasePriceCents,
		&order.RefundGuaranteeFeeCents, &order.Supplier, &order.PackageTitle,
		&order.Destination, &order.StartDate, &order.EndDate,
		&order.Adults, &order.Children, &order.SmsSent, &order.SmsSentAt,
		&order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}
	return &order, nil
}

// ListBySession returns all orders for a given session.
func (s *OrderService) ListBySession(ctx context.Context, sessionID uuid.UUID) ([]Order, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, session_id, payment_id, quote_id, trace_id, order_no, status,
		       contact_name, contact_phone, total_amount_cents, base_price_cents,
		       refund_guarantee_fee_cents, supplier, package_title, destination,
		       start_date, end_date, adults, children, sms_sent, sms_sent_at,
		       created_at, updated_at
		FROM orders WHERE session_id = $1 ORDER BY created_at DESC`, sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(
			&o.ID, &o.SessionID, &o.PaymentID, &o.QuoteID,
			&o.TraceID, &o.OrderNo, &o.Status,
			&o.ContactName, &o.ContactPhone,
			&o.TotalAmountCents, &o.BasePriceCents,
			&o.RefundGuaranteeFeeCents, &o.Supplier, &o.PackageTitle,
			&o.Destination, &o.StartDate, &o.EndDate,
			&o.Adults, &o.Children, &o.SmsSent, &o.SmsSentAt,
			&o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, o)
	}

	return orders, rows.Err()
}

// CreateFromPayment implements payment.OrderCreator interface.
// Called automatically by CallbackProcessor after successful payment.
func (s *OrderService) CreateFromPayment(ctx context.Context, sessionID, paymentID, quoteID, traceID string) error {
	sid, err := uuid.Parse(sessionID)
	if err != nil {
		return fmt.Errorf("parse session_id: %w", err)
	}
	pid, err := uuid.Parse(paymentID)
	if err != nil {
		return fmt.Errorf("parse payment_id: %w", err)
	}
	qid, err := uuid.Parse(quoteID)
	if err != nil {
		return fmt.Errorf("parse quote_id: %w", err)
	}

	_, err = s.CreateOrder(ctx, sid, pid, qid, traceID)
	return err
}

// Valid status transitions for refund requests.
var refundableStatuses = map[string]bool{
	"created":    true,
	"confirmed":  true,
	"fulfilling": true,
	"completed":  true,
}

// RequestRefund updates order status to 'refund_requested'.
func (s *OrderService) RequestRefund(ctx context.Context, orderID uuid.UUID, traceID string) error {
	var currentStatus string
	err := s.db.QueryRowContext(ctx,
		`SELECT status FROM orders WHERE id = $1`, orderID,
	).Scan(&currentStatus)
	if err != nil {
		return fmt.Errorf("get order status: %w", err)
	}

	if !refundableStatuses[currentStatus] {
		return fmt.Errorf("cannot request refund from status %q", currentStatus)
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE orders SET status = 'refund_requested', trace_id = $1, updated_at = NOW()
		WHERE id = $2`,
		traceID, orderID,
	)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}

	return nil
}

package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func tenantIDFromHeader(r *http.Request) (uuid.UUID, error) {
	raw := r.Header.Get("X-Tenant-ID")
	if raw == "" {
		return uuid.Nil, errors.New("X-Tenant-ID header is required")
	}

	tenantID, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.New("X-Tenant-ID must be a valid UUID")
	}
	return tenantID, nil
}

func parseUUID(raw string) (uuid.UUID, error) {
	return uuid.Parse(raw)
}

func toPGUUID(id uuid.UUID) pgtype.UUID {
	var out pgtype.UUID
	copy(out.Bytes[:], id[:])
	out.Valid = true
	return out
}

func toPGText(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
}

func toPGTimestamptz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: true}
}

func pgUUIDToString(value pgtype.UUID) string {
	if !value.Valid {
		return ""
	}
	return uuid.UUID(value.Bytes).String()
}

func pgTextToString(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func pgDateToString(value pgtype.Date) string {
	if !value.Valid {
		return ""
	}
	return value.Time.Format("2006-01-02")
}

func pgTimestampToString(value pgtype.Timestamptz) string {
	if !value.Valid {
		return ""
	}
	return value.Time.UTC().Format(time.RFC3339)
}

func queryPageLimit(r *http.Request, defaultLimit int32) (int32, int32) {
	page := int32(1)
	limit := defaultLimit

	if raw := r.URL.Query().Get("page"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			page = int32(parsed)
		}
	}
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 200 {
			limit = int32(parsed)
		}
	}

	offset := (page - 1) * limit
	return offset, limit
}

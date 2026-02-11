package handlers

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	dbgen "sfa/backend/internal/db/sqlc"
)

func (h FeaturePackHandler) CreateIntegrationEvent(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromHeader(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_tenant_id", err.Error())
		return
	}

	var req struct {
		Provider            string         `json:"provider"`
		IntegrationType     string         `json:"integrationType"`
		ExternalEventID     string         `json:"externalEventId"`
		EventType           string         `json:"eventType"`
		Payload             map[string]any `json:"payload"`
		LinkedAccountID     string         `json:"linkedAccountId"`
		LinkedContactID     string         `json:"linkedContactId"`
		LinkedOpportunityID string         `json:"linkedOpportunityId"`
		OccurredAt          string         `json:"occurredAt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json body")
		return
	}

	provider, err := parseProvider(req.Provider)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_provider", err.Error())
		return
	}
	integrationType, err := parseIntegrationType(req.IntegrationType)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_integration_type", err.Error())
		return
	}

	occurredAt, err := parseOptionalTimestamp(req.OccurredAt)
	if err != nil || !occurredAt.Valid {
		writeError(w, http.StatusBadRequest, "invalid_occurred_at", "occurredAt must be RFC3339")
		return
	}

	payload, err := json.Marshal(req.Payload)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_payload", "payload must be serializable")
		return
	}

	var linkedAccountID pgtype.UUID
	if strings.TrimSpace(req.LinkedAccountID) != "" {
		id, parseErr := parseUUID(req.LinkedAccountID)
		if parseErr != nil {
			writeError(w, http.StatusBadRequest, "invalid_linked_account_id", "linkedAccountId must be UUID")
			return
		}
		linkedAccountID = toPGUUID(id)
	}

	var linkedContactID pgtype.UUID
	if strings.TrimSpace(req.LinkedContactID) != "" {
		id, parseErr := parseUUID(req.LinkedContactID)
		if parseErr != nil {
			writeError(w, http.StatusBadRequest, "invalid_linked_contact_id", "linkedContactId must be UUID")
			return
		}
		linkedContactID = toPGUUID(id)
	}

	var linkedOpportunityID pgtype.UUID
	if strings.TrimSpace(req.LinkedOpportunityID) != "" {
		id, parseErr := parseUUID(req.LinkedOpportunityID)
		if parseErr != nil {
			writeError(w, http.StatusBadRequest, "invalid_linked_opportunity_id", "linkedOpportunityId must be UUID")
			return
		}
		linkedOpportunityID = toPGUUID(id)
	}

	var row dbgen.IntegrationEvent
	if err := h.Store.WithTenantTx(r.Context(), tenantID, func(q *dbgen.Queries) error {
		var queryErr error
		row, queryErr = q.CreateIntegrationEvent(r.Context(), dbgen.CreateIntegrationEventParams{
			TenantID:            toPGUUID(tenantID),
			Provider:            provider,
			IntegrationType:     integrationType,
			ExternalEventID:     toPGText(req.ExternalEventID),
			EventType:           req.EventType,
			Payload:             payload,
			LinkedAccountID:     linkedAccountID,
			LinkedContactID:     linkedContactID,
			LinkedOpportunityID: linkedOpportunityID,
			OccurredAt:          occurredAt,
		})
		return queryErr
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "integration_event_create_failed", "failed to create integration event")
		return
	}

	payloadOut := map[string]any{}
	_ = json.Unmarshal(row.Payload, &payloadOut)
	writeJSON(w, http.StatusCreated, map[string]any{
		"data": map[string]any{
			"id":                  row.ID,
			"provider":            string(row.Provider),
			"integrationType":     string(row.IntegrationType),
			"externalEventId":     pgTextToString(row.ExternalEventID),
			"eventType":           row.EventType,
			"payload":             payloadOut,
			"linkedAccountId":     pgUUIDToString(row.LinkedAccountID),
			"linkedContactId":     pgUUIDToString(row.LinkedContactID),
			"linkedOpportunityId": pgUUIDToString(row.LinkedOpportunityID),
			"occurredAt":          pgTimestampToString(row.OccurredAt),
		},
	})
}

func (h FeaturePackHandler) ListIntegrationEvents(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromHeader(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_tenant_id", err.Error())
		return
	}

	offset, limit := queryPageLimit(r, 20)
	var rows []dbgen.IntegrationEvent
	if err := h.Store.WithTenantTx(r.Context(), tenantID, func(q *dbgen.Queries) error {
		var queryErr error
		rows, queryErr = q.ListIntegrationEvents(r.Context(), dbgen.ListIntegrationEventsParams{
			TenantID:    toPGUUID(tenantID),
			OffsetCount: offset,
			LimitCount:  limit,
		})
		return queryErr
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "integration_event_list_failed", "failed to load integration events")
		return
	}

	data := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		payload := map[string]any{}
		_ = json.Unmarshal(row.Payload, &payload)
		data = append(data, map[string]any{
			"id":                  row.ID,
			"provider":            string(row.Provider),
			"integrationType":     string(row.IntegrationType),
			"externalEventId":     pgTextToString(row.ExternalEventID),
			"eventType":           row.EventType,
			"payload":             payload,
			"linkedAccountId":     pgUUIDToString(row.LinkedAccountID),
			"linkedContactId":     pgUUIDToString(row.LinkedContactID),
			"linkedOpportunityId": pgUUIDToString(row.LinkedOpportunityID),
			"occurredAt":          pgTimestampToString(row.OccurredAt),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": data})
}

func (h FeaturePackHandler) exportAccountsCSV(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromHeader(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_tenant_id", err.Error())
		return
	}

	var rows []dbgen.ExportAccountsRowsRow
	if err := h.Store.WithTenantTx(r.Context(), tenantID, func(q *dbgen.Queries) error {
		var queryErr error
		rows, queryErr = q.ExportAccountsRows(r.Context(), toPGUUID(tenantID))
		return queryErr
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "accounts_export_failed", "failed to export accounts")
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=accounts.csv")
	writer := csv.NewWriter(w)
	_ = writer.Write([]string{
		"id", "owner_user_id", "name", "industry", "website", "phone", "status", "memo", "created_at", "updated_at",
	})
	for _, row := range rows {
		_ = writer.Write([]string{
			pgUUIDToString(row.ID),
			pgUUIDToString(row.OwnerUserID),
			row.Name,
			pgTextToString(row.Industry),
			pgTextToString(row.Website),
			pgTextToString(row.Phone),
			string(row.Status),
			pgTextToString(row.Memo),
			pgTimestampToString(row.CreatedAt),
			pgTimestampToString(row.UpdatedAt),
		})
	}
	writer.Flush()
}

func (h FeaturePackHandler) exportOpportunitiesCSV(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromHeader(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_tenant_id", err.Error())
		return
	}

	var rows []dbgen.ExportOpportunitiesRowsRow
	if err := h.Store.WithTenantTx(r.Context(), tenantID, func(q *dbgen.Queries) error {
		var queryErr error
		rows, queryErr = q.ExportOpportunitiesRows(r.Context(), toPGUUID(tenantID))
		return queryErr
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "opportunities_export_failed", "failed to export opportunities")
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=opportunities.csv")
	writer := csv.NewWriter(w)
	_ = writer.Write([]string{
		"id", "account_id", "contact_id", "owner_user_id", "name", "stage", "probability", "amount", "expected_close_date", "next_action_at", "next_action_note", "created_at", "updated_at",
	})
	for _, row := range rows {
		_ = writer.Write([]string{
			pgUUIDToString(row.ID),
			pgUUIDToString(row.AccountID),
			pgUUIDToString(row.ContactID),
			pgUUIDToString(row.OwnerUserID),
			row.Name,
			string(row.Stage),
			strconv.Itoa(int(row.Probability)),
			strconv.FormatFloat(row.Amount, 'f', 2, 64),
			pgDateToString(row.ExpectedCloseDate),
			pgTimestampToString(row.NextActionAt),
			pgTextToString(row.NextActionNote),
			pgTimestampToString(row.CreatedAt),
			pgTimestampToString(row.UpdatedAt),
		})
	}
	writer.Flush()
}

func (h FeaturePackHandler) importAccountsCSV(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromHeader(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_tenant_id", err.Error())
		return
	}
	records, err := readCSVRecords(r)
	if err != nil || len(records) < 2 {
		writeError(w, http.StatusBadRequest, "invalid_csv", "csv must contain header and at least one row")
		return
	}

	headers := buildCSVHeaderIndex(records[0])
	inserted := 0
	rowErrors := []string{}
	if err := h.Store.WithTenantTx(r.Context(), tenantID, func(q *dbgen.Queries) error {
		for i, rec := range records[1:] {
			rowNo := i + 2
			ownerRaw := csvCell(rec, headers, "owner_user_id")
			name := csvCell(rec, headers, "name")
			if ownerRaw == "" || name == "" {
				rowErrors = append(rowErrors, "row "+strconv.Itoa(rowNo)+": owner_user_id and name are required")
				continue
			}
			ownerID, parseErr := parseUUID(ownerRaw)
			if parseErr != nil {
				rowErrors = append(rowErrors, "row "+strconv.Itoa(rowNo)+": invalid owner_user_id")
				continue
			}
			status, parseErr := parseAccountStatus(csvCell(rec, headers, "status"))
			if parseErr != nil {
				rowErrors = append(rowErrors, "row "+strconv.Itoa(rowNo)+": invalid status")
				continue
			}
			_, queryErr := q.CreateAccount(r.Context(), dbgen.CreateAccountParams{
				TenantID:    toPGUUID(tenantID),
				OwnerUserID: toPGUUID(ownerID),
				Name:        name,
				Industry:    toPGText(csvCell(rec, headers, "industry")),
				Website:     toPGText(csvCell(rec, headers, "website")),
				Phone:       toPGText(csvCell(rec, headers, "phone")),
				Status:      status,
				Memo:        toPGText(csvCell(rec, headers, "memo")),
				CreatedBy:   toPGUUID(ownerID),
			})
			if queryErr != nil {
				rowErrors = append(rowErrors, "row "+strconv.Itoa(rowNo)+": "+queryErr.Error())
				continue
			}
			inserted++
		}
		return nil
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "accounts_import_failed", "failed to import accounts")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"inserted": inserted, "errors": rowErrors})
}

func (h FeaturePackHandler) importOpportunitiesCSV(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromHeader(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_tenant_id", err.Error())
		return
	}
	records, err := readCSVRecords(r)
	if err != nil || len(records) < 2 {
		writeError(w, http.StatusBadRequest, "invalid_csv", "csv must contain header and at least one row")
		return
	}

	headers := buildCSVHeaderIndex(records[0])
	inserted := 0
	rowErrors := []string{}
	if err := h.Store.WithTenantTx(r.Context(), tenantID, func(q *dbgen.Queries) error {
		for i, rec := range records[1:] {
			rowNo := i + 2
			accountRaw := csvCell(rec, headers, "account_id")
			ownerRaw := csvCell(rec, headers, "owner_user_id")
			name := csvCell(rec, headers, "name")
			if accountRaw == "" || ownerRaw == "" || name == "" {
				rowErrors = append(rowErrors, "row "+strconv.Itoa(rowNo)+": account_id, owner_user_id, name are required")
				continue
			}
			accountID, err1 := parseUUID(accountRaw)
			ownerID, err2 := parseUUID(ownerRaw)
			if err1 != nil || err2 != nil {
				rowErrors = append(rowErrors, "row "+strconv.Itoa(rowNo)+": invalid account_id or owner_user_id")
				continue
			}

			stage, stageErr := parseOpportunityStage(csvCell(rec, headers, "stage"))
			if stageErr != nil {
				rowErrors = append(rowErrors, "row "+strconv.Itoa(rowNo)+": invalid stage")
				continue
			}

			expectedClose := pgtype.Date{}
			if rawDate := csvCell(rec, headers, "expected_close_date"); rawDate != "" {
				if t, parseErr := time.Parse("2006-01-02", rawDate); parseErr == nil {
					expectedClose = pgtype.Date{Time: t, Valid: true}
				} else {
					rowErrors = append(rowErrors, "row "+strconv.Itoa(rowNo)+": invalid expected_close_date")
					continue
				}
			}

			created, queryErr := q.CreateOpportunity(r.Context(), dbgen.CreateOpportunityParams{
				TenantID:          toPGUUID(tenantID),
				AccountID:         toPGUUID(accountID),
				ContactID:         pgtype.UUID{},
				OwnerUserID:       toPGUUID(ownerID),
				Name:              name,
				Stage:             stage,
				Probability:       parseInt16(csvCell(rec, headers, "probability"), 0),
				Amount:            parseFloat(csvCell(rec, headers, "amount"), 0),
				ExpectedCloseDate: expectedClose,
				Memo:              toPGText(csvCell(rec, headers, "memo")),
				CreatedBy:         toPGUUID(ownerID),
			})
			if queryErr != nil {
				rowErrors = append(rowErrors, "row "+strconv.Itoa(rowNo)+": "+queryErr.Error())
				continue
			}

			if nextActionRaw := csvCell(rec, headers, "next_action_at"); nextActionRaw != "" {
				nextActionAt, parseErr := time.Parse(time.RFC3339, nextActionRaw)
				if parseErr != nil {
					rowErrors = append(rowErrors, "row "+strconv.Itoa(rowNo)+": invalid next_action_at")
				} else {
					_, _ = q.UpdateOpportunityNextAction(r.Context(), dbgen.UpdateOpportunityNextActionParams{
						NextActionAt:   toPGTimestamptz(nextActionAt.UTC()),
						NextActionNote: toPGText(csvCell(rec, headers, "next_action_note")),
						TenantID:       toPGUUID(tenantID),
						OpportunityID:  created.ID,
					})
				}
			}
			inserted++
		}
		return nil
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "opportunities_import_failed", "failed to import opportunities")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"inserted": inserted, "errors": rowErrors})
}

func parseAccountStatus(raw string) (dbgen.NullAccountStatusEnum, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "prospect":
		return dbgen.NullAccountStatusEnum{AccountStatusEnum: dbgen.AccountStatusEnumProspect, Valid: true}, nil
	case "active":
		return dbgen.NullAccountStatusEnum{AccountStatusEnum: dbgen.AccountStatusEnumActive, Valid: true}, nil
	case "inactive":
		return dbgen.NullAccountStatusEnum{AccountStatusEnum: dbgen.AccountStatusEnumInactive, Valid: true}, nil
	default:
		return dbgen.NullAccountStatusEnum{}, strconv.ErrSyntax
	}
}

func parseOpportunityStage(raw string) (dbgen.NullOpportunityStageEnum, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "new_lead":
		return dbgen.NullOpportunityStageEnum{OpportunityStageEnum: dbgen.OpportunityStageEnumNewLead, Valid: true}, nil
	case "qualified":
		return dbgen.NullOpportunityStageEnum{OpportunityStageEnum: dbgen.OpportunityStageEnumQualified, Valid: true}, nil
	case "proposal":
		return dbgen.NullOpportunityStageEnum{OpportunityStageEnum: dbgen.OpportunityStageEnumProposal, Valid: true}, nil
	case "negotiation":
		return dbgen.NullOpportunityStageEnum{OpportunityStageEnum: dbgen.OpportunityStageEnumNegotiation, Valid: true}, nil
	case "closed_won":
		return dbgen.NullOpportunityStageEnum{OpportunityStageEnum: dbgen.OpportunityStageEnumClosedWon, Valid: true}, nil
	case "closed_lost":
		return dbgen.NullOpportunityStageEnum{OpportunityStageEnum: dbgen.OpportunityStageEnumClosedLost, Valid: true}, nil
	default:
		return dbgen.NullOpportunityStageEnum{}, strconv.ErrSyntax
	}
}

func buildCSVHeaderIndex(header []string) map[string]int {
	index := make(map[string]int, len(header))
	for i, h := range header {
		cleaned := strings.TrimPrefix(strings.TrimSpace(h), "\ufeff")
		index[strings.ToLower(cleaned)] = i
	}
	return index
}

func csvCell(row []string, index map[string]int, key string) string {
	pos, ok := index[key]
	if !ok || pos < 0 || pos >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[pos])
}
